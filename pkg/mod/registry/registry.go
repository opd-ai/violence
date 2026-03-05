// Package registry provides mod distribution and management infrastructure.
package registry

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/opd-ai/violence/pkg/mod"
	"github.com/sirupsen/logrus"
)

// Registry manages mod storage, validation, and distribution.
type Registry struct {
	db          *sql.DB
	storagePath string
	maxModSize  int64
	mu          sync.RWMutex
}

// ModRecord represents stored mod metadata.
type ModRecord struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	SHA256      string    `json:"sha256"`
	Size        int64     `json:"size"`
	UploadedAt  time.Time `json:"uploaded_at"`
	Downloads   int       `json:"downloads"`
}

// NewRegistry creates a new mod registry with database and storage.
func NewRegistry(db *sql.DB, storagePath string) (*Registry, error) {
	// Create storage directory
	if err := os.MkdirAll(storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	r := &Registry{
		db:          db,
		storagePath: storagePath,
		maxModSize:  10 * 1024 * 1024, // 10MB default
	}

	// Initialize database schema
	if err := r.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return r, nil
}

// SetMaxModSize configures maximum mod file size in bytes.
func (r *Registry) SetMaxModSize(size int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.maxModSize = size
}

// Close closes the registry database connection.
func (r *Registry) Close() error {
	return r.db.Close()
}

// initSchema creates database tables if they don't exist.
func (r *Registry) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS mods (
		name TEXT NOT NULL,
		version TEXT NOT NULL,
		author TEXT NOT NULL,
		description TEXT,
		tags TEXT,
		sha256 TEXT NOT NULL,
		size INTEGER NOT NULL,
		uploaded_at DATETIME NOT NULL,
		downloads INTEGER DEFAULT 0,
		PRIMARY KEY (name, version)
	);
	CREATE INDEX IF NOT EXISTS idx_mods_author ON mods(author);
	CREATE INDEX IF NOT EXISTS idx_mods_uploaded ON mods(uploaded_at DESC);
	`
	_, err := r.db.Exec(schema)
	return err
}

// HandleUpload processes mod upload requests with validation and virus scanning.
func (r *Registry) HandleUpload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	manifest, wasmData, err := r.parseUploadRequest(w, req)
	if err != nil {
		return
	}

	checksum, err := r.validateAndStoreFiles(w, manifest, wasmData)
	if err != nil {
		return
	}

	if err := r.saveModMetadata(w, manifest, wasmData, checksum); err != nil {
		return
	}

	r.sendUploadSuccess(w, manifest, checksum)
}

// parseUploadRequest extracts and validates manifest and WASM files from the upload request.
func (r *Registry) parseUploadRequest(w http.ResponseWriter, req *http.Request) (*mod.Manifest, []byte, error) {
	if err := req.ParseMultipartForm(r.maxModSize); err != nil {
		logrus.WithError(err).Warn("Failed to parse multipart form")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return nil, nil, err
	}

	file, header, err := req.FormFile("wasm")
	if err != nil {
		http.Error(w, "Missing wasm file", http.StatusBadRequest)
		return nil, nil, err
	}
	defer file.Close()

	if header.Size > r.maxModSize {
		http.Error(w, fmt.Sprintf("File too large (max %d bytes)", r.maxModSize), http.StatusRequestEntityTooLarge)
		return nil, nil, fmt.Errorf("file too large")
	}

	manifestFile, _, err := req.FormFile("manifest")
	if err != nil {
		http.Error(w, "Missing manifest file", http.StatusBadRequest)
		return nil, nil, err
	}
	defer manifestFile.Close()

	manifest, err := r.parseManifest(w, manifestFile)
	if err != nil {
		return nil, nil, err
	}

	wasmData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read WASM file", http.StatusInternalServerError)
		return nil, nil, err
	}

	return manifest, wasmData, nil
}

// parseManifest reads and validates a manifest JSON file.
func (r *Registry) parseManifest(w http.ResponseWriter, manifestFile multipart.File) (*mod.Manifest, error) {
	manifestData, err := io.ReadAll(manifestFile)
	if err != nil {
		http.Error(w, "Failed to read manifest", http.StatusBadRequest)
		return nil, err
	}

	var manifest mod.Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		logrus.WithError(err).Warn("Invalid manifest JSON")
		http.Error(w, "Invalid manifest JSON", http.StatusBadRequest)
		return nil, err
	}

	if err := manifest.Validate(); err != nil {
		logrus.WithError(err).Warn("Manifest validation failed")
		http.Error(w, fmt.Sprintf("Invalid manifest: %v", err), http.StatusBadRequest)
		return nil, err
	}

	return &manifest, nil
}

// validateAndStoreFiles validates WASM data and stores it to disk.
func (r *Registry) validateAndStoreFiles(w http.ResponseWriter, manifest *mod.Manifest, wasmData []byte) (string, error) {
	if err := validateWASM(wasmData); err != nil {
		logrus.WithError(err).Warn("WASM validation failed")
		http.Error(w, fmt.Sprintf("Invalid WASM file: %v", err), http.StatusBadRequest)
		return "", err
	}

	if err := virusScanStub(wasmData); err != nil {
		logrus.WithError(err).Warn("Virus scan failed")
		http.Error(w, "Security scan failed", http.StatusForbidden)
		return "", err
	}

	hash := sha256.Sum256(wasmData)
	checksum := hex.EncodeToString(hash[:])

	modPath := filepath.Join(r.storagePath, fmt.Sprintf("%s-%s.wasm", manifest.Name, manifest.Version))
	if err := os.WriteFile(modPath, wasmData, 0o644); err != nil {
		logrus.WithError(err).Error("Failed to write WASM file")
		http.Error(w, "Storage error", http.StatusInternalServerError)
		return "", err
	}

	return checksum, nil
}

// saveModMetadata inserts mod metadata into the database.
func (r *Registry) saveModMetadata(w http.ResponseWriter, manifest *mod.Manifest, wasmData []byte, checksum string) error {
	tagsJSON, _ := json.Marshal(manifest.Tags)
	_, err := r.db.Exec(`
		INSERT OR REPLACE INTO mods (name, version, author, description, tags, sha256, size, uploaded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, manifest.Name, manifest.Version, manifest.Author, manifest.Description, string(tagsJSON), checksum, len(wasmData), time.Now())
	if err != nil {
		logrus.WithError(err).Error("Failed to insert mod record")
		modPath := filepath.Join(r.storagePath, fmt.Sprintf("%s-%s.wasm", manifest.Name, manifest.Version))
		os.Remove(modPath)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"system_name": "mod_registry",
		"mod_name":    manifest.Name,
		"version":     manifest.Version,
		"author":      manifest.Author,
		"size":        len(wasmData),
		"sha256":      checksum,
	}).Info("Mod uploaded successfully")

	return nil
}

// sendUploadSuccess sends a successful upload response.
func (r *Registry) sendUploadSuccess(w http.ResponseWriter, manifest *mod.Manifest, checksum string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"name":    manifest.Name,
		"version": manifest.Version,
		"sha256":  checksum,
	})
}

// HandleSearch processes mod search requests with filtering.
func (r *Registry) HandleSearch(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := req.URL.Query()
	name := query.Get("name")
	author := query.Get("author")
	tag := query.Get("tag")

	// Build SQL query dynamically
	var conditions []string
	var args []interface{}

	if name != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+name+"%")
	}
	if author != "" {
		conditions = append(conditions, "author = ?")
		args = append(args, author)
	}
	if tag != "" {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, "%"+tag+"%")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	sqlQuery := fmt.Sprintf(`
		SELECT name, version, author, description, tags, sha256, size, uploaded_at, downloads
		FROM mods %s
		ORDER BY uploaded_at DESC
		LIMIT 50
	`, whereClause)

	rows, err := r.db.Query(sqlQuery, args...)
	if err != nil {
		logrus.WithError(err).Error("Search query failed")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []ModRecord
	for rows.Next() {
		var rec ModRecord
		var tagsJSON string
		err := rows.Scan(&rec.Name, &rec.Version, &rec.Author, &rec.Description, &tagsJSON, &rec.SHA256, &rec.Size, &rec.UploadedAt, &rec.Downloads)
		if err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &rec.Tags)
		results = append(results, rec)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"count":   len(results),
	})
}

// HandleDownload serves mod WASM files and increments download counter.
func (r *Registry) HandleDownload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract mod name and version from URL path
	path := strings.TrimPrefix(req.URL.Path, "/download/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		http.Error(w, "Invalid download path (expected /download/{name}/{version})", http.StatusBadRequest)
		return
	}

	name, version := parts[0], parts[1]

	// Verify mod exists in database
	var sha256 string
	err := r.db.QueryRow("SELECT sha256 FROM mods WHERE name = ? AND version = ?", name, version).Scan(&sha256)
	if err == sql.ErrNoRows {
		http.Error(w, "Mod not found", http.StatusNotFound)
		return
	} else if err != nil {
		logrus.WithError(err).Error("Database query failed")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Increment download counter
	_, err = r.db.Exec("UPDATE mods SET downloads = downloads + 1 WHERE name = ? AND version = ?", name, version)
	if err != nil {
		logrus.WithError(err).Warn("Failed to update download counter")
	}

	// Serve WASM file
	modPath := filepath.Join(r.storagePath, fmt.Sprintf("%s-%s.wasm", name, version))
	if _, err := os.Stat(modPath); os.IsNotExist(err) {
		http.Error(w, "WASM file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/wasm")
	w.Header().Set("X-Mod-SHA256", sha256)
	http.ServeFile(w, req, modPath)

	logrus.WithFields(logrus.Fields{
		"system_name": "mod_registry",
		"mod_name":    name,
		"version":     version,
	}).Debug("Mod downloaded")
}

// validateWASM performs basic WASM magic number validation.
func validateWASM(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("file too small to be valid WASM")
	}

	// WASM magic number: 0x00 0x61 0x73 0x6D
	if data[0] != 0x00 || data[1] != 0x61 || data[2] != 0x73 || data[3] != 0x6D {
		return fmt.Errorf("invalid WASM magic number")
	}

	// WASM version: 0x01 0x00 0x00 0x00 (version 1)
	if data[4] != 0x01 || data[5] != 0x00 || data[6] != 0x00 || data[7] != 0x00 {
		return fmt.Errorf("unsupported WASM version")
	}

	return nil
}

// virusScanStub is a placeholder for future virus scanning integration.
// In production, this would integrate with ClamAV or similar.
func virusScanStub(data []byte) error {
	// Check for trivial malicious patterns (stub implementation)
	// Real implementation would use proper antivirus scanning
	if len(data) > 50*1024*1024 {
		return fmt.Errorf("file suspiciously large")
	}
	return nil
}

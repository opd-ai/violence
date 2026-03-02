package registry

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/opd-ai/violence/pkg/mod"
)

func setupTestRegistry(t *testing.T) (*Registry, func()) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	storagePath := filepath.Join(tmpDir, "storage")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	reg, err := NewRegistry(db, storagePath)
	if err != nil {
		db.Close()
		t.Fatalf("Failed to create registry: %v", err)
	}

	cleanup := func() {
		reg.Close()
	}

	return reg, cleanup
}

func createValidWASM() []byte {
	// WASM magic number (0x00 0x61 0x73 0x6D) + version 1 (0x01 0x00 0x00 0x00)
	wasm := []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00}
	// Add some dummy content
	wasm = append(wasm, make([]byte, 100)...)
	return wasm
}

func createValidManifest() mod.Manifest {
	return mod.Manifest{
		Name:        "test-mod",
		Version:     "1.0.0",
		Author:      "test-author",
		Description: "A test mod",
		Tags:        []string{"test", "example"},
	}
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (string, string)
		wantErr bool
	}{
		{
			name: "valid_paths",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "test.db")
				storagePath := filepath.Join(tmpDir, "storage")
				return dbPath, storagePath
			},
			wantErr: false,
		},
		{
			name: "creates_storage_directory",
			setup: func(t *testing.T) (string, string) {
				tmpDir := t.TempDir()
				dbPath := filepath.Join(tmpDir, "test.db")
				storagePath := filepath.Join(tmpDir, "nested", "storage")
				return dbPath, storagePath
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath, storagePath := tt.setup(t)

			db, err := sql.Open("sqlite3", dbPath)
			if err != nil {
				t.Fatalf("Failed to open database: %v", err)
			}
			defer db.Close()

			reg, err := NewRegistry(db, storagePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRegistry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer reg.Close()

				// Verify storage directory exists
				if _, err := os.Stat(storagePath); os.IsNotExist(err) {
					t.Errorf("Storage directory not created: %v", err)
				}

				// Verify database schema
				var count int
				err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='mods'").Scan(&count)
				if err != nil {
					t.Errorf("Failed to query schema: %v", err)
				}
				if count != 1 {
					t.Errorf("Expected mods table to exist, got count=%d", count)
				}
			}
		})
	}
}

func TestValidateWASM(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid_wasm",
			data:    []byte{0x00, 0x61, 0x73, 0x6D, 0x01, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "too_small",
			data:    []byte{0x00, 0x61, 0x73},
			wantErr: true,
		},
		{
			name:    "invalid_magic",
			data:    []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "invalid_version",
			data:    []byte{0x00, 0x61, 0x73, 0x6D, 0x02, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "empty",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWASM(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateWASM() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVirusScanStub(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "normal_size",
			data:    make([]byte, 1024),
			wantErr: false,
		},
		{
			name:    "max_allowed",
			data:    make([]byte, 50*1024*1024),
			wantErr: false,
		},
		{
			name:    "too_large",
			data:    make([]byte, 51*1024*1024),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := virusScanStub(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("virusScanStub() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleUpload(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		setupRequest func(t *testing.T) *http.Request
		wantStatus   int
	}{
		{
			name:   "method_not_allowed",
			method: http.MethodGet,
			setupRequest: func(t *testing.T) *http.Request {
				return httptest.NewRequest(http.MethodGet, "/upload", nil)
			},
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "valid_upload",
			method: http.MethodPost,
			setupRequest: func(t *testing.T) *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Add WASM file
				wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
				wasmPart.Write(createValidWASM())

				// Add manifest file
				manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
				manifest := createValidManifest()
				manifestJSON, _ := json.Marshal(manifest)
				manifestPart.Write(manifestJSON)

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:   "missing_wasm",
			method: http.MethodPost,
			setupRequest: func(t *testing.T) *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Only add manifest
				manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
				manifest := createValidManifest()
				manifestJSON, _ := json.Marshal(manifest)
				manifestPart.Write(manifestJSON)

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "missing_manifest",
			method: http.MethodPost,
			setupRequest: func(t *testing.T) *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Only add WASM
				wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
				wasmPart.Write(createValidWASM())

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid_manifest",
			method: http.MethodPost,
			setupRequest: func(t *testing.T) *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Add WASM file
				wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
				wasmPart.Write(createValidWASM())

				// Add invalid manifest
				manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
				manifestPart.Write([]byte(`{"name": "INVALID-NAME"}`))

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid_wasm",
			method: http.MethodPost,
			setupRequest: func(t *testing.T) *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)

				// Add invalid WASM file
				wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
				wasmPart.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})

				// Add valid manifest
				manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
				manifest := createValidManifest()
				manifestJSON, _ := json.Marshal(manifest)
				manifestPart.Write(manifestJSON)

				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, cleanup := setupTestRegistry(t)
			defer cleanup()

			req := tt.setupRequest(t)
			w := httptest.NewRecorder()

			reg.HandleUpload(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleUpload() status = %d, want %d", w.Code, tt.wantStatus)
				t.Logf("Response body: %s", w.Body.String())
			}

			// Verify successful upload
			if tt.wantStatus == http.StatusCreated {
				var response map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["status"] != "success" {
					t.Errorf("Expected success status, got %v", response["status"])
				}
				if response["sha256"] == nil || response["sha256"] == "" {
					t.Errorf("Expected SHA256 in response")
				}
			}
		})
	}
}

func TestHandleSearch(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Upload test mods
	mods := []mod.Manifest{
		{
			Name:        "mod-one",
			Version:     "1.0.0",
			Author:      "author-a",
			Description: "First test mod",
			Tags:        []string{"combat", "weapons"},
		},
		{
			Name:        "mod-two",
			Version:     "1.0.0",
			Author:      "author-b",
			Description: "Second test mod",
			Tags:        []string{"graphics", "ui"},
		},
		{
			Name:        "mod-three",
			Version:     "2.0.0",
			Author:      "author-a",
			Description: "Third test mod",
			Tags:        []string{"combat", "ai"},
		},
	}

	for _, manifest := range mods {
		uploadTestMod(t, reg, manifest)
	}

	tests := []struct {
		name       string
		query      string
		wantCount  int
		checkNames []string
	}{
		{
			name:       "search_all",
			query:      "",
			wantCount:  3,
			checkNames: []string{"mod-one", "mod-two", "mod-three"},
		},
		{
			name:       "search_by_name",
			query:      "name=mod-one",
			wantCount:  1,
			checkNames: []string{"mod-one"},
		},
		{
			name:       "search_by_author",
			query:      "author=author-a",
			wantCount:  2,
			checkNames: []string{"mod-one", "mod-three"},
		},
		{
			name:       "search_by_tag",
			query:      "tag=combat",
			wantCount:  2,
			checkNames: []string{"mod-one", "mod-three"},
		},
		{
			name:      "search_no_results",
			query:     "name=nonexistent",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/search?"+tt.query, nil)
			w := httptest.NewRecorder()

			reg.HandleSearch(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("HandleSearch() status = %d, want %d", w.Code, http.StatusOK)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			count := int(response["count"].(float64))
			if count != tt.wantCount {
				t.Errorf("Got %d results, want %d", count, tt.wantCount)
			}

			if tt.checkNames != nil {
				results := response["results"].([]interface{})
				foundNames := make(map[string]bool)
				for _, r := range results {
					result := r.(map[string]interface{})
					foundNames[result["name"].(string)] = true
				}
				for _, name := range tt.checkNames {
					if !foundNames[name] {
						t.Errorf("Expected to find mod %q in results", name)
					}
				}
			}
		})
	}
}

func TestHandleDownload(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	manifest := createValidManifest()
	uploadTestMod(t, reg, manifest)

	tests := []struct {
		name       string
		path       string
		wantStatus int
		checkHash  bool
	}{
		{
			name:       "valid_download",
			path:       "/download/test-mod/1.0.0",
			wantStatus: http.StatusOK,
			checkHash:  true,
		},
		{
			name:       "not_found",
			path:       "/download/nonexistent/1.0.0",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid_path",
			path:       "/download/test-mod",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "method_not_allowed",
			path:       "/download/test-mod/1.0.0",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.name == "method_not_allowed" {
				method = http.MethodPost
			}

			req := httptest.NewRequest(method, tt.path, nil)
			w := httptest.NewRecorder()

			reg.HandleDownload(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("HandleDownload() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.checkHash {
				hash := w.Header().Get("X-Mod-SHA256")
				if hash == "" {
					t.Errorf("Expected SHA256 header in response")
				}
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/wasm" {
					t.Errorf("Expected Content-Type application/wasm, got %s", contentType)
				}
			}
		})
	}
}

func TestSetMaxModSize(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Test setting max size
	newSize := int64(5 * 1024 * 1024)
	reg.SetMaxModSize(newSize)

	if reg.maxModSize != newSize {
		t.Errorf("SetMaxModSize() = %d, want %d", reg.maxModSize, newSize)
	}

	// Test upload with oversized file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create file larger than max
	wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
	largeWASM := createValidWASM()
	largeWASM = append(largeWASM, make([]byte, 6*1024*1024)...)
	wasmPart.Write(largeWASM)

	manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
	manifest := createValidManifest()
	manifestJSON, _ := json.Marshal(manifest)
	manifestPart.Write(manifestJSON)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	reg.HandleUpload(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status %d for oversized file, got %d", http.StatusRequestEntityTooLarge, w.Code)
	}
}

func uploadTestMod(t *testing.T, reg *Registry, manifest mod.Manifest) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
	wasmPart.Write(createValidWASM())

	manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
	manifestJSON, _ := json.Marshal(manifest)
	manifestPart.Write(manifestJSON)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	reg.HandleUpload(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Failed to upload test mod: status=%d, body=%s", w.Code, w.Body.String())
	}
}

func TestDownloadIncrementsCounter(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	manifest := createValidManifest()
	uploadTestMod(t, reg, manifest)

	// Download twice
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/download/test-mod/1.0.0", nil)
		w := httptest.NewRecorder()
		reg.HandleDownload(w, req)
	}

	// Verify download count
	var downloads int
	err := reg.db.QueryRow("SELECT downloads FROM mods WHERE name = ? AND version = ?", "test-mod", "1.0.0").Scan(&downloads)
	if err != nil {
		t.Fatalf("Failed to query downloads: %v", err)
	}

	if downloads != 2 {
		t.Errorf("Expected 2 downloads, got %d", downloads)
	}
}

func TestUploadReplaceExisting(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	manifest := createValidManifest()

	// Upload first version
	uploadTestMod(t, reg, manifest)

	// Upload same version again (should replace)
	uploadTestMod(t, reg, manifest)

	// Verify only one record exists
	var count int
	err := reg.db.QueryRow("SELECT COUNT(*) FROM mods WHERE name = ? AND version = ?", manifest.Name, manifest.Version).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query count: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 mod record after replacement, got %d", count)
	}
}

func TestSearchMethodNotAllowed(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/search", nil)
	w := httptest.NewRecorder()

	reg.HandleSearch(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d for POST to /search, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandleUploadCleanupOnError(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	// Close database to force error during INSERT
	reg.db.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
	wasmPart.Write(createValidWASM())

	manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
	manifest := createValidManifest()
	manifestJSON, _ := json.Marshal(manifest)
	manifestPart.Write(manifestJSON)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	reg.HandleUpload(w, req)

	// Should get database error
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d for database error, got %d", http.StatusInternalServerError, w.Code)
	}

	// Verify WASM file was cleaned up
	modPath := filepath.Join(reg.storagePath, "test-mod-1.0.0.wasm")
	if _, err := os.Stat(modPath); !os.IsNotExist(err) {
		t.Errorf("Expected WASM file to be cleaned up on error")
	}
}

func TestHandleUploadInvalidJSON(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
	wasmPart.Write(createValidWASM())

	manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
	manifestPart.Write([]byte("{invalid json"))

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	reg.HandleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleUploadReadError(t *testing.T) {
	reg, cleanup := setupTestRegistry(t)
	defer cleanup()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a reader that will fail
	wasmPart, _ := writer.CreateFormFile("wasm", "test.wasm")
	io.WriteString(wasmPart, "short")

	manifestPart, _ := writer.CreateFormFile("manifest", "mod.json")
	manifest := createValidManifest()
	manifestJSON, _ := json.Marshal(manifest)
	manifestPart.Write(manifestJSON)

	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	reg.HandleUpload(w, req)

	// Should fail WASM validation
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for invalid WASM, got %d", http.StatusBadRequest, w.Code)
	}
}

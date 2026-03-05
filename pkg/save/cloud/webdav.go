package cloud

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/studio-b12/gowebdav"
)

// WebDAVConfig holds WebDAV backend configuration.
type WebDAVConfig struct {
	URL      string
	Username string
	Password string
	BasePath string
}

// webDAVClient defines the interface for WebDAV operations.
type webDAVClient interface {
	Write(path string, data []byte, perm os.FileMode) error
	Read(path string) ([]byte, error)
	ReadDir(path string) ([]os.FileInfo, error)
	Remove(path string) error
	MkdirAll(path string, perm os.FileMode) error
}

// WebDAVProvider implements Provider for WebDAV storage.
type WebDAVProvider struct {
	client   webDAVClient
	basePath string
}

// NewWebDAVProvider creates a WebDAV cloud save provider.
func NewWebDAVProvider(cfg WebDAVConfig) (*WebDAVProvider, error) {
	if cfg.URL == "" {
		return nil, errors.New("URL is required")
	}

	client := gowebdav.NewClient(cfg.URL, cfg.Username, cfg.Password)
	basePath := strings.TrimSuffix(cfg.BasePath, "/")
	if basePath == "" {
		basePath = "/saves"
	}

	return &WebDAVProvider{
		client:   client,
		basePath: basePath,
	}, nil
}

func (p *WebDAVProvider) key(slotID int) string {
	return fmt.Sprintf("%s/slot-%d.sav", p.basePath, slotID)
}

func (p *WebDAVProvider) metadataKey(slotID int) string {
	return fmt.Sprintf("%s/slot-%d.meta.json", p.basePath, slotID)
}

// Upload uploads save data to WebDAV.
func (p *WebDAVProvider) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	if err := p.client.MkdirAll(p.basePath, 0o755); err != nil {
		return fmt.Errorf("create base directory: %w", err)
	}

	if err := p.client.Write(p.key(slotID), data, 0o644); err != nil {
		return fmt.Errorf("upload save: %w", err)
	}

	if err := p.client.Write(p.metadataKey(slotID), metaJSON, 0o644); err != nil {
		return fmt.Errorf("upload metadata: %w", err)
	}

	return nil
}

// Download retrieves save data from WebDAV.
func (p *WebDAVProvider) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	metadata, err := p.GetMetadata(ctx, slotID)
	if err != nil {
		return nil, SaveMetadata{}, err
	}

	data, err := p.client.Read(p.key(slotID))
	if err != nil {
		if isNotFoundError(err) {
			return nil, SaveMetadata{}, ErrNotFound
		}
		return nil, SaveMetadata{}, fmt.Errorf("read save: %w", err)
	}

	return data, metadata, nil
}

func parseMetadataFilename(filename string) (int, bool) {
	var slotID int
	if _, err := fmt.Sscanf(filename, "slot-%d.meta.json", &slotID); err != nil {
		return 0, false
	}
	return slotID, true
}

// List returns metadata for all saves in WebDAV.
func (p *WebDAVProvider) List(ctx context.Context) ([]SaveMetadata, error) {
	files, err := p.client.ReadDir(p.basePath)
	if err != nil {
		if isNotFoundError(err) {
			return []SaveMetadata{}, nil
		}
		return nil, fmt.Errorf("list files: %w", err)
	}

	var metadatas []SaveMetadata
	seen := make(map[int]bool)

	for _, file := range files {
		slotID, ok := parseMetadataFilename(file.Name())
		if !ok || seen[slotID] {
			continue
		}
		seen[slotID] = true

		if meta, err := p.GetMetadata(ctx, slotID); err == nil {
			metadatas = append(metadatas, meta)
		}
	}

	return metadatas, nil
}

// Delete removes save data from WebDAV.
func (p *WebDAVProvider) Delete(ctx context.Context, slotID int) error {
	if err := p.client.Remove(p.key(slotID)); err != nil && !isNotFoundError(err) {
		return fmt.Errorf("delete save: %w", err)
	}

	if err := p.client.Remove(p.metadataKey(slotID)); err != nil && !isNotFoundError(err) {
		return fmt.Errorf("delete metadata: %w", err)
	}

	return nil
}

// GetMetadata retrieves save metadata from WebDAV.
func (p *WebDAVProvider) GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error) {
	data, err := p.client.Read(p.metadataKey(slotID))
	if err != nil {
		if isNotFoundError(err) {
			return SaveMetadata{}, ErrNotFound
		}
		return SaveMetadata{}, fmt.Errorf("read metadata: %w", err)
	}

	var metadata SaveMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return SaveMetadata{}, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return metadata, nil
}

func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "404") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "Not Found")
}

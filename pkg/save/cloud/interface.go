// Package cloud provides cloud save synchronization interfaces.
package cloud

import (
	"context"
	"time"
)

// SaveMetadata contains metadata for a cloud-stored save file.
type SaveMetadata struct {
	SlotID       int       `json:"slot_id"`
	Timestamp    time.Time `json:"timestamp"`
	Version      string    `json:"version"`
	Genre        string    `json:"genre"`
	Seed         int64     `json:"seed"`
	Size         int64     `json:"size"`
	Checksum     string    `json:"checksum"`
	LastModified time.Time `json:"last_modified"`
}

// ConflictResolution defines how to handle save conflicts.
type ConflictResolution int

const (
	// KeepLocal keeps the local save file and discards cloud version.
	KeepLocal ConflictResolution = iota
	// KeepCloud keeps the cloud save file and discards local version.
	KeepCloud
	// KeepBoth keeps both versions by creating a new slot for cloud data.
	KeepBoth
)

// Provider defines the interface for cloud save backends.
type Provider interface {
	// Upload uploads a save file to cloud storage.
	Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error

	// Download retrieves a save file from cloud storage.
	Download(ctx context.Context, slotID int) (data []byte, metadata SaveMetadata, err error)

	// List returns metadata for all cloud-stored save files.
	List(ctx context.Context) ([]SaveMetadata, error)

	// Delete removes a save file from cloud storage.
	Delete(ctx context.Context, slotID int) error

	// GetMetadata retrieves only metadata without downloading the full file.
	GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error)
}

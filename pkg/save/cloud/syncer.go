package cloud

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrConflict indicates a save conflict between local and cloud versions.
	ErrConflict = errors.New("save conflict detected")
	// ErrNoSlotAvailable indicates no empty slot for KeepBoth resolution.
	ErrNoSlotAvailable = errors.New("no slot available")
)

// Syncer manages cloud save synchronization with conflict resolution.
type Syncer struct {
	provider Provider
	maxSlots int
}

// NewSyncer creates a new cloud save synchronizer.
func NewSyncer(provider Provider, maxSlots int) *Syncer {
	return &Syncer{
		provider: provider,
		maxSlots: maxSlots,
	}
}

// computeChecksum computes SHA256 checksum of data.
func computeChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Upload uploads local save data to cloud storage.
func (s *Syncer) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	metadata.Checksum = computeChecksum(data)
	metadata.Size = int64(len(data))
	metadata.LastModified = time.Now()
	return s.provider.Upload(ctx, slotID, data, metadata)
}

// Download retrieves save data from cloud storage.
func (s *Syncer) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	data, metadata, err := s.provider.Download(ctx, slotID)
	if err != nil {
		return nil, SaveMetadata{}, err
	}
	if computeChecksum(data) != metadata.Checksum {
		return nil, SaveMetadata{}, errors.New("checksum mismatch")
	}
	return data, metadata, nil
}

// Sync synchronizes a save slot between local and cloud.
func (s *Syncer) Sync(ctx context.Context, slotID int, localData []byte, localMeta SaveMetadata, resolution ConflictResolution) error {
	cloudMeta, err := s.provider.GetMetadata(ctx, slotID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("get metadata: %w", err)
	}

	if errors.Is(err, ErrNotFound) {
		return s.Upload(ctx, slotID, localData, localMeta)
	}

	if localMeta.Timestamp.After(cloudMeta.Timestamp) {
		return s.Upload(ctx, slotID, localData, localMeta)
	}

	if cloudMeta.Timestamp.After(localMeta.Timestamp) {
		return ErrConflict
	}

	return nil
}

// List returns metadata for all cloud saves.
func (s *Syncer) List(ctx context.Context) ([]SaveMetadata, error) {
	return s.provider.List(ctx)
}

// Delete removes a save from cloud storage.
func (s *Syncer) Delete(ctx context.Context, slotID int) error {
	return s.provider.Delete(ctx, slotID)
}

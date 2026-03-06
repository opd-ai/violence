# Cloud Save Encryption at Rest

## Overview

The cloud save encryption feature provides AES-256-GCM encryption for save files stored in cloud backends (S3, WebDAV, etc.). This ensures that save data is encrypted before leaving the client and remains encrypted at rest in cloud storage.

## Security Specifications

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Derivation**: PBKDF2-HMAC-SHA256 with 100,000 iterations
- **Key Size**: 256 bits (32 bytes)
- **Salt Size**: 256 bits (32 bytes), randomly generated per encryption
- **Nonce Size**: 96 bits (12 bytes), randomly generated per encryption
- **Authentication**: GCM provides authenticated encryption with 128-bit authentication tag

## Usage

### Basic Example

```go
package main

import (
	"context"
	"github.com/opd-ai/violence/pkg/save/cloud"
)

func main() {
	// Create a base provider (S3, WebDAV, etc.)
	s3Provider, _ := cloud.NewS3Provider(cloud.S3Config{
		Bucket:    "my-game-saves",
		Region:    "us-west-2",
		AccessKey: "...",
		SecretKey: "...",
	})

	// Wrap with encryption
	password := "user-password-from-secure-input"
	encryptedProvider := cloud.NewEncryptedProvider(s3Provider, password)

	// Use normally - all data is automatically encrypted/decrypted
	ctx := context.Background()
	saveData := []byte("... save file data ...")
	metadata := cloud.SaveMetadata{
		SlotID:  1,
		Genre:   "fantasy",
		Version: "1.0",
	}

	// Upload (encrypts automatically)
	err := encryptedProvider.Upload(ctx, 1, saveData, metadata)

	// Download (decrypts automatically)
	data, meta, err := encryptedProvider.Download(ctx, 1)
}
```

### Encryption Format

Encrypted save files use the following binary format:

```
[Salt: 32 bytes][Nonce: 12 bytes][Ciphertext + Auth Tag: variable]
```

- **Salt**: Random salt for PBKDF2 key derivation
- **Nonce**: Random nonce for AES-GCM (never reused)
- **Ciphertext**: Encrypted save data
- **Auth Tag**: 16-byte GCM authentication tag (appended by GCM)

### Password Requirements

The encryption is only as strong as the password. Recommendations:

- **Minimum Length**: 12 characters
- **Complexity**: Mix of uppercase, lowercase, numbers, and symbols
- **Uniqueness**: Don't reuse passwords from other services
- **Storage**: Never hardcode passwords; use secure input methods

### Error Handling

```go
data, _, err := encryptedProvider.Download(ctx, slotID)
if err == cloud.ErrDecryptionFailed {
	// Wrong password or corrupted data
	log.Println("Decryption failed - check password")
} else if err == cloud.ErrInvalidCiphertext {
	// Corrupted or truncated ciphertext
	log.Println("Invalid encrypted data")
} else if err == cloud.ErrNotFound {
	// Save file doesn't exist
	log.Println("Save not found")
}
```

## Architecture

### Component Diagram

```
┌─────────────┐
│ Application │
└──────┬──────┘
       │
       v
┌──────────────────┐
│ EncryptedProvider│ ← Encryption/Decryption Layer
└──────┬───────────┘
       │
       v
┌──────────────┐
│   Provider   │ ← Backend (S3Provider, WebDAVProvider, etc.)
│  (S3/WebDAV) │
└──────┬───────┘
       │
       v
┌──────────────┐
│ Cloud Storage│
└──────────────┘
```

### Wrapper Pattern

`EncryptedProvider` wraps any `Provider` implementation:

```go
type Provider interface {
	Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error
	Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error)
	// ... other methods
}
```

This allows encryption to be added to any backend without modifying the backend implementation.

## Security Considerations

### What is Protected

✅ **Save file contents**: Game progress, inventory, settings  
✅ **Authentication**: GCM provides cryptographic integrity verification  
✅ **Password-based encryption**: Only users with the password can decrypt  

### What is NOT Protected

❌ **Metadata**: Slot ID, timestamps, genre, version are stored unencrypted  
❌ **File existence**: Attackers can see which save slots exist  
❌ **File size**: Approximate save file size is visible (with ~60 byte overhead)  
❌ **Weak passwords**: PBKDF2 iterations slow brute-force but can't prevent weak passwords  

### Threat Model

**Protects Against:**
- Cloud storage provider reading save data
- Attackers with access to cloud storage
- Data breaches of cloud storage
- Unauthorized access to backups

**Does NOT Protect Against:**
- Compromised client (malware, keyloggers)
- Side-channel attacks (timing, power analysis)
- Attacks on the client while game is running
- Social engineering to obtain password

## Performance

### Benchmarks

On modern hardware (approximate):

- **Encryption**: ~200 MB/s for 100KB saves
- **Decryption**: ~250 MB/s for 100KB saves
- **Key Derivation**: ~10ms per operation (100,000 PBKDF2 iterations)

Key derivation happens once per upload/download, so the overhead is minimal for typical save file sizes (<1MB).

### Optimization Tips

1. **Cache keys**: For bulk operations, consider caching the derived key
2. **Async encryption**: Encrypt in background while player continues
3. **Compress first**: Compress save data before encryption for smaller uploads

## Testing

Run the encryption tests:

```bash
go test ./pkg/save/cloud/... -run TestEncrypt -v
```

Test coverage:

- Round-trip encryption/decryption
- Wrong password detection
- Invalid ciphertext handling
- Nonce uniqueness verification
- Key derivation determinism
- Large file encryption (1MB+)

## Migration Guide

### Enabling Encryption for Existing Saves

```go
// 1. Download unencrypted saves
oldProvider := cloud.NewS3Provider(config)
data, metadata, _ := oldProvider.Download(ctx, slotID)

// 2. Create encrypted provider
password := getPasswordFromUser()
encProvider := cloud.NewEncryptedProvider(oldProvider, password)

// 3. Re-upload with encryption
err := encProvider.Upload(ctx, slotID, data, metadata)
```

### Disabling Encryption

```go
// 1. Download and decrypt
encProvider := cloud.NewEncryptedProvider(backend, password)
data, metadata, _ := encProvider.Download(ctx, slotID)

// 2. Upload unencrypted to new backend
newBackend := cloud.NewS3Provider(newConfig)
err := newBackend.Upload(ctx, slotID, data, metadata)
```

## FAQ

**Q: Can I change my password?**  
A: Yes. Download all saves with the old password, then re-upload with a new `EncryptedProvider` using the new password.

**Q: What if I forget my password?**  
A: Encrypted saves are unrecoverable without the correct password. There is no password reset mechanism (by design).

**Q: Is the encryption format compatible across versions?**  
A: Yes. The binary format (salt + nonce + ciphertext) is stable and backward-compatible.

**Q: Can I use different passwords for different slots?**  
A: No, currently one password per provider. Consider using multiple provider instances if needed.

**Q: Does this work with WebDAV and S3?**  
A: Yes, `EncryptedProvider` wraps any `Provider` implementation.

## References

- [AES-GCM Specification](https://csrc.nist.gov/publications/detail/sp/800-38d/final)
- [PBKDF2 RFC 2898](https://tools.ietf.org/html/rfc2898)
- [Go crypto/cipher Documentation](https://pkg.go.dev/crypto/cipher)

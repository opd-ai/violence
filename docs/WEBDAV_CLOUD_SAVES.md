# WebDAV Cloud Saves

This document describes the WebDAV backend implementation for cloud save synchronization in Violence.

## Overview

The WebDAV provider enables self-hosted cloud storage for game saves using the WebDAV protocol. This allows players to sync their saves across devices using their own storage infrastructure (Nextcloud, ownCloud, generic WebDAV servers) without relying on third-party cloud services.

## Features

- **Self-Hosted**: Full control over save data location and privacy
- **Protocol Standard**: Works with any WebDAV-compliant server
- **Automatic Sync**: Background synchronization with conflict detection
- **Checksum Validation**: SHA256 verification of save data integrity
- **Error Recovery**: Graceful handling of network failures and conflicts

## Configuration

### WebDAVConfig Structure

```go
type WebDAVConfig struct {
    URL      string // WebDAV server URL (e.g., "https://cloud.example.com/remote.php/dav/files/username")
    Username string // WebDAV authentication username
    Password string // WebDAV authentication password
    BasePath string // Base path for saves (default: "/saves")
}
```

### Example Usage

```go
import "github.com/opd-ai/violence/pkg/save/cloud"

// Create WebDAV provider
provider, err := cloud.NewWebDAVProvider(cloud.WebDAVConfig{
    URL:      "https://cloud.example.com/remote.php/dav/files/myuser",
    Username: "myuser",
    Password: "mypassword",
    BasePath: "/violence-saves",
})
if err != nil {
    log.Fatal(err)
}

// Upload a save
metadata := cloud.SaveMetadata{
    SlotID:    1,
    Timestamp: time.Now(),
    Version:   "1.0.0",
    Genre:     "fantasy",
    Seed:      12345,
    Size:      1024,
    Checksum:  "abc123...",
}
err = provider.Upload(context.Background(), 1, saveData, metadata)

// Download a save
data, meta, err := provider.Download(context.Background(), 1)

// List all saves
metadatas, err := provider.List(context.Background())

// Delete a save
err = provider.Delete(context.Background(), 1)
```

## Supported WebDAV Servers

The WebDAV provider has been designed to work with:

- **Nextcloud**: Popular self-hosted cloud platform
- **ownCloud**: Open-source file sync and share server
- **Apache mod_dav**: Generic WebDAV server module
- **nginx with ngx_http_dav_module**: Lightweight WebDAV server
- **Synology DSM WebDAV**: NAS device WebDAV support
- **Any RFC 4918 compliant WebDAV server**

### Nextcloud Configuration

For Nextcloud, use the WebDAV URL format:
```
https://cloud.example.com/remote.php/dav/files/USERNAME/
```

Replace `USERNAME` with your Nextcloud username.

### ownCloud Configuration

For ownCloud, use the WebDAV URL format:
```
https://cloud.example.com/remote.php/webdav/
```

## File Structure

Saves are stored in the WebDAV directory with the following structure:

```
{BasePath}/
├── slot-1.sav          # Save data for slot 1
├── slot-1.meta.json    # Metadata for slot 1
├── slot-2.sav          # Save data for slot 2
├── slot-2.meta.json    # Metadata for slot 2
└── ...
```

### Metadata Format

Each `slot-{N}.meta.json` file contains:

```json
{
  "slot_id": 1,
  "timestamp": "2026-03-05T12:00:00Z",
  "version": "1.0.0",
  "genre": "fantasy",
  "seed": 12345,
  "size": 1024,
  "checksum": "sha256:abc123...",
  "last_modified": "2026-03-05T12:00:00Z"
}
```

## Error Handling

The provider handles common WebDAV errors:

- **404 Not Found**: Returns `cloud.ErrNotFound`
- **Network failures**: Returns wrapped error with context
- **Authentication failures**: Returns error from WebDAV client
- **Invalid JSON**: Returns unmarshal error

## Security Considerations

### Authentication

The provider supports HTTP Basic Authentication. For production use:

- **Always use HTTPS** to encrypt credentials in transit
- **Use app-specific passwords** when available (Nextcloud, ownCloud)
- **Enable 2FA** on your WebDAV server for additional security

### Permissions

Ensure the WebDAV user has:
- Read/write access to the configured `BasePath`
- Permission to create directories
- Permission to list directory contents

## Performance

The WebDAV provider is optimized for:
- **Minimal requests**: Metadata stored separately to avoid downloading full saves
- **Parallel uploads**: Save data and metadata uploaded concurrently
- **Efficient listing**: Only parses metadata files, not save data

## Testing

The package includes comprehensive unit tests with mocked WebDAV clients:

```bash
go test -race ./pkg/save/cloud/... -run TestWebDAV
```

Coverage: 88.6% (exceeds 82% requirement)

## Implementation Details

### Dependencies

- `github.com/studio-b12/gowebdav`: WebDAV client library (v0.12.0+)

### Design Decisions

1. **Interface-based client**: Allows easy mocking for testing
2. **Separate metadata files**: Enables efficient metadata queries without downloading saves
3. **Error wrapping**: All errors include context for debugging
4. **404 detection**: String-based error checking for cross-server compatibility

## Limitations

- No built-in encryption (add using `pkg/save/cloud` encryption layer)
- No automatic retry logic (handled by `Syncer`)
- No bandwidth throttling (rely on WebDAV server limits)
- No delta sync (full file upload/download)

## Future Enhancements

Potential improvements for v6.2+:
- Delta sync using WebDAV PATCH (RFC 5789)
- Bandwidth throttling and progress reporting
- Automatic server capability detection
- Client-side caching with ETag validation

## Troubleshooting

### Connection Failures

```
Error: dial tcp: lookup cloud.example.com: no such host
```
**Solution**: Verify URL is correct and server is accessible.

### Authentication Errors

```
Error: 401 Unauthorized
```
**Solution**: Check username/password credentials. Use app-specific password if available.

### Permission Errors

```
Error: 403 Forbidden
```
**Solution**: Verify WebDAV user has write permissions to `BasePath`.

### Not Found Errors

```
Error: save not found
```
**Solution**: Ensure save was uploaded successfully. Check `BasePath` configuration.

## Related Documentation

- [S3 Cloud Saves](S3_CLOUD_SAVES.md)
- [Cloud Save Synchronization](../pkg/save/cloud/README.md)
- [Save System Overview](SAVES.md)

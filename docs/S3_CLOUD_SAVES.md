# S3-Compatible Cloud Save Backend

## Overview

The S3 cloud save backend enables cross-platform save synchronization using any S3-compatible object storage service including:
- Amazon S3
- MinIO (self-hosted)
- Backblaze B2
- DigitalOcean Spaces
- Wasabi
- Cloudflare R2

## Configuration

### AWS S3
```go
import "github.com/opd-ai/violence/pkg/save/cloud"

provider, err := cloud.NewS3Provider(cloud.S3Config{
    Region:          "us-east-1",
    Bucket:          "my-game-saves",
    AccessKeyID:     "AKIAIOSFODNN7EXAMPLE",
    SecretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
})
```

### MinIO (Self-Hosted)
```go
provider, err := cloud.NewS3Provider(cloud.S3Config{
    Endpoint:        "http://localhost:9000",
    Bucket:          "game-saves",
    AccessKeyID:     "minioadmin",
    SecretAccessKey: "minioadmin",
    Region:          "us-east-1", // Required even for MinIO
})
```

### Backblaze B2
```go
provider, err := cloud.NewS3Provider(cloud.S3Config{
    Endpoint:        "https://s3.us-west-004.backblazeb2.com",
    Bucket:          "my-b2-bucket",
    AccessKeyID:     "your-application-key-id",
    SecretAccessKey: "your-application-key",
    Region:          "us-west-004",
})
```

## Usage

### Upload Save
```go
syncer := cloud.NewSyncer(provider, 10) // 10 max slots

metadata := cloud.SaveMetadata{
    SlotID:    1,
    Timestamp: time.Now(),
    Version:   "1.0.0",
    Genre:     "fantasy",
    Seed:      12345,
}

err = syncer.Upload(context.Background(), 1, saveData, metadata)
```

### Download Save
```go
data, metadata, err := syncer.Download(context.Background(), 1)
if err != nil {
    log.Fatal(err)
}
```

### List All Saves
```go
saves, err := syncer.List(context.Background())
for _, meta := range saves {
    fmt.Printf("Slot %d: Version %s, Genre %s\n", 
        meta.SlotID, meta.Version, meta.Genre)
}
```

### Synchronize Local and Cloud
```go
err = syncer.Sync(context.Background(), slotID, localData, localMeta, cloud.KeepCloud)
if errors.Is(err, cloud.ErrConflict) {
    // Handle conflict - let user choose resolution
}
```

## Storage Layout

The S3 backend uses the following key structure:
```
saves/
  slot-1.sav          # Save data binary
  slot-1.meta.json    # Metadata JSON
  slot-2.sav
  slot-2.meta.json
  ...
```

## Security Considerations

1. **Credentials**: Never hardcode credentials. Use environment variables or AWS credentials file.
2. **Bucket Permissions**: Create dedicated buckets with restrictive IAM policies.
3. **Encryption**: The backend stores data as-is. Use S3 server-side encryption or client-side encryption for sensitive saves.
4. **Checksums**: All downloads are verified against SHA256 checksums to detect corruption.

## Performance

- **Upload**: Single PutObject call per save + metadata
- **Download**: Single GetObject call with checksum verification
- **List**: Single ListObjectsV2 call with filtering
- **Delete**: Two DeleteObject calls (save + metadata)

## Error Handling

The provider returns standard errors:
- `cloud.ErrNotFound`: Save does not exist
- `cloud.ErrConflict`: Synchronization conflict detected
- AWS SDK errors: Network, authentication, permission issues

## Testing

Run tests with:
```bash
go test ./pkg/save/cloud/... -v
```

For integration testing with a real S3 backend:
```bash
# Start MinIO locally
docker run -p 9000:9000 minio/minio server /data

# Run your application with S3Config pointing to localhost:9000
```

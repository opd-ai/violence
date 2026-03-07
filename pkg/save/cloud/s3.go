//go:build !js

package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

// S3Config holds S3-compatible backend configuration.
type S3Config struct {
	Region          string
	Bucket          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
}

// S3Provider implements Provider for S3-compatible storage.
type S3Provider struct {
	client *s3.Client
	bucket string
}

func buildConfigOpts(cfg S3Config) []func(*config.LoadOptions) error {
	var opts []func(*config.LoadOptions) error
	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}
	return opts
}

func buildClientOpts(endpoint string) []func(*s3.Options) {
	if endpoint == "" {
		return nil
	}
	return []func(*s3.Options){func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	}}
}

// NewS3Provider creates an S3-compatible cloud save provider.
func NewS3Provider(cfg S3Config) (*S3Provider, error) {
	if cfg.Bucket == "" {
		return nil, errors.New("bucket is required")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), buildConfigOpts(cfg)...)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return &S3Provider{
		client: s3.NewFromConfig(awsCfg, buildClientOpts(cfg.Endpoint)...),
		bucket: cfg.Bucket,
	}, nil
}

func (p *S3Provider) key(slotID int) string {
	return fmt.Sprintf("saves/slot-%d.sav", slotID)
}

func (p *S3Provider) metadataKey(slotID int) string {
	return fmt.Sprintf("saves/slot-%d.meta.json", slotID)
}

// Upload uploads save data to S3.
func (p *S3Provider) Upload(ctx context.Context, slotID int, data []byte, metadata SaveMetadata) error {
	metaJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &p.bucket,
		Key:    aws.String(p.key(slotID)),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("upload save: %w", err)
	}

	_, err = p.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &p.bucket,
		Key:    aws.String(p.metadataKey(slotID)),
		Body:   bytes.NewReader(metaJSON),
	})
	if err != nil {
		return fmt.Errorf("upload metadata: %w", err)
	}

	return nil
}

// Download retrieves save data from S3.
func (p *S3Provider) Download(ctx context.Context, slotID int) ([]byte, SaveMetadata, error) {
	metadata, err := p.GetMetadata(ctx, slotID)
	if err != nil {
		return nil, SaveMetadata{}, err
	}

	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    aws.String(p.key(slotID)),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, SaveMetadata{}, ErrNotFound
		}
		return nil, SaveMetadata{}, fmt.Errorf("get object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, SaveMetadata{}, fmt.Errorf("read body: %w", err)
	}

	return data, metadata, nil
}

func parseSlotID(key *string) (int, bool) {
	if key == nil {
		return 0, false
	}
	var slotID int
	if _, err := fmt.Sscanf(*key, "saves/slot-%d.meta.json", &slotID); err != nil {
		return 0, false
	}
	return slotID, true
}

// List returns metadata for all saves in S3.
func (p *S3Provider) List(ctx context.Context) ([]SaveMetadata, error) {
	result, err := p.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: &p.bucket,
		Prefix: aws.String("saves/"),
	})
	if err != nil {
		return nil, fmt.Errorf("list objects: %w", err)
	}

	var metadatas []SaveMetadata
	seen := make(map[int]bool)

	for _, obj := range result.Contents {
		slotID, ok := parseSlotID(obj.Key)
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

// Delete removes save data from S3.
func (p *S3Provider) Delete(ctx context.Context, slotID int) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &p.bucket,
		Key:    aws.String(p.key(slotID)),
	})
	if err != nil {
		return fmt.Errorf("delete save: %w", err)
	}

	_, err = p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: &p.bucket,
		Key:    aws.String(p.metadataKey(slotID)),
	})
	if err != nil {
		return fmt.Errorf("delete metadata: %w", err)
	}

	return nil
}

// GetMetadata retrieves save metadata from S3.
func (p *S3Provider) GetMetadata(ctx context.Context, slotID int) (SaveMetadata, error) {
	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &p.bucket,
		Key:    aws.String(p.metadataKey(slotID)),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		var apiErr smithy.APIError
		if errors.As(err, &nsk) || (errors.As(err, &apiErr) && apiErr.ErrorCode() == "NoSuchKey") {
			return SaveMetadata{}, ErrNotFound
		}
		return SaveMetadata{}, fmt.Errorf("get metadata: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return SaveMetadata{}, fmt.Errorf("read metadata: %w", err)
	}

	var metadata SaveMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return SaveMetadata{}, fmt.Errorf("unmarshal metadata: %w", err)
	}

	return metadata, nil
}

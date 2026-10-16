// Package storage provides abstraction interfaces and implementations for object storage backends (e.g., S3/MinIO).
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageProvider defines the contract for interacting with cloud object storage providers like MinIO or AWS S3.
type StorageProvider interface {
	// Upload streams an object from file reader with given size and contentType into the target bucket and object key.
	Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error
	// Delete removes an object identified by bucket and key from the storage backend.
	Delete(ctx context.Context, bucket, key string) error
	// GetPresignedURL creates a time-limited signed URL for downloading or viewing an object.
	GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
	// GetPublicURL returns a direct public URL for accessing an object when the bucket or CDN is publicly accessible.
	GetPublicURL(bucket, key string) string
}

// MinioStorageProvider implements StorageProvider backed by MinIO Go SDK.
type MinioStorageProvider struct {
	// Client is the underlying MinIO S3 API client.
	Client *minio.Client
	// publicURL is the base public endpoint URL (e.g. "https://cdn.example.com" or "http://minio:9000").
	publicURL string
	// knownBuckets caches verified or created buckets in-memory to eliminate redundant BucketExists round-trips.
	knownBuckets sync.Map
}

// NewMinioStorageProvider initializes and configures a new MinioStorageProvider instance.
func NewMinioStorageProvider(endpoint, accessKey, secretKey, publicURL string) (*MinioStorageProvider, error) {
	useSSL := false
	trimmedEndpoint := strings.TrimSpace(endpoint)

	if strings.Contains(trimmedEndpoint, "://") {
		u, err := url.Parse(trimmedEndpoint)
		if err != nil {
			return nil, fmt.Errorf("parse minio endpoint URL: %w", err)
		}
		if strings.EqualFold(u.Scheme, "https") {
			useSSL = true
		}
		trimmedEndpoint = u.Host
	} else {
		trimmedEndpoint = strings.TrimRight(trimmedEndpoint, "/")
	}

	cleanPublicURL := strings.TrimSpace(publicURL)
	cleanPublicURL = strings.TrimRight(cleanPublicURL, "/")
	if cleanPublicURL != "" && !strings.Contains(cleanPublicURL, "://") {
		cleanPublicURL = "http://" + cleanPublicURL
	}

	minioClient, err := minio.New(trimmedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	return &MinioStorageProvider{
		Client:    minioClient,
		publicURL: cleanPublicURL,
	}, nil
}

// ensureBucket checks if the specified bucket exists, and creates it if it does not.
// It caches verified buckets in sync.Map to avoid redundant network round-trips on every upload,
// handles race conditions if another worker created the bucket concurrently, and specifies region options.
func (m *MinioStorageProvider) ensureBucket(ctx context.Context, bucket string) error {
	if _, ok := m.knownBuckets.Load(bucket); ok {
		return nil
	}

	exists, err := m.Client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}

	if !exists {
		err = m.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
			Region: "us-east-1",
		})
		if err != nil {
			errResp := minio.ToErrorResponse(err)
			if errResp.Code != "BucketAlreadyOwnedByYou" && errResp.Code != "BucketAlreadyExists" {
				return fmt.Errorf("make bucket: %w", err)
			}
		}
	}

	m.knownBuckets.Store(bucket, struct{}{})
	return nil
}

// Upload uploads an object to the specified bucket and key using streaming PutObject.
// Sanitizes the object key against path traversal and leading slashes, and configures part size for unknown sizes.
func (m *MinioStorageProvider) Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error {
	cleanKey := path.Clean(key)
	cleanKey = strings.TrimPrefix(cleanKey, "/")
	if strings.HasPrefix(cleanKey, "../") || cleanKey == ".." || cleanKey == "." || cleanKey == "" {
		return fmt.Errorf("invalid object key: %s", key)
	}

	if err := m.ensureBucket(ctx, bucket); err != nil {
		return fmt.Errorf("ensure bucket: %w", err)
	}

	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if size <= 0 {
		opts.PartSize = 16 * 1024 * 1024
	}

	_, err := m.Client.PutObject(ctx, bucket, cleanKey, file, size, opts)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}

// Delete removes an object identified by bucket and key from the storage backend.
func (m *MinioStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	cleanKey := strings.TrimPrefix(path.Clean(key), "/")
	if cleanKey == "" || cleanKey == "." {
		return fmt.Errorf("invalid object key: %s", key)
	}

	err := m.Client.RemoveObject(ctx, bucket, cleanKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}

// GetPresignedURL generates a presigned GET URL for accessing a private object within the duration window.
// If publicURL is configured, internal Docker/cluster hostnames are rewritten to the public origin URL.
func (m *MinioStorageProvider) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	cleanKey := strings.TrimPrefix(path.Clean(key), "/")
	presignedURL, err := m.Client.PresignedGetObject(ctx, bucket, cleanKey, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presigned get object: %w", err)
	}

	if m.publicURL != "" {
		pubURL, parseErr := url.Parse(m.publicURL)
		if parseErr == nil && pubURL.Host != "" {
			presignedURL.Scheme = pubURL.Scheme
			presignedURL.Host = pubURL.Host
		}
	}

	return presignedURL.String(), nil
}

// GetPublicURL returns a direct public URL for an object based on the configured publicURL base.
// Object key path segments are properly URL-escaped to ensure valid URLs with spaces and special characters.
func (m *MinioStorageProvider) GetPublicURL(bucket, key string) string {
	cleanKey := strings.TrimPrefix(path.Clean(key), "/")
	parts := strings.Split(cleanKey, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	escapedKey := strings.Join(parts, "/")

	baseURL := strings.TrimRight(m.publicURL, "/")
	if baseURL == "" {
		return fmt.Sprintf("/%s/%s", bucket, escapedKey)
	}
	return fmt.Sprintf("%s/%s/%s", baseURL, bucket, escapedKey)
}

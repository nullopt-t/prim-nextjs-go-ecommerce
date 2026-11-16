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

// S3StorageProvider is an alias for the S3-compatible storage backend.
type S3StorageProvider = MinioStorageProvider

// MinioStorageProvider implements StorageProvider backed by MinIO Go SDK.
type MinioStorageProvider struct {
	// Client is the underlying MinIO S3 API client.
	Client *minio.Client
	// publicURL is the base public endpoint URL (e.g. "https://cdn.example.com" or "http://minio:9000").
	publicURL string
	// region is the S3 bucket region (defaults to "us-east-1").
	region string
	// knownBuckets caches verified or created buckets in-memory to eliminate redundant BucketExists round-trips.
	knownBuckets sync.Map
}

// StorageOption configures optional MinioStorageProvider settings.
type StorageOption func(*minioStorageConfig)

type minioStorageConfig struct {
	useSSL *bool
	region string
}

// WithSSL explicitly sets whether to use HTTPS/TLS for S3 requests.
func WithSSL(ssl bool) StorageOption {
	return func(c *minioStorageConfig) {
		c.useSSL = &ssl
	}
}

// WithRegion specifies the target S3 bucket region (e.g. "eu-west-1", "me-central-1").
func WithRegion(region string) StorageOption {
	return func(c *minioStorageConfig) {
		if strings.TrimSpace(region) != "" {
			c.region = strings.TrimSpace(region)
		}
	}
}

// NewMinioStorageProvider initializes and configures a new MinioStorageProvider instance.
func NewMinioStorageProvider(endpoint, accessKey, secretKey, publicURL string, opts ...StorageOption) (*MinioStorageProvider, error) {
	cfg := minioStorageConfig{
		region: "us-east-1",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

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
		lower := strings.ToLower(trimmedEndpoint)
		if strings.Contains(lower, "amazonaws.com") ||
			strings.Contains(lower, "r2.cloudflarestorage.com") ||
			strings.Contains(lower, "storage.googleapis.com") {
			useSSL = true
		}
	}

	if cfg.useSSL != nil {
		useSSL = *cfg.useSSL
	}

	cleanPublicURL := strings.TrimSpace(publicURL)
	cleanPublicURL = strings.TrimRight(cleanPublicURL, "/")
	if cleanPublicURL != "" && !strings.Contains(cleanPublicURL, "://") && !strings.HasPrefix(cleanPublicURL, "/") {
		cleanPublicURL = "http://" + cleanPublicURL
	}

	// Align internal container hosts (e.g. "minio:9000", "floci:4566", "localstack:4566") to public browser endpoints ("localhost:...")
	// so assets served to browsers load without ERR_CONNECTION_REFUSED.
	if cleanPublicURL != "" && strings.Contains(cleanPublicURL, "://") {
		if u, err := url.Parse(cleanPublicURL); err == nil && (u.Hostname() == "minio" || u.Hostname() == "floci" || u.Hostname() == "localstack") {
			port := u.Port()
			if port == "" {
				if u.Hostname() == "minio" {
					port = "9000"
				} else {
					port = "4566"
				}
			}
			u.Host = "localhost:" + port
			cleanPublicURL = u.String()
		}
	}

	minioClient, err := minio.New(trimmedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: cfg.region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	return &MinioStorageProvider{
		Client:    minioClient,
		publicURL: cleanPublicURL,
		region:    cfg.region,
	}, nil
}

// NewS3StorageProvider initializes an S3-compatible (AWS S3 / LocalStack / R2) storage provider.
func NewS3StorageProvider(endpoint, accessKey, secretKey, publicURL string, opts ...StorageOption) (*S3StorageProvider, error) {
	return NewMinioStorageProvider(endpoint, accessKey, secretKey, publicURL, opts...)
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
			Region: m.region,
		})
		if err != nil {
			errResp := minio.ToErrorResponse(err)
			// Ignore if bucket already exists or if IAM policy forbids CreateBucket on pre-provisioned buckets
			if errResp.Code != "BucketAlreadyOwnedByYou" &&
				errResp.Code != "BucketAlreadyExists" &&
				errResp.Code != "AccessDenied" &&
				errResp.Code != "MethodNotAllowed" {
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
		size = -1
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
// Prevents duplicate bucket segments if publicURL already includes the bucket in its hostname or path.
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

	// Avoid duplicate bucket in path if publicURL already embeds the bucket in its host or path.
	if u, err := url.Parse(baseURL); err == nil {
		bucketLower := strings.ToLower(bucket)
		hostLower := strings.ToLower(u.Hostname())
		// e.g. "my-bucket.s3.eu-west-1.amazonaws.com" or "my-bucket.r2.cloudflarestorage.com"
		if strings.HasPrefix(hostLower, bucketLower+".") {
			return fmt.Sprintf("%s/%s", baseURL, escapedKey)
		}
		// e.g. "https://cdn.example.com/my-bucket"
		pathTrimmed := strings.Trim(strings.ToLower(u.Path), "/")
		if pathTrimmed == bucketLower || strings.HasSuffix(pathTrimmed, "/"+bucketLower) {
			return fmt.Sprintf("%s/%s", baseURL, escapedKey)
		}
	}

	return fmt.Sprintf("%s/%s/%s", baseURL, bucket, escapedKey)
}

// NoopProvider implements StorageProvider with no-op operations, ideal for tests.
type NoopProvider struct{}

func (NoopProvider) Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error {
	return nil
}

func (NoopProvider) Delete(ctx context.Context, bucket, key string) error {
	return nil
}

func (NoopProvider) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	return fmt.Sprintf("http://localhost/%s/%s", bucket, key), nil
}

func (NoopProvider) GetPublicURL(bucket, key string) string {
	return fmt.Sprintf("http://localhost/%s/%s", bucket, key)
}

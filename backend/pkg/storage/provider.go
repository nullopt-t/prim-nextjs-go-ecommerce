package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type StorageProvider interface {
	Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error
	Delete(ctx context.Context, bucket, key string) error
	GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
	GetPublicURL(bucket, key string) string
}

type MinioStorageProvider struct {
	client    *minio.Client
	publicURL string
}

func NewMinioStorageProvider(endpoint, accessKey, secretKey, publicURL string) (*MinioStorageProvider, error) {
	// Initialize minio client object.
	useSSL := false // Set true if using HTTPS
	
	// Quick hack for minio endpoint (strip http://)
	if len(endpoint) > 7 && endpoint[:7] == "http://" {
		endpoint = endpoint[7:]
	} else if len(endpoint) > 8 && endpoint[:8] == "https://" {
		endpoint = endpoint[8:]
		useSSL = true
	}

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init minio client: %w", err)
	}

	return &MinioStorageProvider{
		client:    minioClient,
		publicURL: publicURL,
	}, nil
}

func (m *MinioStorageProvider) ensureBucket(ctx context.Context, bucket string) error {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		err = m.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MinioStorageProvider) Upload(ctx context.Context, bucket, key string, file io.Reader, size int64, contentType string) error {
	if err := m.ensureBucket(ctx, bucket); err != nil {
		return fmt.Errorf("ensure bucket: %w", err)
	}

	_, err := m.client.PutObject(ctx, bucket, key, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}

	return nil
}

func (m *MinioStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	err := m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}

func (m *MinioStorageProvider) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucket, key, expires, nil)
	if err != nil {
		return "", fmt.Errorf("presigned get object: %w", err)
	}
	return url.String(), nil
}

func (m *MinioStorageProvider) GetPublicURL(bucket, key string) string {
	return fmt.Sprintf("%s/%s/%s", m.publicURL, bucket, key)
}

package service

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const streamUploadPartSize = 5 * 1024 * 1024

type ObjectStorage interface {
	Upload(ctx context.Context, objectKey, contentType string, size int64, reader io.Reader) error
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectKey string) error
	Exists(ctx context.Context, objectKey string) (bool, error)
}

type MinIOStorage struct {
	client *minio.Client
	bucket string
}

type MinIOOptions struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func NewMinIOStorage(ctx context.Context, opts MinIOOptions) (*MinIOStorage, error) {
	client, err := minio.New(opts.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""),
		Secure: opts.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	storage := &MinIOStorage{client: client, bucket: opts.Bucket}
	if err := storage.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *MinIOStorage) Upload(ctx context.Context, objectKey, contentType string, size int64, reader io.Reader) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}
	if size < 0 {
		opts.PartSize = streamUploadPartSize
	}

	_, err := s.client.PutObject(ctx, s.bucket, objectKey, reader, size, opts)
	if err != nil {
		return fmt.Errorf("upload object: %w", err)
	}

	return nil
}

func (s *MinIOStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("download object: %w", err)
	}

	return object, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, objectKey string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object: %w", err)
	}

	return nil
}

func (s *MinIOStorage) Exists(ctx context.Context, objectKey string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}

	minioErr := minio.ToErrorResponse(err)
	if minioErr.Code == "NoSuchKey" || minioErr.Code == "NotFound" {
		return false, nil
	}

	return false, fmt.Errorf("stat object: %w", err)
}

func (s *MinIOStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if exists {
		return nil
	}

	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	return nil
}

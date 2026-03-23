package media

import (
	"context"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/Tattsum/blog/backend/internal/application/upload"
	"google.golang.org/api/option"
)

type GCSStorage struct {
	client        *storage.Client
	bucket        string
	publicBaseURL string
}

func NewGCSStorage(ctx context.Context, bucket string, publicBaseURL string, opts ...option.ClientOption) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	bucket = strings.TrimSpace(bucket)
	publicBaseURL = strings.TrimSuffix(strings.TrimSpace(publicBaseURL), "/")
	return &GCSStorage{client: client, bucket: bucket, publicBaseURL: publicBaseURL}, nil
}

func (s *GCSStorage) Put(ctx context.Context, key, contentType string, body io.Reader) (publicURL string, err error) {
	key = strings.TrimPrefix(key, "/")
	obj := s.client.Bucket(s.bucket).Object(key)
	w := obj.NewWriter(ctx)
	w.ContentType = contentType
	if _, err := io.Copy(w, body); err != nil {
		_ = w.Close()
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + key, nil
	}
	return "https://storage.googleapis.com/" + s.bucket + "/" + key, nil
}

func (s *GCSStorage) Close() error { return s.client.Close() }

var _ upload.MediaStorage = (*GCSStorage)(nil)

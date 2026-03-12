package media

import (
	"context"
	"io"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/Tattsum/blog/backend/internal/application/upload"
	"google.golang.org/api/option"
)

// GCSStorage は Google Cloud Storage に保存し、公開 URL を返す。
// バケットはコンソールで公開読取に設定するか、Uniform bucket-level access で AllUsers に読取権限を付与すること。
type GCSStorage struct {
	client *storage.Client
	bucket string
}

// NewGCSStorage は bucket 名で GCS クライアントを生成する。ctx はクライアント初期化にのみ使用する。
func NewGCSStorage(ctx context.Context, bucket string, opts ...option.ClientOption) (*GCSStorage, error) {
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &GCSStorage{client: client, bucket: strings.TrimSpace(bucket)}, nil
}

// Put は body を key でバケットに書き込む。公開 URL は https://storage.googleapis.com/bucket/key 形式で返す。
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
	return "https://storage.googleapis.com/" + s.bucket + "/" + key, nil
}

// Close はクライアントを閉じる。
func (s *GCSStorage) Close() error { return s.client.Close() }

var _ upload.MediaStorage = (*GCSStorage)(nil)

package media

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tattsum/blog/backend/internal/application/upload"
)

type LocalStorage struct {
	dir     string
	baseURL string
}

func NewLocalStorage(dir, baseURL string) (*LocalStorage, error) {
	dir = strings.TrimSuffix(filepath.Clean(dir), string(os.PathSeparator))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &LocalStorage{dir: dir, baseURL: baseURL}, nil
}

func (s *LocalStorage) Put(ctx context.Context, key, contentType string, body io.Reader) (publicURL string, err error) {
	if strings.ContainsAny(key, "/\\") {
		key = filepath.Base(key)
	}
	fpath := filepath.Join(s.dir, key)
	f, err := os.Create(fpath)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	if _, err := io.Copy(f, body); err != nil {
		_ = os.Remove(fpath)
		return "", err
	}
	if s.baseURL != "" {
		return s.baseURL + "/uploads/" + key, nil
	}
	return "/uploads/" + key, nil
}

func (s *LocalStorage) Dir() string { return s.dir }

var _ upload.MediaStorage = (*LocalStorage)(nil)

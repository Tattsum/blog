package media

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Tattsum/blog/backend/internal/application/upload"
)

// LocalStorage はローカルディスクに保存し、baseURL + /uploads/key を公開 URL として返す。
type LocalStorage struct {
	dir     string // 保存先ディレクトリ（絶対パス推奨）
	baseURL string // 例: https://api.example.com（末尾スラッシュなし）
}

// NewLocalStorage は dir に保存する LocalStorage を返す。baseURL が空の場合は /uploads/ からの相対パスのみ返す（ハンドラ側で結合する想定）。
func NewLocalStorage(dir, baseURL string) (*LocalStorage, error) {
	dir = strings.TrimSuffix(filepath.Clean(dir), string(os.PathSeparator))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &LocalStorage{dir: dir, baseURL: baseURL}, nil
}

// Put は body を key で dir に保存する。key に含まれるディレクトリ区切りは無視しフラットに保存する。
func (s *LocalStorage) Put(ctx context.Context, key, contentType string, body io.Reader) (publicURL string, err error) {
	// パストラバーサル対策: key はファイル名のみ（スラッシュ・バックスラッシュ禁止）
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

// Dir は保存先ディレクトリを返す（FileServer 用）。
func (s *LocalStorage) Dir() string { return s.dir }

var _ upload.MediaStorage = (*LocalStorage)(nil)

package upload

import (
	"context"
	"io"
)

// MediaStorage はメディアファイルの保存先を抽象化するインターフェース。
// 実装例: ローカルディスク、GCS、R2（S3 互換）。
type MediaStorage interface {
	// Put は body を key で保存し、公開 URL を返す。
	// contentType は MIME タイプ（例: image/jpeg）。
	Put(ctx context.Context, key, contentType string, body io.Reader) (publicURL string, err error)
}

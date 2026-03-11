package rpc

import (
	"errors"
	"log/slog"

	"connectrpc.com/connect"
)

// errInternal はクライアントに返す汎用メッセージ（実装詳細は含めない）。
var errInternal = errors.New("internal error")

// MapHandlerError はハンドラ内で repo / infra から返った error を connect 用に正規化する。
// - 既に *connect.Error ならそのまま返す（二重ラップしない）
// - それ以外は CodeInternal にラップし、元エラーの文言をクライアントに渡さない（サーバー側では slog で記録）
func MapHandlerError(err error) error {
	if err == nil {
		return nil
	}
	var ce *connect.Error
	if errors.As(err, &ce) {
		return err
	}
	slog.Error("rpc handler error", "err", err)
	return connect.NewError(connect.CodeInternal, errInternal)
}

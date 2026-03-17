package rpc

import (
	"errors"
	"log/slog"
	"strings"

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

	// Vertex AI / Publisher model 系の失敗は「内部エラー」で隠すと復旧不能になりやすいので、
	// 管理画面からの操作で分かるように前提条件エラーとして返す（詳細はログ側に残す）。
	// 例: "Publisher Model ... was not found or your project does not have access to it."
	msg := err.Error()
	if strings.Contains(msg, "Publisher Model") && strings.Contains(msg, "does not have access") {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("AI が利用できません（Vertex AI のモデル利用権限/有効化を確認してください）"))
	}

	return connect.NewError(connect.CodeInternal, errInternal)
}

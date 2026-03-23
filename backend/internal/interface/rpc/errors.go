package rpc

import (
	"errors"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
)

var errInternal = errors.New("internal error")

func MapHandlerError(err error) error {
	if err == nil {
		return nil
	}
	var ce *connect.Error
	if errors.As(err, &ce) {
		return err
	}
	slog.Error("rpc handler error", "err", err)
	msg := err.Error()
	if strings.Contains(msg, "Publisher Model") && strings.Contains(msg, "does not have access") {
		return connect.NewError(connect.CodeFailedPrecondition, errors.New("AI が利用できません（Vertex AI のモデル利用権限/有効化を確認してください）"))
	}

	return connect.NewError(connect.CodeInternal, errInternal)
}

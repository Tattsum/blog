package rpc

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"connectrpc.com/connect"
)

const adminKeyHeader = "X-Admin-Key"

// requireAdmin は管理者キーが設定されている場合、リクエストヘッダの X-Admin-Key が一致することを要求する。
// adminKey が空の場合は常に PermissionDenied を返す（本番ではキー設定を必須とする）。
func requireAdmin(adminKey string, header http.Header) error {
	if adminKey == "" {
		return connect.NewError(connect.CodePermissionDenied, errors.New("admin API key not configured"))
	}
	if header.Get(adminKeyHeader) != adminKey {
		return connect.NewError(connect.CodePermissionDenied, errors.New("invalid or missing X-Admin-Key"))
	}
	return nil
}

// Slugify はタイトルや名前を URL 用スラグに変換する（小文字・空白をハイフン・英数字とハイフンのみ）。
func Slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	var prevHyphen bool
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			b.WriteRune(r)
			prevHyphen = false
		} else if (r == ' ' || r == '-') && !prevHyphen {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

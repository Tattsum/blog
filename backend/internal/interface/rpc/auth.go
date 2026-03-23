package rpc

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"connectrpc.com/connect"
	"github.com/gosimple/unidecode"
)

const (
	adminKeyHeader      = "X-Admin-Key"
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// bearerToken は Authorization: Bearer <token> からトークン部分を返す。
func bearerToken(header http.Header) string {
	v := header.Get(authorizationHeader)
	if !strings.HasPrefix(v, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(v[len(bearerPrefix):])
}

// RequireAdminOrSession は X-Admin-Key の一致、または有効な Bearer セッションのいずれかで管理者を許可する。
// HTTP ハンドラなど Connect 外から使う場合はこの名前で呼ぶ。
func RequireAdminOrSession(adminKey string, header http.Header, sessionStore SessionStore) error {
	return requireAdminOrSession(adminKey, header, sessionStore)
}

// requireAdminOrSession は X-Admin-Key の一致、または有効な Bearer セッションのいずれかで管理者を許可する。
// sessionStore が nil の場合は Bearer チェックを行わず requireAdmin と同様にキーのみ検証する。
func requireAdminOrSession(adminKey string, header http.Header, sessionStore SessionStore) error {
	if adminKey != "" && header.Get(adminKeyHeader) == adminKey {
		return nil
	}
	if sessionStore != nil {
		token := bearerToken(header)
		if token != "" {
			if _, ok := sessionStore.Get(token); ok {
				return nil
			}
		}
	}
	if adminKey == "" {
		return connect.NewError(connect.CodePermissionDenied, errors.New("admin API key not configured"))
	}
	return connect.NewError(connect.CodePermissionDenied, errors.New("invalid or missing X-Admin-Key or session"))
}

func Slugify(s string) string {
	s = unidecode.Unidecode(s)
	s = strings.TrimSpace(strings.ToLower(s))
	var b strings.Builder
	var prevHyphen bool
	for _, r := range s {
		if (unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_') && r <= unicode.MaxASCII {
			b.WriteRune(r)
			prevHyphen = false
		} else if (r == ' ' || r == '-') && !prevHyphen {
			b.WriteRune('-')
			prevHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

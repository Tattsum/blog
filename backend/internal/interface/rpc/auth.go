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

func bearerToken(header http.Header) string {
	v := header.Get(authorizationHeader)
	if !strings.HasPrefix(v, bearerPrefix) {
		return ""
	}
	return strings.TrimSpace(v[len(bearerPrefix):])
}

func RequireAdminOrSession(adminKey string, header http.Header, sessionStore SessionStore) error {
	return requireAdminOrSession(adminKey, header, sessionStore)
}

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

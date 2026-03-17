package upload

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"connectrpc.com/connect"
	uploadapp "github.com/Tattsum/blog/backend/internal/application/upload"
	"github.com/Tattsum/blog/backend/internal/interface/rpc"
	"github.com/google/uuid"
)

// Handler は POST /upload で multipart ファイルを受け取り、ストレージに保存して公開 URL を JSON で返す。
type Handler struct {
	storage      uploadapp.MediaStorage
	adminKey     string
	sessionStore rpc.SessionStore
}

// NewHandler は Handler を返す。
func NewHandler(storage uploadapp.MediaStorage, adminKey string, sessionStore rpc.SessionStore) *Handler {
	return &Handler{storage: storage, adminKey: adminKey, sessionStore: sessionStore}
}

// ServeHTTP は POST のみ受け付ける。multipart form の "file" を保存し、{"url": "..."} を返す。
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if err := rpc.RequireAdminOrSession(h.adminKey, r.Header, h.sessionStore); err != nil {
		code := http.StatusForbidden
		if connect.CodeOf(err) == connect.CodePermissionDenied {
			code = http.StatusForbidden
		}
		writeJSONError(w, code, "unauthorized")
		return
	}

	// 10MB + 100MB より少し大きめで multipart を制限（本体は後で ValidateMedia でチェック）
	const maxMultipartMem = 110 << 20
	if err := r.ParseMultipartForm(maxMultipartMem); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "missing or invalid file field")
		return
	}
	defer func() { _ = file.Close() }()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	size := max(header.Size, 0)
	if err := uploadapp.ValidateMedia(contentType, size); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	ext := extFromFilename(header.Filename)
	if ext == "" {
		ext = extFromContentType(contentType)
	}
	if ext == "" {
		writeJSONError(w, http.StatusBadRequest, "unsupported file type")
		return
	}
	key := uuid.Must(uuid.NewRandom()).String() + ext

	ctx := r.Context()
	publicURL, err := h.storage.Put(ctx, key, contentType, file)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "upload failed")
		return
	}
	// 相対パスで返ってきた場合はリクエストの scheme+host を前置
	if strings.HasPrefix(publicURL, "/") && r.URL != nil {
		scheme := "https"
		if r.TLS == nil {
			scheme = "http"
		}
		if v := r.Header.Get("X-Forwarded-Proto"); v != "" {
			scheme = v
		}
		publicURL = scheme + "://" + r.Host + publicURL
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(struct {
		URL string `json:"url"`
	}{URL: publicURL})
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: msg})
}

func extFromFilename(filename string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	allowed := map[string]string{
		"jpg": ".jpg", "jpeg": ".jpeg", "png": ".png", "gif": ".gif", "webp": ".webp",
		"mp4": ".mp4", "webm": ".webm",
	}
	if e, ok := allowed[ext]; ok {
		return e
	}
	return ""
}

func extFromContentType(contentType string) string {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if i := strings.Index(contentType, ";"); i >= 0 {
		contentType = contentType[:i]
	}
	m := map[string]string{
		"image/jpeg": ".jpg", "image/png": ".png", "image/gif": ".gif", "image/webp": ".webp",
		"video/mp4": ".mp4", "video/webm": ".webm",
	}
	if e, ok := m[contentType]; ok {
		return e
	}
	return ""
}

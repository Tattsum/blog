package upload

import (
	"errors"
	"strings"
)

// ファイルサイズ上限（バイト）
const (
	MaxImageSize = 10 << 20  // 10 MiB
	MaxVideoSize = 100 << 20 // 100 MiB
)

// 許可する MIME タイプ（小文字）。拡張子検証は呼び出し側で行う。
var (
	allowedImageMIMEs = map[string]int64{
		"image/jpeg": MaxImageSize,
		"image/png":  MaxImageSize,
		"image/gif":  MaxImageSize,
		"image/webp": MaxImageSize,
	}
	allowedVideoMIMEs = map[string]int64{
		"video/mp4":  MaxVideoSize,
		"video/webm": MaxVideoSize,
	}
)

// ValidateMedia は contentType と size を検証する。許可されない場合は error を返す。
func ValidateMedia(contentType string, size int64) error {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if contentType == "" {
		return errors.New("content-type is required")
	}
	// パラメータ付き (e.g. image/jpeg; charset=utf-8) は先頭部分のみ使用
	if i := strings.Index(contentType, ";"); i >= 0 {
		contentType = strings.TrimSpace(contentType[:i])
	}
	max, ok := allowedImageMIMEs[contentType]
	if ok {
		if size <= 0 || size > max {
			return errors.New("image size exceeds limit (max 10MB)")
		}
		return nil
	}
	max, ok = allowedVideoMIMEs[contentType]
	if ok {
		if size <= 0 || size > max {
			return errors.New("video size exceeds limit (max 100MB)")
		}
		return nil
	}
	return errors.New("unsupported media type: allowed image (jpeg,png,gif,webp) and video (mp4,webm)")
}

// AllowedExtensions は許可する拡張子のリスト（小文字、ドットなし）。UI の accept 用。
func AllowedExtensions() []string {
	return []string{"jpg", "jpeg", "png", "gif", "webp", "mp4", "webm"}
}

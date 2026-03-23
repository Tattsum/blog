package upload

import (
	"errors"
	"strings"
)

const (
	MaxImageSize = 10 << 20  // 10 MiB
	MaxVideoSize = 100 << 20 // 100 MiB
)

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

func ValidateMedia(contentType string, size int64) error {
	contentType = strings.TrimSpace(strings.ToLower(contentType))
	if contentType == "" {
		return errors.New("content-type is required")
	}
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

func AllowedExtensions() []string {
	return []string{"jpg", "jpeg", "png", "gif", "webp", "mp4", "webm"}
}

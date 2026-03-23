package upload

import (
	"context"
	"io"
)

type MediaStorage interface {
	Put(ctx context.Context, key, contentType string, body io.Reader) (publicURL string, err error)
}

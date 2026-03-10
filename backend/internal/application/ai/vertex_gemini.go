package ai

import (
	"context"
	"time"

	"github.com/Tattsum/blog/backend/internal/infrastructure/vertexai"
)

const defaultGenerateTimeout = 90 * time.Second

// VertexGemini は Vertex 上の Gemini を TextGenerator としてラップする。
type VertexGemini struct {
	client *vertexai.Client
}

// NewVertexGemini は client が nil のとき nil を返す（呼び出し側でフォールバック）。
func NewVertexGemini(client *vertexai.Client) *VertexGemini {
	if client == nil {
		return nil
	}
	return &VertexGemini{client: client}
}

// GenerateText は vertexai.Client.GenerateText に委譲し、タイムアウトを付与する。
func (a *VertexGemini) GenerateText(ctx context.Context, prompt string) (string, error) {
	if a == nil || a.client == nil {
		return "", nil
	}
	ctx, cancel := context.WithTimeout(ctx, defaultGenerateTimeout)
	defer cancel()
	return a.client.GenerateText(ctx, prompt)
}

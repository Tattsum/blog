package ai

import (
	"context"
	"time"

	"github.com/Tattsum/blog/backend/internal/infrastructure/vertexai"
)

const defaultGenerateTimeout = 90 * time.Second

type VertexGemini struct {
	client *vertexai.Client
}

func NewVertexGemini(client *vertexai.Client) *VertexGemini {
	if client == nil {
		return nil
	}
	return &VertexGemini{client: client}
}

func (a *VertexGemini) GenerateText(ctx context.Context, prompt string) (string, error) {
	if a == nil || a.client == nil {
		return "", nil
	}
	ctx, cancel := context.WithTimeout(ctx, defaultGenerateTimeout)
	defer cancel()
	return a.client.GenerateText(ctx, prompt)
}

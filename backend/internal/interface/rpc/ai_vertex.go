package rpc

import "context"

// vertexGenerator は Vertex AI 等のテキスト生成を抽象化する（テスト差し替え用）。
type vertexGenerator interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

package ai

import "context"

type TextGenerator interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

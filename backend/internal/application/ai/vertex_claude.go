package ai

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/vertex"
	"golang.org/x/oauth2/google"
)

const (
	defaultClaudeModel     = anthropic.ModelClaudeSonnet4_5_20250929
	defaultClaudeMaxTokens = int64(4096)
	maxClaudePromptRunes   = 60_000
)

// VertexClaude は Vertex AI 上の Claude（Partner）を TextGenerator として呼び出す。
// IAM: Cloud Run 実行 SA に roles/aiplatform.user。リージョンで利用可能なモデルは GCP コンソール要確認。
type VertexClaude struct {
	client anthropic.Client
	model  anthropic.Model
}

// NewVertexClaude は ADC で Vertex 上の Claude クライアントを構築する。
// project / region が空ならエラー。model が空なら defaultClaudeModel。
func NewVertexClaude(ctx context.Context, project, region, model string) (*VertexClaude, error) {
	project = strings.TrimSpace(project)
	region = strings.TrimSpace(region)
	if project == "" || region == "" {
		return nil, errors.New("ai: vertex claude requires project and region")
	}
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, err
	}
	client := anthropic.NewClient(vertex.WithCredentials(ctx, region, project, creds))
	m := anthropic.Model(strings.TrimSpace(model))
	if m == "" {
		m = defaultClaudeModel
	}
	return &VertexClaude{client: client, model: m}, nil
}

// NewVertexClaudeFromEnv は GOOGLE_CLOUD_PROJECT / GOOGLE_CLOUD_LOCATION（または GCP_REGION）と
// 任意の VERTEX_CLAUDE_MODEL で VertexClaude を返す。project が無い場合は nil, nil。
func NewVertexClaudeFromEnv(ctx context.Context) (*VertexClaude, error) {
	project := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if project == "" {
		return nil, nil
	}
	region := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_LOCATION"))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("GCP_REGION"))
	}
	if region == "" {
		region = "us-central1"
	}
	model := strings.TrimSpace(os.Getenv("VERTEX_CLAUDE_MODEL"))
	return NewVertexClaude(ctx, project, region, model)
}

// GenerateText は Messages API で 1 ターンの user メッセージを送り、assistant のテキストを連結して返す。
func (a *VertexClaude) GenerateText(ctx context.Context, prompt string) (string, error) {
	if a == nil {
		return "", nil
	}
	prompt = truncateRunes(prompt, maxClaudePromptRunes)
	if prompt == "" {
		return "", nil
	}
	ctx, cancel := context.WithTimeout(ctx, defaultGenerateTimeout)
	defer cancel()

	msg, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     a.model,
		MaxTokens: defaultClaudeMaxTokens,
		Messages: []anthropic.MessageParam{{
			Role: anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{{
				OfText: &anthropic.TextBlockParam{Text: prompt},
			}},
		}},
	})
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, block := range msg.Content {
		tb := block.AsText()
		if tb.Text != "" {
			b.WriteString(tb.Text)
		}
	}
	return strings.TrimSpace(b.String()), nil
}

func truncateRunes(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "\n\n…(truncated)"
}

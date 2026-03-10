// Package vertexai wraps Vertex AI (Gemini) via google.golang.org/genai.
// Cloud Run では実行 SA に roles/aiplatform.user が必要。
package vertexai

import (
	"context"
	"errors"
	"os"
	"strings"

	"google.golang.org/genai"
)

const (
	// DefaultModel は asia-northeast1 等で利用しやすい Flash 系（必要に応じて env で上書き）。
	DefaultModel = "gemini-2.0-flash-001"
	// maxPromptRunes は入力が極端に長い場合のトークン抑制（ざっくり制限）。
	maxPromptRunes = 60_000
)

// Client は Vertex 上の Gemini を呼び出す薄いラッパー。
type Client struct {
	genai *genai.Client
	model string
}

// Config は New 用設定。
type Config struct {
	Project  string // GCP プロジェクト ID（必須）
	Location string // リージョン（例: asia-northeast1）
	Model    string // 省略時 DefaultModel
}

// New は Vertex AI クライアントを返す。project が空ならエラー。
func New(ctx context.Context, cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.Project) == "" {
		return nil, errors.New("vertexai: project is required")
	}
	loc := strings.TrimSpace(cfg.Location)
	if loc == "" {
		loc = "us-central1"
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = DefaultModel
	}
	c, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  cfg.Project,
		Location: loc,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		return nil, err
	}
	return &Client{genai: c, model: model}, nil
}

// NewFromEnv は GOOGLE_CLOUD_PROJECT と GOOGLE_CLOUD_LOCATION（任意）から Client を構築する。
// VERTEX_GEMINI_MODEL でモデル上書き可。project が無ければ (nil, nil) を返す（呼び出し側でフォールバック）。
func NewFromEnv(ctx context.Context) (*Client, error) {
	project := strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if project == "" {
		return nil, nil
	}
	cfg := Config{
		Project:  project,
		Location: strings.TrimSpace(os.Getenv("GOOGLE_CLOUD_LOCATION")),
		Model:    strings.TrimSpace(os.Getenv("VERTEX_GEMINI_MODEL")),
	}
	if cfg.Location == "" {
		cfg.Location = os.Getenv("GCP_REGION") // Makefile 等で使っている場合
	}
	return New(ctx, cfg)
}

// GenerateText は単一プロンプトでテキスト生成を行い、最初の候補テキストを返す。
func (c *Client) GenerateText(ctx context.Context, prompt string) (string, error) {
	if c == nil || c.genai == nil {
		return "", errors.New("vertexai: client is nil")
	}
	prompt = truncate(prompt, maxPromptRunes)
	if prompt == "" {
		return "", nil
	}
	contents := []*genai.Content{
		genai.NewContentFromText(prompt, genai.RoleUser),
	}
	resp, err := c.genai.Models.GenerateContent(ctx, c.model, contents, nil)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("vertexai: empty response")
	}
	out := resp.Text()
	return strings.TrimSpace(out), nil
}

func truncate(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "\n\n…(truncated)"
}

package vertexai

import (
	"context"
	"errors"
	"os"
	"strings"

	"google.golang.org/genai"
)

const (
	DefaultModel   = "gemini-2.0-flash-001"
	maxPromptRunes = 60_000
)

type Client struct {
	genai *genai.Client
	model string
}

type Config struct {
	Project  string
	Location string
	Model    string
}

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

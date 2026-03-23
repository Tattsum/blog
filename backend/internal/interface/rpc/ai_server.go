package rpc

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/Tattsum/blog/backend/internal/application/ai"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

type AIServer struct {
	blogv1connect.UnimplementedAIServiceHandler
	adminKey        string
	sessionStore    SessionStore
	defaultProvider string
	gemini          ai.TextGenerator
	claude          ai.TextGenerator
}

func NewAIServer(adminKey string, sessionStore SessionStore, provider string, gemini, claude ai.TextGenerator) *AIServer {
	return &AIServer{
		adminKey:        adminKey,
		sessionStore:    sessionStore,
		defaultProvider: strings.ToLower(strings.TrimSpace(provider)),
		gemini:          gemini,
		claude:          claude,
	}
}

func (s *AIServer) pickGenerator(h map[string][]string) (provider string, gen ai.TextGenerator, specified bool) {
	get := func(key string) string {
		v := h[key]
		if len(v) == 0 {
			v = h[http.CanonicalHeaderKey(key)]
		}
		if len(v) == 0 {
			return ""
		}
		return v[0]
	}
	p := strings.ToLower(strings.TrimSpace(get("X-AI-Provider")))
	specified = p != ""
	if p == "" {
		p = s.defaultProvider
	}
	switch p {
	case "vertex-claude", "claude":
		return "claude", s.claude, specified
	case "", "vertex-gemini", "gemini":
		return "gemini", s.gemini, specified
	default:
		return p, nil, specified
	}
}

func (s *AIServer) Summarize(ctx context.Context, req *connect.Request[blogv1.SummarizeRequest]) (*connect.Response[blogv1.SummarizeResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	text := strings.TrimSpace(req.Msg.GetText())
	if text == "" {
		return connect.NewResponse(&blogv1.SummarizeResponse{Summary: ""}), nil
	}
	maxSentences := req.Msg.GetMaxSentences()
	if maxSentences <= 0 || maxSentences > 10 {
		maxSentences = 3
	}
	provider, gen, specified := s.pickGenerator(req.Header())
	if specified && gen == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("AI プロバイダ %q が利用できません", provider))
	}
	if gen != nil {
		prompt := fmt.Sprintf(
			"次の文章を日本語で、おおよそ%d文以内の要約にまとめてください。要約の本文だけを出力し、前置きや見出しは付けないでください。\n\n---\n%s",
			maxSentences, text,
		)
		summary, err := gen.GenerateText(ctx, prompt)
		if err != nil {
			return nil, MapHandlerError(err)
		}
		return connect.NewResponse(&blogv1.SummarizeResponse{Summary: summary}), nil
	}
	summary := summarizeText(text, int(maxSentences))
	return connect.NewResponse(&blogv1.SummarizeResponse{Summary: summary}), nil
}

func (s *AIServer) DraftSupport(ctx context.Context, req *connect.Request[blogv1.DraftSupportRequest]) (*connect.Response[blogv1.DraftSupportResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(req.Msg.GetPrompt())
	body := req.Msg.GetCurrentBody()

	provider, gen, specified := s.pickGenerator(req.Header())
	if specified && gen == nil {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("AI プロバイダ %q が利用できません", provider))
	}
	if gen != nil {
		userPrompt := fmt.Sprintf(
			"あなたはブログ記事の下書き支援を行います。ユーザーの指示に従い、マークダウン本文として使える案だけを出力してください。説明文や「以下のとおり」などのメタ文は不要です。\n\n【指示】\n%s\n\n【現在の本文】\n%s",
			prompt, body,
		)
		if prompt == "" && body == "" {
			userPrompt = "短いブログ記事の導入段落を1つ、マークダウンで書いてください。"
		}
		suggested, err := gen.GenerateText(ctx, userPrompt)
		if err != nil {
			return nil, MapHandlerError(err)
		}
		return connect.NewResponse(&blogv1.DraftSupportResponse{SuggestedBody: suggested}), nil
	}

	var builder strings.Builder
	if prompt != "" {
		builder.WriteString("【指示】")
		builder.WriteString(prompt)
		builder.WriteString("\n\n")
	}
	builder.WriteString("【提案本文】\n")
	if body == "" {
		builder.WriteString("ここに本文の下書き案を記述してください。")
	} else {
		builder.WriteString(body)
	}

	return connect.NewResponse(&blogv1.DraftSupportResponse{
		SuggestedBody: builder.String(),
	}), nil
}

func summarizeText(text string, n int) string {
	separators := []string{"。", "．", ".", "！", "!", "？", "?"}
	for _, sep := range separators {
		text = strings.ReplaceAll(text, sep, sep+"\n")
	}
	lines := strings.Split(text, "\n")
	var sentences []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		sentences = append(sentences, line)
		if len(sentences) >= n {
			break
		}
	}
	if len(sentences) == 0 {
		return ""
	}
	summary := strings.Join(sentences, "。")
	if !strings.HasSuffix(summary, "。") {
		summary += "。"
	}
	return summary
}

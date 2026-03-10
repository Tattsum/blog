package rpc

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

// AIServer は AIService の connect-go ハンドラ実装。
// vertex が nil のときはローカル要約／プレースホルダにフォールバックする。
type AIServer struct {
	blogv1connect.UnimplementedAIServiceHandler
	adminKey     string
	sessionStore SessionStore
	vertex       vertexGenerator
}

// NewAIServer は AIServer を返す。認証は X-Admin-Key または Bearer セッションのいずれかで行う。
// vertex に Vertex 等の実装を渡すと Summarize / DraftSupport で利用する（nil なら従来のローカル動作）。
func NewAIServer(adminKey string, sessionStore SessionStore, vertex vertexGenerator) *AIServer {
	return &AIServer{adminKey: adminKey, sessionStore: sessionStore, vertex: vertex}
}

// Summarize は本文の先頭から指定文数ぶんの文を抽出する簡易要約を行う。
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
	if s.vertex != nil {
		prompt := fmt.Sprintf(
			"次の文章を日本語で、おおよそ%d文以内の要約にまとめてください。要約の本文だけを出力し、前置きや見出しは付けないでください。\n\n---\n%s",
			maxSentences, text,
		)
		summary, err := s.vertex.GenerateText(ctx, prompt)
		if err != nil {
			return nil, MapHandlerError(err)
		}
		return connect.NewResponse(&blogv1.SummarizeResponse{Summary: summary}), nil
	}
	summary := summarizeText(text, int(maxSentences))
	return connect.NewResponse(&blogv1.SummarizeResponse{Summary: summary}), nil
}

// DraftSupport は現在の本文に対して、プロンプトを前置した提案本文を返す簡易実装。
func (s *AIServer) DraftSupport(ctx context.Context, req *connect.Request[blogv1.DraftSupportRequest]) (*connect.Response[blogv1.DraftSupportResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(req.Msg.GetPrompt())
	body := req.Msg.GetCurrentBody()

	if s.vertex != nil {
		userPrompt := fmt.Sprintf(
			"あなたはブログ記事の下書き支援を行います。ユーザーの指示に従い、マークダウン本文として使える案だけを出力してください。説明文や「以下のとおり」などのメタ文は不要です。\n\n【指示】\n%s\n\n【現在の本文】\n%s",
			prompt, body,
		)
		if prompt == "" && body == "" {
			userPrompt = "短いブログ記事の導入段落を1つ、マークダウンで書いてください。"
		}
		suggested, err := s.vertex.GenerateText(ctx, userPrompt)
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

// summarizeText は句点や改行で区切って最大 n 文を返す。
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

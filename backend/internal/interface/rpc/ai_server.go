package rpc

import (
	"context"
	"strings"

	"connectrpc.com/connect"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

// AIServer は AIService の connect-go ハンドラ実装（現状はローカルダミー実装）。
type AIServer struct {
	blogv1connect.UnimplementedAIServiceHandler
	adminKey string
}

// NewAIServer は AIServer を返す。adminKey が空の場合は呼び出しを拒否する。
func NewAIServer(adminKey string) *AIServer {
	return &AIServer{adminKey: adminKey}
}

// Summarize は本文の先頭から指定文数ぶんの文を抽出する簡易要約を行う。
func (s *AIServer) Summarize(ctx context.Context, req *connect.Request[blogv1.SummarizeRequest]) (*connect.Response[blogv1.SummarizeResponse], error) {
	if err := requireAdmin(s.adminKey, req.Header()); err != nil {
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
	summary := summarizeText(text, int(maxSentences))
	return connect.NewResponse(&blogv1.SummarizeResponse{Summary: summary}), nil
}

// DraftSupport は現在の本文に対して、プロンプトを前置した提案本文を返す簡易実装。
func (s *AIServer) DraftSupport(ctx context.Context, req *connect.Request[blogv1.DraftSupportRequest]) (*connect.Response[blogv1.DraftSupportResponse], error) {
	if err := requireAdmin(s.adminKey, req.Header()); err != nil {
		return nil, err
	}
	prompt := strings.TrimSpace(req.Msg.GetPrompt())
	body := req.Msg.GetCurrentBody()

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

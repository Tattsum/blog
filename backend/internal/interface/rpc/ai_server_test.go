package rpc_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"

	"github.com/Tattsum/blog/backend/internal/application/ai"
	rpcpkg "github.com/Tattsum/blog/backend/internal/interface/rpc"
)

func TestAIServer_Proofread(t *testing.T) {
	longText := strings.Repeat("あ", 100_001)
	okGen := &fakeTextGenerator{out: "  特に問題は見つかりませんでした  "}

	tests := []struct {
		name     string
		adminKey string
		header   func(h http.Header)
		text     string
		gen      ai.TextGenerator
		wantCode connect.Code
	}{
		{
			name:     "no_credentials",
			adminKey: "secret",
			header:   func(http.Header) {},
			text:     "本文",
			gen:      okGen,
			wantCode: connect.CodePermissionDenied,
		},
		{
			name:     "wrong_admin_key",
			adminKey: "secret",
			header:   func(h http.Header) { h.Set("X-Admin-Key", "wrong") },
			text:     "本文",
			gen:      okGen,
			wantCode: connect.CodePermissionDenied,
		},
		{
			name:     "empty_text",
			adminKey: "k",
			header:   func(h http.Header) { h.Set("X-Admin-Key", "k") },
			text:     "  ",
			gen:      okGen,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name:     "text_too_long",
			adminKey: "k",
			header:   func(h http.Header) { h.Set("X-Admin-Key", "k") },
			text:     longText,
			gen:      okGen,
			wantCode: connect.CodeInvalidArgument,
		},
		{
			name:     "ai_not_configured",
			adminKey: "k",
			header:   func(h http.Header) { h.Set("X-Admin-Key", "k") },
			text:     "本文",
			gen:      nil,
			wantCode: connect.CodeFailedPrecondition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := rpcpkg.NewAIServer(tt.adminKey, nil, "gemini", tt.gen, nil)
			req := connect.NewRequest(&blogv1.ProofreadRequest{Text: tt.text})
			tt.header(req.Header())
			_, err := srv.Proofread(context.Background(), req)
			if err == nil {
				t.Fatal("expected error")
			}
			if connect.CodeOf(err) != tt.wantCode {
				t.Fatalf("code = %v, want %v", connect.CodeOf(err), tt.wantCode)
			}
		})
	}
}

func TestAIServer_Proofread_success(t *testing.T) {
	want := "特に問題は見つかりませんでした"
	srv := rpcpkg.NewAIServer("k", nil, "gemini", &fakeTextGenerator{out: "  " + want + "  "}, nil)
	req := connect.NewRequest(&blogv1.ProofreadRequest{Text: "テスト本文"})
	req.Header().Set("X-Admin-Key", "k")
	res, err := srv.Proofread(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if res.Msg.GetReport() != want {
		t.Fatalf("report = %q, want %q", res.Msg.GetReport(), want)
	}
}

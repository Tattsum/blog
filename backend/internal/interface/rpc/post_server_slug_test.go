package rpc_test

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"

	rpcpkg "github.com/Tattsum/blog/backend/internal/interface/rpc"
)

type fakeTextGenerator struct {
	out string
	err error
}

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[-_][a-z0-9]+)*$`)

func (f fakeTextGenerator) GenerateText(ctx context.Context, prompt string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.out, nil
}

func TestPostServer_generateSlug_fallbackOnInvalidAIOutput(t *testing.T) {
	titles := []string{
		"日本語タイトル",
		"もう一つの例",
	}

	tests := []struct {
		name     string
		aiOutput string
		aiErr    error
		wantSlug string
	}{
		{
			name:     "empty_after_slugify",
			aiOutput: "----",
			wantSlug: rpcpkg.Slugify(titles[0]),
		},
		{
			name:     "too_long",
			aiOutput: strings.Repeat("a", 90),
			wantSlug: rpcpkg.Slugify(titles[0]),
		},
		{
			name:     "ai_error_fallback",
			aiErr:    errors.New("boom"),
			wantSlug: rpcpkg.Slugify(titles[0]),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := fakeTextGenerator{out: tt.aiOutput, err: tt.aiErr}
			s := rpcpkg.NewPostServer(nil, "", nil, "gemini", fake, nil)

			header := map[string][]string{}
			title := titles[0]
			got := s.GenerateSlugForTitle(context.Background(), header, title)
			if got != tt.wantSlug {
				t.Fatalf("generateSlug(%q) = %q, want %q", title, got, tt.wantSlug)
			}
		})
	}
}

func TestPostServer_generateSlug_returnsNormalizedCandidate(t *testing.T) {
	s := rpcpkg.NewPostServer(nil, "", nil, "gemini", fakeTextGenerator{
		// 生成物に大文字等が混じっていても Slugify で正規化される想定。
		out: "Hello 日本語 World",
	}, nil)

	got := s.GenerateSlugForTitle(context.Background(), map[string][]string{}, "日本語タイトル")
	if got == "" || len(got) > 80 || !slugPattern.MatchString(got) {
		t.Fatalf("generateSlug returned invalid slug: %q", got)
	}

	// Slugify によって少なくとも空ではなく、バリデーション通過することを確認する。
}

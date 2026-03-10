package vertexai

import (
	"context"
	"strings"
	"testing"
)

func TestNew_requiresProject(t *testing.T) {
	_, err := New(context.Background(), Config{})
	if err == nil {
		t.Fatal("expected error when project empty")
	}
}

func TestTruncate_addsMarkerWhenOverMax(t *testing.T) {
	long := strings.Repeat("a", maxPromptRunes+100)
	out := truncate(long, maxPromptRunes)
	if !strings.Contains(out, "truncated") {
		t.Fatal("expected truncated marker in output")
	}
	if len(out) >= len(long) {
		t.Fatal("expected shorter output")
	}
}

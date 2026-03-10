package ai

import (
	"context"
	"testing"
)

func TestNewVertexGemini_nilClient(t *testing.T) {
	if g := NewVertexGemini(nil); g != nil {
		t.Fatal("expected nil when client is nil")
	}
}

func TestVertexGemini_GenerateText_nilReceiverSafe(t *testing.T) {
	var g *VertexGemini
	_, err := g.GenerateText(context.Background(), "x")
	if err != nil {
		t.Fatal("nil receiver should not error")
	}
}

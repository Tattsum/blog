package ai

import (
	"context"
	"testing"
)

func TestNewVertexClaude_requiresProjectRegion(t *testing.T) {
	_, err := NewVertexClaude(context.Background(), "", "us-central1", "")
	if err == nil {
		t.Fatal("expected error without project")
	}
	_, err = NewVertexClaude(context.Background(), "p", "", "")
	if err == nil {
		t.Fatal("expected error without region")
	}
}

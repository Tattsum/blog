package post

import (
	"testing"
	"time"
)

func TestPost_IsPublished(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{"draft is not published", StatusDraft, false},
		{"published is published", StatusPublished, true},
		{"unspecified is not published", StatusUnspecified, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Post{Status: tt.status}
			if got := p.IsPublished(); got != tt.want {
				t.Errorf("IsPublished() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPost_Publish(t *testing.T) {
	at := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	p := &Post{
		ID:        "post-1",
		Status:    StatusDraft,
		UpdatedAt: at.Add(-time.Hour),
	}
	p.Publish(at)
	if p.Status != StatusPublished {
		t.Errorf("Status = %v, want Published", p.Status)
	}
	if p.PublishedAt == nil || !p.PublishedAt.Equal(at) {
		t.Errorf("PublishedAt = %v, want %v", p.PublishedAt, at)
	}
	if !p.UpdatedAt.Equal(at) {
		t.Errorf("UpdatedAt = %v, want %v", p.UpdatedAt, at)
	}
}

func TestPost_Unpublish(t *testing.T) {
	at := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	pub := at.Add(-time.Hour)
	p := &Post{
		ID:          "post-1",
		Status:      StatusPublished,
		PublishedAt: &pub,
		UpdatedAt:   pub,
	}
	p.Unpublish(at)
	if p.Status != StatusDraft {
		t.Errorf("Status = %v, want Draft", p.Status)
	}
	if p.PublishedAt != nil {
		t.Errorf("PublishedAt should be nil after unpublish, got %v", p.PublishedAt)
	}
	if !p.UpdatedAt.Equal(at) {
		t.Errorf("UpdatedAt = %v, want %v", p.UpdatedAt, at)
	}
}

func TestSlug_String(t *testing.T) {
	s := Slug("my-post")
	if s.String() != "my-post" {
		t.Errorf("Slug.String() = %q, want %q", s.String(), "my-post")
	}
}

package rpc

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Hello World", "hello-world"},
		{"  trim  spaces  ", "trim-spaces"},
		{"UPPERCASE", "uppercase"},
		{"Already-slug", "already-slug"},
		{"", ""},
		{"a", "a"},
		{"---", ""},
		{"Hello   World", "hello-world"},
		{"日本語", "日本語"},
		{"Go 1.21", "go-121"},
	}
	for _, tt := range tests {
		got := Slugify(tt.in)
		if got != tt.want {
			t.Errorf("Slugify(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

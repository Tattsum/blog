package rpc

import (
	"errors"
	"regexp"
	"unicode/utf8"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:[-_][a-z0-9]+)*$`)

// validatePostFields は記事の入力値を検証する。
func validatePostFields(title, slug, body, summary string, tagIDs []string) error {
	titleLen := utf8.RuneCountInString(title)
	if titleLen == 0 || titleLen > 120 {
		return errors.New("title must be between 1 and 120 characters")
	}
	if slug == "" {
		return errors.New("slug is required")
	}
	if len(slug) > 80 || !slugPattern.MatchString(slug) {
		return errors.New("slug must match pattern [a-z0-9_-]{1,80}")
	}
	if utf8.RuneCountInString(summary) > 300 {
		return errors.New("summary must be at most 300 characters")
	}
	if utf8.RuneCountInString(body) > 100000 {
		return errors.New("body_markdown is too long")
	}
	if len(tagIDs) > 50 {
		return errors.New("too many tag_ids")
	}
	for _, id := range tagIDs {
		if id == "" {
			return errors.New("tag_id must not be empty")
		}
		if utf8.RuneCountInString(id) > 64 {
			return errors.New("tag_id is too long")
		}
	}
	return nil
}

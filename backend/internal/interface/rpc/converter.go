package rpc

import (
	"time"

	"github.com/Tattsum/blog/backend/internal/domain/post"
	"github.com/Tattsum/blog/backend/internal/domain/tag"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
)

func timeToRFC3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}

func postStatusToProto(s post.Status) blogv1.Post_Status {
	switch s {
	case post.StatusDraft:
		return blogv1.Post_DRAFT
	case post.StatusPublished:
		return blogv1.Post_PUBLISHED
	default:
		return blogv1.Post_STATUS_UNSPECIFIED
	}
}

// PostToProto はドメインの Post を API 用の proto Post に変換する。
func PostToProto(p *post.Post) *blogv1.Post {
	if p == nil {
		return nil
	}
	out := &blogv1.Post{
		Id:           p.ID,
		Title:        p.Title,
		Slug:         p.Slug.String(),
		BodyMarkdown: p.BodyMarkdown,
		Summary:      p.Summary,
		TagIds:       p.TagIDs,
		Status:       postStatusToProto(p.Status),
		CreatedAt:    timeToRFC3339(p.CreatedAt),
		UpdatedAt:    timeToRFC3339(p.UpdatedAt),
	}
	if p.PublishedAt != nil {
		out.PublishedAt = timeToRFC3339(*p.PublishedAt)
	}
	return out
}

// TagToProto はドメインの Tag を API 用の proto Tag に変換する。
func TagToProto(t *tag.Tag) *blogv1.Tag {
	if t == nil {
		return nil
	}
	return &blogv1.Tag{
		Id:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug.String(),
		CreatedAt: timeToRFC3339(t.CreatedAt),
	}
}

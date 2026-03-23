package post

import "time"

type Status int32

const (
	StatusUnspecified Status = 0
	StatusDraft       Status = 1
	StatusPublished   Status = 2
)

type Slug string

func (s Slug) String() string { return string(s) }

type Post struct {
	ID           string
	Title        string
	Slug         Slug
	BodyMarkdown string
	Summary      string
	ThumbnailURL string
	TagIDs       []string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PublishedAt  *time.Time
}

func (p *Post) IsPublished() bool {
	return p.Status == StatusPublished
}

func (p *Post) Publish(at time.Time) {
	p.Status = StatusPublished
	p.PublishedAt = &at
	p.UpdatedAt = at
}

func (p *Post) Unpublish(at time.Time) {
	p.Status = StatusDraft
	p.PublishedAt = nil
	p.UpdatedAt = at
}

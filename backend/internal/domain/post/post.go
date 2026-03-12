package post

import "time"

// Status は記事の公開状態を表す。
type Status int32

const (
	StatusUnspecified Status = 0
	StatusDraft       Status = 1
	StatusPublished   Status = 2
)

// Slug は記事・タグのURL用識別子（値オブジェクト）。
type Slug string

// String はスラグの文字列表現を返す。
func (s Slug) String() string { return string(s) }

// Post は記事のドメインエンティティ。
type Post struct {
	ID           string
	Title        string
	Slug         Slug
	BodyMarkdown string
	Summary      string
	ThumbnailURL string // サムネイル画像の URL（空の場合は未設定）
	TagIDs       []string
	Status       Status
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PublishedAt  *time.Time
}

// IsPublished は公開済みかどうかを返す。
func (p *Post) IsPublished() bool {
	return p.Status == StatusPublished
}

// Publish は記事を公開状態にし、公開日時を設定する。
func (p *Post) Publish(at time.Time) {
	p.Status = StatusPublished
	p.PublishedAt = &at
	p.UpdatedAt = at
}

// Unpublish は記事を下書きに戻す。
func (p *Post) Unpublish(at time.Time) {
	p.Status = StatusDraft
	p.PublishedAt = nil
	p.UpdatedAt = at
}

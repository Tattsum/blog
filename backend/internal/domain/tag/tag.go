package tag

import "time"

// Slug はタグのURL用識別子（値オブジェクト）。
type Slug string

// String はスラグの文字列表現を返す。
func (s Slug) String() string { return string(s) }

// Tag はタグのドメインエンティティ。
type Tag struct {
	ID        string
	Name      string
	Slug      Slug
	CreatedAt time.Time
}

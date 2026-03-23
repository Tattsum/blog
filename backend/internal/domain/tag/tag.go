package tag

import "time"

type Slug string

func (s Slug) String() string { return string(s) }

type Tag struct {
	ID        string
	Name      string
	Slug      Slug
	CreatedAt time.Time
}

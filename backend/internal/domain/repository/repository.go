package repository

import (
	"context"
	"time"

	"github.com/Tattsum/blog/backend/internal/domain/post"
	"github.com/Tattsum/blog/backend/internal/domain/tag"
	"github.com/Tattsum/blog/backend/internal/domain/user"
)

type ListPostsFilter struct {
	Status   post.Status
	Page     int32
	PageSize int32
	TagID    string
}

type PostRepository interface {
	Create(ctx context.Context, p *post.Post) error
	GetByID(ctx context.Context, id string) (*post.Post, error)
	GetBySlug(ctx context.Context, slug post.Slug) (*post.Post, error)
	List(ctx context.Context, filter ListPostsFilter) ([]*post.Post, int64, error)
	Update(ctx context.Context, p *post.Post) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, page, pageSize int32) ([]*post.Post, int64, error)
}

type TagRepository interface {
	Create(ctx context.Context, t *tag.Tag) error
	GetByID(ctx context.Context, id string) (*tag.Tag, error)
	GetBySlug(ctx context.Context, slug tag.Slug) (*tag.Tag, error)
	List(ctx context.Context, page, pageSize int32) ([]*tag.Tag, int64, error)
	Delete(ctx context.Context, id string) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email user.Email) (*user.User, error)
	VerifyCredentials(ctx context.Context, email user.Email, plainPassword string) (*user.User, error)
}

type Clock interface {
	Now() time.Time
}

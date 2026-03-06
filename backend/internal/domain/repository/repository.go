package repository

import (
	"context"
	"time"

	"github.com/Tattsum/blog/backend/internal/domain/post"
	"github.com/Tattsum/blog/backend/internal/domain/tag"
	"github.com/Tattsum/blog/backend/internal/domain/user"
)

// ListPostsFilter は記事一覧の絞り込み条件。
type ListPostsFilter struct {
	Status   post.Status
	Page     int32
	PageSize int32
	TagID    string // 空でなければこのタグに紐づく記事のみ
}

// PostRepository は記事の永続化を抽象化するリポジトリ。
type PostRepository interface {
	Create(ctx context.Context, p *post.Post) error
	GetByID(ctx context.Context, id string) (*post.Post, error)
	GetBySlug(ctx context.Context, slug post.Slug) (*post.Post, error)
	List(ctx context.Context, filter ListPostsFilter) ([]*post.Post, int64, error)
	Update(ctx context.Context, p *post.Post) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, query string, page, pageSize int32) ([]*post.Post, int64, error)
}

// TagRepository はタグの永続化を抽象化するリポジトリ。
type TagRepository interface {
	Create(ctx context.Context, t *tag.Tag) error
	GetByID(ctx context.Context, id string) (*tag.Tag, error)
	GetBySlug(ctx context.Context, slug tag.Slug) (*tag.Tag, error)
	List(ctx context.Context, page, pageSize int32) ([]*tag.Tag, int64, error)
	Delete(ctx context.Context, id string) error
}

// UserRepository は管理者ユーザの永続化を抽象化するリポジトリ。
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email user.Email) (*user.User, error)
	// VerifyCredentials はメールと平文パスワードで認証し、成功時のみユーザを返す。
	VerifyCredentials(ctx context.Context, email user.Email, plainPassword string) (*user.User, error)
}

// Clock は現在時刻の取得を抽象化する（テストで差し替え可能）。
type Clock interface {
	Now() time.Time
}

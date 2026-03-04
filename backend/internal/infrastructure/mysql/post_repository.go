package mysql

import (
	"context"
	"database/sql"
	"time"

	"github.com/Tattsum/blog/backend/internal/domain/post"
	"github.com/Tattsum/blog/backend/internal/domain/repository"
)

// PostRepository は MySQL による PostRepository の実装。
type PostRepository struct {
	db *sql.DB
}

// NewPostRepository は PostRepository を返す。
func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create は記事を1件挿入する。
func (r *PostRepository) Create(ctx context.Context, p *post.Post) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO posts (id, title, slug, body_markdown, summary, status, created_at, updated_at, published_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, p.Title, p.Slug.String(), p.BodyMarkdown, p.Summary, int32(p.Status),
		p.CreatedAt, p.UpdatedAt, nullTime(p.PublishedAt),
	)
	if err != nil {
		return err
	}
	return r.replacePostTags(ctx, p.ID, p.TagIDs)
}

// GetByID は ID で記事を1件取得する。
func (r *PostRepository) GetByID(ctx context.Context, id string) (*post.Post, error) {
	return r.getOne(ctx, `SELECT id, title, slug, body_markdown, summary, status, created_at, updated_at, published_at FROM posts WHERE id = ?`, id)
}

// GetBySlug は slug で記事を1件取得する。
func (r *PostRepository) GetBySlug(ctx context.Context, slug post.Slug) (*post.Post, error) {
	return r.getOne(ctx, `SELECT id, title, slug, body_markdown, summary, status, created_at, updated_at, published_at FROM posts WHERE slug = ?`, slug.String())
}

func (r *PostRepository) getOne(ctx context.Context, query string, arg interface{}) (*post.Post, error) {
	var p post.Post
	var slug string
	var status int32
	var publishedAt sql.NullTime
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&p.ID, &p.Title, &slug, &p.BodyMarkdown, &p.Summary, &status,
		&p.CreatedAt, &p.UpdatedAt, &publishedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.Slug = post.Slug(slug)
	p.Status = post.Status(status)
	p.PublishedAt = nullTimePtr(publishedAt)
	tagIDs, _ := r.getTagIDsByPostID(ctx, p.ID)
	p.TagIDs = tagIDs
	return &p, nil
}

// List は条件に応じて記事一覧と総件数を返す。
func (r *PostRepository) List(ctx context.Context, filter repository.ListPostsFilter) ([]*post.Post, int64, error) {
	offset := (int64(filter.Page) - 1) * int64(filter.PageSize)
	if offset < 0 {
		offset = 0
	}
	limit := filter.PageSize
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var count int64
	countQuery := `SELECT COUNT(*) FROM posts WHERE 1=1`
	countArgs := []interface{}{}
	if filter.Status != post.StatusUnspecified {
		countQuery += ` AND status = ?`
		countArgs = append(countArgs, int32(filter.Status))
	}
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&count); err != nil {
		return nil, 0, err
	}

	listQuery := `SELECT id, title, slug, body_markdown, summary, status, created_at, updated_at, published_at FROM posts WHERE 1=1`
	listArgs := []interface{}{}
	if filter.Status != post.StatusUnspecified {
		listQuery += ` AND status = ?`
		listArgs = append(listArgs, int32(filter.Status))
	}
	listQuery += ` ORDER BY updated_at DESC LIMIT ? OFFSET ?`
	listArgs = append(listArgs, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var list []*post.Post
	for rows.Next() {
		var p post.Post
		var slug string
		var status int32
		var publishedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.Title, &slug, &p.BodyMarkdown, &p.Summary, &status, &p.CreatedAt, &p.UpdatedAt, &publishedAt); err != nil {
			return nil, 0, err
		}
		p.Slug = post.Slug(slug)
		p.Status = post.Status(status)
		p.PublishedAt = nullTimePtr(publishedAt)
		tagIDs, _ := r.getTagIDsByPostID(ctx, p.ID)
		p.TagIDs = tagIDs
		list = append(list, &p)
	}
	return list, count, rows.Err()
}

// Update は記事を1件更新する。
func (r *PostRepository) Update(ctx context.Context, p *post.Post) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE posts SET title=?, slug=?, body_markdown=?, summary=?, status=?, updated_at=?, published_at=? WHERE id=?`,
		p.Title, p.Slug.String(), p.BodyMarkdown, p.Summary, int32(p.Status), p.UpdatedAt, nullTime(p.PublishedAt), p.ID,
	)
	if err != nil {
		return err
	}
	return r.replacePostTags(ctx, p.ID, p.TagIDs)
}

// Delete は ID で記事を1件削除する。
func (r *PostRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}

// Search は全文検索（LIKE）で記事一覧と総件数を返す。本番では FULLTEXT 等を検討。
func (r *PostRepository) Search(ctx context.Context, query string, page, pageSize int32) ([]*post.Post, int64, error) {
	offset := (int64(page) - 1) * int64(pageSize)
	if offset < 0 {
		offset = 0
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	like := "%" + query + "%"

	var count int64
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM posts WHERE status=2 AND (title LIKE ? OR body_markdown LIKE ? OR summary LIKE ?)`,
		like, like, like,
	).Scan(&count)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, slug, body_markdown, summary, status, created_at, updated_at, published_at FROM posts
		 WHERE status=2 AND (title LIKE ? OR body_markdown LIKE ? OR summary LIKE ?)
		 ORDER BY updated_at DESC LIMIT ? OFFSET ?`,
		like, like, like, pageSize, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var list []*post.Post
	for rows.Next() {
		var p post.Post
		var slug string
		var status int32
		var publishedAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.Title, &slug, &p.BodyMarkdown, &p.Summary, &status, &p.CreatedAt, &p.UpdatedAt, &publishedAt); err != nil {
			return nil, 0, err
		}
		p.Slug = post.Slug(slug)
		p.Status = post.Status(status)
		p.PublishedAt = nullTimePtr(publishedAt)
		tagIDs, _ := r.getTagIDsByPostID(ctx, p.ID)
		p.TagIDs = tagIDs
		list = append(list, &p)
	}
	return list, count, rows.Err()
}

func (r *PostRepository) getTagIDsByPostID(ctx context.Context, postID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT tag_id FROM post_tags WHERE post_id = ? ORDER BY tag_id`, postID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *PostRepository) replacePostTags(ctx context.Context, postID string, tagIDs []string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM post_tags WHERE post_id = ?`, postID); err != nil {
		return err
	}
	for _, tagID := range tagIDs {
		if _, err := r.db.ExecContext(ctx, `INSERT INTO post_tags (post_id, tag_id) VALUES (?, ?)`, postID, tagID); err != nil {
			return err
		}
	}
	return nil
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func nullTimePtr(n sql.NullTime) *time.Time {
	if !n.Valid {
		return nil
	}
	return &n.Time
}

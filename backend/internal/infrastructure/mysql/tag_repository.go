package mysql

import (
	"context"
	"database/sql"

	"github.com/Tattsum/blog/backend/internal/domain/tag"
)

type TagRepository struct {
	db *sql.DB
}

func NewTagRepository(db *sql.DB) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, t *tag.Tag) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tags (id, name, slug, created_at) VALUES (?, ?, ?, ?)`,
		t.ID, t.Name, t.Slug.String(), t.CreatedAt,
	)
	return err
}

func (r *TagRepository) GetByID(ctx context.Context, id string) (*tag.Tag, error) {
	var t tag.Tag
	var slug string
	err := r.db.QueryRowContext(ctx, `SELECT id, name, slug, created_at FROM tags WHERE id = ?`, id).Scan(
		&t.ID, &t.Name, &slug, &t.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t.Slug = tag.Slug(slug)
	return &t, nil
}

func (r *TagRepository) GetBySlug(ctx context.Context, slug tag.Slug) (*tag.Tag, error) {
	var t tag.Tag
	var slugStr string
	err := r.db.QueryRowContext(ctx, `SELECT id, name, slug, created_at FROM tags WHERE slug = ?`, slug.String()).Scan(
		&t.ID, &t.Name, &slugStr, &t.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t.Slug = tag.Slug(slugStr)
	return &t, nil
}

func (r *TagRepository) List(ctx context.Context, page, pageSize int32) ([]*tag.Tag, int64, error) {
	offset := max((int64(page)-1)*int64(pageSize), 0)
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var count int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tags`).Scan(&count); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, name, slug, created_at FROM tags ORDER BY name LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var list []*tag.Tag
	for rows.Next() {
		var t tag.Tag
		var slug string
		if err := rows.Scan(&t.ID, &t.Name, &slug, &t.CreatedAt); err != nil {
			return nil, 0, err
		}
		t.Slug = tag.Slug(slug)
		list = append(list, &t)
	}
	return list, count, rows.Err()
}

func (r *TagRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ?`, id)
	return err
}

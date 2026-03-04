package mysql

import (
	"context"
	"database/sql"

	"github.com/Tattsum/blog/backend/internal/domain/user"
)

// UserRepository は MySQL による UserRepository の実装。
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository は UserRepository を返す。
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID は ID でユーザを1件取得する。
func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	var email string
	err := r.db.QueryRowContext(ctx, `SELECT id, email, display_name FROM users WHERE id = ?`, id).Scan(
		&u.ID, &email, &u.DisplayName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Email = user.Email(email)
	return &u, nil
}

// GetByEmail はメールアドレスでユーザを1件取得する。
func (r *UserRepository) GetByEmail(ctx context.Context, email user.Email) (*user.User, error) {
	var u user.User
	var emailStr string
	err := r.db.QueryRowContext(ctx, `SELECT id, email, display_name FROM users WHERE email = ?`, email.String()).Scan(
		&u.ID, &emailStr, &u.DisplayName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.Email = user.Email(emailStr)
	return &u, nil
}

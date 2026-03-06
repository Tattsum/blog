package mysql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Tattsum/blog/backend/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
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

// VerifyCredentials はメールと平文パスワードで認証する。一致時のみユーザを返す。
func (r *UserRepository) VerifyCredentials(ctx context.Context, email user.Email, plainPassword string) (*user.User, error) {
	var u user.User
	var emailStr string
	var passwordHash []byte
	err := r.db.QueryRowContext(ctx, `SELECT id, email, display_name, password_hash FROM users WHERE email = ?`, email.String()).Scan(
		&u.ID, &emailStr, &u.DisplayName, &passwordHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword(passwordHash, []byte(plainPassword)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, nil
		}
		return nil, err
	}
	u.Email = user.Email(emailStr)
	return &u, nil
}

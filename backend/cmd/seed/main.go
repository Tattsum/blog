// 初回管理者ユーザを users テーブルに登録するシードコマンド。
// 使用例: SEED_ADMIN_EMAIL=admin@example.com SEED_ADMIN_PASSWORD=yourpassword DATABASE_DSN="..." go run ./backend/cmd/seed
package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Tattsum/blog/backend/internal/infrastructure/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	dsn := os.Getenv("DATABASE_DSN")
	email := strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL"))
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	displayName := strings.TrimSpace(os.Getenv("SEED_ADMIN_DISPLAY_NAME"))

	if dsn == "" {
		log.Fatal("DATABASE_DSN is required")
	}
	if email == "" {
		log.Fatal("SEED_ADMIN_EMAIL is required")
	}
	if password == "" {
		log.Fatal("SEED_ADMIN_PASSWORD is required")
	}
	if len(password) < 8 {
		log.Fatal("SEED_ADMIN_PASSWORD must be at least 8 characters")
	}

	db, err := mysql.NewDB(mysql.Config{
		DSN:             dsn,
		MaxOpenConns:    2,
		MaxIdleConns:    1,
		ConnMaxLifetime: time.Minute,
	})
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("db close: %v", err)
		}
	}()

	ctx := context.Background()
	var existingID string
	err = db.QueryRowContext(ctx, `SELECT id FROM users WHERE email = ?`, email).Scan(&existingID)
	if err == nil {
		log.Printf("user already exists: %s (id=%s)", email, existingID)
		return
	}
	if err != sql.ErrNoRows {
		log.Fatalf("check user: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt: %v", err)
	}

	id := uuid.New().String()
	_, err = db.ExecContext(ctx,
		`INSERT INTO users (id, email, display_name, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(6), NOW(6))`,
		id, email, displayName, hash,
	)
	if err != nil {
		log.Fatalf("insert user: %v", err)
	}
	log.Printf("created admin user: %s (id=%s)", email, id)
}

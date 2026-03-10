package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	appai "github.com/Tattsum/blog/backend/internal/application/ai"
	"github.com/Tattsum/blog/backend/internal/infrastructure/mysql"
	"github.com/Tattsum/blog/backend/internal/infrastructure/vertexai"
	"github.com/Tattsum/blog/backend/internal/interface/rpc"
	blogv1connect "github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

func main() {
	initLogging()

	addr := ":" + envOrDefault("PORT", "8080")

	mux := http.NewServeMux()
	// 生存確認: /health を本番で使用（Cloud Run は末尾 "z" のパスを予約しており /healthz は 404 になる。https://cloud.google.com/run/docs/known-issues#reserved_url_paths）
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           requestLog(securityHeaders(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("listening", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	dsn := os.Getenv("DATABASE_DSN")
	if dsn != "" {
		db, err := mysql.NewDB(mysql.Config{
			DSN:             dsn,
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
		})
		if err != nil {
			slog.Error("database", "err", err)
			os.Exit(1)
		}
		defer func() {
			if err := db.Close(); err != nil {
				slog.Error("db close", "err", err)
			}
		}()

		postRepo := mysql.NewPostRepository(db)
		tagRepo := mysql.NewTagRepository(db)
		userRepo := mysql.NewUserRepository(db)
		sessionStore := rpc.NewMemSessionStore()
		adminKey := os.Getenv("ADMIN_API_KEY")

		var textGen appai.TextGenerator
		ctxBg := context.Background()
		provider := strings.ToLower(strings.TrimSpace(os.Getenv("AI_PROVIDER")))
		if provider == "vertex-claude" || provider == "claude" {
			if cl, err := appai.NewVertexClaudeFromEnv(ctxBg); err != nil {
				slog.Warn("vertex claude disabled", "err", err)
			} else if cl != nil {
				textGen = cl
				slog.Info("ai provider", "name", "vertex-claude")
			}
		}
		if textGen == nil {
			if vc, err := vertexai.NewFromEnv(ctxBg); err != nil {
				slog.Warn("vertex gemini disabled", "err", err)
			} else if vc != nil {
				textGen = appai.NewVertexGemini(vc)
				slog.Info("ai provider", "name", "vertex-gemini")
			}
		}

		postPath, postHandler := blogv1connect.NewPostServiceHandler(rpc.NewPostServer(postRepo, adminKey, sessionStore))
		tagPath, tagHandler := blogv1connect.NewTagServiceHandler(rpc.NewTagServer(tagRepo, adminKey, sessionStore))
		aiPath, aiHandler := blogv1connect.NewAIServiceHandler(rpc.NewAIServer(adminKey, sessionStore, textGen))
		authPath, authHandler := blogv1connect.NewAuthServiceHandler(rpc.NewAuthServer(userRepo, sessionStore, 24*time.Hour))

		mux.Handle(postPath, postHandler)
		mux.Handle(tagPath, tagHandler)
		mux.Handle(aiPath, aiHandler)
		mux.Handle(authPath, authHandler)
		slog.Info("rpc handlers registered")
	} else {
		slog.Warn("DATABASE_DSN not set; RPC handlers not registered")
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

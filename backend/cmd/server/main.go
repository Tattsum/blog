package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tattsum/blog/backend/internal/infrastructure/mysql"
	"github.com/Tattsum/blog/backend/internal/interface/rpc"
	blogv1connect "github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

func main() {
	addr := ":" + envOrDefault("PORT", "8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	dsn := os.Getenv("DATABASE_DSN")
	if dsn != "" {
		db, err := mysql.NewDB(mysql.Config{
			DSN:             dsn,
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Hour,
		})
		if err != nil {
			log.Fatalf("database: %v", err)
		}
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("db close: %v", err)
			}
		}()

		postRepo := mysql.NewPostRepository(db)
		tagRepo := mysql.NewTagRepository(db)
		userRepo := mysql.NewUserRepository(db)
		sessionStore := rpc.NewMemSessionStore()
		adminKey := os.Getenv("ADMIN_API_KEY")

		postPath, postHandler := blogv1connect.NewPostServiceHandler(rpc.NewPostServer(postRepo, adminKey, sessionStore))
		tagPath, tagHandler := blogv1connect.NewTagServiceHandler(rpc.NewTagServer(tagRepo, adminKey, sessionStore))
		aiPath, aiHandler := blogv1connect.NewAIServiceHandler(rpc.NewAIServer(adminKey, sessionStore))
		authPath, authHandler := blogv1connect.NewAuthServiceHandler(rpc.NewAuthServer(userRepo, sessionStore, 24*time.Hour))

		mux.Handle(postPath, postHandler)
		mux.Handle(tagPath, tagHandler)
		mux.Handle(aiPath, aiHandler)
		mux.Handle(authPath, authHandler)
		log.Print("RPC handlers registered (PostService, TagService, AIService, AuthService)")
	} else {
		log.Print("DATABASE_DSN not set; RPC handlers not registered")
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
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

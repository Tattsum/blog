package rpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/Tattsum/blog/backend/internal/domain/repository"
	"github.com/Tattsum/blog/backend/internal/domain/user"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

var (
	errMissingCredentials      = errors.New("email and password are required")
	errInvalidCredentials      = errors.New("invalid email or password")
	errMissingSession          = errors.New("missing or invalid Authorization header")
	errInvalidOrExpiredSession = errors.New("session invalid or expired")
)

const defaultSessionDur = 24 * time.Hour

// AuthServer は AuthService の connect-go ハンドラ実装。
type AuthServer struct {
	blogv1connect.UnimplementedAuthServiceHandler
	userRepo        repository.UserRepository
	sessionStore    SessionStore
	sessionDuration time.Duration
}

// NewAuthServer は AuthServer を返す。
func NewAuthServer(userRepo repository.UserRepository, sessionStore SessionStore, sessionDuration time.Duration) *AuthServer {
	if sessionDuration <= 0 {
		sessionDuration = defaultSessionDur
	}
	return &AuthServer{
		userRepo:        userRepo,
		sessionStore:    sessionStore,
		sessionDuration: sessionDuration,
	}
}

// Login はメール・パスワードで認証し、セッショントークンを返す。
func (s *AuthServer) Login(ctx context.Context, req *connect.Request[blogv1.LoginRequest]) (*connect.Response[blogv1.LoginResponse], error) {
	email := strings.TrimSpace(req.Msg.GetEmail())
	password := req.Msg.GetPassword()
	if email == "" || password == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errMissingCredentials)
	}
	u, err := s.userRepo.VerifyCredentials(ctx, user.Email(email), password)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	if u == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errInvalidCredentials)
	}
	expiresAt := time.Now().Add(s.sessionDuration)
	token := s.sessionStore.Create(u.ID, expiresAt)
	return connect.NewResponse(&blogv1.LoginResponse{
		SessionToken: token,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
	}), nil
}

// Logout はリクエストに含まれる Bearer トークンを無効化する。
func (s *AuthServer) Logout(ctx context.Context, req *connect.Request[blogv1.LogoutRequest]) (*connect.Response[blogv1.LogoutResponse], error) {
	token := bearerToken(req.Header())
	if token != "" {
		s.sessionStore.Delete(token)
	}
	return connect.NewResponse(&blogv1.LogoutResponse{}), nil
}

// GetMe は Authorization: Bearer <token> からセッションを解決し、現在のユーザ情報を返す。
func (s *AuthServer) GetMe(ctx context.Context, req *connect.Request[blogv1.GetMeRequest]) (*connect.Response[blogv1.GetMeResponse], error) {
	token := bearerToken(req.Header())
	if token == "" {
		return nil, connect.NewError(connect.CodeUnauthenticated, errMissingSession)
	}
	userID, ok := s.sessionStore.Get(token)
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, errInvalidOrExpiredSession)
	}
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	if u == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errInvalidOrExpiredSession)
	}
	return connect.NewResponse(&blogv1.GetMeResponse{
		Id:          u.ID,
		Email:       u.Email.String(),
		DisplayName: u.DisplayName,
	}), nil
}

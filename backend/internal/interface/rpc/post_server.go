package rpc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"connectrpc.com/connect"
	"github.com/Tattsum/blog/backend/internal/application/ai"
	"github.com/Tattsum/blog/backend/internal/domain/post"
	"github.com/Tattsum/blog/backend/internal/domain/repository"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
	"github.com/google/uuid"
)

// PostServer は PostService の connect-go ハンドラ実装。
type PostServer struct {
	blogv1connect.UnimplementedPostServiceHandler
	posts           repository.PostRepository
	adminKey        string
	sessionStore    SessionStore
	defaultProvider string
	gemini          ai.TextGenerator
	claude          ai.TextGenerator
}

// NewPostServer は PostServer を返す。認証は X-Admin-Key または Bearer セッションのいずれかで行う。
func NewPostServer(posts repository.PostRepository, adminKey string, sessionStore SessionStore, provider string, gemini, claude ai.TextGenerator) *PostServer {
	return &PostServer{
		posts:           posts,
		adminKey:        adminKey,
		sessionStore:    sessionStore,
		defaultProvider: strings.ToLower(strings.TrimSpace(provider)),
		gemini:          gemini,
		claude:          claude,
	}
}

// ListPosts は記事一覧を返す。未認証時は status=published のみ許可。
func (s *PostServer) ListPosts(ctx context.Context, req *connect.Request[blogv1.ListPostsRequest]) (*connect.Response[blogv1.ListPostsResponse], error) {
	page := max(req.Msg.GetPage(), 1)
	pageSize := req.Msg.GetPageSize()
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	statusStr := req.Msg.GetStatus()
	var statusFilter post.Status
	switch statusStr {
	case "published", "":
		statusFilter = post.StatusPublished
	case "draft":
		if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
			return nil, err
		}
		statusFilter = post.StatusDraft
	default:
		statusFilter = post.StatusPublished
	}
	filter := repository.ListPostsFilter{
		Status:   statusFilter,
		Page:     page,
		PageSize: pageSize,
		TagID:    strings.TrimSpace(req.Msg.GetTagId()),
	}
	list, total, err := s.posts.List(ctx, filter)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	posts := make([]*blogv1.Post, 0, len(list))
	for _, p := range list {
		posts = append(posts, PostToProto(p))
	}
	return connect.NewResponse(&blogv1.ListPostsResponse{
		Posts:      posts,
		TotalCount: int32(total),
	}), nil
}

// GetPost は ID または slug で記事を1件返す。未認証時は公開記事のみ。
func (s *PostServer) GetPost(ctx context.Context, req *connect.Request[blogv1.GetPostRequest]) (*connect.Response[blogv1.GetPostResponse], error) {
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	if p == nil {
		p, err = s.posts.GetBySlug(ctx, post.Slug(id))
		if err != nil {
			return nil, MapHandlerError(err)
		}
	}
	if p == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
	}
	if !p.IsPublished() {
		if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&blogv1.GetPostResponse{Post: PostToProto(p)}), nil
}

// CreatePost は記事を下書きで作成する。管理者キー必須。
func (s *PostServer) CreatePost(ctx context.Context, req *connect.Request[blogv1.CreatePostRequest]) (*connect.Response[blogv1.CreatePostResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Msg.GetTitle())
	slug := strings.TrimSpace(req.Msg.GetSlug())
	if slug == "" {
		slug = s.generateSlug(ctx, req.Header(), title)
	}
	body := req.Msg.GetBodyMarkdown()
	summary := req.Msg.GetSummary()
	thumbnailURL := strings.TrimSpace(req.Msg.GetThumbnailUrl())
	tagIDs := req.Msg.GetTagIds()
	if err := validatePostFields(title, slug, body, summary, thumbnailURL, tagIDs); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	now := time.Now()
	p := &post.Post{
		ID:           uuid.New().String(),
		Title:        title,
		Slug:         post.Slug(slug),
		BodyMarkdown: req.Msg.GetBodyMarkdown(),
		Summary:      req.Msg.GetSummary(),
		ThumbnailURL: thumbnailURL,
		TagIDs:       req.Msg.GetTagIds(),
		Status:       post.StatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.posts.Create(ctx, p); err != nil {
		return nil, MapHandlerError(err)
	}
	return connect.NewResponse(&blogv1.CreatePostResponse{Post: PostToProto(p)}), nil
}

func (s *PostServer) pickSlugGenerator(h map[string][]string) (provider string, gen ai.TextGenerator, specified bool) {
	get := func(key string) string {
		v := h[key]
		if len(v) == 0 {
			v = h[http.CanonicalHeaderKey(key)]
		}
		if len(v) == 0 {
			return ""
		}
		return v[0]
	}
	p := strings.ToLower(strings.TrimSpace(get("X-AI-Provider")))
	specified = p != ""
	if p == "" {
		p = s.defaultProvider
	}
	switch p {
	case "vertex-claude", "claude":
		return "claude", s.claude, specified
	case "", "vertex-gemini", "gemini":
		return "gemini", s.gemini, specified
	default:
		return p, nil, specified
	}
}

func (s *PostServer) generateSlug(ctx context.Context, header map[string][]string, title string) string {
	if !containsJapanese(title) {
		return Slugify(title)
	}

	_, gen, specified := s.pickSlugGenerator(header)
	if specified && gen == nil {
		return Slugify(title)
	}
	if gen == nil {
		return Slugify(title)
	}

	prompt := fmt.Sprintf(
		"次のタイトルを、意味を反映した英語の短いURLスラグに変換してください。出力は小文字の英数字とハイフンのみ、スペース禁止、最大80文字。説明文や引用符は不要で、スラグ文字列だけを返してください。\n\nタイトル:\n%s",
		title,
	)
	slugRaw, err := gen.GenerateText(ctx, prompt)
	if err != nil {
		return Slugify(title)
	}
	slugRaw = strings.TrimSpace(slugRaw)
	if slugRaw == "" {
		return Slugify(title)
	}
	candidate := Slugify(slugRaw)
	if candidate == "" || len(candidate) > 80 || !slugPattern.MatchString(candidate) {
		return Slugify(title)
	}
	return candidate
}

func (s *PostServer) GenerateSlugForTitle(ctx context.Context, header map[string][]string, title string) string {
	return s.generateSlug(ctx, header, title)
}

func containsJapanese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
			return true
		}
	}
	return false
}

// UpdatePost は記事を更新する。管理者キー必須。
func (s *PostServer) UpdatePost(ctx context.Context, req *connect.Request[blogv1.UpdatePostRequest]) (*connect.Response[blogv1.UpdatePostResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	if p == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
	}
	if req.Msg.Title != nil {
		p.Title = strings.TrimSpace(*req.Msg.Title)
	}
	if req.Msg.Slug != nil {
		p.Slug = post.Slug(strings.TrimSpace(*req.Msg.Slug))
	}
	if req.Msg.BodyMarkdown != nil {
		p.BodyMarkdown = *req.Msg.BodyMarkdown
	}
	if req.Msg.Summary != nil {
		p.Summary = *req.Msg.Summary
	}
	if req.Msg.ThumbnailUrl != nil {
		p.ThumbnailURL = strings.TrimSpace(*req.Msg.ThumbnailUrl)
	}
	if req.Msg.TagIds != nil {
		p.TagIDs = req.Msg.TagIds
	}
	if err := validatePostFields(p.Title, p.Slug.String(), p.BodyMarkdown, p.Summary, p.ThumbnailURL, p.TagIDs); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	p.UpdatedAt = time.Now()
	if err := s.posts.Update(ctx, p); err != nil {
		return nil, MapHandlerError(err)
	}
	return connect.NewResponse(&blogv1.UpdatePostResponse{Post: PostToProto(p)}), nil
}

// DeletePost は記事を削除する。管理者キー必須。
func (s *PostServer) DeletePost(ctx context.Context, req *connect.Request[blogv1.DeletePostRequest]) (*connect.Response[blogv1.DeletePostResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	if err := s.posts.Delete(ctx, id); err != nil {
		return nil, MapHandlerError(err)
	}
	return connect.NewResponse(&blogv1.DeletePostResponse{}), nil
}

// SearchPosts は公開記事を全文検索する。未認証でも利用可能。
func (s *PostServer) SearchPosts(ctx context.Context, req *connect.Request[blogv1.SearchPostsRequest]) (*connect.Response[blogv1.SearchPostsResponse], error) {
	query := strings.TrimSpace(req.Msg.GetQuery())
	if query == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("query is required"))
	}
	page := max(req.Msg.GetPage(), 1)
	pageSize := req.Msg.GetPageSize()
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.posts.Search(ctx, query, page, pageSize)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	posts := make([]*blogv1.Post, 0, len(list))
	for _, p := range list {
		posts = append(posts, PostToProto(p))
	}
	return connect.NewResponse(&blogv1.SearchPostsResponse{
		Posts:      posts,
		TotalCount: int32(total),
	}), nil
}

// PublishPost は記事を公開または下書きに戻す。管理者キー必須。
func (s *PostServer) PublishPost(ctx context.Context, req *connect.Request[blogv1.PublishPostRequest]) (*connect.Response[blogv1.PublishPostResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	if p == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
	}
	now := time.Now()
	if req.Msg.GetUnpublish() {
		p.Unpublish(now)
	} else {
		p.Publish(now)
	}
	if err := s.posts.Update(ctx, p); err != nil {
		return nil, MapHandlerError(err)
	}
	return connect.NewResponse(&blogv1.PublishPostResponse{Post: PostToProto(p)}), nil
}

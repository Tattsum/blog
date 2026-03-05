package rpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/Tattsum/blog/backend/internal/domain/post"
	"github.com/Tattsum/blog/backend/internal/domain/repository"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
	"github.com/google/uuid"
)

// PostServer は PostService の connect-go ハンドラ実装。
type PostServer struct {
	blogv1connect.UnimplementedPostServiceHandler
	posts    repository.PostRepository
	adminKey string
}

// NewPostServer は PostServer を返す。adminKey が空の場合は作成・更新・削除・公開は PermissionDenied となる。
func NewPostServer(posts repository.PostRepository, adminKey string) *PostServer {
	return &PostServer{posts: posts, adminKey: adminKey}
}

// ListPosts は記事一覧を返す。未認証時は status=published のみ許可。
func (s *PostServer) ListPosts(ctx context.Context, req *connect.Request[blogv1.ListPostsRequest]) (*connect.Response[blogv1.ListPostsResponse], error) {
	page := req.Msg.GetPage()
	if page < 1 {
		page = 1
	}
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
		if err := requireAdmin(s.adminKey, req.Header()); err != nil {
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
	}
	list, total, err := s.posts.List(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if p == nil {
		p, err = s.posts.GetBySlug(ctx, post.Slug(id))
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}
	if p == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
	}
	if !p.IsPublished() {
		if err := requireAdmin(s.adminKey, req.Header()); err != nil {
			return nil, err
		}
	}
	return connect.NewResponse(&blogv1.GetPostResponse{Post: PostToProto(p)}), nil
}

// CreatePost は記事を下書きで作成する。管理者キー必須。
func (s *PostServer) CreatePost(ctx context.Context, req *connect.Request[blogv1.CreatePostRequest]) (*connect.Response[blogv1.CreatePostResponse], error) {
	if err := requireAdmin(s.adminKey, req.Header()); err != nil {
		return nil, err
	}
	title := strings.TrimSpace(req.Msg.GetTitle())
	if title == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("title is required"))
	}
	slug := strings.TrimSpace(req.Msg.GetSlug())
	if slug == "" {
		slug = Slugify(title)
	}
	if slug == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("slug could not be generated from title"))
	}
	now := time.Now()
	p := &post.Post{
		ID:           uuid.New().String(),
		Title:        title,
		Slug:         post.Slug(slug),
		BodyMarkdown: req.Msg.GetBodyMarkdown(),
		Summary:      req.Msg.GetSummary(),
		TagIDs:       req.Msg.GetTagIds(),
		Status:       post.StatusDraft,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.posts.Create(ctx, p); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&blogv1.CreatePostResponse{Post: PostToProto(p)}), nil
}

// UpdatePost は記事を更新する。管理者キー必須。
func (s *PostServer) UpdatePost(ctx context.Context, req *connect.Request[blogv1.UpdatePostRequest]) (*connect.Response[blogv1.UpdatePostResponse], error) {
	if err := requireAdmin(s.adminKey, req.Header()); err != nil {
		return nil, err
	}
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if p == nil {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
	}
	if req.Msg.Title != nil {
		p.Title = strings.TrimSpace(*req.Msg.Title)
		if p.Title == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("title cannot be empty"))
		}
	}
	if req.Msg.Slug != nil {
		p.Slug = post.Slug(strings.TrimSpace(*req.Msg.Slug))
		if string(p.Slug) == "" {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("slug cannot be empty"))
		}
	}
	if req.Msg.BodyMarkdown != nil {
		p.BodyMarkdown = *req.Msg.BodyMarkdown
	}
	if req.Msg.Summary != nil {
		p.Summary = *req.Msg.Summary
	}
	if req.Msg.TagIds != nil {
		p.TagIDs = req.Msg.TagIds
	}
	p.UpdatedAt = time.Now()
	if err := s.posts.Update(ctx, p); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&blogv1.UpdatePostResponse{Post: PostToProto(p)}), nil
}

// DeletePost は記事を削除する。管理者キー必須。
func (s *PostServer) DeletePost(ctx context.Context, req *connect.Request[blogv1.DeletePostRequest]) (*connect.Response[blogv1.DeletePostResponse], error) {
	if err := requireAdmin(s.adminKey, req.Header()); err != nil {
		return nil, err
	}
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	if err := s.posts.Delete(ctx, id); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&blogv1.DeletePostResponse{}), nil
}

// SearchPosts は公開記事を全文検索する。未認証でも利用可能。
func (s *PostServer) SearchPosts(ctx context.Context, req *connect.Request[blogv1.SearchPostsRequest]) (*connect.Response[blogv1.SearchPostsResponse], error) {
	query := strings.TrimSpace(req.Msg.GetQuery())
	if query == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("query is required"))
	}
	page := req.Msg.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.Msg.GetPageSize()
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.posts.Search(ctx, query, page, pageSize)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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
	if err := requireAdmin(s.adminKey, req.Header()); err != nil {
		return nil, err
	}
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	p, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
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
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&blogv1.PublishPostResponse{Post: PostToProto(p)}), nil
}

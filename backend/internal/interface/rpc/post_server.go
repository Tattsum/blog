package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/Tattsum/blog/backend/internal/domain/post"
	"github.com/Tattsum/blog/backend/internal/domain/repository"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

// PostServer は PostService の connect-go ハンドラ実装。
type PostServer struct {
	blogv1connect.UnimplementedPostServiceHandler
	posts repository.PostRepository
}

// NewPostServer は PostServer を返す。
func NewPostServer(posts repository.PostRepository) *PostServer {
	return &PostServer{posts: posts}
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
		return nil, connect.NewError(connect.CodePermissionDenied, errors.New("listing draft posts requires authentication"))
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
		return nil, connect.NewError(connect.CodeNotFound, errors.New("post not found"))
	}
	return connect.NewResponse(&blogv1.GetPostResponse{Post: PostToProto(p)}), nil
}

// CreatePost は未実装（認証実装後に追加）。
func (s *PostServer) CreatePost(context.Context, *connect.Request[blogv1.CreatePostRequest]) (*connect.Response[blogv1.CreatePostResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("CreatePost not yet implemented"))
}

// UpdatePost は未実装。
func (s *PostServer) UpdatePost(context.Context, *connect.Request[blogv1.UpdatePostRequest]) (*connect.Response[blogv1.UpdatePostResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("UpdatePost not yet implemented"))
}

// DeletePost は未実装。
func (s *PostServer) DeletePost(context.Context, *connect.Request[blogv1.DeletePostRequest]) (*connect.Response[blogv1.DeletePostResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("DeletePost not yet implemented"))
}

// SearchPosts は未実装。
func (s *PostServer) SearchPosts(context.Context, *connect.Request[blogv1.SearchPostsRequest]) (*connect.Response[blogv1.SearchPostsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("SearchPosts not yet implemented"))
}

// PublishPost は未実装。
func (s *PostServer) PublishPost(context.Context, *connect.Request[blogv1.PublishPostRequest]) (*connect.Response[blogv1.PublishPostResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("PublishPost not yet implemented"))
}

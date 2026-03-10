package rpc

import (
	"context"
	"errors"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/Tattsum/blog/backend/internal/domain/repository"
	"github.com/Tattsum/blog/backend/internal/domain/tag"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
	"github.com/google/uuid"
)

// TagServer は TagService の connect-go ハンドラ実装。
type TagServer struct {
	blogv1connect.UnimplementedTagServiceHandler
	tags         repository.TagRepository
	adminKey     string
	sessionStore SessionStore
}

// NewTagServer は TagServer を返す。認証は X-Admin-Key または Bearer セッションのいずれかで行う。
func NewTagServer(tags repository.TagRepository, adminKey string, sessionStore SessionStore) *TagServer {
	return &TagServer{tags: tags, adminKey: adminKey, sessionStore: sessionStore}
}

// ListTags はタグ一覧を返す。
func (s *TagServer) ListTags(ctx context.Context, req *connect.Request[blogv1.ListTagsRequest]) (*connect.Response[blogv1.ListTagsResponse], error) {
	page := req.Msg.GetPage()
	if page < 1 {
		page = 1
	}
	pageSize := req.Msg.GetPageSize()
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	list, total, err := s.tags.List(ctx, page, pageSize)
	if err != nil {
		return nil, MapHandlerError(err)
	}
	tags := make([]*blogv1.Tag, 0, len(list))
	for _, t := range list {
		tags = append(tags, TagToProto(t))
	}
	return connect.NewResponse(&blogv1.ListTagsResponse{
		Tags:       tags,
		TotalCount: int32(total),
	}), nil
}

// CreateTag はタグを1件作成する。管理者キー必須。
func (s *TagServer) CreateTag(ctx context.Context, req *connect.Request[blogv1.CreateTagRequest]) (*connect.Response[blogv1.CreateTagResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Msg.GetName())
	if name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name is required"))
	}
	slug := strings.TrimSpace(req.Msg.GetSlug())
	if slug == "" {
		slug = Slugify(name)
	}
	if slug == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("slug could not be generated from name"))
	}
	now := time.Now()
	t := &tag.Tag{
		ID:        uuid.New().String(),
		Name:      name,
		Slug:      tag.Slug(slug),
		CreatedAt: now,
	}
	if err := s.tags.Create(ctx, t); err != nil {
		return nil, MapHandlerError(err)
	}
	return connect.NewResponse(&blogv1.CreateTagResponse{Tag: TagToProto(t)}), nil
}

// DeleteTag はタグを削除する。管理者キー必須。
func (s *TagServer) DeleteTag(ctx context.Context, req *connect.Request[blogv1.DeleteTagRequest]) (*connect.Response[blogv1.DeleteTagResponse], error) {
	if err := requireAdminOrSession(s.adminKey, req.Header(), s.sessionStore); err != nil {
		return nil, err
	}
	id := req.Msg.GetId()
	if id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("id is required"))
	}
	if err := s.tags.Delete(ctx, id); err != nil {
		return nil, MapHandlerError(err)
	}
	return connect.NewResponse(&blogv1.DeleteTagResponse{}), nil
}

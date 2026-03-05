package rpc

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/Tattsum/blog/backend/internal/domain/repository"
	blogv1 "github.com/Tattsum/blog/gen/blog/v1"
	"github.com/Tattsum/blog/gen/blog/v1/blogv1connect"
)

// TagServer は TagService の connect-go ハンドラ実装。
type TagServer struct {
	blogv1connect.UnimplementedTagServiceHandler
	tags repository.TagRepository
}

// NewTagServer は TagServer を返す。
func NewTagServer(tags repository.TagRepository) *TagServer {
	return &TagServer{tags: tags}
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
		return nil, connect.NewError(connect.CodeInternal, err)
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

// CreateTag は未実装（認証実装後に追加）。
func (s *TagServer) CreateTag(context.Context, *connect.Request[blogv1.CreateTagRequest]) (*connect.Response[blogv1.CreateTagResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("CreateTag not yet implemented"))
}

// DeleteTag は未実装。
func (s *TagServer) DeleteTag(context.Context, *connect.Request[blogv1.DeleteTagRequest]) (*connect.Response[blogv1.DeleteTagResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("DeleteTag not yet implemented"))
}

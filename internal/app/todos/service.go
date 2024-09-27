package todos

import (
	"connectrpc.com/connect"
	"context"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	v1 "go.azuka.me/connect-playground/internal/pkg/proto/playground/v1"
	"go.azuka.me/connect-playground/internal/pkg/proto/playground/v1/v1connect"
)

type Service struct {
	logger *otelzap.Logger
	v1connect.UnimplementedTodoServiceHandler
}

func New(logger *otelzap.Logger) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) List(
	ctx context.Context,
	req *connect.Request[v1.TodoServiceListRequest]) (*connect.Response[v1.TodoServiceListResponse], error) {
	return connect.NewResponse(&v1.TodoServiceListResponse{
		Items: []*v1.TodoServiceCreateResponse{
			{
				Name: "Example 1",
			},
		},
		Pagination: &v1.Pagination{
			Total: 2,
		},
	}), nil
}

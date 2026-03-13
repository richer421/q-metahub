package metadata

import "context"

// Service 定义 metahub 元数据领域层的最小边界。
// 当前具体业务仍由 app 层直接组织，后续领域规则应逐步下沉到这里。
type Service interface {
	Name() string
	Health(ctx context.Context) error
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) Name() string {
	return "metadata"
}

func (s *service) Health(_ context.Context) error {
	return nil
}

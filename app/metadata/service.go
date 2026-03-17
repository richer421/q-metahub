package metadata

import (
	"context"

	"github.com/richer421/q-metahub/app/metadata/vo"
	domainmetadata "github.com/richer421/q-metahub/domain/metadata"
)

type Service struct {
	domain domainmetadata.Service
}

func NewService() *Service {
	return &Service{
		domain: domainmetadata.NewService(),
	}
}

func (s *Service) GetBusinessUnitFullSpec(ctx context.Context, businessUnitID int64) (*vo.BusinessUnitFullSpecDTO, error) {
	aggregate, err := loadBusinessUnitAggregate(ctx, businessUnitID)
	if err != nil {
		return nil, err
	}
	return aggregate.toFullSpecDTO(), nil
}

func (s *Service) GetDeployPlanFullSpec(ctx context.Context, deployPlanID int64) (*vo.DeployPlanAggregateDTO, error) {
	aggregate, err := loadDeployPlanAggregate(ctx, deployPlanID)
	if err != nil {
		return nil, err
	}
	return aggregate.toDTO(), nil
}

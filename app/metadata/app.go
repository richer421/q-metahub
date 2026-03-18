package metadata

import (
	"context"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
)

type app struct{}

var App = new(app)

func (s *app) GetDeployPlan(ctx context.Context, deployPlanID int64) (*vo.DeployPlanAggregateVO, error) {
	aggregate, err := loadDeployPlanAggregate(ctx, deployPlanID)
	if err != nil {
		return nil, err
	}
	return aggregate.toVO(), nil
}

func (s *app) CreateInstanceOAM(ctx context.Context, createReq vo.CreateInstanceOAMReq) (vo.InstanceOAMVO, error) {
	oam := convert2OAM(createReq)
	q := dao.Q.WithContext(ctx)
	err := q.InstanceOAM.Create(&oam)
	if err != nil {
		return vo.InstanceOAMVO{}, err
	}
	return convertToInstOAMVO(oam), nil
}

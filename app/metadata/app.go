package metadata

import (
	"context"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
)

type App struct {
	ctx context.Context
}

func NewApp(ctx context.Context) *App {
	return &App{
		ctx: ctx,
	}
}

func (s *App) GetDeployPlan(deployPlanID int64) (*vo.DeployPlanAggregateVO, error) {
	aggregate, err := loadDeployPlanAggregate(s.ctx, deployPlanID)
	if err != nil {
		return nil, err
	}
	return aggregate.toVO(), nil
}

func (s *App) CreateInstanceOAM(createReq vo.CreateInstanceOAMReq) (vo.InstanceOAMVO, error) {
	oam := convert2OAM(createReq)
	q := dao.Q.WithContext(s.ctx)
	err := q.InstanceOAM.Create(&oam)
	if err != nil {
		return vo.InstanceOAMVO{}, err
	}
	return convertToInstOAMVO(oam), nil
}

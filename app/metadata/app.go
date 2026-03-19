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

func (s *app) ListBusinessUnits(ctx context.Context, page int, pageSize int, keyword string) (*vo.BusinessUnitPageDTO, error) {
	return listBusinessUnits(ctx, page, pageSize, keyword)
}

func (s *app) CreateBusinessUnit(ctx context.Context, req vo.CreateBusinessUnitReq) (*vo.BusinessUnitVO, error) {
	return createBusinessUnit(ctx, req)
}

func (s *app) UpdateBusinessUnit(ctx context.Context, businessUnitID int64, req vo.UpdateBusinessUnitReq) (*vo.BusinessUnitVO, error) {
	return updateBusinessUnit(ctx, businessUnitID, req)
}

func (s *app) DeleteBusinessUnit(ctx context.Context, businessUnitID int64) error {
	return deleteBusinessUnit(ctx, businessUnitID)
}

func (s *app) ListBusinessUnitCIConfigs(ctx context.Context, businessUnitID int64, page int, pageSize int, keyword string) (*vo.CIConfigPageVO, error) {
	return listBusinessUnitCIConfigs(ctx, businessUnitID, page, pageSize, keyword)
}

func (s *app) GetCIConfig(ctx context.Context, ciConfigID int64) (*vo.CIConfigVO, error) {
	return getCIConfig(ctx, ciConfigID)
}

func (s *app) CreateBusinessUnitCIConfig(ctx context.Context, businessUnitID int64, req vo.CreateCIConfigReq) (*vo.CIConfigVO, error) {
	return createBusinessUnitCIConfig(ctx, businessUnitID, req)
}

func (s *app) UpdateCIConfig(ctx context.Context, ciConfigID int64, req vo.UpdateCIConfigReq) (*vo.CIConfigVO, error) {
	return updateCIConfig(ctx, ciConfigID, req)
}

func (s *app) DeleteCIConfig(ctx context.Context, ciConfigID int64) error {
	return deleteCIConfig(ctx, ciConfigID)
}

func (s *app) ListBusinessUnitCDConfigs(ctx context.Context, businessUnitID int64, req vo.CDConfigListReq) (*vo.CDConfigPageDTO, error) {
	return listBusinessUnitCDConfigs(ctx, businessUnitID, req)
}

func (s *app) GetCDConfig(ctx context.Context, id int64) (*vo.CDConfigFrontendVO, error) {
	return getCDConfig(ctx, id)
}

func (s *app) CreateBusinessUnitCDConfig(ctx context.Context, businessUnitID int64, req vo.UpsertCDConfigReq) (*vo.CDConfigFrontendVO, error) {
	return createBusinessUnitCDConfig(ctx, businessUnitID, req)
}

func (s *app) UpdateCDConfig(ctx context.Context, id int64, req vo.UpsertCDConfigReq) (*vo.CDConfigFrontendVO, error) {
	return updateCDConfig(ctx, id, req)
}

func (s *app) DeleteCDConfig(ctx context.Context, id int64) error {
	return deleteCDConfig(ctx, id)
}

func (s *app) ListInstanceOAMTemplates(ctx context.Context) []vo.InstanceOAMTemplateDTO {
	return listInstanceOAMTemplates(ctx)
}

func (s *app) ListBusinessUnitInstanceOAMs(ctx context.Context, businessUnitID int64, page int, pageSize int, env string, keyword string) (*vo.InstanceOAMPageDTO, error) {
	return listBusinessUnitInstanceOAMs(ctx, businessUnitID, page, pageSize, env, keyword)
}

func (s *app) CreateBusinessUnitInstanceOAM(ctx context.Context, businessUnitID int64, req vo.CreateInstanceOAMFromTemplateReq) (*vo.InstanceOAMDTO, error) {
	return createBusinessUnitInstanceOAM(ctx, businessUnitID, req)
}

func (s *app) UpdateInstanceOAM(ctx context.Context, instanceOAMID int64, req vo.UpdateInstanceOAMReq) (*vo.InstanceOAMDTO, error) {
	return updateInstanceOAM(ctx, instanceOAMID, req)
}

func (s *app) DeleteInstanceOAM(ctx context.Context, instanceOAMID int64) error {
	return deleteInstanceOAM(ctx, instanceOAMID)
}

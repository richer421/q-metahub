package metadata

import (
	"context"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

type businessUnitAggregate struct {
	project         *model.Project
	businessUnit    *model.BusinessUnit
	ciConfigs       []*model.CIConfig
	cdConfigs       []*model.CDConfig
	instanceConfigs []*model.InstanceOAM
	deployPlans     []*model.DeployPlan
}

func loadBusinessUnitAggregate(ctx context.Context, businessUnitID int64) (*businessUnitAggregate, error) {
	businessUnit, err := dao.Q.WithContext(ctx).BusinessUnit.Where(dao.BusinessUnit.ID.Eq(businessUnitID)).First()
	if err != nil {
		return nil, err
	}
	project, err := dao.Q.WithContext(ctx).Project.Where(dao.Project.ID.Eq(businessUnit.ProjectID)).First()
	if err != nil {
		return nil, err
	}
	ciConfigs, err := dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.BusinessUnitID.Eq(businessUnitID)).Find()
	if err != nil {
		return nil, err
	}
	cdConfigs, err := dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.BusinessUnitID.Eq(businessUnitID)).Find()
	if err != nil {
		return nil, err
	}
	instanceConfigs, err := dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.BusinessUnitID.Eq(businessUnitID)).Find()
	if err != nil {
		return nil, err
	}
	deployPlans, err := dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.BusinessUnitID.Eq(businessUnitID)).Find()
	if err != nil {
		return nil, err
	}

	return &businessUnitAggregate{
		project:         project,
		businessUnit:    businessUnit,
		ciConfigs:       ciConfigs,
		cdConfigs:       cdConfigs,
		instanceConfigs: instanceConfigs,
		deployPlans:     deployPlans,
	}, nil
}

func (a *businessUnitAggregate) toFullSpecDTO() *vo.BusinessUnitFullSpecDTO {
	return &vo.BusinessUnitFullSpecDTO{
		Project:         toProjectDTO(a.project),
		BusinessUnit:    toBusinessUnitDTO(a.businessUnit),
		CIConfigs:       mapCIConfigs(a.ciConfigs),
		CDConfigs:       mapCDConfigs(a.cdConfigs),
		InstanceConfigs: mapInstanceConfigs(a.instanceConfigs),
		DeployPlans:     mapDeployPlans(a.deployPlans),
	}
}

type deployPlanAggregate struct {
	project        *model.Project
	businessUnit   *model.BusinessUnit
	ciConfig       *model.CIConfig
	cdConfig       *model.CDConfig
	instanceConfig *model.InstanceOAM
	deployPlan     *model.DeployPlan
}

func loadDeployPlanAggregate(ctx context.Context, deployPlanID int64) (*deployPlanAggregate, error) {
	deployPlan, err := dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.ID.Eq(deployPlanID)).First()
	if err != nil {
		return nil, err
	}
	businessUnit, err := dao.Q.WithContext(ctx).BusinessUnit.Where(dao.BusinessUnit.ID.Eq(deployPlan.BusinessUnitID)).First()
	if err != nil {
		return nil, err
	}
	project, err := dao.Q.WithContext(ctx).Project.Where(dao.Project.ID.Eq(businessUnit.ProjectID)).First()
	if err != nil {
		return nil, err
	}
	ciConfig, err := dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.ID.Eq(deployPlan.CIConfigID)).First()
	if err != nil {
		return nil, err
	}
	cdConfig, err := dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.ID.Eq(deployPlan.CDConfigID)).First()
	if err != nil {
		return nil, err
	}
	instanceConfig, err := dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.ID.Eq(deployPlan.InstanceConfigID)).First()
	if err != nil {
		return nil, err
	}

	return &deployPlanAggregate{
		project:        project,
		businessUnit:   businessUnit,
		ciConfig:       ciConfig,
		cdConfig:       cdConfig,
		instanceConfig: instanceConfig,
		deployPlan:     deployPlan,
	}, nil
}

func (a *deployPlanAggregate) toDTO() *vo.DeployPlanAggregateDTO {
	return aggregateDTO(a.project, a.businessUnit, a.ciConfig, a.cdConfig, a.instanceConfig, a.deployPlan)
}

package metadata

import (
	"context"

	"github.com/richer421/q-metahub/app/metadata/vo"
	domainmetadata "github.com/richer421/q-metahub/domain/metadata"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

type Service struct {
	domain domainmetadata.Service
}

func NewService() *Service {
	return &Service{
		domain: domainmetadata.NewService(),
	}
}

func (s *Service) CreateDeployPlanAggregate(ctx context.Context, req *vo.CreateDeployPlanAggregateReq) (*vo.DeployPlanAggregateDTO, error) {
	var result *vo.DeployPlanAggregateDTO

	err := dao.Q.Transaction(func(tx *dao.Query) error {
		q := tx.WithContext(ctx)

		project := &model.Project{
			GitID:   req.Project.GitID,
			Name:    req.Project.Name,
			RepoURL: req.Project.RepoURL,
		}
		if err := q.Project.Create(project); err != nil {
			return err
		}

		businessUnit := &model.BusinessUnit{
			Name:        req.BusinessUnit.Name,
			Description: req.BusinessUnit.Description,
			ProjectID:   project.ID,
		}
		if err := q.BusinessUnit.Create(businessUnit); err != nil {
			return err
		}

		ciConfig, err := buildCIConfig(req, businessUnit.ID)
		if err != nil {
			return err
		}
		if err := q.CIConfig.Create(ciConfig); err != nil {
			return err
		}

		cdConfig, err := buildCDConfig(req, businessUnit.ID)
		if err != nil {
			return err
		}
		if err := q.CDConfig.Create(cdConfig); err != nil {
			return err
		}

		instanceConfig, err := buildInstanceConfig(req, businessUnit.ID)
		if err != nil {
			return err
		}
		if err := q.InstanceConfig.Create(instanceConfig); err != nil {
			return err
		}

		deployPlan := &model.DeployPlan{
			Name:             req.DeployPlan.Name,
			Description:      req.DeployPlan.Description,
			BusinessUnitID:   businessUnit.ID,
			CIConfigID:       ciConfig.ID,
			CDConfigID:       cdConfig.ID,
			InstanceConfigID: instanceConfig.ID,
		}
		if err := q.DeployPlan.Create(deployPlan); err != nil {
			return err
		}

		result = aggregateDTO(project, businessUnit, ciConfig, cdConfig, instanceConfig, deployPlan)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
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

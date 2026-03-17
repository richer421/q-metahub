package metadata

import (
	"context"
	"strings"

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

		instanceOAM, err := buildInstanceOAM(req, businessUnit.ID)
		if err != nil {
			return err
		}
		if err := q.InstanceOAM.Create(instanceOAM); err != nil {
			return err
		}

		deployPlan := &model.DeployPlan{
			Name:           req.DeployPlan.Name,
			Description:    req.DeployPlan.Description,
			BusinessUnitID: businessUnit.ID,
			CIConfigID:     ciConfig.ID,
			CDConfigID:     cdConfig.ID,
			InstanceOAMID:  instanceOAM.ID,
		}
		if err := q.DeployPlan.Create(deployPlan); err != nil {
			return err
		}

		result = aggregateDTO(project, businessUnit, ciConfig, cdConfig, instanceOAM, deployPlan)
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

func (s *Service) ListInstanceOAMs(ctx context.Context, businessUnitID int64, env string, keyword string) ([]vo.InstanceOAMDTO, error) {
	query := dao.Q.WithContext(ctx).
		InstanceOAM.
		Where(dao.InstanceOAM.BusinessUnitID.Eq(businessUnitID))

	normalizedEnv := strings.TrimSpace(env)
	if normalizedEnv != "" {
		query = query.Where(dao.InstanceOAM.Env.Eq(normalizedEnv))
	}

	normalizedKeyword := strings.TrimSpace(keyword)
	if normalizedKeyword != "" {
		query = query.Where(dao.InstanceOAM.Name.Like("%" + normalizedKeyword + "%"))
	}

	entities, err := query.Order(dao.InstanceOAM.UpdatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}

	return mapInstanceOAMs(entities), nil
}

func (s *Service) CreateInstanceOAM(ctx context.Context, businessUnitID int64, req *vo.CreateInstanceOAMReq) (*vo.InstanceOAMDTO, error) {
	var oamApplication model.OAMApplication
	if err := convertJSONMap(req.OAMApplication, &oamApplication); err != nil {
		return nil, err
	}

	var frontendPayload model.InstanceOAMPayload
	if err := convertJSONMap(req.FrontendPayload, &frontendPayload); err != nil {
		return nil, err
	}

	schemaVersion := strings.TrimSpace(req.SchemaVersion)
	if schemaVersion == "" {
		schemaVersion = "v1alpha1"
	}

	entity := &model.InstanceOAM{
		Name:            strings.TrimSpace(req.Name),
		BusinessUnitID:  businessUnitID,
		Env:             strings.TrimSpace(req.Env),
		SchemaVersion:   schemaVersion,
		OAMApplication:  oamApplication,
		FrontendPayload: frontendPayload,
	}

	if err := dao.Q.WithContext(ctx).InstanceOAM.Create(entity); err != nil {
		return nil, err
	}

	dto := toInstanceOAMDTO(entity)
	return &dto, nil
}

func (s *Service) UpdateInstanceOAM(ctx context.Context, instanceOAMID int64, req *vo.UpdateInstanceOAMReq) (*vo.InstanceOAMDTO, error) {
	entity, err := dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.ID.Eq(instanceOAMID)).First()
	if err != nil {
		return nil, err
	}

	var oamApplication model.OAMApplication
	if err := convertJSONMap(req.OAMApplication, &oamApplication); err != nil {
		return nil, err
	}

	var frontendPayload model.InstanceOAMPayload
	if err := convertJSONMap(req.FrontendPayload, &frontendPayload); err != nil {
		return nil, err
	}

	schemaVersion := strings.TrimSpace(req.SchemaVersion)
	if schemaVersion == "" {
		schemaVersion = entity.SchemaVersion
	}
	if schemaVersion == "" {
		schemaVersion = "v1alpha1"
	}

	entity.Name = strings.TrimSpace(req.Name)
	entity.Env = strings.TrimSpace(req.Env)
	entity.SchemaVersion = schemaVersion
	entity.OAMApplication = oamApplication
	entity.FrontendPayload = frontendPayload

	if err := dao.Q.WithContext(ctx).InstanceOAM.Save(entity); err != nil {
		return nil, err
	}

	dto := toInstanceOAMDTO(entity)
	return &dto, nil
}

package metadata

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
	"gorm.io/gorm"
)

type deployPlanAggregate struct {
	project      *model.Project
	businessUnit *model.BusinessUnit
	ciConfig     *model.CIConfig
	cdConfig     *model.CDConfig
	instanceOAM  *model.InstanceOAM
	deployPlan   *model.DeployPlan
}

type deployPlanRelatedResources struct {
	ciConfig    *model.CIConfig
	cdConfig    *model.CDConfig
	instanceOAM *model.InstanceOAM
}

type instanceView struct {
	Name string
	Env  string
}

func loadDeployPlanAggregate(ctx context.Context, deployPlanID int64) (*deployPlanAggregate, error) {
	q := dao.Q.WithContext(ctx)

	deployPlan, err := q.DeployPlan.
		Where(dao.DeployPlan.ID.Eq(deployPlanID)).
		First()
	if err != nil {
		return nil, fmt.Errorf("query deploy_plan id=%d: %w", deployPlanID, err)
	}

	businessUnit, err := q.BusinessUnit.
		Where(dao.BusinessUnit.ID.Eq(deployPlan.BusinessUnitID)).
		First()
	if err != nil {
		return nil, fmt.Errorf("query business_unit id=%d: %w", deployPlan.BusinessUnitID, err)
	}

	project, err := q.Project.
		Where(dao.Project.ID.Eq(businessUnit.ProjectID)).
		First()
	if err != nil {
		return nil, fmt.Errorf("query project id=%d: %w", businessUnit.ProjectID, err)
	}

	ciConfig, err := q.CIConfig.
		Where(dao.CIConfig.ID.Eq(deployPlan.CIConfigID)).
		First()
	if err != nil {
		return nil, fmt.Errorf("query ci_config id=%d: %w", deployPlan.CIConfigID, err)
	}

	cdConfig, err := q.CDConfig.
		Where(dao.CDConfig.ID.Eq(deployPlan.CDConfigID)).
		First()
	if err != nil {
		return nil, fmt.Errorf("query cd_config id=%d: %w", deployPlan.CDConfigID, err)
	}

	instanceOAM, err := q.InstanceOAM.
		Where(dao.InstanceOAM.ID.Eq(deployPlan.InstanceOAMID)).
		First()
	if err != nil {
		return nil, fmt.Errorf("query instance_oam id=%d: %w", deployPlan.InstanceOAMID, err)
	}

	return &deployPlanAggregate{
		project:      project,
		businessUnit: businessUnit,
		ciConfig:     ciConfig,
		cdConfig:     cdConfig,
		instanceOAM:  instanceOAM,
		deployPlan:   deployPlan,
	}, nil
}

func (a *deployPlanAggregate) toVO() *vo.DeployPlanAggregateVO {
	return &vo.DeployPlanAggregateVO{
		Project:      toProjectVO(a.project),
		BusinessUnit: toBusinessUnitVO(a.businessUnit),
		CIConfig:     toCIConfigVO(a.ciConfig),
		CDConfig:     toCDConfigVO(a.cdConfig),
		InstanceOAM:  convertToInstOAMVO(*a.instanceOAM),
		DeployPlan:   toDeployPlanVO(a.deployPlan),
	}
}

func toProjectVO(in *model.Project) vo.ProjectVO {
	return vo.ProjectVO{
		ID:        in.ID,
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
		GitID:     in.GitID,
		Name:      in.Name,
		RepoURL:   in.RepoURL,
	}
}

func toBusinessUnitVO(in *model.BusinessUnit) vo.BusinessUnitVO {
	return vo.BusinessUnitVO{
		ID:          in.ID,
		CreatedAt:   in.CreatedAt,
		UpdatedAt:   in.UpdatedAt,
		Name:        in.Name,
		Description: in.Description,
		ProjectID:   in.ProjectID,
	}
}

func toCIConfigVO(in *model.CIConfig) vo.CIConfigVO {
	return vo.CIConfigVO{
		ID:             in.ID,
		CreatedAt:      in.CreatedAt,
		UpdatedAt:      in.UpdatedAt,
		Name:           in.Name,
		BusinessUnitID: in.BusinessUnitID,
		ImageTagRule: vo.CIConfigImageTagRuleVO{
			Type:          in.ImageTagRule.Type,
			Template:      in.ImageTagRule.Template,
			WithTimestamp: in.ImageTagRule.WithTimestamp,
			WithCommit:    in.ImageTagRule.WithCommit,
		},
		BuildSpec: vo.CIConfigBuildSpecVO{
			Branch:         in.BuildSpec.Branch,
			Tag:            in.BuildSpec.Tag,
			CommitID:       in.BuildSpec.CommitID,
			MakefilePath:   defaultString(in.BuildSpec.MakefilePath, "./Makefile"),
			MakeCommand:    normalizeCIBuildCommand(in.BuildSpec.MakeCommand),
			DockerfilePath: defaultString(in.BuildSpec.DockerfilePath, "./Dockerfile"),
			DockerContext:  in.BuildSpec.DockerContext,
			BuildArgs:      in.BuildSpec.BuildArgs,
		},
		DeployPlanRefCount: 0,
	}
}

func toCDConfigVO(in *model.CDConfig) vo.CDConfigVO {
	out := vo.CDConfigVO{
		ID:             in.ID,
		CreatedAt:      in.CreatedAt,
		UpdatedAt:      in.UpdatedAt,
		Name:           in.Name,
		BusinessUnitID: in.BusinessUnitID,
		RenderEngine:   in.RenderEngine,
		ValuesYAML:     in.ValuesYAML,
		ReleaseStrategy: map[string]any{
			"deployment_mode": in.ReleaseStrategy.DeploymentMode,
			"batch_rule": map[string]any{
				"batch_count":  in.ReleaseStrategy.BatchRule.BatchCount,
				"batch_ratio":  in.ReleaseStrategy.BatchRule.BatchRatio,
				"trigger_type": in.ReleaseStrategy.BatchRule.TriggerType,
				"interval":     in.ReleaseStrategy.BatchRule.Interval,
			},
		},
	}
	if in.GitOps != nil {
		out.GitOps = map[string]any{
			"enabled":       in.GitOps.Enabled,
			"repo_url":      in.GitOps.RepoURL,
			"branch":        in.GitOps.Branch,
			"app_root":      in.GitOps.AppRoot,
			"manifest_root": in.GitOps.ManifestRoot,
		}
	}
	return out
}

func toDeployPlanVO(in *model.DeployPlan) vo.DeployPlanVO {
	return vo.DeployPlanVO{
		ID:             in.ID,
		CreatedAt:      in.CreatedAt,
		UpdatedAt:      in.UpdatedAt,
		Name:           in.Name,
		Description:    in.Description,
		BusinessUnitID: in.BusinessUnitID,
		CIConfigID:     in.CIConfigID,
		CDConfigID:     in.CDConfigID,
		InstanceOAMID:  in.InstanceOAMID,
	}
}

func listBusinessUnitDeployPlans(ctx context.Context, businessUnitID int64, page int, pageSize int, keyword string) (*vo.DeployPlanPageDTO, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	query := dao.Q.WithContext(ctx).
		DeployPlan.
		Where(dao.DeployPlan.BusinessUnitID.Eq(businessUnitID))
	if normalizedKeyword := strings.TrimSpace(keyword); normalizedKeyword != "" {
		query = query.Where(dao.DeployPlan.Name.Like("%" + normalizedKeyword + "%"))
	}

	offset := (page - 1) * pageSize
	rows, total, err := query.Order(dao.DeployPlan.UpdatedAt.Desc()).FindByPage(offset, pageSize)
	if err != nil {
		return nil, fmt.Errorf("list deploy plans by business_unit_id=%d: %w", businessUnitID, err)
	}

	ciNameMap, err := loadCIConfigNamesByDeployPlanRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	cdNameMap, err := loadCDConfigNamesByDeployPlanRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	instanceMap, err := loadInstanceViewsByDeployPlanRows(ctx, rows)
	if err != nil {
		return nil, err
	}

	items := make([]vo.DeployPlanFrontendVO, 0, len(rows))
	for _, row := range rows {
		instanceMeta := instanceMap[row.InstanceOAMID]
		items = append(items, vo.DeployPlanFrontendVO{
			CreatedAt:      row.CreatedAt,
			UpdatedAt:      row.UpdatedAt,
			ID:             row.ID,
			BusinessUnitID: row.BusinessUnitID,
			Name:           row.Name,
			Env:            instanceMeta.Env,
			CIConfigName:   ciNameMap[row.CIConfigID],
			CDConfigName:   cdNameMap[row.CDConfigID],
			InstanceName:   instanceMeta.Name,
			LastStatus:     "pending",
			LastTime:       row.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}

	return &vo.DeployPlanPageDTO{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func getDeployPlan(ctx context.Context, deployPlanID int64) (*vo.DeployPlanFrontendVO, error) {
	plan, err := dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.ID.Eq(deployPlanID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("deploy plan not found")
		}
		return nil, err
	}

	relations, err := loadDeployPlanRelatedResources(ctx, plan.BusinessUnitID, plan.CIConfigID, plan.CDConfigID, plan.InstanceOAMID)
	if err != nil {
		return nil, err
	}

	res := toDeployPlanFrontendDetailVO(plan, relations)
	return &res, nil
}

func createBusinessUnitDeployPlan(ctx context.Context, businessUnitID int64, req vo.UpsertDeployPlanReq) (*vo.DeployPlanFrontendVO, error) {
	if err := ensureBusinessUnitExists(ctx, businessUnitID); err != nil {
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if _, err := dao.Q.WithContext(ctx).DeployPlan.
		Where(dao.DeployPlan.BusinessUnitID.Eq(businessUnitID)).
		Where(dao.DeployPlan.Name.Eq(name)).
		First(); err == nil {
		return nil, fmt.Errorf("deploy plan name already exists in current business unit")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	relations, err := loadDeployPlanRelatedResources(ctx, businessUnitID, req.CIConfigID, req.CDConfigID, req.InstanceOAMID)
	if err != nil {
		return nil, err
	}

	entity := &model.DeployPlan{
		Name:           name,
		Description:    strings.TrimSpace(req.Description),
		BusinessUnitID: businessUnitID,
		CIConfigID:     req.CIConfigID,
		CDConfigID:     req.CDConfigID,
		InstanceOAMID:  req.InstanceOAMID,
	}

	if err := dao.Q.WithContext(ctx).DeployPlan.Create(entity); err != nil {
		return nil, err
	}

	res := toDeployPlanFrontendDetailVO(entity, relations)
	return &res, nil
}

func updateDeployPlan(ctx context.Context, deployPlanID int64, req vo.UpsertDeployPlanReq) (*vo.DeployPlanFrontendVO, error) {
	current, err := dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.ID.Eq(deployPlanID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("deploy plan not found")
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}

	if name != current.Name {
		if _, err := dao.Q.WithContext(ctx).DeployPlan.
			Where(dao.DeployPlan.BusinessUnitID.Eq(current.BusinessUnitID)).
			Where(dao.DeployPlan.Name.Eq(name)).
			Not(dao.DeployPlan.ID.Eq(deployPlanID)).
			First(); err == nil {
			return nil, fmt.Errorf("deploy plan name already exists in current business unit")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	relations, err := loadDeployPlanRelatedResources(ctx, current.BusinessUnitID, req.CIConfigID, req.CDConfigID, req.InstanceOAMID)
	if err != nil {
		return nil, err
	}

	current.Name = name
	current.Description = strings.TrimSpace(req.Description)
	current.CIConfigID = req.CIConfigID
	current.CDConfigID = req.CDConfigID
	current.InstanceOAMID = req.InstanceOAMID

	if err := dao.Q.WithContext(ctx).DeployPlan.Save(current); err != nil {
		return nil, err
	}

	res := toDeployPlanFrontendDetailVO(current, relations)
	return &res, nil
}

func deleteDeployPlan(ctx context.Context, deployPlanID int64) error {
	current, err := dao.Q.WithContext(ctx).DeployPlan.Where(dao.DeployPlan.ID.Eq(deployPlanID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("deploy plan not found")
		}
		return err
	}

	_, err = dao.Q.WithContext(ctx).DeployPlan.Delete(current)
	return err
}

func ensureBusinessUnitExists(ctx context.Context, businessUnitID int64) error {
	if _, err := dao.Q.WithContext(ctx).BusinessUnit.Where(dao.BusinessUnit.ID.Eq(businessUnitID)).First(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("business unit not found")
		}
		return err
	}
	return nil
}

func loadDeployPlanRelatedResources(
	ctx context.Context,
	businessUnitID int64,
	ciConfigID int64,
	cdConfigID int64,
	instanceOAMID int64,
) (*deployPlanRelatedResources, error) {
	if ciConfigID <= 0 {
		return nil, fmt.Errorf("ci_config_id is required")
	}
	if cdConfigID <= 0 {
		return nil, fmt.Errorf("cd_config_id is required")
	}
	if instanceOAMID <= 0 {
		return nil, fmt.Errorf("instance_oam_id is required")
	}

	ciConfig, err := dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.ID.Eq(ciConfigID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ci config not found")
		}
		return nil, err
	}
	if ciConfig.BusinessUnitID != businessUnitID {
		return nil, fmt.Errorf("ci config does not belong to current business unit")
	}

	cdConfig, err := dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.ID.Eq(cdConfigID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("cd config not found")
		}
		return nil, err
	}
	if cdConfig.BusinessUnitID != businessUnitID {
		return nil, fmt.Errorf("cd config does not belong to current business unit")
	}

	instanceOAM, err := dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.ID.Eq(instanceOAMID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("instance oam not found")
		}
		return nil, err
	}
	if instanceOAM.BusinessUnitID != businessUnitID {
		return nil, fmt.Errorf("instance oam does not belong to current business unit")
	}

	return &deployPlanRelatedResources{
		ciConfig:    ciConfig,
		cdConfig:    cdConfig,
		instanceOAM: instanceOAM,
	}, nil
}

func toDeployPlanFrontendDetailVO(plan *model.DeployPlan, relations *deployPlanRelatedResources) vo.DeployPlanFrontendVO {
	return vo.DeployPlanFrontendVO{
		CreatedAt:      plan.CreatedAt,
		UpdatedAt:      plan.UpdatedAt,
		ID:             plan.ID,
		BusinessUnitID: plan.BusinessUnitID,
		Name:           plan.Name,
		Description:    plan.Description,
		CIConfigID:     plan.CIConfigID,
		CDConfigID:     plan.CDConfigID,
		InstanceOAMID:  plan.InstanceOAMID,
		Env:            relations.instanceOAM.Env,
		CIConfigName:   relations.ciConfig.Name,
		CDConfigName:   relations.cdConfig.Name,
		InstanceName:   relations.instanceOAM.Name,
		LastStatus:     "pending",
		LastTime:       plan.UpdatedAt.Format(time.RFC3339),
	}
}

func loadCIConfigNamesByDeployPlanRows(ctx context.Context, rows []*model.DeployPlan) (map[int64]string, error) {
	ids := collectDeployPlanForeignKeys(rows, func(row *model.DeployPlan) int64 {
		return row.CIConfigID
	})
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}

	ciRows, err := dao.Q.WithContext(ctx).CIConfig.Where(dao.CIConfig.ID.In(ids...)).Find()
	if err != nil {
		return nil, fmt.Errorf("list ci configs by ids: %w", err)
	}

	result := make(map[int64]string, len(ciRows))
	for _, row := range ciRows {
		result[row.ID] = row.Name
	}
	return result, nil
}

func loadCDConfigNamesByDeployPlanRows(ctx context.Context, rows []*model.DeployPlan) (map[int64]string, error) {
	ids := collectDeployPlanForeignKeys(rows, func(row *model.DeployPlan) int64 {
		return row.CDConfigID
	})
	if len(ids) == 0 {
		return map[int64]string{}, nil
	}

	cdRows, err := dao.Q.WithContext(ctx).CDConfig.Where(dao.CDConfig.ID.In(ids...)).Find()
	if err != nil {
		return nil, fmt.Errorf("list cd configs by ids: %w", err)
	}

	result := make(map[int64]string, len(cdRows))
	for _, row := range cdRows {
		result[row.ID] = row.Name
	}
	return result, nil
}

func loadInstanceViewsByDeployPlanRows(ctx context.Context, rows []*model.DeployPlan) (map[int64]instanceView, error) {
	ids := collectDeployPlanForeignKeys(rows, func(row *model.DeployPlan) int64 {
		return row.InstanceOAMID
	})
	if len(ids) == 0 {
		return map[int64]instanceView{}, nil
	}

	instanceRows, err := dao.Q.WithContext(ctx).InstanceOAM.Where(dao.InstanceOAM.ID.In(ids...)).Find()
	if err != nil {
		return nil, fmt.Errorf("list instance oams by ids: %w", err)
	}

	result := make(map[int64]instanceView, len(instanceRows))
	for _, row := range instanceRows {
		result[row.ID] = instanceView{
			Name: row.Name,
			Env:  row.Env,
		}
	}
	return result, nil
}

func collectDeployPlanForeignKeys(rows []*model.DeployPlan, getter func(*model.DeployPlan) int64) []int64 {
	dedup := make(map[int64]struct{})
	for _, row := range rows {
		if row == nil {
			continue
		}
		id := getter(row)
		if id <= 0 {
			continue
		}
		dedup[id] = struct{}{}
	}

	ids := make([]int64, 0, len(dedup))
	for id := range dedup {
		ids = append(ids, id)
	}
	return ids
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func normalizeCIBuildCommand(command string) string {
	if command == "" || command == "build" {
		return "make build"
	}
	return command
}

package metadata

import (
	"context"
	"fmt"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/dao"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

type deployPlanAggregate struct {
	project      *model.Project
	businessUnit *model.BusinessUnit
	ciConfig     *model.CIConfig
	cdConfig     *model.CDConfig
	instanceOAM  *model.InstanceOAM
	deployPlan   *model.DeployPlan
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
		ImageRegistry:  in.ImageRegistry,
		ImageRepo:      in.ImageRepo,
		FullImageRepo:  in.ImageRegistry + "/" + in.ImageRepo,
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
			MakeCommand:    in.BuildSpec.MakeCommand,
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

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

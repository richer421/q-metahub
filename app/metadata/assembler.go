package metadata

import (
	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

func buildInstanceOAM(req vo.CreateInstanceOAMReq, businessUnitID int64) *model.InstanceOAM {
	form := frontendVOToForm(req.FrontendPayload, req.Name, req.Env)
	oamApplication := normalizeOAMApplication(formToOAM(form), req.Name)

	schemaVersion := req.SchemaVersion
	if schemaVersion == "" {
		schemaVersion = defaultSchemaVersion
	}

	return &model.InstanceOAM{
		Name:           req.Name,
		BusinessUnitID: businessUnitID,
		Env:            req.Env,
		SchemaVersion:  schemaVersion,
		OAMApplication: oamApplication,
	}
}

func aggregateDTO(project *model.Project, businessUnit *model.BusinessUnit, ciConfig *model.CIConfig, cdConfig *model.CDConfig, instanceOAM *model.InstanceOAM, deployPlan *model.DeployPlan) *vo.DeployPlanAggregateDTO {
	return &vo.DeployPlanAggregateDTO{
		Project:      toProjectDTO(project),
		BusinessUnit: toBusinessUnitDTO(businessUnit),
		CIConfig:     toCIConfigDTO(ciConfig),
		CDConfig:     toCDConfigDTO(cdConfig),
		InstanceOAM:  toInstanceOAMDTO(instanceOAM),
		DeployPlan:   toDeployPlanDTO(deployPlan),
	}
}

func toProjectDTO(in *model.Project) vo.ProjectDTO {
	return vo.ProjectDTO{
		ID:        in.ID,
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
		GitID:     in.GitID,
		Name:      in.Name,
		RepoURL:   in.RepoURL,
	}
}

func toBusinessUnitDTO(in *model.BusinessUnit) vo.BusinessUnitDTO {
	return vo.BusinessUnitDTO{
		ID:          in.ID,
		CreatedAt:   in.CreatedAt,
		UpdatedAt:   in.UpdatedAt,
		Name:        in.Name,
		Description: in.Description,
		ProjectID:   in.ProjectID,
	}
}

func toCIConfigDTO(in *model.CIConfig) vo.CIConfigDTO {
	return vo.CIConfigDTO{
		ID:             in.ID,
		CreatedAt:      in.CreatedAt,
		UpdatedAt:      in.UpdatedAt,
		Name:           in.Name,
		BusinessUnitID: in.BusinessUnitID,
		ImageRegistry:  in.ImageRegistry,
		ImageRepo:      in.ImageRepo,
		ImageTagRule: map[string]any{
			"type": in.ImageTagRule.Type,
		},
		BuildSpec: map[string]any{
			"branch":          derefString(in.BuildSpec.Branch),
			"tag":             derefString(in.BuildSpec.Tag),
			"commit_id":       derefString(in.BuildSpec.CommitID),
			"makefile_path":   in.BuildSpec.MakefilePath,
			"make_command":    in.BuildSpec.MakeCommand,
			"dockerfile_path": in.BuildSpec.DockerfilePath,
			"docker_context":  in.BuildSpec.DockerContext,
			"build_args":      in.BuildSpec.BuildArgs,
		},
	}
}

func toCDConfigDTO(in *model.CDConfig) vo.CDConfigDTO {
	dto := vo.CDConfigDTO{
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
		dto.GitOps = map[string]any{
			"enabled":       in.GitOps.Enabled,
			"repo_url":      in.GitOps.RepoURL,
			"branch":        in.GitOps.Branch,
			"app_root":      in.GitOps.AppRoot,
			"manifest_root": in.GitOps.ManifestRoot,
		}
	}
	return dto
}

func toInstanceOAMDTO(in *model.InstanceOAM) vo.InstanceOAMDTO {
	frontendPayload := formToFrontendVO(oamToForm(in.OAMApplication, in.Name, in.Env))
	return vo.InstanceOAMDTO{
		ID:              in.ID,
		CreatedAt:       in.CreatedAt,
		UpdatedAt:       in.UpdatedAt,
		Name:            in.Name,
		BusinessUnitID:  in.BusinessUnitID,
		Env:             in.Env,
		SchemaVersion:   in.SchemaVersion,
		OAMApplication:  in.OAMApplication,
		FrontendPayload: frontendPayload,
	}
}

func toDeployPlanDTO(in *model.DeployPlan) vo.DeployPlanDTO {
	return vo.DeployPlanDTO{
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

func mapCIConfigs(items []*model.CIConfig) []vo.CIConfigDTO {
	out := make([]vo.CIConfigDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toCIConfigDTO(item))
	}
	return out
}

func mapCDConfigs(items []*model.CDConfig) []vo.CDConfigDTO {
	out := make([]vo.CDConfigDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toCDConfigDTO(item))
	}
	return out
}

func mapInstanceOAMs(items []*model.InstanceOAM) []vo.InstanceOAMDTO {
	out := make([]vo.InstanceOAMDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toInstanceOAMDTO(item))
	}
	return out
}

func mapDeployPlans(items []*model.DeployPlan) []vo.DeployPlanDTO {
	out := make([]vo.DeployPlanDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toDeployPlanDTO(item))
	}
	return out
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

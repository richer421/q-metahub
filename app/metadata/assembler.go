package metadata

import (
	"encoding/json"

	"github.com/richer421/q-metahub/app/metadata/vo"
	"github.com/richer421/q-metahub/infra/mysql/model"
)

func buildCIConfig(req *vo.CreateDeployPlanAggregateReq, businessUnitID int64) (*model.CIConfig, error) {
	var imageTagRule model.ImageTagRule
	if err := convertJSONMap(req.CIConfig.ImageTagRule, &imageTagRule); err != nil {
		return nil, err
	}
	var buildSpec model.BuildSpec
	if err := convertJSONMap(req.CIConfig.BuildSpec, &buildSpec); err != nil {
		return nil, err
	}
	return &model.CIConfig{
		Name:           req.CIConfig.Name,
		BusinessUnitID: businessUnitID,
		ImageRegistry:  req.CIConfig.ImageRegistry,
		ImageRepo:      req.CIConfig.ImageRepo,
		ImageTagRule:   imageTagRule,
		BuildSpec:      buildSpec,
	}, nil
}

func buildCDConfig(req *vo.CreateDeployPlanAggregateReq, businessUnitID int64) (*model.CDConfig, error) {
	var releaseStrategy model.ReleaseStrategy
	if err := convertJSONMap(req.CDConfig.ReleaseStrategy, &releaseStrategy); err != nil {
		return nil, err
	}
	var gitOps model.GitOpsConfig
	if len(req.CDConfig.GitOps) > 0 {
		if err := convertJSONMap(req.CDConfig.GitOps, &gitOps); err != nil {
			return nil, err
		}
	}
	return &model.CDConfig{
		Name:            req.CDConfig.Name,
		BusinessUnitID:  businessUnitID,
		RenderEngine:    req.CDConfig.RenderEngine,
		ValuesYAML:      req.CDConfig.ValuesYAML,
		ReleaseStrategy: releaseStrategy,
		GitOps:          &gitOps,
	}, nil
}

func buildInstanceConfig(req *vo.CreateDeployPlanAggregateReq, businessUnitID int64) (*model.InstanceOAM, error) {
	var oamApplication model.OAMApplication
	if err := convertJSONMap(req.InstanceConfig.OAMApplication, &oamApplication); err != nil {
		return nil, err
	}
	var frontendPayload model.InstanceOAMPayload
	if err := convertJSONMap(req.InstanceConfig.FrontendPayload, &frontendPayload); err != nil {
		return nil, err
	}

	schemaVersion := req.InstanceConfig.SchemaVersion
	if schemaVersion == "" {
		schemaVersion = "v1alpha1"
	}

	return &model.InstanceOAM{
		Name:            req.InstanceConfig.Name,
		BusinessUnitID:  businessUnitID,
		Env:             req.InstanceConfig.Env,
		SchemaVersion:   schemaVersion,
		OAMApplication:  oamApplication,
		FrontendPayload: frontendPayload,
	}, nil
}

func convertJSONMap(input map[string]any, target any) error {
	if input == nil {
		return nil
	}
	data, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func aggregateDTO(project *model.Project, businessUnit *model.BusinessUnit, ciConfig *model.CIConfig, cdConfig *model.CDConfig, instanceConfig *model.InstanceOAM, deployPlan *model.DeployPlan) *vo.DeployPlanAggregateDTO {
	return &vo.DeployPlanAggregateDTO{
		Project:        toProjectDTO(project),
		BusinessUnit:   toBusinessUnitDTO(businessUnit),
		CIConfig:       toCIConfigDTO(ciConfig),
		CDConfig:       toCDConfigDTO(cdConfig),
		InstanceConfig: toInstanceConfigDTO(instanceConfig),
		DeployPlan:     toDeployPlanDTO(deployPlan),
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

func toInstanceConfigDTO(in *model.InstanceOAM) vo.InstanceConfigDTO {
	return vo.InstanceConfigDTO{
		ID:              in.ID,
		CreatedAt:       in.CreatedAt,
		UpdatedAt:       in.UpdatedAt,
		Name:            in.Name,
		BusinessUnitID:  in.BusinessUnitID,
		Env:             in.Env,
		SchemaVersion:   in.SchemaVersion,
		OAMApplication:  modelToMap(in.OAMApplication),
		FrontendPayload: modelToMap(in.FrontendPayload),
	}
}

func toDeployPlanDTO(in *model.DeployPlan) vo.DeployPlanDTO {
	return vo.DeployPlanDTO{
		ID:               in.ID,
		CreatedAt:        in.CreatedAt,
		UpdatedAt:        in.UpdatedAt,
		Name:             in.Name,
		Description:      in.Description,
		BusinessUnitID:   in.BusinessUnitID,
		CIConfigID:       in.CIConfigID,
		CDConfigID:       in.CDConfigID,
		InstanceConfigID: in.InstanceConfigID,
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

func mapInstanceConfigs(items []*model.InstanceOAM) []vo.InstanceConfigDTO {
	out := make([]vo.InstanceConfigDTO, 0, len(items))
	for _, item := range items {
		out = append(out, toInstanceConfigDTO(item))
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

func modelToMap(v any) map[string]any {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

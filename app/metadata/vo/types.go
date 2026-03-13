package vo

import "time"

type CreateProjectReq struct {
	GitID   int64  `json:"git_id"`
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
}

type CreateBusinessUnitReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateCIConfigReq struct {
	Name          string         `json:"name"`
	ImageRegistry string         `json:"image_registry"`
	ImageRepo     string         `json:"image_repo"`
	ImageTagRule  map[string]any `json:"image_tag_rule"`
	BuildSpec     map[string]any `json:"build_spec"`
}

type CreateCDConfigReq struct {
	Name            string         `json:"name"`
	RenderEngine    string         `json:"render_engine"`
	ValuesYAML      string         `json:"values_yaml"`
	ReleaseStrategy map[string]any `json:"release_strategy"`
	GitOps          map[string]any `json:"git_ops"`
}

type CreateInstanceConfigReq struct {
	Name            string         `json:"name"`
	Env             string         `json:"env"`
	InstanceType    string         `json:"instance_type"`
	Spec            map[string]any `json:"spec"`
	AttachResources map[string]any `json:"attach_resources"`
}

type CreateDeployPlanReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateDeployPlanAggregateReq struct {
	Project        CreateProjectReq        `json:"project"`
	BusinessUnit   CreateBusinessUnitReq   `json:"business_unit"`
	CIConfig       CreateCIConfigReq       `json:"ci_config"`
	CDConfig       CreateCDConfigReq       `json:"cd_config"`
	InstanceConfig CreateInstanceConfigReq `json:"instance_config"`
	DeployPlan     CreateDeployPlanReq     `json:"deploy_plan"`
}

type ProjectDTO struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	GitID     int64     `json:"git_id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
}

type BusinessUnitDTO struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ProjectID   int64     `json:"project_id"`
}

type CIConfigDTO struct {
	ID             int64          `json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Name           string         `json:"name"`
	BusinessUnitID int64          `json:"business_unit_id"`
	ImageRegistry  string         `json:"image_registry"`
	ImageRepo      string         `json:"image_repo"`
	ImageTagRule   map[string]any `json:"image_tag_rule"`
	BuildSpec      map[string]any `json:"build_spec"`
}

type CDConfigDTO struct {
	ID              int64          `json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Name            string         `json:"name"`
	BusinessUnitID  int64          `json:"business_unit_id"`
	RenderEngine    string         `json:"render_engine"`
	ValuesYAML      string         `json:"values_yaml"`
	ReleaseStrategy map[string]any `json:"release_strategy"`
	GitOps          map[string]any `json:"git_ops,omitempty"`
}

type InstanceConfigDTO struct {
	ID              int64          `json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Name            string         `json:"name"`
	BusinessUnitID  int64          `json:"business_unit_id"`
	Env             string         `json:"env"`
	InstanceType    string         `json:"instance_type"`
	Spec            map[string]any `json:"spec"`
	AttachResources map[string]any `json:"attach_resources"`
}

type DeployPlanDTO struct {
	ID               int64     `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	BusinessUnitID   int64     `json:"business_unit_id"`
	CIConfigID       int64     `json:"ci_config_id"`
	CDConfigID       int64     `json:"cd_config_id"`
	InstanceConfigID int64     `json:"instance_config_id"`
}

type DeployPlanAggregateDTO struct {
	Project        ProjectDTO        `json:"project"`
	BusinessUnit   BusinessUnitDTO   `json:"business_unit"`
	CIConfig       CIConfigDTO       `json:"ci_config"`
	CDConfig       CDConfigDTO       `json:"cd_config"`
	InstanceConfig InstanceConfigDTO `json:"instance_config"`
	DeployPlan     DeployPlanDTO     `json:"deploy_plan"`
}

type BusinessUnitFullSpecDTO struct {
	Project         ProjectDTO          `json:"project"`
	BusinessUnit    BusinessUnitDTO     `json:"business_unit"`
	CIConfigs       []CIConfigDTO       `json:"ci_configs"`
	CDConfigs       []CDConfigDTO       `json:"cd_configs"`
	InstanceConfigs []InstanceConfigDTO `json:"instance_configs"`
	DeployPlans     []DeployPlanDTO     `json:"deploy_plans"`
}

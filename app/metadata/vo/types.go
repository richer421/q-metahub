package vo

import (
	"time"

	"github.com/richer421/q-metahub/infra/mysql/model"
)

type CreateInstanceOAMReq struct {
	BusinessUnitID  int64                     `json:"business_unit_id"`
	InstanceName    string                    `json:"instance_name"`
	Env             string                    `json:"env"`
	SchemaVersion   string                    `json:"schema_version"`
	FrontendPayload InstanceFrontendPayloadVO `json:"frontend_payload"`
}

// InstanceFrontendPayloadVO is UI-facing form schema.
type InstanceFrontendPayloadVO struct {
	Basic    InstanceBasicVO    `json:"basic,omitempty"`
	Extended InstanceExtendedVO `json:"extended,omitempty"`
	Advanced InstanceAdvancedVO `json:"advanced,omitempty"`
}

type InstanceBasicVO struct {
	Name string `json:"name,omitempty"`
	Env  string `json:"env,omitempty"`
}

type InstanceExtendedVO struct {
	NetworkMode string `json:"network_mode,omitempty"`
	Ports       []int  `json:"ports,omitempty"`
}

type InstanceAdvancedVO struct {
	Replicas *int32 `json:"replicas,omitempty"`
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

type InstanceOAMDTO struct {
	ID              int64                     `json:"id"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	Name            string                    `json:"name"`
	BusinessUnitID  int64                     `json:"business_unit_id"`
	Env             string                    `json:"env"`
	SchemaVersion   string                    `json:"schema_version"`
	OAMApplication  model.OAMApplication      `json:"oam_application"`
	FrontendPayload InstanceFrontendPayloadVO `json:"frontend_payload"`
}

type DeployPlanDTO struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	BusinessUnitID int64     `json:"business_unit_id"`
	CIConfigID     int64     `json:"ci_config_id"`
	CDConfigID     int64     `json:"cd_config_id"`
	InstanceOAMID  int64     `json:"instance_oam_id"`
}

type DeployPlanAggregateDTO struct {
	Project      ProjectDTO      `json:"project"`
	BusinessUnit BusinessUnitDTO `json:"business_unit"`
	CIConfig     CIConfigDTO     `json:"ci_config"`
	CDConfig     CDConfigDTO     `json:"cd_config"`
	InstanceOAM  InstanceOAMDTO  `json:"instance_oam"`
	DeployPlan   DeployPlanDTO   `json:"deploy_plan"`
}

type BusinessUnitFullSpecDTO struct {
	Project      ProjectDTO       `json:"project"`
	BusinessUnit BusinessUnitDTO  `json:"business_unit"`
	CIConfigs    []CIConfigDTO    `json:"ci_configs"`
	CDConfigs    []CDConfigDTO    `json:"cd_configs"`
	InstanceOAMs []InstanceOAMDTO `json:"instance_oams"`
	DeployPlans  []DeployPlanDTO  `json:"deploy_plans"`
}

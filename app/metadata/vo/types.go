package vo

import (
	"time"

	"github.com/richer421/q-metahub/infra/mysql/model"
)

type CreateInstanceOAMReq struct {
	BusinessUnitID  int64                     `json:"business_unit_id"`
	InstanceName    string                    `json:"instance_name"`
	Env             string                    `json:"env"`
	FrontendPayload InstanceFrontendPayloadVO `json:"frontend_payload"`
}

type CreateInstanceOAMFromTemplateReq struct {
	Name        string `json:"name"`
	Env         string `json:"env"`
	TemplateKey string `json:"template_key"`
}

// InstanceFrontendPayloadVO is UI-facing form schema.
type InstanceFrontendPayloadVO struct {
	Basic    InstanceBasicVO     `json:"basic"`
	Extended *InstanceExtendedVO `json:"extended,omitempty"`
}

type InstanceBasicVO struct {
	Replicas  *int32              `json:"replicas,omitempty"`
	Container InstanceContainerVO `json:"container"` // 业务容器的基础配置
}

type InstanceContainerVO struct {
	Name  string  `json:"name,omitempty"`
	Image string  `json:"image,omitempty"`
	Ports []int32 `json:"ports,omitempty"`
}

type InstanceExtendedVO struct {
	NetworkMode  string  `json:"network_mode,omitempty"`
	ServicePorts []int32 `json:"service_ports,omitempty"`
}

type ProjectVO struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	GitID     int64     `json:"git_id"`
	Name      string    `json:"name"`
	RepoURL   string    `json:"repo_url"`
}

type BusinessUnitVO struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ProjectID   int64     `json:"project_id"`
}

type CIConfigVO struct {
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

type CDConfigVO struct {
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

type CDConfigListReq struct {
	Page           int    `form:"page"`
	PageSize       int    `form:"page_size"`
	Keyword        string `form:"keyword"`
	ReleaseRegion  string `form:"release_region"`
	ReleaseEnv     string `form:"release_env"`
	DeploymentMode string `form:"deployment_mode"`
}

type UpsertCDConfigReq struct {
	Name                 string    `json:"name"`
	ReleaseRegion        string    `json:"release_region"`
	ReleaseEnv           string    `json:"release_env"`
	DeploymentMode       string    `json:"deployment_mode"`
	TrafficBatchCount    *int      `json:"traffic_batch_count,omitempty"`
	TrafficRatioList     []float64 `json:"traffic_ratio_list,omitempty"`
	ManualAdjust         *bool     `json:"manual_adjust,omitempty"`
	AdjustTimeoutSeconds *int      `json:"adjust_timeout_seconds,omitempty"`
}

type CDConfigFrontendVO struct {
	ID                   int64     `json:"id"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Name                 string    `json:"name"`
	BusinessUnitID       int64     `json:"business_unit_id"`
	ReleaseRegion        string    `json:"release_region"`
	ReleaseEnv           string    `json:"release_env"`
	DeploymentMode       string    `json:"deployment_mode"`
	StrategySummary      string    `json:"strategy_summary"`
	TrafficBatchCount    *int      `json:"traffic_batch_count,omitempty"`
	TrafficRatioList     []float64 `json:"traffic_ratio_list,omitempty"`
	ManualAdjust         *bool     `json:"manual_adjust,omitempty"`
	AdjustTimeoutSeconds *int      `json:"adjust_timeout_seconds,omitempty"`
}

type CDConfigPageDTO struct {
	Items    []CDConfigFrontendVO `json:"items"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

type InstanceOAMVO struct {
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

type InstanceOAMDTO struct {
	ID              int64          `json:"id"`
	BusinessUnitID  int64          `json:"business_unit_id"`
	Name            string         `json:"name"`
	Env             string         `json:"env"`
	SchemaVersion   string         `json:"schema_version"`
	OAMApplication  map[string]any `json:"oam_application"`
	FrontendPayload map[string]any `json:"frontend_payload"`
}

type InstanceOAMPageDTO struct {
	Items    []InstanceOAMDTO `json:"items"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

type InstanceOAMTemplateDTO struct {
	Key           string `json:"key"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Replicas      int32  `json:"replicas"`
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

type UpdateInstanceOAMReq struct {
	Name            string         `json:"name"`
	Env             string         `json:"env"`
	SchemaVersion   string         `json:"schema_version"`
	OAMApplication  map[string]any `json:"oam_application"`
	FrontendPayload map[string]any `json:"frontend_payload"`
}

type DeployPlanVO struct {
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

type DeployPlanAggregateVO struct {
	Project      ProjectVO      `json:"project"`
	BusinessUnit BusinessUnitVO `json:"business_unit"`
	CIConfig     CIConfigVO     `json:"ci_config"`
	CDConfig     CDConfigVO     `json:"cd_config"`
	InstanceOAM  InstanceOAMVO  `json:"instance_oam"`
	DeployPlan   DeployPlanVO   `json:"deploy_plan"`
}

type BusinessUnitFullSpecVO struct {
	Project      ProjectVO       `json:"project"`
	BusinessUnit BusinessUnitVO  `json:"business_unit"`
	CIConfigs    []CIConfigVO    `json:"ci_configs"`
	CDConfigs    []CDConfigVO    `json:"cd_configs"`
	InstanceOAMs []InstanceOAMVO `json:"instance_oams"`
	DeployPlans  []DeployPlanVO  `json:"deploy_plans"`
}

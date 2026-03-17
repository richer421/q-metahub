package oam

// DeployPlanSpecDTO defines the stable open model used by downstream services
// (e.g. q-deploy) when reading deploy plan full-spec from q-metahub.
type DeployPlanSpecDTO struct {
	Project      ProjectDTO      `json:"project"`
	BusinessUnit BusinessUnitDTO `json:"business_unit"`
	CDConfig     CDConfigDTO     `json:"cd_config"`
	InstanceOAM  InstanceOAMDTO  `json:"instance_oam"`
	DeployPlan   DeployPlanDTO   `json:"deploy_plan"`
}

type ProjectDTO struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	RepoURL string `json:"repo_url"`
}

type BusinessUnitDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CDConfigDTO struct {
	ID              int64          `json:"id"`
	GitOps          map[string]any `json:"git_ops"`
	ReleaseStrategy map[string]any `json:"release_strategy"`
}

type DeployPlanDTO struct {
	ID            int64 `json:"id"`
	CDConfigID    int64 `json:"cd_config_id"`
	InstanceOAMID int64 `json:"instance_oam_id"`
}

type InstanceOAMDTO struct {
	ID             int64          `json:"id"`
	Env            string         `json:"env"`
	SchemaVersion  string         `json:"schema_version"`
	OAMApplication OAMApplication `json:"oam_application"`
}

type OAMApplication struct {
	APIVersion string          `json:"apiVersion"`
	Kind       string          `json:"kind"`
	Metadata   *OAMObjectMeta  `json:"metadata,omitempty"`
	Component  OAMPodComponent `json:"component"`
	Traits     *OAMTraits      `json:"traits,omitempty"`
}

type OAMObjectMeta struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type OAMPodComponent struct {
	Name       string           `json:"name"`
	Type       string           `json:"type"`
	Properties OAMPodProperties `json:"properties"`
}

type OAMPodProperties struct {
	MainContainer MainContainer `json:"mainContainer"`
}

type MainContainer struct {
	Container
}

type Container struct {
	Name      string         `json:"name"`
	Image     string         `json:"image,omitempty"`
	Command   string         `json:"command,omitempty"`
	Args      []string       `json:"args,omitempty"`
	Env       []OAMEnvVar    `json:"env,omitempty"`
	Ports     []int32        `json:"ports,omitempty"`
	Resources *ResourceQuota `json:"resources,omitempty"`
}

type ResourceQuota struct {
	Cpu    *CpuQuota    `json:"cpu,omitempty"`
	Memory *MemoryQuota `json:"memory,omitempty"`
}

type CpuQuota struct {
	Request string `json:"request"`
	Limit   string `json:"limit"`
}

type MemoryQuota struct {
	Request string `json:"request"`
	Limit   string `json:"limit"`
}

type OAMEnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type OAMTraits struct {
	Network *NetworkTrait `json:"network,omitempty"`
	Scaling *ScalingTrait `json:"scaling,omitempty"`
}

type ScalingTrait struct {
	Replicas int32 `json:"replicas"`
}

type NetworkTrait struct {
	Type            string           `json:"type"`
	K8sServiceTrait *K8sServiceTrait `json:"k8sServiceTrait,omitempty"`
}

type K8sServiceTrait struct {
	Ports []int `json:"ports,omitempty"`
}

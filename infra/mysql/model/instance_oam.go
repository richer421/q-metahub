package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// InstanceOAM 实例配置（OAM 前端交互模型）
// 该模型用于存储前端可视化编辑后的 OAM 数据。
type InstanceOAM struct {
	BaseModel
	Name           string         `gorm:"column:name;type:varchar(128);not null;uniqueIndex:uk_instance_oam_identity" json:"name"`
	BusinessUnitID int64          `gorm:"column:business_unit_id;not null;index;uniqueIndex:uk_instance_oam_identity" json:"business_unit_id"`
	Env            string         `gorm:"column:env;type:varchar(32);not null;index;uniqueIndex:uk_instance_oam_identity" json:"env"` // dev/test/gray/prod
	SchemaVersion  string         `gorm:"column:schema_version;type:varchar(32);not null;default:v1alpha1" json:"schema_version"`
	OAMApplication OAMApplication `gorm:"column:oam_application;type:json;not null" json:"oam_application"` // OAM-Lite: 单组件+分类traits
}

func (InstanceOAM) TableName() string {
	return "instance_oams"
}

// OAMApplication 采用 OAM-Lite 模型：
// 仅保留单组件(component) + 分类 traits，去掉 policy/workflow 复杂度。
type OAMApplication struct {
	APIVersion string          `json:"apiVersion"` // q.oam/v1alpha1
	Kind       string          `json:"kind"`       // InstanceApplication
	Metadata   *OAMObjectMeta  `json:"metadata,omitempty"`
	Component  OAMPodComponent `json:"component"`        // 单组件
	Traits     *OAMTraits      `json:"traits,omitempty"` // 分类 trait
}

type OAMObjectMeta struct {
	Name        string            `json:"name,omitempty"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type OAMPodComponent struct {
	Name       string           `json:"name"`       // 组件名
	Type       string           `json:"type"`       // 固定：pod
	Properties OAMPodProperties `json:"properties"` // 这里只聚焦基础配置
}

const OAMComponentTypePod = "pod"

type OAMPodProperties struct {
	MainContainer MainContainer `json:"mainContainer"` // 主容器
}

type MainContainer struct {
	Container
}

type ContainerTrait struct {
	Container
}

type Container struct {
	Name      string         `json:"name"`
	Image     string         `json:"image,omitempty"` // 前端不输入，后端可写 IMAGE 占位符
	Command   string         `json:"command,omitempty"`
	Args      []string       `json:"args,omitempty"`
	Env       []OAMEnvVar    `json:"env,omitempty"`
	Ports     []int32        `json:"ports,omitempty"`
	Resources *ResourceQuota `json:"resources,omitempty"` // 示例: cpuLimit/memLimit/cpuRequest/memRequest
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

// OAMTraits 以能力域分类，贴合前端“网络/存储/配置”等分层设计。
type OAMTraits struct {
	Network        *NetworkTrait    `json:"network,omitempty"`        // 网络特性（k8s service/apisix 等）
	Scaling        *ScalingTrait    `json:"scaling,omitempty"`        // 弹性伸缩（副本数）
	Storage        *StorageTrait    `json:"storage,omitempty"`        // 存储特性（volume/mount）
	Config         *ConfigTrait     `json:"config,omitempty"`         // 配置特性（env/configmap/secret）
	Sidecars       []ContainerTrait `json:"sidecars,omitempty"`       // 副容器
	InitContainers []ContainerTrait `json:"initContainers,omitempty"` // init 容器
	Extensions     map[string]any   `json:"extensions,omitempty"`     // 扩展特性，预留
}

type ScalingTrait struct {
	Replicas int32 `json:"replicas"`
}

type StorageTrait struct{}

type ConfigTrait struct{}

type NetworkTrait struct {
	Type            string           `json:"type"`
	K8sServiceTrait *K8sServiceTrait `json:"k8sServiceTrait,omitempty"`
	ApiSixTrait     *ApiSixTrait     `json:"apiSixTrait,omitempty"`
}

type K8sServiceTrait struct {
	Ports []int `json:"ports,omitempty"`
}

type ApiSixTrait struct {
	// 预留
}

// InstanceOAMPayload 存储前端编辑态结构（基础/扩展/高级），用于无损回显。
func (o OAMApplication) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *OAMApplication) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan OAMApplication: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, o)
}

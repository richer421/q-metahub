package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// InstanceConfig 实例配置 - 运行态配置
// 前端配置本质是 YAML 的 UI 化，这里直接存 K8s 原生结构
type InstanceConfig struct {
	BaseModel
	Name            string                 `gorm:"column:name;type:varchar(64);not null" json:"name"`
	BusinessUnitID  int64                  `gorm:"column:business_unit_id;not null;index" json:"business_unit_id"`          // 关联业务单元
	Env             string                 `gorm:"column:env;type:varchar(32);not null;index" json:"env"`               // dev/test/gray/prod
	InstanceType    string                 `gorm:"column:instance_type;type:varchar(32);not null" json:"instance_type"` // deployment/statefulset/job/cronjob/pod
	Spec            InstanceSpec           `gorm:"column:spec;type:json;not null" json:"spec"`                          // K8s 原生工作负载 Spec
	AttachResources InstanceAttachResources `gorm:"column:attach_resources;type:json;default:'{}'" json:"attach_resources"` // 配套附加资源（ConfigMap/Secret/Service）
}

func (InstanceConfig) TableName() string {
	return "instance_configs"
}

// InstanceSpec 实例规格，包装 K8s 原生工作负载 Spec
// 根据 InstanceType 不同，存储对应的 K8s Spec
type InstanceSpec struct {
	Deployment  *appsv1.DeploymentSpec  `json:"deployment,omitempty"`
	StatefulSet *appsv1.StatefulSetSpec `json:"statefulSet,omitempty"`
	Job         *batchv1.JobSpec        `json:"job,omitempty"`
	CronJob     *batchv1.CronJobSpec    `json:"cronJob,omitempty"`
	Pod         *corev1.PodSpec         `json:"pod,omitempty"` // 直接运行 Pod，无控制器
}

// InstanceAttachResources 实例附加资源
// 核心：用 Map 替代切片，按名称索引，简洁且易查询
type InstanceAttachResources struct {
	ConfigMaps map[string]corev1.ConfigMap `json:"configMaps,omitempty"` // 配置字典
	Secrets    map[string]corev1.Secret    `json:"secrets,omitempty"`    // 密钥
	Services   map[string]corev1.Service   `json:"services,omitempty"`   // 服务
}

// ========== InstanceSpec 数据库序列化/反序列化 ==========

func (s InstanceSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *InstanceSpec) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan InstanceSpec: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}

// ========== InstanceAttachResources 数据库序列化/反序列化 ==========

func (ar InstanceAttachResources) Value() (driver.Value, error) {
	return json.Marshal(ar)
}

func (ar *InstanceAttachResources) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan InstanceAttachResources: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, ar)
}

// ========== 便捷方法 ==========

// GetPodSpec 获取底层 PodSpec（所有工作负载类型都有 PodSpec）
func (s *InstanceSpec) GetPodSpec() *corev1.PodSpec {
	switch {
	case s.Deployment != nil:
		return &s.Deployment.Template.Spec
	case s.StatefulSet != nil:
		return &s.StatefulSet.Template.Spec
	case s.Job != nil:
		return &s.Job.Template.Spec
	case s.CronJob != nil:
		return &s.CronJob.JobTemplate.Spec.Template.Spec
	case s.Pod != nil:
		return s.Pod
	default:
		return nil
	}
}

// GetConfigMap 根据名称获取关联的 ConfigMap
func (ar *InstanceAttachResources) GetConfigMap(name string) *corev1.ConfigMap {
	if ar.ConfigMaps == nil {
		return nil
	}
	cm, ok := ar.ConfigMaps[name]
	if !ok {
		return nil
	}
	return &cm
}

// GetSecret 根据名称获取关联的 Secret
func (ar *InstanceAttachResources) GetSecret(name string) *corev1.Secret {
	if ar.Secrets == nil {
		return nil
	}
	secret, ok := ar.Secrets[name]
	if !ok {
		return nil
	}
	return &secret
}

// GetService 根据名称获取关联的 Service
func (ar *InstanceAttachResources) GetService(name string) *corev1.Service {
	if ar.Services == nil {
		return nil
	}
	svc, ok := ar.Services[name]
	if !ok {
		return nil
	}
	return &svc
}

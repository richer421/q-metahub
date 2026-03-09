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
// 前端配置本质是 yaml 的 UI 化，这里直接存 K8s 原生结构
type InstanceConfig struct {
	BaseModel
	Name         string        `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Env          string        `gorm:"column:env;type:varchar(32);not null;index" json:"env"` // dev/test/gray/prod
	ClusterID    int64         `gorm:"column:cluster_id;not null;index" json:"cluster_id"`
	InstanceType string        `gorm:"column:instance_type;type:varchar(32);not null" json:"instance_type"` // deployment/statefulset/job/cronjob
	Spec         InstanceSpec  `gorm:"column:spec;type:json;not null" json:"spec"`                         // K8s 原生 Spec
}

func (InstanceConfig) TableName() string {
	return "instance_configs"
}

// InstanceSpec 实例规格，包装 K8s 原生结构
// 根据 InstanceType 不同，存储对应的 K8s Spec
type InstanceSpec struct {
	Deployment  *appsv1.DeploymentSpec  `json:"deployment,omitempty"`
	StatefulSet *appsv1.StatefulSetSpec `json:"statefulSet,omitempty"`
	Job         *batchv1.JobSpec        `json:"job,omitempty"`
	CronJob     *batchv1.CronJobSpec    `json:"cronJob,omitempty"`
	Pod         *corev1.PodSpec         `json:"pod,omitempty"` // 直接运行 Pod，无控制器
}

// Value 实现 driver.Valuer 接口
func (s InstanceSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
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

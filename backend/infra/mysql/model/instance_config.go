package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// InstanceConfig 实例配置 - 运行态配置
type InstanceConfig struct {
	BaseModel
	Name         string        `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Env          string        `gorm:"column:env;type:varchar(32);not null;index" json:"env"` // dev/test/gray/prod
	ClusterID    int64         `gorm:"column:cluster_id;not null;index" json:"cluster_id"`
	InstanceType string        `gorm:"column:instance_type;type:varchar(32);not null" json:"instance_type"` // deployment/statefulset/job
	Replicas     int           `gorm:"column:replicas;default:1" json:"replicas"`
	ResourceQuota ResourceQuota `gorm:"column:resource_quota;type:json;not null" json:"resource_quota"`
}

func (InstanceConfig) TableName() string {
	return "instance_configs"
}

// ResourceQuota 资源配额 - CPU和内存
type ResourceQuota struct {
	CPU    string `json:"cpu"`    // CPU资源配额，如"500m"、"1"等
	Memory string `json:"memory"` // 内存资源配额，如"256Mi"、"1Gi"等
}

// Value 实现 driver.Valuer 接口，用于将结构体转换为数据库值
func (r ResourceQuota) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan 实现 sql.Scanner 接口，用于将数据库值转换为结构体
func (r *ResourceQuota) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ResourceQuota: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, r)
}

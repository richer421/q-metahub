package model

// InstanceConfig 实例配置 - 运行态配置
type InstanceConfig struct {
	BaseModel
	EnvironmentID   uint `gorm:"column:environment_id;not null;index" json:"environment_id"`
	ResourceQuotaID uint `gorm:"column:resource_quota_id;not null" json:"resource_quota_id"`
}

func (InstanceConfig) TableName() string {
	return "instance_configs"
}

package model

// CDConfig CD配置 - 部署相关配置
type CDConfig struct {
	BaseModel
	TriggerType  string `gorm:"column:trigger_type;type:varchar(32);not null" json:"trigger_type"`   // manual/auto
	DeployEngine string `gorm:"column:deploy_engine;type:varchar(32);not null" json:"deploy_engine"` // helm/kustomize/custom
}

// CanaryConfig 金丝雀发布配置
type CanaryConfig struct {
	BaseModel
	CDConfigID     int64   `gorm:"column:cd_config_id;not null;index" json:"cd_config_id"`
	Weight         int32   `gorm:"column:weight;default:0" json:"weight"`                    // 流量权重 0-100
	MaxUnavailable string  `gorm:"column:max_unavailable;type:varchar(32)" json:"max_unavailable"` // 最大不可用比例
}

func (CDConfig) TableName() string {
	return "cd_configs"
}

func (CanaryConfig) TableName() string {
	return "canary_configs"
}

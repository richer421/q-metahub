package model

// DeployPlan 部署计划 - 完整配置包，聚合CI/CD/实例配置
// 核心语义
type DeployPlan struct {
	BaseModel
	BusinessUnitID   uint   `gorm:"column:business_unit_id;not null;index" json:"business_unit_id"`
	CIConfigID       uint   `gorm:"column:ci_config_id;not null" json:"ci_config_id"`
	CDConfigID       uint   `gorm:"column:cd_config_id;not null" json:"cd_config_id"`
	InstanceConfigID uint   `gorm:"column:instance_config_id;not null" json:"instance_config_id"`
	Name             string `gorm:"column:name;type:varchar(64);not null" json:"name"`
}

func (DeployPlan) TableName() string {
	return "deploy_plans"
}

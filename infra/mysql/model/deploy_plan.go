package model

// DeployPlan 部署计划 - 完整配置包，聚合CI/CD/实例配置
type DeployPlan struct {
	BaseModel
	Name             string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description      string `gorm:"column:description;type:varchar(255)" json:"description"`
	BusinessUnitID   int64  `gorm:"column:business_unit_id;not null;index" json:"business_unit_id"`
	CIConfigID       int64  `gorm:"column:ci_config_id;not null" json:"ci_config_id"`
	CDConfigID       int64  `gorm:"column:cd_config_id;not null" json:"cd_config_id"`
	InstanceConfigID int64  `gorm:"column:instance_config_id;not null" json:"instance_config_id"`
}

func (DeployPlan) TableName() string {
	return "deploy_plans"
}

package model

// CDConfig CD配置 - 部署相关配置
type CDConfig struct {
	BaseModel
	DeployStrategy string `gorm:"column:deploy_strategy;type:varchar(32);not null" json:"deploy_strategy"` // publish/update/rollback
	RenderEngine   string `gorm:"column:render_engine;type:varchar(32);not null" json:"render_engine"`     // helm/kustomize/custom
	WorkEngine     string `gorm:"column:work_engine;type:varchar(32);not null" json:"work_engine"`         // k8s/docker/ssh
	RenderConfig   string `gorm:"column:render_config;type:json" json:"render_config"`
}

func (CDConfig) TableName() string {
	return "cd_configs"
}

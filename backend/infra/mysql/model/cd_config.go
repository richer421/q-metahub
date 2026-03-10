package model

// CDConfig CD配置 - 部署相关配置
// 定义如何将包产物部署到目标环境
type CDConfig struct {
	BaseModel
	Name           string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	TriggerType    string `gorm:"column:trigger_type;type:varchar(32);not null;default:manual" json:"trigger_type"` // manual/auto
	DeployStrategy string `gorm:"column:deploy_strategy;type:varchar(32);not null;default:rolling" json:"deploy_strategy"` // rolling/canary/bluegreen
	RenderEngine   string `gorm:"column:render_engine;type:varchar(32);not null;default:helm" json:"render_engine"`        // helm/kustomize/custom
	ValuesYAML     string `gorm:"column:values_yaml;type:text" json:"values_yaml"`                                         // Helm values 或等效配置
}

func (CDConfig) TableName() string {
	return "cd_configs"
}

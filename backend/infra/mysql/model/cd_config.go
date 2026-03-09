package model

// CDConfig CD配置 - 部署相关配置
type CDConfig struct {
	BaseModel
	TriggerType  string // manual/auto
	DeployEngine string // 发布引擎

}

// CanaryConfig 金丝雀发布配置 - 金丝雀发布相关配置
type CanaryConfig struct {
}

func (CDConfig) TableName() string {
	return "cd_configs"
}

package model

// CDConfig CD配置 - 部署相关配置
type CDConfig struct {
	BaseModel
}

func (CDConfig) TableName() string {
	return "cd_configs"
}

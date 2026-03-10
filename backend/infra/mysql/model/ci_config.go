package model

// CIConfig CI配置 - 代码构建相关配置
type CIConfig struct {
	BaseModel
}

func (CIConfig) TableName() string {
	return "ci_configs"
}

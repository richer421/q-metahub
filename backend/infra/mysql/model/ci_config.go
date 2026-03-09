package model

// CIConfig CI配置 - 代码构建相关配置
type CIConfig struct {
	BaseModel
	BuildParams string `gorm:"column:build_params;type:json" json:"build_params"`
	EnvVars     string `gorm:"column:env_vars;type:json" json:"env_vars"`
	BuildScript string `gorm:"column:build_script;type:text" json:"build_script"`
}

func (CIConfig) TableName() string {
	return "ci_configs"
}

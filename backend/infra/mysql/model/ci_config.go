package model

// CIConfig CI配置 - 代码构建相关配置
// 定义如何从代码构建出包产物
type CIConfig struct {
	BaseModel
	Name         string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Dockerfile   string `gorm:"column:dockerfile;type:varchar(255);default:Dockerfile" json:"dockerfile"` // Dockerfile 路径
	BuildContext string `gorm:"column:build_context;type:varchar(255);default:." json:"build_context"`    // 构建上下文
	ImageTag     string `gorm:"column:image_tag;type:varchar(128)" json:"image_tag"`                      // 镜像标签模板
	BuildArgs    string `gorm:"column:build_args;type:json" json:"build_args"`                            // 构建参数
	EnvVars      string `gorm:"column:env_vars;type:json" json:"env_vars"`                                // 环境变量
}

func (CIConfig) TableName() string {
	return "ci_configs"
}

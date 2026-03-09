package model

// Dependency 依赖 - 中间件与基础能力
type Dependency struct {
	BaseModel
	Name   string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Type   string `gorm:"column:type;type:varchar(32);not null;index" json:"type"` // mysql/redis/mq
	Config string `gorm:"column:config;type:json" json:"config"`
	Shared bool   `gorm:"column:shared;default:false" json:"shared"`
}

func (Dependency) TableName() string {
	return "dependencies"
}

package model

// Environment 环境 - 部署目标环境
type Environment struct {
	BaseModel
	Name string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Code string `gorm:"column:code;type:varchar(32);uniqueIndex;not null" json:"code"` // dev/test/gray/prod
}

func (Environment) TableName() string {
	return "environments"
}

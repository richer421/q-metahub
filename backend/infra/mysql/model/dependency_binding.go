package model

// DependencyBinding 依赖绑定 - 实例配置与依赖的绑定关系
type DependencyBinding struct {
	BaseModel
	InstanceConfigID int64 `gorm:"column:instance_config_id;not null;index" json:"instance_config_id"`
	DependencyID     int64 `gorm:"column:dependency_id;not null;index" json:"dependency_id"`
}

func (DependencyBinding) TableName() string {
	return "dependency_bindings"
}

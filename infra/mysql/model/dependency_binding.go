package model

// DependencyBinding 依赖绑定 - 实例配置与依赖的绑定关系
type DependencyBinding struct {
	BaseModel
	// 暂时不管
}

func (DependencyBinding) TableName() string {
	return "dependency_bindings"
}

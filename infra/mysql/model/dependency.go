package model

// Dependency 依赖 - 中间件与基础能力
type Dependency struct {
	BaseModel
	// 暂时不管
}

func (Dependency) TableName() string {
	return "dependencies"
}

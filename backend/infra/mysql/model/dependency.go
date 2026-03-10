package model

// Dependency 依赖 - 中间件与基础能力
type Dependency struct {
	BaseModel
}

func (Dependency) TableName() string {
	return "dependencies"
}

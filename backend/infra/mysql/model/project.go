package model

// Project 代码仓库
type Project struct {
	BaseModel
	GitID   int64
	Name    string // 项目名称
	RepoURL string // 仓库地址
}

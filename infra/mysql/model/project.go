package model

// Project 项目 - 代码仓库
type Project struct {
	BaseModel
	GitID   int64  `gorm:"column:git_id;not null;uniqueIndex" json:"git_id"`
	Name    string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	RepoURL string `gorm:"column:repo_url;type:varchar(255);not null" json:"repo_url"`
}

func (Project) TableName() string {
	return "projects"
}

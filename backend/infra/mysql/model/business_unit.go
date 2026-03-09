package model

// BusinessUnit 业务单元 - 面向业务的独立交付单元
type BusinessUnit struct {
	BaseModel
	Name        string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Code        string `gorm:"column:code;type:varchar(32);uniqueIndex;not null" json:"code"`
	RepoURL     string `gorm:"column:repo_url;type:varchar(255);not null" json:"repo_url"`
	Description string `gorm:"column:description;type:varchar(255)" json:"description"`
}

func (BusinessUnit) TableName() string {
	return "business_units"
}

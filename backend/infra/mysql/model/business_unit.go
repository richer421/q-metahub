package model

// BusinessUnit 业务单元 - 面向业务的独立交付单元
type BusinessUnit struct {
	BaseModel
	Name        string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Description string `gorm:"column:description;type:varchar(255);not null" json:"description"`
	ProjectID   int64  `gorm:"column:project_id;not null" json:"project_id"`
}

func (BusinessUnit) TableName() string {
	return "business_units"
}

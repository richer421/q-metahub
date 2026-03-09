package model

// ResourceQuota 资源配额 - 硬件资源定义
type ResourceQuota struct {
	BaseModel
	Name    string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	CPU     string `gorm:"column:cpu;type:varchar(16);not null" json:"cpu"`
	Memory  string `gorm:"column:memory;type:varchar(16);not null" json:"memory"`
	Storage string `gorm:"column:storage;type:varchar(16)" json:"storage"`
	Network string `gorm:"column:network;type:varchar(64)" json:"network"`
}

func (ResourceQuota) TableName() string {
	return "resource_quotas"
}

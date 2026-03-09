package model

// DeployPlan 部署计划 - 完整配置包，聚合CI/CD/实例配置
type DeployPlan struct {
	BaseModel
	BusinessUnitID   int64 // 所属业务单元
	CIConfigID       int64
	CDConfigID       int64
	InstanceConfigID int64
	Name             string
	Description      string
}

func (DeployPlan) TableName() string {
	return "deploy_plans"
}

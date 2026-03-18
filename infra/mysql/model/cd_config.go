package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// CDConfig CD配置 - 部署相关配置
// 定义如何将包产物部署到目标环境
type CDConfig struct {
	BaseModel
	Name            string          `gorm:"column:name;type:varchar(64);not null" json:"name"`
	BusinessUnitID  int64           `gorm:"column:business_unit_id;not null;index" json:"business_unit_id"`  // 关联业务单元
	ReleaseRegion   string          `gorm:"column:release_region;type:varchar(32);not null;default:'cn-east'" json:"release_region"`
	ReleaseEnv      string          `gorm:"column:release_env;type:varchar(32);not null;default:'dev'" json:"release_env"`
	RenderEngine    string          `gorm:"column:render_engine;type:varchar(32);not null;default:helm" json:"render_engine"` // helm/kustomize/custom
	ValuesYAML      string          `gorm:"column:values_yaml;type:text" json:"values_yaml"`                                  // Helm values 或等效配置
	ReleaseStrategy ReleaseStrategy `gorm:"column:release_strategy;type:json;not null" json:"release_strategy"`             // 发布策略
	GitOps          *GitOpsConfig   `gorm:"column:git_ops;type:json" json:"git_ops,omitempty"`                               // GitOps 配置
}

func (CDConfig) TableName() string {
	return "cd_configs"
}

// GitOpsConfig GitOps 发布配置 - 定义 Git 仓库及目录结构
type GitOpsConfig struct {
	Enabled      bool   `json:"enabled"`       // 是否启用 GitOps
	RepoURL      string `json:"repo_url"`      // GitOps 仓库地址
	Branch       string `json:"branch"`        // 分支，如 main
	AppRoot      string `json:"app_root"`      // Application 根目录，默认 "apps"
	ManifestRoot string `json:"manifest_root"` // manifests 根目录，默认 "manifests"
}

func (g GitOpsConfig) Value() (driver.Value, error) {
	return json.Marshal(g)
}

func (g *GitOpsConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan GitOpsConfig: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, g)
}

// ReleaseStrategy 发布策略（聚焦核心：滚动/蓝绿/金丝雀）
type ReleaseStrategy struct {
	DeploymentMode   DeploymentMode     `json:"deployment_mode"`                // 发布模式：rolling/blue_green/canary
	BatchRule        BatchRule          `json:"batch_rule"`                     // 通用分批规则
	CanaryTrafficRule *CanaryTrafficRule `json:"canary_traffic_rule,omitempty"` // 金丝雀专属流量规则（仅canary模式）
}

// DeploymentMode 发布模式
type DeploymentMode string

const (
	DeploymentModeRolling   DeploymentMode = "rolling"    // 滚动发布
	DeploymentModeBlueGreen DeploymentMode = "blue_green" // 蓝绿发布
	DeploymentModeCanary    DeploymentMode = "canary"     // 金丝雀发布
)

// BatchRule 通用分批规则（控制实例发布节奏）
type BatchRule struct {
	BatchCount  int         `json:"batch_count"`  // 总批次
	BatchRatio  []float64   `json:"batch_ratio"`  // 每批实例比例（总和=1）
	TriggerType TriggerType `json:"trigger_type"` // 批次触发方式
	Interval    int         `json:"interval"`     // 批次间隔（秒）
}

// TriggerType 批次触发类型
type TriggerType string

const (
	TriggerTypeAuto   TriggerType = "auto"   // 自动执行
	TriggerTypeManual TriggerType = "manual" // 手动确认
)

// CanaryTrafficRule 金丝雀专属流量规则
type CanaryTrafficRule struct {
	TrafficBatchCount int       `json:"traffic_batch_count"` // 流量分批数
	TrafficRatioList  []float64 `json:"traffic_ratio_list"`  // 每批流量比例（总和≤1）
	ManualAdjust      bool      `json:"manual_adjust"`       // 是否允许手动调整
	AdjustTimeout     int       `json:"adjust_timeout"`      // 手动调整超时（秒）
}

// Value 实现 driver.Valuer 接口
func (r ReleaseStrategy) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan 实现 sql.Scanner 接口
func (r *ReleaseStrategy) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ReleaseStrategy: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, r)
}

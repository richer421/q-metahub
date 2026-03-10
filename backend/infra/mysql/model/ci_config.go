package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// CIConfig CI配置 - 代码构建相关配置
// 核心设计：代码拉取维度（Branch/Tag/CommitID 三选一，互斥）
type CIConfig struct {
	BaseModel
	Name string       `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Spec CIConfigSpec `gorm:"column:spec;type:json;not null" json:"spec"`
}

func (CIConfig) TableName() string {
	return "ci_configs"
}

// CIConfigSpec CI配置详情
// 一体化打包构建流程：代码拉取 → make build → docker build → 输出镜像产物
type CIConfigSpec struct {
	// ========== 代码拉取维度（三选一，互斥） ==========
	RepoURL  string  `json:"repo_url"`            // 项目仓库地址
	Branch   *string `json:"branch,omitempty"`    // 构建分支（如 main）
	Tag      *string `json:"tag,omitempty"`       // 构建标签（如 v1.0.0）
	CommitID *string `json:"commit_id,omitempty"` // 构建 Commit ID

	// ========== 构建配置 ==========
	MakefilePath   string `json:"makefile_path,omitempty"`   // Makefile路径，默认 ./Makefile
	MakeCommand    string `json:"make_command,omitempty"`    // 编译命令，默认 build
	DockerfilePath string `json:"dockerfile_path,omitempty"` // Dockerfile路径，默认 ./Dockerfile
	DockerImage    string `json:"docker_image"`              // 输出镜像产物（必填，如 registry.xxx.com/app:v1.0.0）
}

// Value 实现 driver.Valuer 接口
func (s CIConfigSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *CIConfigSpec) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan CIConfigSpec: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}

// ValidateRef 校验代码拉取字段（确保仅设置一个）
func (s *CIConfigSpec) ValidateRef() error {
	setCount := 0
	if s.Branch != nil && *s.Branch != "" {
		setCount++
	}
	if s.Tag != nil && *s.Tag != "" {
		setCount++
	}
	if s.CommitID != nil && *s.CommitID != "" {
		setCount++
	}

	if setCount > 1 {
		return fmt.Errorf("仅能设置 Branch/Tag/CommitID 中的一个")
	}
	if setCount == 0 {
		return fmt.Errorf("必须设置 Branch/Tag/CommitID 中的一个")
	}
	return nil
}

// GetRef 获取实际使用的代码引用
func (s *CIConfigSpec) GetRef() (refType string, refValue string) {
	if s.Branch != nil && *s.Branch != "" {
		return "branch", *s.Branch
	}
	if s.Tag != nil && *s.Tag != "" {
		return "tag", *s.Tag
	}
	if s.CommitID != nil && *s.CommitID != "" {
		return "commit", *s.CommitID
	}
	return "", ""
}

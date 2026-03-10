package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// CIConfig CI配置 - 代码构建相关配置
// 核心改进：拆分代码拉取维度（Branch/Tag/CommitID 三选一，互斥）
type CIConfig struct {
	BaseModel
	Name string `gorm:"column:name;type:varchar(64);not null" json:"name"`
	Spec CIConfigSpec `gorm:"column:spec;type:json;not null" json:"spec"`
}

func (CIConfig) TableName() string {
	return "ci_configs"
}

// CIConfigSpec CI配置详情
type CIConfigSpec struct {
	// ========== 基础必选信息（拉取代码） ==========
	RepoURL string `json:"repo_url"` // 项目仓库地址

	// ========== 细化的代码拉取维度（三选一，互斥） ==========
	Branch   *string `json:"branch,omitempty"`   // 构建分支（如 main、feature/xxx）
	Tag      *string `json:"tag,omitempty"`      // 构建标签（如 v1.0.0）
	CommitID *string `json:"commit_id,omitempty"` // 构建 Commit ID（精准构建某次提交）

	// ========== 兼容字段（废弃） ==========
	// Deprecated: 建议使用 Branch/Tag/CommitID
	TargetRef string `json:"target_ref,omitempty"`

	// ========== 构建配置（约定默认值） ==========
	ArtifactOutputDir  string `json:"artifact_output_dir,omitempty"`  // 产物输出目录，默认 ./dist
	MakefilePath       string `json:"makefile_path,omitempty"`        // Makefile路径，默认 ./Makefile
	DockerfilePath     string `json:"dockerfile_path,omitempty"`      // Dockerfile路径，默认 ./Dockerfile
	DockerIgnorePath   string `json:"docker_ignore_path,omitempty"`   // .dockerignore路径
	MakeCommand        string `json:"make_command,omitempty"`         // Make执行命令，默认 build
	DockerImage        string `json:"docker_image"`                   // Docker镜像名称+标签（必填）
	DockerBuildContext string `json:"docker_build_context,omitempty"` // Docker构建上下文，默认 ./
	BuildTimeout       int    `json:"build_timeout,omitempty"`        // 构建超时（分钟），默认 30
	CleanWorkspace     bool   `json:"clean_workspace,omitempty"`      // 清理Workspace，默认 true
	VerifyArtifact     bool   `json:"verify_artifact,omitempty"`      // 验证打包产物，默认 true
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
		return fmt.Errorf("仅能设置 Branch/Tag/CommitID 中的一个，不可同时设置")
	}
	if setCount == 0 && s.TargetRef == "" {
		return fmt.Errorf("必须设置 Branch/Tag/CommitID 中的一个，或设置兼容字段 TargetRef")
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
	if s.TargetRef != "" {
		// 简单判断是否为标签（以v开头）
		if len(s.TargetRef) > 0 && s.TargetRef[0] == 'v' {
			return "tag", s.TargetRef
		}
		return "branch", s.TargetRef
	}
	return "", ""
}

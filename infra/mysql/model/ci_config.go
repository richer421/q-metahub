package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CIConfig CI配置 - 代码构建相关配置
// 核心设计：代码拉取维度（Branch/Tag/CommitID 三选一，互斥）
type CIConfig struct {
	BaseModel
	Name           string       `gorm:"column:name;type:varchar(64);not null" json:"name"`
	BusinessUnitID int64        `gorm:"column:business_unit_id;not null;index" json:"business_unit_id"` // 关联业务单元（间接关联 Project）
	ImageRegistry  string       `gorm:"column:image_registry;type:varchar(255);not null" json:"image_registry"` // 镜像仓库地址（如 registry.example.com）
	ImageRepo      string       `gorm:"column:image_repo;type:varchar(255);not null" json:"image_repo"`        // 镜像仓库路径（如 library/myapp）
	ImageTagRule   ImageTagRule `gorm:"column:image_tag_rule;type:json;not null" json:"image_tag_rule"`         // 镜像标签规则
	BuildSpec      BuildSpec    `gorm:"column:build_spec;type:json;not null" json:"build_spec"`                 // 构建配置详情
}

func (CIConfig) TableName() string {
	return "ci_configs"
}

// FullImageRef 生成完整镜像引用（带 tag）
func (c *CIConfig) FullImageRef(tag string) string {
	return fmt.Sprintf("%s/%s:%s", c.ImageRegistry, c.ImageRepo, tag)
}

// ImageTagRule 镜像标签规则
type ImageTagRule struct {
	// 标签类型：branch（分支名）、tag（Git标签）、commit（短commit）、timestamp（时间戳）、custom���自定义）
	Type string `json:"type"` // branch/tag/commit/timestamp/custom
	// 自定义模板（当 type=custom 时使用），支持变量：${branch}, ${tag}, ${commit}, ${timestamp}, ${version}
	Template string `json:"template,omitempty"`
	// 是否添加时间戳后缀
	WithTimestamp bool `json:"with_timestamp,omitempty"`
	// 是否添加 commit 短 hash 后缀
	WithCommit bool `json:"with_commit,omitempty"`
}

// Value 实现 driver.Valuer 接口
func (r ImageTagRule) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan 实现 sql.Scanner 接口
func (r *ImageTagRule) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ImageTagRule: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, r)
}

// GenerateTag 根据规则生成镜像标签
func (r *ImageTagRule) GenerateTag(branch, tag, commit string) string {
	ts := fmt.Sprintf("%d", time.Now().Unix())

	switch r.Type {
	case "branch":
		result := branch
		if r.WithCommit && commit != "" {
			result += "-" + commit[:8]
		}
		if r.WithTimestamp {
			result += "-" + ts
		}
		return result
	case "tag":
		return tag
	case "commit":
		return commit[:8]
	case "timestamp":
		return ts
	case "custom":
		// 替换模板变量
		result := r.Template
		result = strings.ReplaceAll(result, "${branch}", branch)
		result = strings.ReplaceAll(result, "${tag}", tag)
		result = strings.ReplaceAll(result, "${commit}", commit[:8])
		result = strings.ReplaceAll(result, "${timestamp}", ts)
		return result
	default:
		return "latest"
	}
}

// BuildSpec 构建配置详情
// 一体化打包构建流程：代码拉取 → make build → docker build → 输出镜像产物
type BuildSpec struct {
	// ========== 代码拉取维度（三选一，互斥） ==========
	// 注：RepoURL 通过 BusinessUnit -> Project 获取，无需重复存储
	Branch   *string `json:"branch,omitempty"`    // 构建分支（如 main）
	Tag      *string `json:"tag,omitempty"`       // 构建标签（如 v1.0.0）
	CommitID *string `json:"commit_id,omitempty"` // 构建 Commit ID

	// ========== 构建配置 ==========
	MakefilePath   string `json:"makefile_path,omitempty"`   // Makefile路径，默认 ./Makefile
	MakeCommand    string `json:"make_command,omitempty"`    // 编译命令，默认 build
	DockerfilePath string `json:"dockerfile_path,omitempty"` // Dockerfile路径，默认 ./Dockerfile
	DockerContext  string `json:"docker_context,omitempty"`  // Docker 构建上下文路径，默认 .
	BuildArgs      map[string]string `json:"build_args,omitempty"` // Docker 构建参数
}

// Value 实现 driver.Valuer 接口
func (s BuildSpec) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan 实现 sql.Scanner 接口
func (s *BuildSpec) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan BuildSpec: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}

// ValidateRef 校验代码拉取字段（确保仅设置一个）
func (s *BuildSpec) ValidateRef() error {
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
func (s *BuildSpec) GetRef() (refType string, refValue string) {
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

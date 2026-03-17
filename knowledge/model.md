# 核心数据模型

系统核心实体及其关系。新增实体时在此注册。

## BaseModel（通用基础字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | int64 | 主键，自增 |
| CreatedAt | time.Time | 创建时间，自动填充 |
| UpdatedAt | time.Time | 更新时间，自动填充 |

## 实体清单

### Project（项目）

- **表名**: projects
- **说明**: 代码仓库

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| GitID | int64 | NOT NULL, UNIQUE | Git 平台项目 ID |
| Name | varchar(64) | NOT NULL | 项目名称 |
| RepoURL | varchar(255) | NOT NULL | 仓库地址 |

### BusinessUnit（业务单元）

- **表名**: business_units
- **说明**: 面向业务的独立交付单元

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 业务单元名称 |
| Description | varchar(255) | - | 描述 |
| ProjectID | int64 | NOT NULL, INDEX | 关联项目 |

### CIConfig（CI配置）

- **表名**: ci_configs
- **说明**: 代码构建配置，一体化打包构建流程

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 配置名称 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 关联业务单元 |
| ImageRegistry | varchar(255) | NOT NULL | 镜像仓库地址（如 registry.example.com） |
| ImageRepo | varchar(255) | NOT NULL | 镜像仓库路径（如 library/myapp） |
| ImageTagRule | json | NOT NULL | 镜像标签规则 |
| BuildSpec | json | NOT NULL | 构建配置详情 |

**ImageTagRule 结构**（JSON 字段）：

| 字段 | 类型 | 说明 |
|------|------|------|
| Type | string | 标签类型：branch/tag/commit/timestamp/custom |
| Template | string | 自定义模板（当 type=custom 时使用） |
| WithTimestamp | bool | 是否添加时间戳后缀 |
| WithCommit | bool | 是否添加 commit 短 hash 后缀 |

**BuildSpec 结构**（JSON 字段）：

| 字段 | 类型 | 说明 |
|------|------|------|
| Branch | *string | 构建分支（三选一） |
| Tag | *string | 构建标签（三选一） |
| CommitID | *string | 构建 Commit ID（三选一） |
| MakefilePath | string | Makefile路径，默认 ./Makefile |
| MakeCommand | string | 编译命令，默认 build |
| DockerfilePath | string | Dockerfile路径，默认 ./Dockerfile |
| DockerContext | string | Docker 构建上下文路径，默认 . |
| BuildArgs | map | Docker 构建参数 |

### CDConfig（CD配置）

- **表名**: cd_configs
- **说明**: 部署配置，定义发布策略

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 配置名称 |
| RenderEngine | varchar(32) | NOT NULL | 渲染引擎：helm/kustomize/custom |
| ValuesYAML | text | - | Helm values 配置 |
| ReleaseStrategy | json | NOT NULL | 发布策略 |

**ReleaseStrategy 结构**（JSON 字段）：

| 字段 | 类型 | 说明 |
|------|------|------|
| DeploymentMode | string | 发布模式：rolling/blue_green/canary |
| BatchRule | BatchRule | 分批规则 |
| CanaryTrafficRule | *CanaryTrafficRule | 金丝雀流量规则（仅 canary 模式） |

### InstanceOAM（实例配置）

- **表名**: instance_oams
- **说明**: 运行态配置，存储 OAM-Lite 应用模型与前端编辑态

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(128) | NOT NULL, UNIQUE(与BusinessUnitID/Env组合) | 配置名称 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 关联业务单元 |
| Env | varchar(32) | NOT NULL, INDEX | 环境：dev/test/gray/prod |
| SchemaVersion | varchar(32) | NOT NULL, DEFAULT 'v1alpha1' | 模型版本 |
| OAMApplication | json | NOT NULL | OAM-Lite 应用定义（component + traits） |
| FrontendPayload | json | NOT NULL | 前端编辑态回显数据 |

### DeployPlan（部署计划）

- **表名**: deploy_plans
- **说明**: 完整配置包，聚合 CI/CD/InstanceOAM 配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 计划名称 |
| Description | varchar(255) | - | 描述 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 关联业务单元 |
| CIConfigID | int64 | NOT NULL | 关联 CI 配置 |
| CDConfigID | int64 | NOT NULL | 关联 CD 配置 |
| InstanceOAMID | int64 | NOT NULL | 关联实例配置 |

### Dependency（依赖）

- **表名**: dependencies
- **说明**: 中间件与基础能力（TODO: 暂不实现）

## 实体关系

```
Project 1:N BusinessUnit
BusinessUnit 1:N DeployPlan
BusinessUnit 1:1 CIConfig（默认CI配置）
DeployPlan 1:1 CIConfig
DeployPlan 1:1 CDConfig
DeployPlan 1:1 InstanceOAM
```

## 数据流
```
用户 → Project → BusinessUnit → CIConfig → BuildArtifact(q-ci) → Release(q-deploy)
```

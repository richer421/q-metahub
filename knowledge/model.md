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

- **表名**: `projects`
- **说明**: 代码仓库元信息

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| GitID | int64 | NOT NULL, UNIQUE | Git 平台项目 ID |
| Name | varchar(64) | NOT NULL | 项目名称 |
| RepoURL | varchar(255) | NOT NULL | 仓库地址 |

### BusinessUnit（业务单元）

- **表名**: `business_units`
- **说明**: 面向业务的独立交付单元

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 业务单元名称 |
| Description | varchar(255) | - | 描述 |
| ProjectID | int64 | NOT NULL, INDEX | 关联项目 |

### CIConfig（CI 配置）

- **表名**: `ci_configs`
- **说明**: 构建配置（代码拉取 + 构建 + 镜像产物规则）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 配置名称 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 关联业务单元 |
| ImageRegistry | varchar(255) | NOT NULL | 镜像仓库地址 |
| ImageRepo | varchar(255) | NOT NULL | 镜像仓库路径 |
| ImageTagRule | json | NOT NULL | 镜像标签规则 |
| BuildSpec | json | NOT NULL | 构建配置详情 |

### CDConfig（CD 配置）

- **表名**: `cd_configs`
- **说明**: 发布配置，前端仅编辑业务字段，后端补齐渲染与 GitOps 默认规则

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 配置名称 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 关联业务单元 |
| ReleaseRegion | varchar(32) | NOT NULL, DEFAULT `cn-east` | 发布区域，内部存编码（如 `cn-east`） |
| ReleaseEnv | varchar(32) | NOT NULL, DEFAULT `dev` | 发布环境，内部存编码（如 `dev`） |
| RenderEngine | varchar(32) | NOT NULL | 渲染引擎：helm/kustomize/custom |
| ValuesYAML | text | - | 渲染参数 |
| ReleaseStrategy | json | NOT NULL | 发布策略 |
| GitOps | json | - | GitOps 配置 |

说明：

- 前端展示字段使用中文标签：发布区域（华东/华北/新加坡）、发布环境（开发/测试/灰度/生产）、发布策略（滚动发布/金丝雀发布）。
- `RenderEngine`、`ValuesYAML`、`GitOps` 不对前端透出编辑能力，创建/更新时由后端按内置规则回填。
- `ReleaseStrategy` 当前支持：
  - 滚动发布：后端使用默认批次规则。
  - 金丝雀发布：额外持久化批次流量、是否允许手动调整、手动调整超时等参数。

### InstanceOAM（实例配置）

- **表名**: `instance_oams`
- **说明**: 运行态配置权威来源（OAM-Lite）

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(128) | NOT NULL, UNIQUE(与 BusinessUnitID/Env 组合) | 实例名称 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 关联业务单元 |
| Env | varchar(32) | NOT NULL, INDEX | 环境：dev/test/gray/prod |
| SchemaVersion | varchar(32) | NOT NULL, DEFAULT 'v1alpha1' | 模型版本 |
| OAMApplication | json | NOT NULL | OAM-Lite 应用定义 |

说明：`frontend_payload` 不再持久化在 `instance_oams` 表中，仅作为 API 输入/输出桥接视图。

### DeployPlan（部署计划）

- **表名**: `deploy_plans`
- **说明**: 发布配置聚合关系，连接 CI/CD/InstanceOAM

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

- **表名**: `dependencies`
- **说明**: 中间件与基础能力占位模型（字段待业务扩展）

### DependencyBinding（依赖绑定）

- **表名**: `dependency_bindings`
- **说明**: 实例配置与依赖的绑定关系占位模型（字段待业务扩展）

## 实体关系

```
Project 1:N BusinessUnit
BusinessUnit 1:N CIConfig
BusinessUnit 1:N CDConfig
BusinessUnit 1:N InstanceOAM
BusinessUnit 1:N DeployPlan
DeployPlan 1:1 CIConfig
DeployPlan 1:1 CDConfig
DeployPlan 1:1 InstanceOAM
```

## 数据流

```
UI FrontendPayload -> q-metahub(metadata) -> InstanceOAM.OAMApplication
q-deploy / MCP -> q-metahub(open-model) -> DeployPlanSpecVO
```

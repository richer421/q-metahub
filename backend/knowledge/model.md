# 核心数据模型

系统核心实体及其关��。新增实体时在此注册。

## BaseModel（通用基础字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | int64 | 主键，自增 |
| CreatedAt | time.Time | 创建时间，自动填充 |
| UpdatedAt | time.Time | 更新时间，自动填充 |

## 实体清单

### CIConfig（CI配置）

- **表名**: ci_configs
- **说明**: 代码构建配置，一体化打包构建流程

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 配置名称 |
| Spec | json | NOT NULL | 构建配置详情 |

**CIConfigSpec 结构**（JSON 字段）：

| 字段 | 类型 | 说明 |
|------|------|------|
| RepoURL | string | 项目仓库地址（必填） |
| Branch | *string | 构建分支（三选一） |
| Tag | *string | 构建标签（三选一） |
| CommitID | *string | 构建 Commit ID（三选一） |
| MakefilePath | string | Makefile路径，默认 ./Makefile |
| MakeCommand | string | 编译命令，默认 build |
| DockerfilePath | string | Dockerfile路径，默认 ./Dockerfile |
| DockerImage | string | 输出镜像产物（必填） |

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

**BatchRule 结构**：

| 字段 | 类型 | 说明 |
|------|------|------|
| BatchCount | int | 总批次 |
| BatchRatio | []float64 | 每批实例比例 |
| TriggerType | string | 触发方式：auto/manual |
| Interval | int | 批次间隔（秒） |

### InstanceConfig（实例配置）

- **表名**: instance_configs
- **说明**: 运行态配置，存储 K8s 原生 Spec

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 配置名称 |
| Env | varchar(32) | NOT NULL, INDEX | 环境：dev/test/gray/prod |
| InstanceType | varchar(32) | NOT NULL | 实例类型：deployment/statefulset/job/cronjob/pod |
| Spec | json | NOT NULL | K8s 原生工作负载 Spec |
| AttachResources | json | DEFAULT '{}' | 附加资源（ConfigMap/Secret/Service） |

**InstanceSpec 结构**（JSON 字段）：

| InstanceType | K8s Spec |
|--------------|----------|
| deployment | appsv1.DeploymentSpec |
| statefulset | appsv1.StatefulSetSpec |
| job | batchv1.JobSpec |
| cronjob | batchv1.CronJobSpec |
| pod | corev1.PodSpec |

**InstanceAttachResources 结构**：

| 字段 | 类型 | 说明 |
|------|------|------|
| ConfigMaps | map[string]ConfigMap | 配置字典（按名称索引） |
| Secrets | map[string]Secret | 密钥（按名称索引） |
| Services | map[string]Service | 服务（按名称索引） |

### DeployPlan（部署计划）

- **表名**: deploy_plans
- **说明**: 完整配置包，聚合 CI/CD/Instance 配置

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 计划名称 |
| Description | varchar(255) | - | 描述 |
| BusinessUnitID | int64 | NOT NULL, INDEX | 关联业务单元 |
| CIConfigID | int64 | NOT NULL | 关联 CI 配置 |
| CDConfigID | int64 | NOT NULL | 关联 CD 配置 |
| InstanceConfigID | int64 | NOT NULL | 关联实例配置 |

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

### Dependency（依赖）

- **表名**: dependencies
- **说明**: 中间件与基础能力（TODO: 暂不实现）

## 实体关系

```
Project 1:N BusinessUnit
BusinessUnit 1:N DeployPlan
DeployPlan 1:1 CIConfig
DeployPlan 1:1 CDConfig
DeployPlan 1:1 InstanceConfig
```

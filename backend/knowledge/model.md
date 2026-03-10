# 核心数据模型

系统核心实体及其关系。新增实体时在此注册。

## BaseModel（通用基础字段）

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 主键，自增 |
| CreatedAt | time.Time | 创建时间，自动填充 |
| UpdatedAt | time.Time | 更新时间，自动填充 |

## 实体清单

### InstanceConfig（实例配置）

- **表名**: instance_configs
- **说明**: 运行态配置，存储 K8s 原生 Spec

| 字段 | 类型 | 约束 | 说明 |
|------|------|------|------|
| (BaseModel) | - | - | 嵌入通用基础字段 |
| Name | varchar(64) | NOT NULL | 配置名称 |
| Env | varchar(32) | NOT NULL, INDEX | 环境：dev/test/gray/prod |
| InstanceType | varchar(32) | NOT NULL | 实例类型：deployment/statefulset/job/cronjob |
| Spec | json | NOT NULL | K8s 原生 Spec（见下方说明） |

**InstanceSpec 结构**（JSON 字段）：

根据 InstanceType 存储对应的 K8s 原生 Spec：

| InstanceType | K8s Spec |
|--------------|----------|
| deployment | appsv1.DeploymentSpec |
| statefulset | appsv1.StatefulSetSpec |
| job | batchv1.JobSpec |
| cronjob | batchv1.CronJobSpec |

**设计要点**：
- 前端配置本质是 YAML 的 UI 化
- 后端直接存储 K8s 原生结构（通过 JSON 序列化）
- 无需维护额外的模型转换逻辑

## 实体关系

```
BusinessUnit 1:N DeployPlan
DeployPlan 1:1 CIConfig
DeployPlan 1:1 CDConfig
DeployPlan 1:1 InstanceConfig
```

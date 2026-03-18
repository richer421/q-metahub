# 核心抽象

元数据中心的关键抽象与领域概念。

## 部署计划核心语义

**部署计划**是发布配置聚合体：在一个业务单元下，关联构建配置（CI）、发布配置（CD）和实例运行配置（InstanceOAM）。

```
BusinessUnit
  ├── CIConfig
  ├── CDConfig
  ├── InstanceOAM (OAMApplication)
  └── DeployPlan (关联 CI/CD/InstanceOAM)
```

## 实例配置抽象

**后端事实模型**：
- `InstanceOAM.OAMApplication`（持久化）
- `InstanceOAM.SchemaVersion`

**接口桥接模型**：
- `FrontendPayload` 作为 UI 交互视图
- 在接口边界完成 `FrontendPayload <-> OAMApplication` 转换
- `FrontendPayload` 不作为数据库持久化事实来源

**open-model 输出语义**：
- `DeployPlanSpecVO` 是对下游系统暴露的稳定部署规格
- 该规格由 DeployPlan 聚合体派生，不等价于数据库写模型

## 路由语义抽象

API 按语义分组：

- `metadata` 组（写入）：
  - `POST /v1/metadata/instance-oams`
- `open-model` 组（读取）：
  - `GET /v1/open-model/deploy-plans/:id`

## 与外部系统关系

- q-ci：消费 CI 配置并执行构建
- q-deploy：消费 open-model 部署规格并执行发布
- q-metahub：只负责元数据管理与规格输出，不负责执行

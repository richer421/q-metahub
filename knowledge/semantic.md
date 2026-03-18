# 核心语义

## 项目

- **名称**: q-metahub
- **定位**: DevOps 平台的元数据中心，承载部署计划相关的权威元数据
- **领域**: 发布/部署元数据管理与对外模型输出

## 系统边界

做什么：
- 维护发布链路核心元数据（Project / BusinessUnit / CIConfig / CDConfig / InstanceOAM / DeployPlan）
- 接收实例前端视图模型并转换为 OAM-Lite 持久化模型（`instance_oams.oam_application`）
- 对外提供稳定 open-model 读取能力，给下游系统消费部署规格
- 提供 MCP 工具用于读取 open-model 部署计划与服务日志
- 维护前端视图模型、持久化 OAM 与 open-model 之间的语义边界

不做什么：
- 不执行构建（由 q-ci 负责）
- 不执行部署（由 q-deploy 负责）
- 不承担运行时状态采集或控制面职责
- 不将前端编辑态直接作为数据库持久化事实来源

## 核心业务概念

| 术语 | 含义 |
|------|------|
| BusinessUnit | 业务交付单元，组织 DeployPlan/CI/CD/InstanceOAM 的归属 |
| DeployPlan | 发布配置聚合体，关联 CI/CD/InstanceOAM |
| InstanceOAM | 实例配置实体，数据库中持久化 `oam_application` |
| FrontendPayload | 前端视图模型，仅作为接口输入/输出桥接，不作为持久化事实 |
| OAMApplication | 后端权威的实例结构，供发布链路与 open-model 输出使用 |
| OpenModel | 对下游暴露的稳定部署规格读模型（非内部写模型） |

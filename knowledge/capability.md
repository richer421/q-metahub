# 核心业务能力

描述系统解决的业务问题和核心能力边界。

## 设计原则

- 写模型与读模型分离：metadata 写入、open-model 对外读取
- OAM 作为后端事实来源：前端 payload 只用于桥接转换
- 接口语义按路由分组：`/v1/metadata/*` 与 `/v1/open-model/*`

## 能力清单

### 元数据写入（Metadata）

提供实例 OAM 的写入入口：

- `POST /v1/metadata/instance-oams`
- 入参：`CreateInstanceOAMReq`（包含前端视图模型）
- 行为：将前端视图模型转换为 OAM-Lite 后写入 `instance_oams`

### 对外读模型（Open Model）

提供部署计划聚合读模型：

- `GET /v1/open-model/deploy-plans/:id`
- 输出：`DeployPlanSpecVO`（稳定字段，面向下游系统）
- 数据来源：DeployPlan 聚合关联的 CI/CD/InstanceOAM

### MCP 能力

通过 `q-metahub mcp` 启动 MCP Server（stdio 模式），工具包括：

- `read_logs`：读取日志文件
- `get_open_model_deploy_plan`：按 `deploy_plan_id` 返回 open-model 部署计划

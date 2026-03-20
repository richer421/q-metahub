# q-metahub Agent Style Guide

本文件记录当前仓库偏好的代码风格与实现约定，后续 AI 改动默认遵循。

## 1. API 层风格

- 使用函数式 handler，不使用 `MetadataAPI` 这类包裹结构体。
- handler 保持固定流程：
  1. `ShouldBindJSON` / 参数解析
  2. 调用 `metadata.App` 对应方法
  3. `common.Fail` / `common.OK` 返回
- API 只做入参解析与响应封装，不承载业务拼装逻辑。

## 2. App 层风格

- 使用轻量 `App`：`var App = new(app)`，方法显式接收 `context.Context`。
- 不为“可测性”引入厚接口或复杂依赖注入容器。
- 业务逻辑按领域拆文件（例如 `deploy_plan.go`、`oam.go`），避免一个超大 assembler 文件。
- 转换逻辑要分层拆开，避免把整条映射链揉进一个大函数里。

## 3. 路由语义风格

- 路由按语义分组：
  - `/v1/metadata/*`：元数据写入/管理接口
  - `/v1/open-model/*`：对外稳定读取模型接口
- 路由注册集中在 `http/router/v1.go`，避免分散注册。

## 4. 数据访问与错误风格

- 数据访问统一使用：
  - `q := dao.Q.WithContext(ctx)`
  - 链式 `Where(...).First()/Find()`
- 错误需要带上下文，例如：
  - `fmt.Errorf("query deploy_plan id=%d: %w", id, err)`
- 不做过度抽象，优先可读性和可定位性。

## 5. VO 与转换风格

- 前后端交互使用强类型 VO，不使用 `map[string]any` 承载业务结构。
- 可选结构使用指针（例如 `Extended *InstanceExtendedVO`）。
- OAM 与 VO 之间转换使用显式字段映射，避免隐式反射/魔法映射。
- 默认值在转换函数中就地处理，常量集中定义（`default*`）。
- 前端 payload 只作为接口视图模型；数据库持久化事实以 OAM 为准。

## 6. MCP 风格

- MCP 采用最小可用工具集，保持函数直连 app 逻辑。
- 仅保留必要工具，避免堆叠历史工具和冗余抽象。
- MCP 工具返回结构化 JSON 文本，错误统一包装为 `{"error":"..."}`。

## 7. 变更约束

- 小步修改，优先修正当前语义链路，不做无关重构。
- 若用户已确认路由或语义，后续不擅自回退到旧风格。

# q-metahub

下一代的云平台元数据管理中心，ai & cloud 原生应用

## 技术栈

- **后端**: Go 1.25 + Gin + GORM Gen + Cobra（模块名: `github.com/richer421/q-metahub`）
- **基础设施**: MySQL 8.0 / Redis 7 / Kafka 3.7
- **可观测性**: OpenTelemetry + Jaeger + Prometheus
- **工程化**: Swagger 自动文档 / golangci-lint / Makefile / Air 热重载 / Docker Compose

## 架构

采用轻量分层架构，保持单向依赖：`http → app → infra`

- **http 层**: 请求处理、参数校验、统一响应
- **app 层**: 业务能力编排，VO 转换
- **infra 层**: 技术实现（MySQL/Redis/Kafka）
- **knowledge 层**: 项目自我描述，纯 Markdown，不参与运行时

```
├── main.go                     # 入口，仅调用 cmd.Execute()
├── cmd/                        # CLI 命令层（Cobra）
│   ├── root.go                 #   根命令，加载配置 + 初始化日志，支持 -c 指定配置文件
│   └── server.go               #   server 子命令，启动 HTTP 服务 + 基础设施生命周期 + 优雅关停
├── conf/                       # 配置层
│   ├── conf.go                 #   Config 结构体 + Load() + 全局变量 C
│   └── config.yaml             #   YAML 配置（app/server/mysql/redis/kafka/otel/log）
├── http/                       # HTTP 接口层（Gin）
│   ├── server.go               #   NewServer() 初始化 Engine + 中间件 + 路由
│   ├── router/                 #   路由注册，按版本管理
│   │   ├── router.go           #     Register() 入口，挂载 /healthz /readyz /pprof /api
│   │   └── v1.go               #     v1 版本路由，按模块拆分 registerXxx
│   ├── api/                    #   请求处理器，对接 app 层
│   ├── sdk/                    #   HTTP 客户端 SDK，镜像 API 结构
│   │   ├── client.go           #     NewClient(baseURL)，统一请求处理
│   │   └── <module>.go         #     模块级 SDK 方法
│   ├── common/                 #   通用工具（统一响应 OK/Fail）
│   └── middleware/             #   中间件（logger/recovery/otel）
├── app/                        # 应用层 — 用例编排
│   └── <module>/
│       ├── app.go              #   轻量 app 入口（暴露 package 级 App）
│       └── vo/                 #   值对象（入参/出参 VO）
├── infra/                      # 基础设施层 — 技术实现
│   ├── mysql/
│   │   ├── mysql.go            #     DB 初始化 + OTel 插桩
│   │   ├── model/              #     GORM 模型（BaseModel: ID/CreatedAt/UpdatedAt）
│   │   └── dao/                #     GORM Gen 自动生成（勿手动修改）
│   ├── redis/
│   │   └── redis.go            #     Client 初始化 + Redsync 分布式锁 + OTel
│   └── kafka/
│       └── kafka.go            #     Producer + Consumer 注册（同步/异步）+ 重试 + DLQ
├── pkg/                        # 共享包（跨层使用）
│   ├── logger/                 #     Zap 日志（console/JSON，可选文件轮转）
│   ├── otel/                   #     OpenTelemetry 初始化（Tracer/Meter/Prometheus）
│   └── testutil/               #     测试工具（NewMockDB + NewMockRedis）
├── knowledge/                  # 知识层 — AI 理解项目的入口
│   ├── semantic.md             #   项目定位、系统边界、架构分层
│   ├── capability.md           #   业务能力清单
│   ├── model.md                #   数据模型与实体关系
│   └── abstraction.md          #   核心抽象：统一响应、路由版本、配置、CLI 命令
└── gen/                        # 代码生成
    ├── docs/                   #   Swagger 文档（自动生成）
    └── gorm_gen/
        └── main.go             #   GORM Gen 脚本，离线运行，无需连接数据库
```

## 部署

```
deploy/
├── Dockerfile          # 多阶段构建（Alpine，含 librdkafka）
├── docker-compose.yml          # 全栈：MySQL + Redis + Kafka + OTel Collector + Jaeger + Prometheus
├── otel-collector.yaml
└── prometheus.yml
```

## Makefile

```bash
make build          # 编译 Go 二进制到 bin/q-metahub
make run            # 编译 + 运行 server
make dev            # Air 热重载（make dev CMD=server）
make swagger        # 生成 Swagger 文档到 gen/docs/
make sql            # 根据 model 结构体生成类型安全查询代码到 dao/
make lint           # go vet + golangci-lint 代码检查
make test           # 运行全部测试（-v -count=1）
make cover          # 生成覆盖率报告（coverage.html）
make docker-build   # 构建 Docker 镜像
make docker-up      # 启动 docker-compose 全栈
make docker-down    # 停止 docker-compose
make clean          # 清理 bin/ 目录
```

## 开发约定

### 设计原则

- **语义化编程**: 变量、函数、模块命名准确表达意图，代码即文档
- **拒绝过度设计**: 满足当前需求，预留扩展点但不预留实现
- **遵循现有风格**: 新代码与项目已有代码风格保持一致

### 分层规则

- 严格单向依赖，禁止反向引用
- 所有方法第一个参数为 `context.Context`
- handler 不向 app 透传 `*gin.Context`
- app 层直接组织 DAO 调用与 VO ↔ Model 转换
- 对外暴露给 `http/api` 的 app 方法统一收敛在 `app/<module>/app.go` 中
- `app/<module>/` 下其它文件仅保留模块内的视图转换、查询封装、校验与组装函数
- app 层文件命名统一使用 snake_case（如 `deploy_plan.go`、`instance_oam.go`、`ci_config.go`）
- 不为简单链路引入额外 service wrapper、接口分层或依赖注入容器

### 命名规范

- Go 包名: snake_case（如 `hello_world`）
- API 路由: kebab-case（如 `/api/v1/hello-world`）
- 数据库表名: snake_case 复数（如 `hello_worlds`）

### 新增模块流程

1. `infra/mysql/model/` — 定义 GORM 模型（嵌入 `BaseModel`）
2. `gen/gorm_gen/main.go` — 注册模型，`make sql` 生成 DAO
3. `app/<module>/` — 实现 app 方法 + VO + 转换逻辑
4. `http/api/` — 实现函数式 handler（Swagger 注解）
5. `http/router/v1.go` — 注册路由
6. `http/sdk/` — 实现客户端 SDK
7. `knowledge/` — 更新能力清单和数据模型文档
8. `make swagger` — 更新 API 文档

### API 处理器模式

```go
func CreateXxx(c *gin.Context) {
    var req vo.CreateXxxReq
    if err := c.ShouldBindJSON(&req); err != nil {
        common.Fail(c, err)
        return
    }

    res, err := xxx.App.Create(c.Request.Context(), req)
    if err != nil {
        common.Fail(c, err)
        return
    }
    common.OK(c, res)
}
```

### 统一响应格式

```json
{"code": 0, "message": "ok", "data": ...}       // 成功
{"code": -1, "message": "错误信息"}               // 失败
```

### 代码风格补充

- 优先直接、清晰的实现，避免为了“方便测试”额外引入厚抽象
- 数据转换使用强类型视图模型，不使用 `map[string]any` 承载核心业务语义
- 前端视图模型与后端持久化模型在接口边界桥接；后端内部统一传递 OAM/开放模型

### 基础设施生命周期

启动顺序（cmd/server.go）: Config → Logger → OTel → MySQL → Redis → Kafka → HTTP Server
关停: 反序优雅关闭

### 运维端点

- `/healthz` — 存活探针（始终 200）
- `/readyz` — 就绪探针（检查 MySQL/Redis/Kafka，异常返回 503）
- `/debug/pprof/*` — Go 性能分析

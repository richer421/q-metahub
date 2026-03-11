# 核心抽象

元数据中心的关键抽象与领域概念。

## 部署计划核心语义

**部署计划**：通过一定方式构建得到纯净的包产物，配合实例配置（含资源配额），通过指定的发布策略，发布到对应的环境中去。

```
代码仓库 ──构建──> 包产物（纯净）
                      │
                      ├── 实例配置
                      │       ├── 目标环境（dev/test/gray/prod）
                      │       ├── 资源配额（CPU/内存/存储/网络）
                      │       ├── 依赖绑定（MySQL/Redis/MQ）
                      │       └── 其他运行时配置
                      │
                      └── 发布策略
                              ├── 渲染引擎（helm/kustomize/custom）
                              └── 工作引擎（k8s/docker/ssh）
                      │
                      ↓
                 运行态实例
```

**设计要点**：
- **包产物**是纯净的构建结果，不包含运行时配置
- 同一个包产物可配合不同实例配置部署到多环境
- **资源配额**属于实例配置，与环境相关（测试环境少、生产环境多）

## 核心概念定义

- **业务单元**：面向业务的独立交付单元，是租户可独立管理、部署的核心业务载体，其下包含1个代码仓库、若干部署计划。

- **部署计划**：完整的配置包，涵盖从代码构建到实例运行的全链路配置，包含 CI 配置、CD 配置和实例配置。

- **CI配置**：代码构建配置，一体化打包构建流程。代码拉取（Branch/Tag/CommitID 三选一）→ make build → docker build → 输出镜像产物。

- **CD配置**：部署配置，定义发布策略。包含渲染引擎（helm/kustomize/custom）和发布策略（rolling/blue_green/canary + 分批规则）。

- **实例配置**：运行态配置，存储 K8s 原生 Spec（Deployment/StatefulSet/Job/CronJob/Pod），以及附加资源（ConfigMap/Secret/Service）。

- **依赖**：实例运行所需的中间件与基础能力，如 MySQL、Redis、消息队列等（TODO: 暂不实现）。

## 静态结构

```
业务单元（BusinessUnit）
    │
    ├── 代码仓库（RepoURL，唯一）
    │
    └── 部署计划（DeployPlan，多个）
            │
            ├── CI 配置（CIConfig）
            │       ├── 代码拉取（Branch/Tag/CommitID 三选一）
            │       ├── 构建配置（MakefilePath/MakeCommand/DockerfilePath）
            │       └── 输出镜像（DockerImage）
            │
            ├── CD 配置（CDConfig）
            │       ├── 渲染引擎（helm/kustomize/custom）
            │       └── 发布策略（ReleaseStrategy）
            │               ├── 发布模式（rolling/blue_green/canary）
            │               ├── 分批规则（BatchRule）
            │               └── 金丝雀流量规则（CanaryTrafficRule，仅 canary）
            │
            └── 实例配置（InstanceConfig）
                    │
                    ├── 环境配置（dev/test/gray/prod）
                    ├── 实例类型（deployment/statefulset/job/cronjob/pod）
                    ├── Spec（K8s 原生结构，JSON 存储）
                    │       ├── DeploymentSpec
                    │       ├── StatefulSetSpec
                    │       ├── JobSpec
                    │       ├── CronJobSpec
                    │       └── PodSpec
                    └── AttachResources（附加资源）
                            ├── ConfigMaps
                            ├── Secrets
                            └── Services
```

## 动态流转

```
代码仓库 ──提供代码──> 部署计划 ──执行──> 实例（运行态）
                          │
                          ├── CI 配置 → q-ci 构建
                          ├── CD 配置 → q-deploy 部署
                          └── 实例配置 → 实例运行参数
```

## 核心关系说明

1. **业务单元**是顶层业务抽象，包含1个代码仓库和多个部署计划；

2. **部署计划**是完整配置包，聚合 CI配置（构建）、CD配置（部署）、实例配置（运行）；

3. **CI配置**定义代码如何构建，供 q-ci 服务消费；

4. **CD配置**定义如何部署，供 q-deploy 服务消费；

5. **实例配置**定义运行态参数，包含环境、资源、依赖、附加资源；

6. **依赖**可被多个实例配置共享，通过**依赖绑定**建立关联（TODO: 暂不实现）。

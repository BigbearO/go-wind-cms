# Go Wind CMS - AI Coding Rules

> 基于项目代码扫描生成的 AI 编码规范文件，用于指导 AI 助手在 go-wind-cms 项目中的代码生成和修改。

## 1. 项目架构概览

### 1.1 微服务分层

本项目采用 **Kratos 微服务框架**，分为三层：

```
┌─────────────────────────────────────────────────────────────┐
│                      网关层 (BFF)                             │
│  ┌──────────────────────┐  ┌──────────────────────┐        │
│  │   admin-service      │  │    app-service       │        │
│  │   (后台管理网关)      │  │   (前台应用网关)      │        │
│  └──────────┬───────────┘  └──────────┬───────────┘        │
└─────────────┼─────────────────────────┼─────────────────────┘
              │                         │
              │    gRPC 服务间调用        │
              └───────────┬─────────────┘
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                   核心服务层 (Core Service)                   │
│                  core-service (领域服务)                      │
│                                                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐   │
│  │  identity   │ │  permission │ │  content / audit    │   │
│  │  (身份域)    │ │  (权限域)    │ │  (内容/审计域)       │   │
│  └─────────────┘ └─────────────┘ └─────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

### 1.2 服务职责

| 服务 | 服务名 | 职责 | 暴露协议 |
|------|--------|------|----------|
| `app/admin/service` | admin-service | 后台管理 BFF，提供完整 CMS 管理能力 | HTTP REST + gRPC + SSE |
| `app/app/service` | app-service | 前台应用 BFF，面向终端用户 | HTTP REST + gRPC + SSE |
| `app/core/service` | core-service | 核心领域服务，业务逻辑与数据持久化 | gRPC |

**核心原则**：
- Admin/App 服务是 **BFF (Backend-for-Frontend)**，只负责协议转换、参数校验、简单编排
- Core 服务是 **领域服务**，包含所有业务逻辑和数据访问
- **单向依赖**：Admin/App → Core，Core 不依赖其他服务

---

## 2. 代码分层规范

### 2.1 目录结构规范

每个服务模块必须遵循以下目录结构：

```
app/{service_name}/service/
├── cmd/server/
│   ├── main.go              # 服务入口
│   ├── wire.go              # Wire 依赖注入定义
│   └── wire_gen.go          # Wire 生成文件（勿手动修改）
├── configs/
│   ├── server.yaml          # HTTP/gRPC/SSE 服务端配置
│   ├── client.yaml          # gRPC 客户端配置
│   ├── data.yaml            # 数据库/缓存配置
│   ├── registry.yaml        # 服务注册发现配置
│   ├── logger.yaml          # 日志配置
│   ├── trace.yaml           # 链路追踪配置
│   ├── oss.yaml             # 对象存储配置
│   └── remote.yaml          # 远程配置中心
└── internal/
    ├── server/
    │   ├── providers/
    │   │   └── wire_set.go  # Server 层 ProviderSet
    │   ├── grpc_server.go   # gRPC 服务端注册
    │   ├── rest_server.go   # HTTP REST 服务端注册
    │   └── sse_server.go    # SSE 服务端注册
    ├── service/
    │   ├── providers/
    │   │   └── wire_set.go  # Service 层 ProviderSet
    │   └── *_service.go     # 业务 Service 实现
    └── data/
        ├── providers/
        │   └── wire_set.go  # Data 层 ProviderSet
        ├── data.go          # 基础设施客户端创建
        └── *_repo.go        # 仓库接口与实现
```

**关键规则**：
- `internal/` 目录下的代码不可被外部模块导入
- 所有依赖注入通过 `wire.go` + `providers/wire_set.go` 完成
- 禁止在 `wire_set.go` 中写业务逻辑，仅允许构造函数注册

### 2.2 分层职责

#### 2.2.1 Server 层 (Transport)

职责：协议适配、服务注册、中间件编排

```go
// server/grpc_server.go - Core Service 示例
func NewGrpcServer(
    ctx *bootstrap.Context,
    middlewares []middleware.Middleware,
    authenticationService *service.AuthenticationService,
    // ... 注入所有 service
) (*grpc.Server, error) {
    srv, err := rpc.CreateGrpcServer(cfg, middlewares...)
    // 注册所有 gRPC 服务
    authenticationV1.RegisterAuthenticationServiceServer(srv, authenticationService)
    return srv, nil
}
```

**规则**：
- Core Service 必须注册 gRPC Server
- Admin/App Service 必须注册 REST Server + gRPC Server (用于 DTM)
- 中间件在 `NewRestMiddleware` / `NewGrpcMiddleware` 中组装

#### 2.2.2 Service 层 (Business)

职责：业务编排、参数校验、调用 Data 层

**Core Service 的 Service**：

```go
type UserService struct {
    identityV1.UnimplementedUserServiceServer  // 嵌入 gRPC 基础实现

    log *log.Helper

    userRepo           data.UserRepo
    userCredentialRepo *data.UserCredentialRepo
    roleRepo           *data.RoleRepo
    // ... 其他 repo 依赖
}

func NewUserService(
    ctx *bootstrap.Context,
    userRepo data.UserRepo,
    // ... 通过 Wire 注入依赖
) *UserService {
    svc := &UserService{
        log:      ctx.NewLoggerHelper("user/service/core-service"),
        userRepo: userRepo,
        // ...
    }
    svc.init()  // 初始化默认数据
    return svc
}
```

**Admin/App Service 的 Service**（RPC 代理模式）：

```go
type UserService struct {
    adminV1.UserServiceHTTPServer  // 嵌入 HTTP 基础实现

    log *log.Helper

    userServiceClient     identityV1.UserServiceClient      // Core 的 gRPC 客户端
    tenantServiceClient   identityV1.TenantServiceClient
    roleServiceClient     permissionV1.RoleServiceClient
    // ... 其他 Core Service 客户端
}

// List 直接透传
func (s *UserService) List(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.ListUserResponse, error) {
    return s.userServiceClient.List(ctx, req)
}

// Create 进行业务编排
func (s *UserService) Create(ctx context.Context, req *identityV1.CreateUserRequest) (*identityV1.User, error) {
    // 1. 参数校验
    // 2. 权限检查
    // 3. 关联数据校验（如角色存在性）
    // 4. 调用 Core Service
    return s.userServiceClient.Create(ctx, req)
}
```

**规则**：
- Core Service 的 Service 直接操作 Repo，包含业务逻辑
- Admin/App Service 的 Service 只进行**参数校验**和**简单编排**，不实现业务逻辑
- 所有 Service 构造函数必须接收 `*bootstrap.Context` 作为第一个参数
- Logger 必须通过 `ctx.NewLoggerHelper("module/service/layer")` 创建

#### 2.2.3 Data 层 (Data Access)

**Core Service 的 Data 层**：

```go
// 定义接口
type UserRepo interface {
    List(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.ListUserResponse, error)
    Get(ctx context.Context, req *identityV1.GetUserRequest) (*identityV1.User, error)
    Create(ctx context.Context, req *identityV1.CreateUserRequest) (*identityV1.User, error)
    CreateWithTx(ctx context.Context, tx *ent.Tx, data *identityV1.User) (*identityV1.User, error)
    Update(ctx context.Context, req *identityV1.UpdateUserRequest) error
    Delete(ctx context.Context, req *identityV1.DeleteUserRequest) error
    Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error)
    // ... 其他方法
}

// 实现结构体
type userRepo struct {
    entClient *entCrud.EntClient[*ent.Client]
    log       *log.Helper
    mapper    *mapper.CopierMapper[identityV1.User, ent.User]
    repository *entCrud.Repository[...]
}
```

**Admin/App Service 的 Data 层**（RPC 客户端创建）：

```go
func NewUserServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.UserServiceClient {
    cli, err := rpc.CreateGrpcClient(
        ctx.Context(),
        r,
        serviceid.NewDiscoveryName(serviceid.CoreService),  // "go-wind-cms/core-service"
        ctx.GetConfig(),
    )
    if err != nil {
        return nil
    }
    return identityV1.NewUserServiceClient(cli)
}
```

**规则**：
- Core Service：Repo 接口 + 实现，使用 Ent ORM 操作数据库
- Admin/App Service：创建 gRPC Client，通过服务发现调用 Core Service
- 所有客户端创建函数命名规范：`New{ServiceName}Client`
- 服务发现地址格式：`discovery:///{project_name}/{service_name}`

---

## 3. API 定义规范

### 3.1 Proto 文件组织

```
api/protos/
├── admin/service/v1/       # Admin BFF 的 HTTP API 定义
│   ├── i_*.proto           # 接口定义（i = interface）
│   └── admin_*.proto       # Admin 专属消息/错误
├── app/service/v1/         # App BFF 的 HTTP API 定义
│   ├── i_*.proto
│   └── app_*.proto
├── authentication/service/v1/  # Core 领域服务
├── identity/service/v1/
├── permission/service/v1/
├── content/service/v1/
├── comment/service/v1/
├── site/service/v1/
├── media/service/v1/
├── dict/service/v1/
├── audit/service/v1/
├── internal_message/service/v1/
├── resource/service/v1/
├── storage/service/v1/
└── task/service/v1/
```

**规则**：
- BFF 层（admin/app）的 proto 以 `i_` 前缀命名，表示 Interface
- Core 领域服务的 proto 不使用 `i_` 前缀
- 每个领域独立目录，包含 `*_error.proto` 定义领域错误码

### 3.2 Proto 生成规则

使用 `buf` 工具链生成代码：

```yaml
# api/buf.gen.yaml
version: v2
managed:
  enabled: true
plugins:
  - local: protoc-gen-go
    out: gen/go
    opt: paths=source_relative
  - local: protoc-gen-go-grpc
    out: gen/go
    opt: paths=source_relative
  - local: protoc-gen-go-http
    out: gen/go
    opt: paths=source_relative
  - local: protoc-gen-validate
    out: gen/go
    opt: paths=source_relative,lang=go
```

**规则**：
- 生成的代码放在 `api/gen/go/` 目录
- 禁止手动修改生成文件（`*_grpc.pb.go`, `*_http.pb.go` 等）
- 修改 proto 后执行 `make generate` 重新生成

---

## 4. 依赖注入规范 (Wire)

### 4.1 ProviderSet 组织

```go
// service/providers/wire_set.go
//go:build wireinject
// +build wireinject

package providers

import (
    "github.com/google/wire"
    "go-wind-cms/app/admin/service/internal/service"
)

var ProviderSet = wire.NewSet(
    service.NewAuthenticationService,
    service.NewUserService,
    service.NewRoleService,
    // ... 所有 Service 构造函数
)
```

### 4.2 Wire 构建

```go
// cmd/server/wire.go
//go:build wireinject
// +build wireinject

package main

import (
    "github.com/google/wire"
    "github.com/go-kratos/kratos/v2"
    "github.com/tx7do/kratos-bootstrap/bootstrap"

    dataProviders "go-wind-cms/app/admin/service/internal/data/providers"
    serverProviders "go-wind-cms/app/admin/service/internal/server/providers"
    serviceProviders "go-wind-cms/app/admin/service/internal/service/providers"
)

func initApp(*bootstrap.Context) (*kratos.App, func(), error) {
    panic(wire.Build(
        serverProviders.ProviderSet,
        serviceProviders.ProviderSet,
        dataProviders.ProviderSet,
        newApp,
    ))
}
```

**规则**：
- 每个层级的 `wire_set.go` 必须独立定义 `ProviderSet`
- `wire.go` 只负责组装各层 ProviderSet
- 修改依赖后执行 `go generate ./...` 或 `wire` 重新生成 `wire_gen.go`

---

## 5. 服务注册与发现

### 5.1 服务注册

```go
// pkg/serviceid/service_id.go
const (
    AdminService = "admin-service"
    AppService   = "app-service"
    CoreService  = "core-service"
    DTMService   = "dtm-service"
)
```

### 5.2 服务发现配置

```yaml
# configs/registry.yaml
registry:
  type: "etcd"           # 支持 etcd / consul
  etcd:
    endpoints:
      - "localhost:2379"
  consul:
    address: "localhost:8500"
    scheme: "http"
```

### 5.3 客户端创建

```go
// 通过服务发现创建 gRPC 客户端
func NewUserServiceClient(ctx *bootstrap.Context, r registry.Discovery) identityV1.UserServiceClient {
    cli, err := rpc.CreateGrpcClient(
        ctx.Context(),
        r,
        serviceid.NewDiscoveryName(serviceid.CoreService),  // "go-wind-cms/core-service"
        ctx.GetConfig(),
    )
    if err != nil {
        return nil
    }
    return identityV1.NewUserServiceClient(cli)
}
```

**规则**：
- 服务发现名称格式：`{project_name}/{service_name}`
- 所有 gRPC 客户端通过 `rpc.CreateGrpcClient` 统一创建
- 客户端超时在 `configs/client.yaml` 中配置（默认 120s）

---

## 6. 负载均衡

### 6.1 负载均衡策略

本项目使用 **Kratos 内置负载均衡**，默认策略：

| 策略 | 说明 | 配置位置 |
|------|------|----------|
| **WRR (Weighted Round Robin)** | 加权轮询，默认策略 | `rpc.CreateGrpcClient` 内部 |
| **Random** | 随机选择 | 可通过配置切换 |
| **Consistent Hash** | 一致性哈希 | 需要自定义配置 |

### 6.2 实现方式

负载均衡由 `kratos-bootstrap/rpc` 包封装，在创建 gRPC 客户端时自动启用：

```go
// 内部实现（kratos-bootstrap/rpc）
func CreateGrpcClient(ctx context.Context, r registry.Discovery, serviceName string, cfg *conf.Bootstrap) (*grpc.ClientConn, error) {
    // 1. 通过服务发现获取实例列表
    // 2. 使用 WRR 策略进行负载均衡
    // 3. 创建 gRPC 连接
}
```

**规则**：
- 无需手动配置负载均衡策略，使用框架默认值（WRR）
- 服务实例权重通过注册中心（etcd/consul）维护
- 健康检查由注册中心自动处理

---

## 7. 网关实现

### 7.1 网关架构

本项目采用 **BFF (Backend-for-Frontend)** 模式作为网关层：

```
外部请求
    │
    ▼
┌─────────────────────────────────────────┐
│         Admin/App Service               │
│  ┌─────────────────────────────────┐   │
│  │         REST Server             │   │
│  │  ┌─────────────────────────┐   │   │
│  │  │    Middleware Chain     │   │   │
│  │  │  1. Logging             │   │   │
│  │  │  2. API Audit Log       │   │   │
│  │  │  3. Auth (JWT)          │   │   │
│  │  │  4. Authorization       │   │   │
│  │  │  5. Rate Limit          │   │   │
│  │  └─────────────────────────┘   │   │
│  │         ↓                       │   │
│  │  ┌─────────────────────────┐   │   │
│  │  │   Service Handler       │   │   │
│  │  │  (参数校验/简单编排)      │   │   │
│  │  └─────────────────────────┘   │   │
│  └─────────────────────────────────┘   │
│                   │                     │
│                   ▼                     │
│  ┌─────────────────────────────────┐   │
│  │      gRPC Client (Core)         │   │
│  │   通过服务发现调用 core-service  │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

### 7.2 中间件链

#### REST 中间件（Admin Service）

```go
func NewRestMiddleware(
    ctx *bootstrap.Context,
    accessTokenChecker auth.AccessTokenChecker,
    authorizer authzEngine.Engine,
    apiAuditLogServiceClient auditV1.ApiAuditLogServiceClient,
    loginAuditLogServiceClient auditV1.LoginAuditLogServiceClient,
) []middleware.Middleware {
    var ms []middleware.Middleware

    // 1. 日志中间件
    ms = append(ms, logging.Server(ctx.GetLogger()))

    // 2. 审计日志中间件
    ms = append(ms, applogging.Server(
        applogging.WithWriteApiLogFunc(...),
        applogging.WithWriteLoginLogFunc(...),
    ))

    // 3. 认证 + 鉴权中间件（白名单跳过）
    ms = append(ms, selector.Server(
        auth.Server(...),      // JWT 认证
        authz.Server(authorizer),  // 权限校验
    ).Match(rpc.NewRestWhiteListMatcher()).Build())

    return ms
}
```

#### gRPC 中间件（Core Service）

```go
func NewGrpcMiddleware(ctx *bootstrap.Context) []middleware.Middleware {
    var ms []middleware.Middleware
    ms = append(ms, logging.Server(ctx.GetLogger()))
    ms = append(ms, ent.Server())  // Ent Viewer 注入
    return ms
}
```

### 7.3 认证与授权

#### JWT 认证

```yaml
# configs/server.yaml
server:
  rest:
    middleware:
      auth:
        method: "HS256"
        key: "some_api_key"
```

#### 权限校验流程

1. **TokenChecker** 校验 JWT Token 有效性
2. **Auth Middleware** 解析 Token 注入用户信息到 Context
3. **Ent Middleware** 将用户信息转换为 Ent Viewer（用于数据权限）
4. **Authorization Middleware** 使用 Casbin/OPA 进行权限校验

### 7.4 白名单机制

```go
// 登录接口跳过认证
rpc.AddWhiteList(
    adminV1.OperationAuthenticationServiceLogin,
    adminV1.OperationAuthenticationServiceLogout,
)

// App 服务公开接口白名单
rpc.AddWhiteList(
    appV1.OperationAuthenticationServiceLogin,
    appV1.OperationNavigationServiceList,
    appV1.OperationPostServiceList,
    // ... 其他公开接口
)
```

---

## 8. 数据访问规范

### 8.1 Ent ORM 使用

```go
// 创建 Ent 客户端
func NewEntClient(ctx *bootstrap.Context) (*entCrud.EntClient[*ent.Client], func(), error) {
    cli, err := entBootstrap.NewEntClient(cfg, func(drv *sql.Driver) *ent.Client {
        client := ent.NewClient(
            ent.Driver(drv),
            ent.Log(func(a ...any) { l.Debug(a...) }),
        )
        // 自动迁移
        if cfg.Data.Database.GetMigrate() {
            client.Schema.Create(ctx.Context(), migrate.WithForeignKeys(true))
        }
        return client
    })
    return cli, cleanup, err
}
```

### 8.2 Repo 实现规范

```go
type userRepo struct {
    entClient  *entCrud.EntClient[*ent.Client]
    log        *log.Helper
    mapper     *mapper.CopierMapper[identityV1.User, ent.User]
    repository *entCrud.Repository[...]
}

func (r *userRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*identityV1.ListUserResponse, error) {
    builder := r.entClient.Client().User.Query()
    whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
    // ... 查询逻辑
}
```

**规则**：
- 使用 `go-crud/entgo` 封装简化 CRUD 操作
- 枚举类型使用 `mapper.EnumTypeConverter` 进行转换
- 时间类型使用 `copierutil.NewTimeStringConverterPair()` 转换

---

## 9. 配置规范

### 9.1 配置文件

| 文件 | 用途 |
|------|------|
| `server.yaml` | HTTP/gRPC/SSE 服务端配置（地址、超时、中间件开关） |
| `client.yaml` | gRPC 客户端配置（超时、中间件开关） |
| `data.yaml` | 数据库、Redis、Elasticsearch 配置 |
| `registry.yaml` | 服务注册发现配置（etcd/consul） |
| `logger.yaml` | 日志级别、格式、输出配置 |
| `trace.yaml` | OpenTelemetry 链路追踪配置 |
| `oss.yaml` | MinIO 对象存储配置 |
| `remote.yaml` | 远程配置中心（Consul KV） |

### 9.2 服务端配置示例

```yaml
server:
  rest:
    addr: "0.0.0.0:6600"
    timeout: 10s
    enable_swagger: true
    cors:
      origins: ["*"]
      methods: ["GET", "POST", "PUT", "DELETE"]
      headers: ["Authorization", "Content-Type"]
    middleware:
      auth:
        method: "HS256"
        key: "some_api_key"

  grpc:
    addr: "0.0.0.0:0"      # 0 表示随机端口
    timeout: 10s
    middleware:
      enable_logging: true
      enable_recovery: true
      enable_tracing: true
      enable_validate: true
      enable_circuit_breaker: true
      enable_metadata: true

  sse:
    addr: ":6601"
    codec: "json"
    path: "/events"
```

---

## 10. 错误处理规范

### 10.1 错误定义

每个领域在 `*_error.proto` 中定义错误码：

```protobuf
// api/protos/authentication/service/v1/authentication_error.proto
syntax = "proto3";

package authentication.service.v1;

import "errors/errors.proto";

option go_package = "go-wind-cms/api/gen/go/authentication/service/v1;authenticationV1";

enum AuthenticationErrorReason {
    option (errors.default_code) = 500;

    UNAUTHORIZED = 0 [(errors.code) = 401];
    FORBIDDEN = 1 [(errors.code) = 403];
    INVALID_GRANT_TYPE = 2 [(errors.code) = 400];
    // ...
}
```

### 10.2 错误使用

```go
// 生成代码后使用
return nil, authenticationV1.ErrorUnauthorized("invalid token")
return nil, authenticationV1.ErrorForbidden("insufficient authority")
return nil, authenticationV1.ErrorInvalidGrantType("unsupported grant type")
```

**规则**：
- 每个领域独立定义错误码
- 使用 `protoc-gen-go-errors` 生成错误辅助函数
- HTTP 状态码通过 proto 注解指定

---

## 11. 中间件规范

### 11.1 自定义中间件位置

```
pkg/middleware/
├── auth/           # 认证中间件
│   ├── auth.go     # 主中间件
│   ├── token_checker.go
│   ├── context.go
│   └── options.go
├── logging/        # 审计日志中间件
│   ├── logging.go
│   ├── api_audit_log.go
│   └── login_audit_log.go
└── ent/            # Ent Viewer 注入中间件
    └── ent.go
```

### 11.2 中间件开发规范

```go
func Server(opts ...Option) middleware.Middleware {
    op := options{...}
    for _, o := range opts {
        o(&op)
    }

    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req any) (any, error) {
            // 前置处理
            // ...

            reply, err := handler(ctx, req)

            // 后置处理
            // ...

            return reply, err
        }
    }
}
```

---

## 12. 分布式事务 (DTM)

### 12.1 DTM 配置

```go
// server/grpc_server.go
func NewGrpcServer(...) (*grpc.Server, error) {
    srv, err := rpc.CreateGrpcServer(cfg, middlewares...)

    // 获取服务端点
    en, err := srv.Endpoint()

    // 初始化 DTM Grpc 驱动
    workflow.InitGrpc(serviceName.DtmServiceAddress, en.String(), srv.Server)

    return srv, nil
}
```

### 12.2 DTM 驱动

```go
// data/dtm.go
func NewDtmDriver(rr registry.Discovery) {
    dtmdriver.Register(dtmdriver.DriverName, &kratosDriver{discovery: rr})
}
```

---

## 13. 编码规范

### 13.1 命名规范

| 类型 | 规范 | 示例 |
|------|------|------|
| 服务名 | kebab-case | `admin-service`, `core-service` |
| 包名 | 小写，无下划线 | `service`, `data`, `server` |
| 接口名 | 动词+名词+Service | `UserService`, `AuthenticationService` |
| 结构体 | 大驼峰 | `AuthenticationService`, `userRepo` |
| 接口 | 名词+Repo | `UserRepo`, `RoleRepo` |
| 方法 | 大驼峰 | `Create`, `List`, `Get`, `Update`, `Delete` |
| 变量 | 小驼峰 | `userRepo`, `entClient` |
| 常量 | 大写下划线 | `AdminService`, `CoreService` |

### 13.2 包导入规范

```go
import (
    // 标准库
    "context"
    "time"

    // 第三方库
    "github.com/go-kratos/kratos/v2/log"
    "github.com/tx7do/kratos-bootstrap/bootstrap"
    "google.golang.org/protobuf/types/known/emptypb"

    // 项目内部 - api
    adminV1 "go-wind-cms/api/gen/go/admin/service/v1"
    identityV1 "go-wind-cms/api/gen/go/identity/service/v1"

    // 项目内部 - pkg
    "go-wind-cms/pkg/middleware/auth"
    "go-wind-cms/pkg/serviceid"

    // 项目内部 - 当前模块
    "go-wind-cms/app/admin/service/internal/service"
)
```

### 13.3 日志规范

```go
// 创建 Logger
log := ctx.NewLoggerHelper("user/service/core-service")

// 使用
log.Debug("debug message")
log.Infof("user %d created", userID)
log.Errorf("create user failed: %v", err)
```

---

## 14. 测试规范

### 14.1 测试文件位置

```
internal/data/
├── user_repo.go
├── user_repo_test.go       # 与实现文件同目录
└── user_token_cache_test.go
```

### 14.2 测试规范

```go
func TestUserRepo_Create(t *testing.T) {
    // 使用 testcontainer 或内存数据库
    // 或使用 enttest
}
```

---

## 15. 新增服务开发流程

当需要新增一个领域服务时，按以下步骤进行：

### 15.1 定义 Proto API

1. 在 `api/protos/{domain}/service/v1/` 创建 proto 文件
2. 定义 Service、Message、Error
3. 执行 `make generate` 生成代码

### 15.2 Core Service 实现

1. 在 `app/core/service/internal/data/` 创建 Repo 接口和实现
2. 在 `app/core/service/internal/service/` 创建 Service 实现
3. 在 `app/core/service/internal/server/grpc_server.go` 注册 gRPC 服务
4. 在 `app/core/service/internal/service/providers/wire_set.go` 添加 Provider

### 15.3 Admin/App Service 代理

1. 在 `app/admin/service/internal/data/data.go` 创建 gRPC Client
2. 在 `app/admin/service/internal/service/` 创建代理 Service
3. 在 `app/admin/service/internal/server/rest_server.go` 注册 HTTP 服务
4. 在 `app/admin/service/internal/data/providers/wire_set.go` 添加 Client Provider
5. 在 `app/admin/service/internal/service/providers/wire_set.go` 添加 Service Provider

### 15.4 执行 Wire 生成

```bash
cd app/core/service && go generate ./...
cd app/admin/service && go generate ./...
```

---

## 16. 关键约束

1. **禁止循环依赖**：Admin/App → Core，Core 不依赖任何其他服务
2. **禁止跨层调用**：Service 层只能调用同层的 Service 和下一层的 Data/Repo
3. **禁止手动修改生成文件**：所有 `*_gen.go` 文件由工具生成
4. **禁止在 wire_set.go 写业务逻辑**：仅允许注册构造函数
5. **必须使用接口**：Repo 层必须定义接口，Service 层依赖接口而非实现
6. **Context 传递**：所有函数第一个参数必须是 `context.Context`
7. **错误处理**：使用领域错误码，禁止直接返回裸错误

---

*本规则文件基于项目代码扫描自动生成，如有变更请同步更新。*

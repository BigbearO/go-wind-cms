# 环境依赖概览

不需要 DTM。需要 Jaeger（分布式追踪）。

## 必须启动的中间件（Docker）

| 服务 | 用途 | 说明 |
|------|------|------|
| PostgreSQL | 数据库 | 主存储，`docker-compose.libs.yaml` 已配置 |
| Redis | 缓存 | 会话/缓存，必需 |
| Elasticsearch | 全文搜索 | 内容搜索 |
| MinIO | 对象存储 | 文件/图片存储 |
| etcd | 配置中心/服务发现 | Kratos 微服务注册 |
| Jaeger | 分布式追踪 | 链路追踪，必需 |

## 本地开发工具

- Go 1.18+、Node.js 16+、pnpm
- Docker + Docker Compose
- Buf（protobuf 代码生成）
- protoc 插件（`make plugin`）
- gow/kratos CLI（`make cli`）

## 启动命令

```powershell
# 启动所有依赖中间件
docker compose -f docker-compose.libs.yaml up -d

# 初始化项目工具链
make init

# 启动 admin 服务
gow run admin
# 或
cd app/admin/service && make run
```

Jaeger 访问地址：`http://localhost:16686`
etcd 地址：`localhost:2379`
DB 默认库名：`gwc`
默认密码：`*Abcd123456`

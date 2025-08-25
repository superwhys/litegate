# LiteGate

一个轻量级的Go语言API网关服务，提供代理转发、JWT认证、路由匹配等功能。

## 功能特性

- 🚀 **高性能代理转发** - 基于Gin框架的高性能HTTP代理
- 🔐 **JWT身份验证** - 支持JWT token验证和用户信息提取
- 🛣️ **灵活路由匹配** - 使用正则表达式进行精确的URL路由匹配
- ⚖️ **负载均衡** - 支持多个后端地址的随机负载均衡
- ⚡ **配置热重载** - 支持配置文件的热重载，无需重启服务
- ⏱️ **超时控制** - 可配置的请求超时时间
- 🛡️ **CORS支持** - 内置跨域资源共享支持

## 快速开始

### 安装

```bash
git clone https://github.com/superwhys/litegate.git
cd litegate
go mod tidy
go build -o litegate main.go
```

### 运行

```bash
# 使用默认配置运行
./litegate


# 指定配置文件
./litegate -f=./content/config.yaml
```

## 配置说明

### 主配置文件 (config.yaml)

```yaml
gateway:
  services:
    - test  # 允许访问的服务列表
  timeout: 20s  # 全局超时时间
```

### 代理配置文件 (content/proxy/{service}.json)

```json
{
    "proxy": ["http://127.0.0.1:8080"],
    "timeout": "10s",
    "auth": {
        "type": "jwt",
        "source": "$query.token",
        "secret": "your-jwt-secret",
        "claims": {
            "$query.user_id": "user_id",
            "$header.X-User": "userName"
        }
    },
    "routes": [
        {
            "match": "^/api/.*",
            "proxy": ["http://127.0.0.1:8000"],
            "disable_auth": true
        }
    ]
}
```

### 配置参数说明

#### RouteConfig

- `proxy` - 代理地址列表（必填）
- `timeout` - 超时时间（可选，默认30秒）
- `auth` - 身份验证配置（可选）
- `routes` - 路由配置列表（必填）

#### Auth

- `type` - Token类型，固定为"jwt"
- `source` - Token在请求中的位置（如：`$query.token`、`$header.Authorization`）
- `secret` - JWT密钥
- `claims` - JWT解码后数据存储位置映射

#### Route

- `match` - URL匹配正则表达式（必填）
- `proxy` - 代理地址列表
- `timeout` - 超时时间（可选）
- `disable_auth` - 是否禁用身份验证
- `auth` - 身份验证配置覆盖

## 使用示例

### 1. 基本代理转发

```json
{
    "proxy": ["http://backend-service:8080"],
    "routes": [
        {
            "match": "^/api/.*",
            "proxy": ["http://backend-service:8080"]
        }
    ]
}
```

### 2. 带JWT认证的代理

```json
{
    "proxy": ["http://backend-service:8080"],
    "auth": {
        "type": "jwt",
        "source": "$header.Authorization",
        "secret": "your-secret-key",
        "claims": {
            "$header.X-User-ID": "user_id",
            "$header.X-User-Name": "user_name"
        }
    },
    "routes": [
        {
            "match": "^/api/.*",
            "proxy": ["http://backend-service:8080"]
        }
    ]
}
```

### 3. 负载均衡配置

```json
{
    "proxy": ["http://backend1:8080", "http://backend2:8080", "http://backend3:8080"],
    "routes": [
        {
            "match": "^/api/.*",
            "proxy": ["http://backend1:8080", "http://backend2:8080"]
        }
    ]
}
```

## API接口

### 调试接口

- `GET /debug/config` - 获取当前所有配置信息
- `GET /debug/config/:serviceName` - 获取指定路由信息

## 开发

### 依赖

- Go 1.25.0+
- Gin Web框架
- JWT库
- 其他依赖见 `go.mod`

### 构建

```bash
go build -o litegate main.go
```

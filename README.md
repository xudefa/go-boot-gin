# go-boot-gin

[![Go Version](https://img.shields.io/github/go-mod/go-version/xudefa/go-boot-gin)](https://go.dev/) [![License](https://img.shields.io/github/license/xudefa/go-boot-gin)](./LICENSE) [![Build Status](https://img.shields.io/github/actions/workflow/status/xudefa/go-boot-gin/test.yml?branch=master)](https://github.com/xudefa/go-boot-gin/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/xudefa/go-boot-gin.svg)](https://pkg.go.dev/github.com/xudefa/go-boot-gin) [![Go Report Card](https://goreportcard.com/badge/github.com/xudefa/go-boot-gin)](https://goreportcard.com/report/github.com/xudefa/go-boot-gin)

基于 [go-boot](https://github.com/xudefa/go-boot) 的 Gin Web 框架集成模块。将 Gin 无缝集成到 go-boot 的 IoC 容器和自动配置体系中，提供声明式的路由注册、中间件配置和优雅启停能力。

> 设计理念：遵循 go-boot 的开发规范，将 Gin 作为 `net.Server` 接口的实现，通过自动配置实现零代码启动 Web 服务。

## 整体架构

```
┌───────────────────────────────────────────────────────────────────────┐
│                    go-boot ApplicationContext                         │
│  ┌───────────┐ ┌──────────────┐ ┌───────────┐ ┌───────────┐           │
│  │ Container │ │  Environment │ │ Lifecycle │ │ EventBus  │           │
│  └───────────┘ └──────────────┘ └───────────┘ └───────────┘           │
│                       ┌─────────────────────┐                         │
│                       │ AutoConfig Registry │                         │
│                       └─────────────────────┘                         │
└───────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────────┐
                    │     go-boot-gin Starter       │
                    │  ┌─────────────────────────┐  │
                    │  │ GinEngine Bean          │  │
                    │  │ GinServer (net.Server)  │  │
                    │  │ Router Configuration    │  │
                    │  │ Middleware Chain        │  │
                    │  └─────────────────────────┘  │
                    └───────────────────────────────┘
```

## 目录

- [快速开始](#快速开始)
- [功能特性](#功能特性)
- [路由注册](#路由注册)
- [中间件配置](#中间件配置)
- [配置选项](#配置选项)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [贡献](#贡献)
- [许可证](#许可证)

## 快速开始

### 安装

```bash
# 安装核心框架
go get github.com/xudefa/go-boot

# 安装 Gin 集成模块
go get github.com/xudefa/go-boot-gin
```

### 最小示例

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/xudefa/go-boot/boot"
    "github.com/xudefa/go-boot/core"
    ginboot "github.com/xudefa/go-boot-gin/gin"
)

func main() {
    app, err := boot.NewApplication(
        boot.WithAppName("my-web-app"),
        boot.WithVersion("1.0.0"),
    )
    if err != nil {
        panic(err)
    }
    defer app.Stop()

    // 注册 Gin Engine
    app.Container().Register("ginEngine", core.Bean(gin.Default()))

    // 注册路由
    engine := app.Container().Get("ginEngine").(*gin.Engine)
    engine.GET("/hello", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "Hello from go-boot-gin!"})
    })

    // 启动应用（自动启动 Gin 服务器）
    app.Start()

    // 等待终止信号
    app.WaitForSignal()
}
```

## 功能特性

| 特性 | 说明 |
|------|------|
| Gin 集成 | 将 Gin Engine 注册为 Bean，支持依赖注入 |
| net.Server 实现 | GinServer 实现 go-boot 的 `net.Server` 接口 |
| 自动配置 | 通过 `gin.enabled=true` 自动启动 Web 服务器 |
| 优雅启停 | 支持优雅关闭和生命周期管理 |
| 声明式路由 | 支持通过 Handler Bean 声明式注册路由 |
| 中间件链 | 支持全局和路由级中间件配置 |
| 配置驱动 | 端口、模式、超时等均可通过配置控制 |

## 路由注册

### 方式一：直接注册

```go
engine := app.Container().Get("ginEngine").(*gin.Engine)
engine.GET("/users", listUsers)
engine.POST("/users", createUser)
engine.PUT("/users/:id", updateUser)
engine.DELETE("/users/:id", deleteUser)
```

### 方式二：Handler Bean

```go
type UserHandler struct {
    Engine  *gin.Engine   `inject:"ginEngine"`
    Service *UserService  `inject:"userService"`
}

func (h *UserHandler) RegisterRoutes() {
    group := h.Engine.Group("/users")
    group.GET("", h.List)
    group.POST("", h.Create)
    group.GET("/:id", h.GetByID)
    group.PUT("/:id", h.Update)
    group.DELETE("/:id", h.Delete)
}
```

### 方式三：Router 辅助

```go
import "github.com/xudefa/go-boot-gin/router"

r := router.New(app.Container())
r.GET("/health", healthCheck)
r.Group("/api/v1", func(g *router.Group) {
    g.GET("/users", listUsers)
    g.POST("/users", createUser)
})
```

## 中间件配置

### 全局中间件

```go
engine := app.Container().Get("ginEngine").(*gin.Engine)

// 内置中间件
engine.Use(gin.Recovery())
engine.Use(gin.Logger())

// 自定义中间件
engine.Use(func(c *gin.Context) {
    start := time.Now()
    c.Next()
    log.Printf("Request took %v", time.Since(start))
})
```

### CORS 中间件

```go
import "github.com/xudefa/go-boot-gin/middleware"

engine.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins:     []string{"*"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length"},
    AllowCredentials: true,
}))
```

## 配置选项

通过 `boot.WithProperty()` 或配置文件设置：

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `gin.enabled` | `false` | 是否启用 Gin 服务器 |
| `gin.port` | `8080` | 服务器监听端口 |
| `gin.host` | `localhost` | 服务器监听地址 |
| `gin.mode` | `debug` | Gin 模式：debug / release / test |
| `gin.read-timeout` | `10` | 读取超时（秒） |
| `gin.write-timeout` | `10` | 写入超时（秒） |
| `gin.idle-timeout` | `60` | 空闲超时（秒） |
| `gin.shutdown-timeout` | `5` | 优雅关闭超时（秒） |

### 示例配置

```yaml
# application.yml
gin:
  enabled: true
  port: 8080
  mode: debug
  read-timeout: 30
  write-timeout: 30
  idle-timeout: 60
  shutdown-timeout: 10
```

## 项目结构

```
go-boot-gin/
├── gin/                    # Gin 自动配置
│   └── autoconfig.go       # 自动配置注册
├── server/                 # Gin Server 实现
│   ├── server.go           # GinServer 实现 net.Server
│   ├── options.go          # Server 选项配置
│   └── server_test.go      # 单元测试
├── router/                 # 路由注册辅助
│   └── router.go           # 声明式路由注册
├── middleware/             # 中间件
│   ├── security.go         # 安全中间件
│   ├── tracing.go          # 分布式追踪中间件
│   ├── validation.go       # 请求验证中间件
│   └── websocket.go        # WebSocket 适配器
├── README.md
├── LICENSE
└── go.mod
```

## 开发指南

### 构建

```bash
go build ./...
```

### 测试

```bash
go test ./...
go test -cover ./...       # 带覆盖率
go test -race ./...        # 数据竞争检测
```

### 代码规范

```bash
go fmt ./...
golangci-lint run
```

## 贡献

欢迎提交 Issue 和 Pull Request！详细贡献指南请参阅 [CONTRIBUTING.md](./CONTRIBUTING.md)。

## 许可证

本项目采用 MIT 许可证 — 详情请参阅 [LICENSE](./LICENSE) 文件。
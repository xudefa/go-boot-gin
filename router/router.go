// Package router 提供声明式路由注册辅助工具。
//
// 该包简化了 Gin 路由注册流程，支持通过容器管理路由处理器。
//
// 快速开始:
//
//	r := router.New(app.Container())
//	r.WithServer(server)
//	r.GET("/health", healthCheck)
//	r.Group("/api/v1", func(g *router.Group) {
//	    g.GET("/users", listUsers)
//	    g.POST("/users", createUser)
//	})
package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/go-boot-gin/server"
	"github.com/xudefa/go-boot/core"
)

// Router 路由注册辅助器
type Router struct {
	container core.Container
	server    *server.GinServer
}

// New 创建新的路由注册辅助器
//
// 参数:
//   - container: go-boot IoC 容器
//
// 返回值:
//   - *Router: 路由注册辅助器实例
func New(container core.Container) *Router {
	return &Router{
		container: container,
	}
}

// WithServer 设置 Gin 服务器实例
func (r *Router) WithServer(s *server.GinServer) *Router {
	r.server = s
	return r
}

// GET 注册 GET 路由
func (r *Router) GET(path string, handlers ...any) *Router {
	if r.server != nil {
		r.server.GET(path, handlers...)
	}
	return r
}

// POST 注册 POST 路由
func (r *Router) POST(path string, handlers ...any) *Router {
	if r.server != nil {
		r.server.POST(path, handlers...)
	}
	return r
}

// PUT 注册 PUT 路由
func (r *Router) PUT(path string, handlers ...any) *Router {
	if r.server != nil {
		r.server.PUT(path, handlers...)
	}
	return r
}

// DELETE 注册 DELETE 路由
func (r *Router) DELETE(path string, handlers ...any) *Router {
	if r.server != nil {
		r.server.DELETE(path, handlers...)
	}
	return r
}

// PATCH 注册 PATCH 路由
func (r *Router) PATCH(path string, handlers ...any) *Router {
	if r.server != nil {
		r.server.PATCH(path, handlers...)
	}
	return r
}

// HEAD 注册 HEAD 路由
func (r *Router) HEAD(path string, handlers ...any) *Router {
	if r.server != nil {
		r.server.HEAD(path, handlers...)
	}
	return r
}

// OPTIONS 注册 OPTIONS 路由
func (r *Router) OPTIONS(path string, handlers ...any) *Router {
	if r.server != nil {
		r.server.OPTIONS(path, handlers...)
	}
	return r
}

// Group 创建路由组
//
// 参数:
//   - relativePath: 路由组基础路径
//   - configure: 路由组配置函数
//   - handlers: 可选的路由组级中间件
func (r *Router) Group(relativePath string, configure func(g *Group), handlers ...any) *Group {
	if r.server == nil {
		return nil
	}
	g := &Group{
		router: r,
		group:  r.server.Group(relativePath, handlers...),
	}
	if configure != nil {
		configure(g)
	}
	return g
}

// Group 路由组
type Group struct {
	router *Router
	group  *gin.RouterGroup
}

// GET 注册 GET 路由
func (g *Group) GET(path string, handlers ...any) *Group {
	g.group.GET(path, g.wrapHandlers(handlers)...)
	return g
}

// POST 注册 POST 路由
func (g *Group) POST(path string, handlers ...any) *Group {
	g.group.POST(path, g.wrapHandlers(handlers)...)
	return g
}

// PUT 注册 PUT 路由
func (g *Group) PUT(path string, handlers ...any) *Group {
	g.group.PUT(path, g.wrapHandlers(handlers)...)
	return g
}

// DELETE 注册 DELETE 路由
func (g *Group) DELETE(path string, handlers ...any) *Group {
	g.group.DELETE(path, g.wrapHandlers(handlers)...)
	return g
}

// PATCH 注册 PATCH 路由
func (g *Group) PATCH(path string, handlers ...any) *Group {
	g.group.PATCH(path, g.wrapHandlers(handlers)...)
	return g
}

// HEAD 注册 HEAD 路由
func (g *Group) HEAD(path string, handlers ...any) *Group {
	g.group.HEAD(path, g.wrapHandlers(handlers)...)
	return g
}

// OPTIONS 注册 OPTIONS 路由
func (g *Group) OPTIONS(path string, handlers ...any) *Group {
	g.group.OPTIONS(path, g.wrapHandlers(handlers)...)
	return g
}

// Group 创建子路由组
func (g *Group) Group(relativePath string, configure func(g *Group), handlers ...any) *Group {
	subGroup := &Group{
		router: g.router,
		group:  g.group.Group(relativePath, g.wrapHandlers(handlers)...),
	}
	if configure != nil {
		configure(subGroup)
	}
	return subGroup
}

// wrapHandlers 将 any 类型的 handlers 转换为 gin.HandlerFunc
func (g *Group) wrapHandlers(handlers []any) []gin.HandlerFunc {
	result := make([]gin.HandlerFunc, 0, len(handlers))
	for _, h := range handlers {
		switch v := h.(type) {
		case gin.HandlerFunc:
			result = append(result, v)
		case func(*gin.Context):
			result = append(result, v)
		default:
			log.Printf("Warning: unsupported handler type %T for route, skipping", h)
		}
	}
	return result
}

// Engine 返回底层的 Gin Engine 实例，用于更灵活的路由配置
func (r *Router) Engine() *gin.Engine {
	if r.server != nil {
		return r.server.Engine()
	}
	return nil
}

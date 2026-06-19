// Package server 基于 Gin 框架提供 HTTP 服务器实现。
//
// 该包将 Gin 框架与 go-boot 容器系统集成，
// 支持依赖注入、中间件和路由注册。
//
// 定义：
//
//   - Server: HTTP 服务器实现了 net.Server 接口
//   - HandlerFunc: 请求处理函数类型
//   - Option: 服务器配置选项
//
// 快速开始:
//
//	s := server.New()
//	s.GET("/hello", func(c *gin.Context) {
//	    c.String(200, "Hello World")
//	})
//	s.Run()
package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/net"

	"github.com/gin-gonic/gin"
)

// GinServer 是 HTTP 服务器，实现了 net.Server 接口。
//
// 字段说明:
//   - engine: Gin 引擎实例
//   - container: go-boot IoC 容器
//   - config: 服务器通用配置
//   - middleware: 路由级中间件
//   - globalMiddleware: 全局中间件
//   - registerFuncs: 注册函数列表
//   - certFile: TLS 证书文件路径（HTTPS）
//   - keyFile: TLS 私钥文件路径（HTTPS）
type GinServer struct {
	host             string
	port             int
	mode             string
	readTimeout      time.Duration
	writeTimeout     time.Duration
	idleTimeout      time.Duration
	shutdownTimeout  time.Duration
	engine           *gin.Engine
	container        core.Container
	middleware       []any
	registerFuncs    []func(container core.Container) error
	globalMiddleware []any
	httpServer       *http.Server
	certFile         string
	keyFile          string
	mu               sync.RWMutex
}

type ginHandlerContext struct {
	c       *gin.Context
	aborted bool
}

func (h *ginHandlerContext) RequestMethod() string {
	return h.c.Request.Method
}

func (h *ginHandlerContext) RequestURI() string {
	return h.c.Request.URL.RequestURI()
}

func (h *ginHandlerContext) Header(key string) string {
	return h.c.GetHeader(key)
}

func (h *ginHandlerContext) SetStatusCode(code int) {
	h.c.Status(code)
}

func (h *ginHandlerContext) SetHeader(key, value string) {
	h.c.Header(key, value)
}

func (h *ginHandlerContext) AbortWithStatus(code int) {
	h.aborted = true
	h.c.AbortWithStatus(code)
}

func (h *ginHandlerContext) AbortWithStatusJSON(code int, body any) {
	h.aborted = true
	h.c.AbortWithStatusJSON(code, body)
}

func (h *ginHandlerContext) Next() {
	h.c.Next()
}

func (h *ginHandlerContext) IsAborted() bool {
	return h.aborted
}

func (h *ginHandlerContext) Context() context.Context {
	return h.c.Request.Context()
}

func (h *ginHandlerContext) SetContext(ctx context.Context) {
	h.c.Request = h.c.Request.WithContext(ctx)
}

// AdaptMiddleware 将 net.MiddlewareFunc 转换为 gin.HandlerFunc。
func AdaptMiddleware(m net.MiddlewareFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		hc := &ginHandlerContext{c: c}
		m(hc)
	}
}

// New 创建新的 Gin HTTP 服务器。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *GinServer: 配置好的服务器实例
func New(opts ...Option) *GinServer {
	gin.SetMode(gin.DebugMode)

	s := &GinServer{
		host:             "localhost",
		port:             8080,
		mode:             gin.DebugMode,
		readTimeout:      30 * time.Second,
		writeTimeout:     30 * time.Second,
		idleTimeout:      60 * time.Second,
		shutdownTimeout:  5 * time.Second,
		container:        core.New(),
		middleware:       make([]any, 0),
		registerFuncs:    make([]func(container core.Container) error, 0),
		globalMiddleware: make([]any, 0),
	}

	// 先初始化引擎，然后再应用选项
	s.engine = gin.New()
	s.engine.Use(gin.Recovery())

	for _, opt := range opts {
		opt(s)
	}

	gin.SetMode(s.mode)
	s.setupDefaultMiddleware()

	return s
}

// setupDefaultMiddleware 设置默认中间件链
func (s *GinServer) setupDefaultMiddleware() {
	s.globalMiddleware = append(s.globalMiddleware,
		net.RequestIDMiddleware(nil),
		net.ErrorMiddleware(net.ErrorMiddlewareConfig{}),
		net.AccessLogMiddleware(net.AccessLogConfig{}),
		net.CORSMiddleware(net.DefaultCORSConfig()),
	)
}

// Option 是服务器配置选项函数。
type Option func(*GinServer)

// HandlerFunc 是请求处理函数类型。
type HandlerFunc func(ctx *gin.Context)

// Use 向服务器的路由中间件链追加一个中间件。
//
// 参数:
//   - m: 中间件函数 (gin.HandlerFunc 或 func(*gin.Context))
func (s *GinServer) Use(m any) {
	s.middleware = append(s.middleware, m)
}

// UseGlobal 向服务器的全局中间件链追加一个中间件，全局中间件在所有路由之前执行。
//
// 参数:
//   - m: 中间件函数 (gin.HandlerFunc 或 func(*gin.Context))
func (s *GinServer) UseGlobal(m any) {
	s.globalMiddleware = append(s.globalMiddleware, m)
}

// GlobalMiddlewareList 返回全局中间件列表
func (s *GinServer) GlobalMiddlewareList() []any {
	return s.globalMiddleware
}

// Register 注册一个处理函数到容器中，用于依赖注入。
//
// 参数:
//   - fn: 注册函数，接受 core.Container 参数
func (s *GinServer) Register(fn func(container core.Container) error) {
	s.registerFuncs = append(s.registerFuncs, fn)
}

// Container 返回 go-boot 容器实例。
func (s *GinServer) Container() any {
	return s.container
}

// Stop 优雅地停止服务器，等待正在处理的请求完成。
//
// 参数:
//   - ctx: 上下文，用于控制停止的超时
func (s *GinServer) Stop(ctx context.Context) error {
	s.mu.RLock()
	srv := s.httpServer
	s.mu.RUnlock()

	if srv == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, s.shutdownTimeout)
	defer cancel()
	return srv.Shutdown(ctx)
}

// 编译时检查 GinServer 是否实现 net.Server 接口
var _ net.Server = (*GinServer)(nil)

// Engine 返回 Gin 引擎实例。
func (s *GinServer) Engine() *gin.Engine {
	return s.engine
}

// Start 启动 HTTP 服务器并开始监听请求。
//
// 在启动时会：
// 1. 执行所有注册函数
// 2. 配置中间件
// 3. 启动服务器监听指定端口
// 4. 阻塞直到服务器被 Stop() 关闭或发生错误
//
// 注意：该方法会阻塞，信号处理应由 go-boot 的 ApplicationContext 管理。
//
// 返回值:
//   - error: 启动或关闭时的错误
func (s *GinServer) Start() error {
	for _, fn := range s.registerFuncs {
		if err := fn(s.container); err != nil {
			return fmt.Errorf("failed to register function: %w", err)
		}
	}

	s.setupMiddleware()

	if s.mode != "" {
		gin.SetMode(s.mode)
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	log.Printf("Starting Gin server on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}
	s.mu.Lock()
	s.httpServer = srv
	s.mu.Unlock()

	errCh := make(chan error, 1)
	go func() {
		if s.certFile != "" && s.keyFile != "" {
			if err := srv.ListenAndServeTLS(s.certFile, s.keyFile); err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("failed to start HTTPS server: %w", err)
				return
			}
		} else {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("failed to start HTTP server: %w", err)
				return
			}
		}
		errCh <- nil
	}()

	// 等待服务器关闭信号（由 Stop() 方法触发）或启动错误
	err := <-errCh
	if err != nil {
		return err
	}
	// errCh 收到 nil 表示服务器正常关闭
	log.Println("Gin server stopped")
	return nil
}

func (s *GinServer) setupMiddleware() {
	// 应用全局中间件（已在 setupDefaultMiddleware 中初始化）
	for _, m := range s.globalMiddleware {
		switch v := m.(type) {
		case gin.HandlerFunc:
			s.engine.Use(v)
		case func(*gin.Context):
			s.engine.Use(v)
		case net.MiddlewareFunc:
			s.engine.Use(AdaptMiddleware(v))
		}
	}

	// 应用路由级中间件
	for _, m := range s.middleware {
		switch v := m.(type) {
		case gin.HandlerFunc:
			s.engine.Use(v)
		case func(*gin.Context):
			s.engine.Use(v)
		case net.MiddlewareFunc:
			s.engine.Use(AdaptMiddleware(v))
		}
	}
}

// GET 注册 GET 路由。
//
// 参数:
//   - path: 路由路径
//   - handlers: 处理函数
//
// 返回值:
//   - *Server: 服务器实例，支持链式调用
func (s *GinServer) GET(path string, handlers ...any) *GinServer {
	s.engine.GET(path, s.wrapHandlers(handlers)...)
	return s
}

// POST 注册 POST 路由。
//
// 参数:
//   - path: 路由路径
//   - handlers: 处理函数
//
// 返回值:
//   - *Server: 服务器实例，支持链式调用
func (s *GinServer) POST(path string, handlers ...any) *GinServer {
	s.engine.POST(path, s.wrapHandlers(handlers)...)
	return s
}

// PUT 注册 PUT 路由。
//
// 参数:
//   - path: 路由路径
//   - handlers: 处理函数
//
// 返回值:
//   - *Server: 服务器实例，支持链式调用
func (s *GinServer) PUT(path string, handlers ...any) *GinServer {
	s.engine.PUT(path, s.wrapHandlers(handlers)...)
	return s
}

// DELETE 注册 DELETE 路由。
//
// 参数:
//   - path: 路由路径
//   - handlers: 处理函数
//
// 返回值:
//   - *Server: 服务器实例，支持链式调用
func (s *GinServer) DELETE(path string, handlers ...any) *GinServer {
	s.engine.DELETE(path, s.wrapHandlers(handlers)...)
	return s
}

// PATCH 注册 PATCH 路由。
//
// 参数:
//   - path: 路由路径
//   - handlers: 处理函数
//
// 返回值:
//   - *Server: 服务器实例，支持链式调用
func (s *GinServer) PATCH(path string, handlers ...any) *GinServer {
	s.engine.PATCH(path, s.wrapHandlers(handlers)...)
	return s
}

// HEAD 注册 HEAD 路由。
//
// 参数:
//   - path: 路由路径
//   - handlers: 处理函数
//
// 返回值:
//   - *Server: 服务器实例，支持链式调用
func (s *GinServer) HEAD(path string, handlers ...any) *GinServer {
	s.engine.HEAD(path, s.wrapHandlers(handlers)...)
	return s
}

// OPTIONS 注册 OPTIONS 路由。
//
// 参数:
//   - path: 路由路径
//   - handlers: 处理函数
//
// 返回值:
//   - *Server: 服务器实例，支持链式调用
func (s *GinServer) OPTIONS(path string, handlers ...any) *GinServer {
	s.engine.OPTIONS(path, s.wrapHandlers(handlers)...)
	return s
}

// Any 注册接受所有 HTTP 方法的路由。
func (s *GinServer) Any(path string, handlers ...any) *GinServer {
	s.engine.Any(path, s.wrapHandlers(handlers)...)
	return s
}

// Group 创建一个路由组。
//
// 参数:
//   - relativePath: 路由组的基础路径
//   - handlers: 路由组级中间件
//
// 返回值:
//   - *gin.RouterGroup: 路由组
func (s *GinServer) Group(relativePath string, handlers ...any) *gin.RouterGroup {
	return s.engine.Group(relativePath, s.wrapHandlers(handlers)...)
}

func (s *GinServer) wrapHandlers(handlers []any) []gin.HandlerFunc {
	result := make([]gin.HandlerFunc, 0, len(handlers))
	for _, h := range handlers {
		switch v := h.(type) {
		case gin.HandlerFunc:
			result = append(result, v)
		case func(*gin.Context):
			result = append(result, v)
		case HandlerFunc:
			result = append(result, gin.HandlerFunc(v))
		case net.MiddlewareFunc:
			result = append(result, AdaptMiddleware(v))
		default:
			log.Printf("Warning: unsupported handler type %T for route, skipping", h)
		}
	}
	return result
}

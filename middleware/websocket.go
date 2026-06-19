package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GinWebSocketHandler Gin WebSocket 处理器
// 封装 gorilla/websocket 的升级和连接处理逻辑
type GinWebSocketHandler struct {
	upgrader  websocket.Upgrader
	onConnect func(*websocket.Conn)
}

// NewGinWebSocketHandler 创建 Gin WebSocket 处理器
func NewGinWebSocketHandler(opts ...WebSocketOption) *GinWebSocketHandler {
	handler := &GinWebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	for _, opt := range opts {
		opt(handler)
	}
	return handler
}

// WebSocketOption WebSocket 配置选项
type WebSocketOption func(*GinWebSocketHandler)

// WithReadBufferSize 设置读取缓冲区大小
func WithReadBufferSize(size int) WebSocketOption {
	return func(h *GinWebSocketHandler) {
		h.upgrader.ReadBufferSize = size
	}
}

// WithWriteBufferSize 设置写入缓冲区大小
func WithWriteBufferSize(size int) WebSocketOption {
	return func(h *GinWebSocketHandler) {
		h.upgrader.WriteBufferSize = size
	}
}

// WithCheckOrigin 设置跨域检查函数
func WithCheckOrigin(fn func(r *http.Request) bool) WebSocketOption {
	return func(h *GinWebSocketHandler) {
		h.upgrader.CheckOrigin = fn
	}
}

// WithOnConnect 设置连接建立回调
func WithOnConnect(fn func(*websocket.Conn)) WebSocketOption {
	return func(h *GinWebSocketHandler) {
		h.onConnect = fn
	}
}

// Handle 处理 WebSocket 连接请求
func (h *GinWebSocketHandler) Handle(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if h.onConnect != nil {
		h.onConnect(conn)
	}
}

// Middleware 返回 Gin 中间件函数，用于拦截 WebSocket 升级请求
func (h *GinWebSocketHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Upgrade") != "websocket" {
			c.Next()
			return
		}

		h.Handle(c)
		c.Abort()
	}
}

// WebSocketMiddleware 创建 WebSocket 中间件（快捷函数）
func WebSocketMiddleware(opts ...WebSocketOption) gin.HandlerFunc {
	handler := NewGinWebSocketHandler(opts...)
	return handler.Middleware()
}

// WebSocketRoute 创建 WebSocket 路由注册函数（快捷函数）
func WebSocketRoute(path string, opts ...WebSocketOption) func(*gin.Engine) {
	return func(engine *gin.Engine) {
		handler := NewGinWebSocketHandler(opts...)
		engine.GET(path, handler.Handle)
	}
}

// WebSocketGroup 创建 WebSocket 路由组（快捷函数）
func WebSocketGroup(path string, engine *gin.Engine, opts ...WebSocketOption) *gin.RouterGroup {
	handler := NewGinWebSocketHandler(opts...)
	group := engine.Group(path)
	group.GET("", handler.Handle)
	return group
}

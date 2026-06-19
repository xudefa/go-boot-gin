package server

import (
	"time"

	"github.com/xudefa/go-boot/core"

	"github.com/gin-gonic/gin"
)

// WithContainer 设置自定义容器。
func WithContainer(c core.Container) Option {
	return func(s *GinServer) {
		s.container = c
	}
}

// WithMode 设置 Gin 引擎的运行模式。
//
// 参数:
//   - mode: 运行模式，可选值为 "debug"、"release"、"test"
//
// 返回值:
//   - Option: 配置选项函数
//
// 示例:
//
//	s := server.New(server.WithMode("debug"))
func WithMode(mode string) Option {
	return func(s *GinServer) {
		s.mode = mode
		gin.SetMode(mode)
	}
}

// WithHost 设置服务器监听地址。
//
// 参数:
//   - host: 监听地址，如 ":8080" 或 "0.0.0.0:8080"
//
// 返回值:
//   - Option: 配置选项函数
func WithHost(host string) Option {
	return func(s *GinServer) {
		s.host = host
	}
}

// WithPort 设置服务器监听端口。
//
// 参数:
//   - port: 监听端口，如 8080
//
// 返回值:
//   - Option: 配置选项函数
func WithPort(port int) Option {
	return func(s *GinServer) {
		s.port = port
	}
}

// WithShutdownTimeout 设置服务器优雅关机超时时间。
//
// 参数:
//   - timeout: 关机超时时间
//
// 返回值:
//   - Option: 配置选项函数
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(s *GinServer) {
		s.shutdownTimeout = timeout
	}
}

// WithReadTimeout 设置读取请求的超时时间。
//
// 参数:
//   - timeout: 读取超时时间
//
// 返回值:
//   - Option: 配置选项函数
func WithReadTimeout(timeout time.Duration) Option {
	return func(s *GinServer) {
		s.readTimeout = timeout
	}
}

// WithWriteTimeout 设置写入响应的超时时间。
//
// 参数:
//   - timeout: 写入超时时间
//
// 返回值:
//   - Option: 配置选项函数
func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *GinServer) {
		s.writeTimeout = timeout
	}
}

// WithIdleTimeout 设置连接空闲超时时间。
//
// 参数:
//   - timeout: 空闲超时时间
//
// 返回值:
//   - Option: 配置选项函数
func WithIdleTimeout(timeout time.Duration) Option {
	return func(s *GinServer) {
		s.idleTimeout = timeout
	}
}

// WithTLS 配置 HTTPS，设置证书和密钥文件路径。
//
// 参数:
//   - certFile: PEM 格式的证书文件路径
//   - keyFile: PEM 格式的私钥文件路径
//
// 返回值:
//   - Option: 服务器配置选项函数
//
// 示例:
//
//	s := server.New(
//	    server.WithTLS("server.crt", "server.key"),
//	)
//	s.Start() // 使用 HTTPS 启动
func WithTLS(certFile, keyFile string) Option {
	return func(s *GinServer) {
		s.certFile = certFile
		s.keyFile = keyFile
	}
}

// Package gin 提供 Gin HTTP 服务器的自动配置。
//
// 当 gin.enabled=true 时自动启用，从 Environment 中读取 gin.host、gin.mode、
// gin.read-timeout、gin.write-timeout、gin.idle-timeout 等配置项，
// 创建并注册 Gin Server Bean 到 IoC 容器中（Bean ID: ginServer）。
package gin

import (
	"time"

	"github.com/xudefa/go-boot-gin/server"
	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
)

// init 注册 Gin 自动配置，由 gin.enabled=true 条件控制。
func init() {
	boot.RegisterAutoConfig(&GinAutoConfiguration{},
		condition.OnProperty(constants.GinEnabled, constants.ConditionTrue),
	)
}

// GinAutoConfiguration Gin HTTP 服务器的自动配置。
//
// 从 Environment 中读取以下配置项：
//   - gin.host: 监听地址（默认 "0.0.0.0"）
//   - gin.mode: Gin 运行模式（debug/release/test）
//   - gin.port: 监听端口（默认 8080）
//   - gin.read-timeout: 读超时（秒，默认 10）
//   - gin.write-timeout: 写超时（秒，默认 10）
//   - gin.idle-timeout: 空闲超时（秒，默认 60）
//   - gin.shutdown-timeout: 优雅关闭超时（秒，默认 5）
//
// 创建 Gin Server 实例并注册到 IoC 容器中。
// 启用条件：gin.enabled=true
type GinAutoConfiguration struct{}

// Configure 执行自动配置逻辑，创建 Gin Server 并注册为 Bean。
func (g *GinAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	opts := []server.Option{
		server.WithContainer(ctx.Container()),
	}

	if host := env.GetString(constants.GinHost, ""); host != "" {
		opts = append(opts, server.WithHost(host))
	}
	if mode := env.GetString(constants.GinMode, ""); mode != "" {
		opts = append(opts, server.WithMode(mode))
	}

	// 端口配置：优先使用 gin.port，回退到 server.port
	port := env.GetInt("gin.port", 0)
	if port == 0 {
		port = env.GetInt(constants.ServerPort, 8080)
	}
	opts = append(opts, server.WithPort(port))

	opts = append(opts,
		server.WithReadTimeout(time.Duration(env.GetInt(constants.GinReadTimeout, constants.DefaultGinReadTimeout))*time.Second),
		server.WithWriteTimeout(time.Duration(env.GetInt(constants.GinWriteTimeout, constants.DefaultGinWriteTimeout))*time.Second),
		server.WithIdleTimeout(time.Duration(env.GetInt(constants.GinIdleTimeout, constants.DefaultGinIdleTimeout))*time.Second),
	)

	// 优雅关闭超时
	shutdownTimeout := env.GetInt("gin.shutdown-timeout", 5)
	opts = append(opts, server.WithShutdownTimeout(time.Duration(shutdownTimeout)*time.Second))

	s := server.New(opts...)

	if err := ctx.Register(constants.GinServerBeanID,
		core.Bean(s),
		core.Singleton(),
	); err != nil {
		return err
	}

	return nil
}

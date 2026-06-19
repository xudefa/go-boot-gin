package server

import (
	"context"
	stdnet "net"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/xudefa/go-boot/core"
	bootnet "github.com/xudefa/go-boot/net"
)

// TestNew 测试默认方式创建 Gin 服务器，验证 engine 和 container 不为空
func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.engine == nil {
		t.Error("engine should not be nil")
	}
	if s.container == nil {
		t.Error("container should not be nil")
	}
}

// TestNewWithOptions 测试使用 WithContainer 和 WithMode 等选项创建服务器，验证选项生效
func TestNewWithOptions(t *testing.T) {
	container := core.New()
	s := New(
		WithContainer(container),
		WithMode(gin.ReleaseMode),
	)
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.container != container {
		t.Error("container should be set")
	}
	if s.mode != gin.ReleaseMode {
		t.Errorf("expected mode %s, got %s", gin.ReleaseMode, s.mode)
	}
}

// TestNewWithPort 测试使用 WithPort 选项
func TestNewWithPort(t *testing.T) {
	s := New(WithPort(9090))
	if s.port != 9090 {
		t.Errorf("expected port 9090, got %d", s.port)
	}
}

// TestNewWithHost 测试使用 WithHost 选项
func TestNewWithHost(t *testing.T) {
	s := New(WithHost("0.0.0.0"))
	if s.host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", s.host)
	}
}

// TestNewWithTimeouts 测试使用各种超时选项
func TestNewWithTimeouts(t *testing.T) {
	timeout := 10 * time.Second
	s := New(
		WithReadTimeout(timeout),
		WithWriteTimeout(timeout),
		WithIdleTimeout(timeout),
		WithShutdownTimeout(timeout),
	)

	if s.readTimeout != timeout {
		t.Errorf("expected read timeout %v, got %v", timeout, s.readTimeout)
	}
	if s.writeTimeout != timeout {
		t.Errorf("expected write timeout %v, got %v", timeout, s.writeTimeout)
	}
	if s.idleTimeout != timeout {
		t.Errorf("expected idle timeout %v, got %v", timeout, s.idleTimeout)
	}
	if s.shutdownTimeout != timeout {
		t.Errorf("expected shutdown timeout %v, got %v", timeout, s.shutdownTimeout)
	}
}

// TestNewWithTLS 测试 TLS 配置选项
func TestNewWithTLS(t *testing.T) {
	s := New(WithTLS("server.crt", "server.key"))
	if s.certFile != "server.crt" {
		t.Errorf("expected certFile server.crt, got %s", s.certFile)
	}
	if s.keyFile != "server.key" {
		t.Errorf("expected keyFile server.key, got %s", s.keyFile)
	}
}

// TestServer_Use 测试添加路由级别中间件，验证中间件列表长度正确
func TestServer_Use(t *testing.T) {
	s := New()
	s.Use(func(c *gin.Context) {})
	if len(s.middleware) != 1 {
		t.Errorf("expected 1 middleware, got %d", len(s.middleware))
	}
}

// TestServer_UseGlobal 测试添加全局中间件，验证全局中间件列表长度正确
func TestServer_UseGlobal(t *testing.T) {
	s := New()
	// 默认有 4 个全局中间件
	initialLen := len(s.globalMiddleware)
	s.UseGlobal(func(c *gin.Context) {})
	if len(s.globalMiddleware) != initialLen+1 {
		t.Errorf("expected %d globalMiddleware, got %d", initialLen+1, len(s.globalMiddleware))
	}
}

// TestServer_GlobalMiddlewareList 测试获取全局中间件列表
func TestServer_GlobalMiddlewareList(t *testing.T) {
	s := New()
	list := s.GlobalMiddlewareList()
	if list == nil {
		t.Error("GlobalMiddlewareList() should not return nil")
	}
	if len(list) < 4 {
		t.Errorf("expected at least 4 default global middlewares, got %d", len(list))
	}
}

// TestServer_Register 测试注册 IoC 容器初始化函数，验证注册函数列表长度正确
func TestServer_Register(t *testing.T) {
	s := New()
	testFn := func(c core.Container) error {
		return nil
	}
	s.Register(testFn)
	if len(s.registerFuncs) != 1 {
		t.Errorf("expected 1 registerFunc, got %d", len(s.registerFuncs))
	}
}

// TestServer_Container 测试获取 IoC 容器，验证返回的容器与设置的容器一致
func TestServer_Container(t *testing.T) {
	container := core.New()
	s := New(WithContainer(container))
	result := s.Container()
	if result != container {
		t.Error("Container() should return the set container")
	}
}

// TestServer_Engine 测试获取 Gin 原生 Engine，验证不为空
func TestServer_Engine(t *testing.T) {
	s := New()
	engine := s.Engine()
	if engine == nil {
		t.Error("Engine() should return gin engine")
	}
}

// TestServer_GET 测试注册 GET 路由，验证返回自身以支持链式调用
func TestServer_GET(t *testing.T) {
	s := New()
	result := s.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("GET() should return self for chaining")
	}
}

// TestServer_POST 测试注册 POST 路由，验证返回自身以支持链式调用
func TestServer_POST(t *testing.T) {
	s := New()
	result := s.POST("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("POST() should return self for chaining")
	}
}

// TestServer_PUT 测试注册 PUT 路由，验证返回自身以支持链式调用
func TestServer_PUT(t *testing.T) {
	s := New()
	result := s.PUT("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("PUT() should return self for chaining")
	}
}

// TestServer_DELETE 测试注册 DELETE 路由，验证返回自身以支持链式调用
func TestServer_DELETE(t *testing.T) {
	s := New()
	result := s.DELETE("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("DELETE() should return self for chaining")
	}
}

// TestServer_PATCH 测试注册 PATCH 路由
func TestServer_PATCH(t *testing.T) {
	s := New()
	result := s.PATCH("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("PATCH() should return self for chaining")
	}
}

// TestServer_HEAD 测试注册 HEAD 路由
func TestServer_HEAD(t *testing.T) {
	s := New()
	result := s.HEAD("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("HEAD() should return self for chaining")
	}
}

// TestServer_OPTIONS 测试注册 OPTIONS 路由
func TestServer_OPTIONS(t *testing.T) {
	s := New()
	result := s.OPTIONS("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("OPTIONS() should return self for chaining")
	}
}

// TestServer_Any 测试注册任意方法路由，验证返回自身以支持链式调用
func TestServer_Any(t *testing.T) {
	s := New()
	result := s.Any("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})
	if result != s {
		t.Error("Any() should return self for chaining")
	}
}

// TestServer_Group 测试创建路由组，验证返回路由组
func TestServer_Group(t *testing.T) {
	s := New()
	group := s.Group("/api")
	if group == nil {
		t.Error("Group() should return router group")
	}
}

// TestServer_Stop 测试停止服务器（未启动时）
func TestServer_Stop(t *testing.T) {
	s := New()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 未启动时 Stop 应该返回 nil
	err := s.Stop(ctx)
	if err != nil {
		t.Errorf("Stop() on unstarted server should return nil, got %v", err)
	}
}

// TestServer_Stop_WithServer 测试停止已启动的服务器
func TestServer_Stop_WithServer(t *testing.T) {
	// 使用随机可用端口
	listener, err := stdnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("Cannot create listener for test")
	}
	port := listener.Addr().(*stdnet.TCPAddr).Port
	_ = listener.Close()

	s := New(WithPort(port), WithHost("127.0.0.1"))

	// 在 goroutine 中启动服务器
	startDone := make(chan error, 1)
	go func() {
		startDone <- s.Start()
	}()

	// 等待服务器启动
	time.Sleep(500 * time.Millisecond)

	// 停止服务器
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = s.Stop(ctx)
	if err != nil {
		t.Logf("Server stopped with error: %v", err)
	}

	// 等待 Start() 返回
	select {
	case startErr := <-startDone:
		if startErr != nil {
			t.Logf("Start() returned error: %v", startErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Start() did not return after Stop()")
	}
}

// TestServer_Start_InvalidPort 测试在无效端口上启动服务器
func TestServer_Start_InvalidPort(t *testing.T) {
	// 创建一个已经被占用的端口
	listener, err := stdnet.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("Cannot create listener for test")
	}
	defer func() { _ = listener.Close() }()

	addr := listener.Addr().(*stdnet.TCPAddr)

	// 创建另一个服务器尝试使用相同端口
	s := New(WithPort(addr.Port), WithHost("127.0.0.1"))

	// 启动服务器，这应该失败
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.Start()
	}()

	select {
	case err := <-errCh:
		if err == nil {
			t.Error("expected error when starting server on occupied port")
		}
		if !strings.Contains(err.Error(), "address already in use") && !strings.Contains(err.Error(), "bind") {
			t.Logf("Got unexpected error message: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for server start error")
	}
}

// TestWrapHandlers_UnsupportedType 测试 wrapHandlers 对不支持类型的处理
func TestWrapHandlers_UnsupportedType(t *testing.T) {
	s := New()
	// 传入不支持的类型，应该被忽略
	result := s.wrapHandlers([]any{"invalid handler", 123})
	if len(result) != 0 {
		t.Errorf("expected 0 handlers for unsupported types, got %d", len(result))
	}
}

// TestWrapHandlers_SupportedTypes 测试 wrapHandlers 对支持类型的处理
func TestWrapHandlers_SupportedTypes(t *testing.T) {
	s := New()
	handler := func(c *gin.Context) {
		c.String(200, "ok")
	}

	result := s.wrapHandlers([]any{
		gin.HandlerFunc(handler),
		handler,
		HandlerFunc(handler),
	})

	if len(result) != 3 {
		t.Errorf("expected 3 handlers, got %d", len(result))
	}
}

// TestAdaptMiddleware 测试中间件适配
func TestAdaptMiddleware(t *testing.T) {
	netMiddleware := func(ctx bootnet.HandlerContext) {
		// 测试中间件被调用
	}

	ginHandler := AdaptMiddleware(netMiddleware)
	if ginHandler == nil {
		t.Fatal("AdaptMiddleware returned nil")
	}
}

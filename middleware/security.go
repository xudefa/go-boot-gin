package middleware

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/go-boot-gin/server"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/security"
)

// GinSecurityRequest Gin安全请求适配器
// 将Gin框架的*gin.Context适配为security.SecurityRequest接口
type GinSecurityRequest struct {
	c *gin.Context
}

// NewGinSecurityRequest 创建Gin安全请求适配器
func NewGinSecurityRequest(c *gin.Context) *GinSecurityRequest {
	return &GinSecurityRequest{c: c}
}

// GetMethod 获取HTTP请求方法
func (r *GinSecurityRequest) GetMethod() string {
	return r.c.Request.Method
}

// GetURI 获取请求URI路径
func (r *GinSecurityRequest) GetURI() string {
	return r.c.Request.URL.Path
}

// GetHeader 获取请求头值
func (r *GinSecurityRequest) GetHeader(key string) string {
	return r.c.GetHeader(key)
}

// SetAttribute 设置请求属性
func (r *GinSecurityRequest) SetAttribute(key string, value any) {
	r.c.Set(key, value)
}

// GetAttribute 获取请求属性
func (r *GinSecurityRequest) GetAttribute(key string) (any, bool) {
	val, exists := r.c.Get(key)
	return val, exists
}

// GinSecurityResponse Gin安全响应适配器
// 将Gin框架的响应适配为security.SecurityResponse接口
type GinSecurityResponse struct {
	c *gin.Context
}

// NewGinSecurityResponse 创建Gin安全响应适配器
func NewGinSecurityResponse(c *gin.Context) *GinSecurityResponse {
	return &GinSecurityResponse{c: c}
}

// SetStatusCode 设置HTTP响应状态码
func (r *GinSecurityResponse) SetStatusCode(code int) {
	r.c.Status(code)
}

// SetHeader 设置响应头
func (r *GinSecurityResponse) SetHeader(key, value string) {
	r.c.Header(key, value)
}

// Write 写入响应体
func (r *GinSecurityResponse) Write(data []byte) error {
	_, err := r.c.Writer.Write(data)
	return err
}

// GinSecurityFilterChain Gin安全过滤器链
// 包装security.SecurityFilterChain以适配Gin框架
type GinSecurityFilterChain struct {
	securityChain security.SecurityFilterChain
}

// NewGinSecurityFilterChain 创建Gin安全过滤器链
func NewGinSecurityFilterChain(securityChain security.SecurityFilterChain) *GinSecurityFilterChain {
	return &GinSecurityFilterChain{securityChain: securityChain}
}

// DoFilter 执行安全过滤器链
func (f *GinSecurityFilterChain) DoFilter(ctx context.Context, request security.SecurityRequest, response security.SecurityResponse) error {
	return f.securityChain.DoFilter(ctx, request, response)
}

// GinSecurityMiddleware Gin安全中间件
// 将安全过滤器链转换为Gin中间件
type GinSecurityMiddleware struct {
	securityChain security.SecurityFilterChain
}

// NewGinSecurityMiddleware 创建Gin安全中间件
func NewGinSecurityMiddleware(securityChain security.SecurityFilterChain) *GinSecurityMiddleware {
	return &GinSecurityMiddleware{securityChain: securityChain}
}

// HandlerFunc 返回Gin.HandlerFunc格式的中间件处理函数
func (m *GinSecurityMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		request := NewGinSecurityRequest(c)
		response := NewGinSecurityResponse(c)

		securityChain := NewGinSecurityFilterChain(m.securityChain)

		err := securityChain.DoFilter(c.Request.Context(), request, response)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
			return
		}

		c.Next()
	}
}

// WithSecurity 创建Gin服务器的security配置选项
// 参数securityChain: 安全过滤器链，通常通过security.NewHttpSecurity().Build()构建
// 返回Option函数，可传入server.New()进行配置
//
// 使用示例:
//
//	import (
//	    "github.com/xudefa/go-boot/security"
//	    "github.com/xudefa/go-boot-gin/server"
//	    "github.com/xudefa/go-boot-gin/middleware"
//	)
//
//	// 1. 创建用户服务
//	userDetailsService := security.NewInMemoryUserDetailsService()
//	userDetailsService.CreateUser("admin", "password", []string{"ROLE_ADMIN"})
//
//	// 2. 创建密码编码器和认证提供者
//	passwordEncoder := security.NewBCryptPasswordEncoder(10)
//	authProvider := security.NewDaoAuthenticationProvider(userDetailsService, passwordEncoder)
//	authManager := security.NewProviderManager(authProvider)
//
//	// 3. 创建安全元数据源
//	metadataSource := security.NewExpressionBasedFilterInvocationSecurityMetadataSource()
//	metadataSource.AddMapping("/public/**", []string{"permitAll"})
//	metadataSource.AddMapping("/admin/**", []string{"hasRole('ADMIN')"})
//
//	// 4. 构建安全过滤器链
//	httpSecurity := security.NewHttpSecurity()
//	httpSecurity.AuthenticationManager(authManager)
//	httpSecurity.SecurityMetadataSource(metadataSource)
//	httpSecurity.Anonymous()
//	chain, _ := httpSecurity.Build()
//
//	// 5. 创建Gin服务器并应用安全配置
//	s := server.New(
//	    server.WithContainer(container),
//	    middleware.WithSecurity(chain),
//	)
//	s.GET("/public/test", func(c *gin.Context) {
//	    c.String(200, "public endpoint")
//	})
//	s.GET("/admin/test", func(c *gin.Context) {
//	    c.String(200, "admin endpoint")
//	})
func WithSecurity(securityChain security.SecurityFilterChain) server.Option {
	return func(s *server.GinServer) {
		securityMiddleware := NewGinSecurityMiddleware(securityChain)
		s.UseGlobal(securityMiddleware.HandlerFunc())
	}
}

// WithSecurityFromContainer 从容器中获取安全过滤器链并应用
// 参数container: go-boot IoC容器
// 返回Option函数
//
// 使用示例:
//
//	s := server.New(
//	    server.WithContainer(appCtx.Container()),
//	    middleware.WithSecurityFromContainer(appCtx.Container()),
//	)
func WithSecurityFromContainer(container core.Container) server.Option {
	return func(s *server.GinServer) {
		bean, err := container.Get(constants.SecurityFilterChainBeanID)
		if err != nil {
			log.Printf("Warning: failed to get security filter chain from container: %v", err)
			return
		}
		securityChain, ok := bean.(security.SecurityFilterChain)
		if !ok || securityChain == nil {
			log.Printf("Warning: security filter chain bean is nil or invalid type")
			return
		}
		securityMiddleware := NewGinSecurityMiddleware(securityChain)
		s.UseGlobal(securityMiddleware.HandlerFunc())
	}
}

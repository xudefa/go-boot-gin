package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/xudefa/go-boot/tracing"
)

// ginCarrier 实现 tracing.TextMapCarrier 接口，用于在 Gin 请求中提取和注入追踪上下文
type ginCarrier struct {
	c *gin.Context
}

// Get 获取指定键的 header 值
func (c *ginCarrier) Get(key string) string {
	return c.c.GetHeader(key)
}

// Set 设置指定键的 header 值
func (c *ginCarrier) Set(key string, value string) {
	c.c.Request.Header.Set(key, value)
}

// Keys 返回所有 header 键的列表
func (c *ginCarrier) Keys() []string {
	keys := make([]string, 0)
	for k := range c.c.Request.Header {
		keys = append(keys, k)
	}
	return keys
}

// GetTraceID 从 Gin Context 中提取当前 Span 的 TraceID
// 返回空字符串如果没有有效的追踪上下文
func GetTraceID(c *gin.Context) string {
	return getTraceIDFromContext(c.Request.Context())
}

// GetSpanID 从 Gin Context 中提取当前 Span 的 SpanID
// 返回空字符串如果没有有效的追踪上下文
func GetSpanID(c *gin.Context) string {
	return getSpanIDFromContext(c.Request.Context())
}

func getTraceIDFromContext(ctx context.Context) string {
	span := tracing.SpanFromContext(ctx)
	if span == nil || span.GetTraceID() == "" {
		return ""
	}
	return span.GetTraceID()
}

func getSpanIDFromContext(ctx context.Context) string {
	span := tracing.SpanFromContext(ctx)
	if span == nil || span.GetSpanID() == "" {
		return ""
	}
	return span.GetSpanID()
}

// AddTraceToResponseHeaders 将 TraceID 和 SpanID 添加到响应头中
// 便于客户端获取追踪信息进行问题排查
func AddTraceToResponseHeaders(c *gin.Context) {
	traceID := GetTraceID(c)
	spanID := GetSpanID(c)
	if traceID != "" {
		c.Header("X-Trace-ID", traceID)
	}
	if spanID != "" {
		c.Header("X-Span-ID", spanID)
	}
}

// AddTraceToResponseHeadersWithContext 使用指定的上下文将 TraceID 和 SpanID 添加到响应头中
func AddTraceToResponseHeadersWithContext(ctx context.Context, c *gin.Context) {
	traceID := getTraceIDFromContext(ctx)
	spanID := getSpanIDFromContext(ctx)
	if traceID != "" {
		c.Header("X-Trace-ID", traceID)
	}
	if spanID != "" {
		c.Header("X-Span-ID", spanID)
	}
}

// TraceIDMiddleware 简单的中间件，仅将追踪 ID 添加到响应头
// 适用于不需要完整追踪功能但需要暴露追踪 ID 的场景
func TraceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		AddTraceToResponseHeaders(c)
		c.Next()
	}
}

// HTTPServerTracingMiddleware 创建 Gin HTTP 服务器端追踪中间件
// serviceName 参数用于标识服务名称，默认为 "gin-server"
//
// 该中间件提供以下功能：
// 1. 从请求头中提取父追踪上下文
// 2. 创建服务端 Span，记录 HTTP 方法、路径、主机等信息
// 3. 将 Span 上下文注入到请求中
// 4. 在请求结束时记录响应状态码和错误状态
// 5. 将 TraceID/SpanID 添加到响应头
func HTTPServerTracingMiddleware(serviceName ...string) gin.HandlerFunc {
	tracerName := "gin-server"
	if len(serviceName) > 0 {
		tracerName = serviceName[0]
	}

	return func(c *gin.Context) {
		carrier := tracing.NewHTTPHeadersCarrier(c.Request.Header)
		parentCtx := tracing.ExtractTraceContext(context.Background(), carrier)

		spanName := c.Request.URL.Path
		if spanName == "" {
			spanName = "HTTP " + c.Request.Method
		}

		tracer := tracing.GetTracer(tracerName)
		ctx, span := tracer.Start(parentCtx, spanName,
			tracing.WithSpanKind(tracing.SpanKindServer),
			tracing.WithAttribute("http.method", c.Request.Method),
			tracing.WithAttribute("http.target", c.Request.URL.Path),
			tracing.WithAttribute("http.host", c.Request.Host),
		)

		traceID := span.GetTraceID()
		spanID := span.GetSpanID()

		c.Header("X-Trace-ID", traceID)
		c.Header("X-Span-ID", spanID)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		statusCode := c.Writer.Status()
		span.SetAttribute("http.status_code", statusCode)

		if statusCode >= http.StatusInternalServerError {
			span.SetStatus(tracing.SpanStatusError)
		} else {
			span.SetStatus(tracing.SpanStatusOK)
		}

		span.End()
	}
}

// InjectTraceHeaders 将当前追踪上下文注入到请求头中
// 用于客户端向外发起请求时传递追踪信息
func InjectTraceHeaders(c *gin.Context) {
	tracing.InjectTraceContext(c.Request.Context(), &ginCarrier{c: c})
}

// HTTPClientTracingMiddleware 创建 Gin HTTP 客户端追踪中间件
// 用于在向外发起 HTTP 请求时注入追踪上下文
func HTTPClientTracingMiddleware(serviceName ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		InjectTraceHeaders(c)
		c.Next()
	}
}

// 编译期类型断言，确保 ginCarrier 实现了 TextMapCarrier 接口
var _ tracing.TextMapCarrier = (*ginCarrier)(nil)

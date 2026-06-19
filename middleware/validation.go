package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/xudefa/go-boot/validation"
)

// ValidationMiddleware 创建 Gin 验证中间件
func ValidationMiddleware(validator validation.Validator) gin.HandlerFunc {
	return ValidationMiddlewareWithConfig(&validation.MiddlewareConfig{
		Validator: validator,
	})
}

// ValidationMiddlewareWithGroups 创建带验证组的 Gin 中间件
func ValidationMiddlewareWithGroups(validator validation.Validator, groups ...string) gin.HandlerFunc {
	return ValidationMiddlewareWithConfig(&validation.MiddlewareConfig{
		Validator: validator,
		Groups:    groups,
	})
}

// ValidationMiddlewareWithConfig 创建自定义配置的 Gin 中间件
func ValidationMiddlewareWithConfig(config *validation.MiddlewareConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if shouldSkipGinPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		obj := getGinRequestObject(c)
		if obj == nil {
			c.Next()
			return
		}

		var err error
		if groupedValidator, ok := config.Validator.(*validation.GroupedTagValidator); ok && len(config.Groups) > 0 {
			err = groupedValidator.ValidateWithGroups(obj, config.Groups...)
		} else {
			err = config.Validator.Validate(obj)
		}

		if err != nil {
			if config.ErrorHandler != nil {
				config.ErrorHandler(c, err)
			} else {
				defaultGinErrorHandler(c, err)
			}
			c.Abort()
			return
		}

		c.Next()
	}
}

// shouldSkipGinPath 检查是否应该跳过 Gin 路径，支持通配符模式匹配
func shouldSkipGinPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
		if matched, _ := filepath.Match(skipPath, path); matched {
			return true
		}
	}
	return false
}

// getGinRequestObject 从 Gin 上下文获取请求对象
func getGinRequestObject(c *gin.Context) any {
	if c.Request.Method == "GET" {
		return c.Request.URL.Query()
	}

	if c.Request.Body == nil {
		return nil
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil
	}

	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var obj any
	if err := json.Unmarshal(body, &obj); err != nil {
		return nil
	}

	return obj
}

// defaultGinErrorHandler 默认 Gin 错误处理器
func defaultGinErrorHandler(c *gin.Context, err error) {
	c.JSON(400, gin.H{
		"error": err.Error(),
	})
}

// BindAndValidate 绑定并验证请求
func BindAndValidate(c *gin.Context, obj any, validator validation.Validator) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}

	if validator != nil {
		return validator.Validate(obj)
	}

	return nil
}

// BindAndValidateWithGroups 绑定并验证请求（带验证组）
func BindAndValidateWithGroups(c *gin.Context, obj any, validator validation.Validator, groups ...string) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return err
	}

	if groupedValidator, ok := validator.(*validation.GroupedTagValidator); ok && len(groups) > 0 {
		return groupedValidator.ValidateWithGroups(obj, groups...)
	}

	if validator != nil {
		return validator.Validate(obj)
	}

	return nil
}

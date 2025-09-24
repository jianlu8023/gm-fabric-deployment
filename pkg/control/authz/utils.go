package authz

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthzMiddleware 创建权限控制中间件
// @param control *Control 权限控制器
// @return gin.HandlerFunc Gin中间件
func AuthzMiddleware(control *Control) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果权限控制未启用，直接通过
		if !control.config.Enabled {
			c.Next()
			return
		}

		// 获取用户信息
		user := getUserFromContext(c)
		if user == "" {
			// 如果没有用户信息，拒绝访问
			control.logger.Warnf("[authz] no user information found in request")
			handleUnauthorized(c)
			return
		}

		// 获取资源和操作
		obj := getResourceFromRequest(c)
		act := c.Request.Method

		// 检查权限
		ok, err := control.CheckPermission(user, obj, act)
		if err != nil {
			control.logger.Errorf("[authz] permission check error: %v", err)
			handleUnauthorized(c)
			return
		}

		if !ok {
			control.logger.Warnf("[authz] permission denied: user=%s, resource=%s, action=%s", user, obj, act)
			handleUnauthorized(c)
			return
		}

		// 权限检查通过，继续处理请求
		c.Next()
	}
}

// getUserFromContext 从请求上下文中获取用户信息
// @param c *gin.Context Gin上下文
// @return string 用户标识
func getUserFromContext(c *gin.Context) string {
	// 这里是一个示例实现，实际应用中需要根据项目的认证方式获取用户信息
	// 例如从JWT token、session或header中获取

	// 示例1：从header中获取
	if user := c.GetHeader("X-User-Id"); user != "" {
		return user
	}

	// 示例2：从context中获取（假设在其他中间件中设置了用户信息）
	if user, exists := c.Get("userId"); exists {
		if userId, ok := user.(string); ok {
			return userId
		}
	}

	// 示例3：从query参数中获取（仅用于测试，不推荐在生产环境中使用）
	if user := c.Query("user_id"); user != "" {
		return user
	}

	return ""
}

// getResourceFromRequest 从请求中获取资源标识
// @param c *gin.Context Gin上下文
// @return string 资源标识
func getResourceFromRequest(c *gin.Context) string {
	// 获取请求路径作为资源标识
	path := c.Request.URL.Path

	// 可以根据需要对路径进行处理，例如移除上下文路径、参数等
	// 这里简单返回原始路径
	return path
}

// handleUnauthorized 处理未授权的请求
// @param c *gin.Context Gin上下文
func handleUnauthorized(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    http.StatusUnauthorized,
		"message": "Unauthorized",
		"data":    nil,
	})
	c.Abort()
}

// AuthzRoleMiddleware 创建基于角色的权限控制中间件
// @param control *Control 权限控制器
// @param requiredRoles ...string 所需角色列表
// @return gin.HandlerFunc Gin中间件
func AuthzRoleMiddleware(control *Control, requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果权限控制未启用，直接通过
		if !control.config.Enabled {
			c.Next()
			return
		}

		// 获取用户信息
		user := getUserFromContext(c)
		if user == "" {
			control.logger.Warnf("[authz] no user information found in request")
			handleUnauthorized(c)
			return
		}

		// 获取用户的所有角色
		roles, err := control.GetRolesForUser(user)
		if err != nil {
			control.logger.Errorf("[authz] get roles error: %v", err)
			handleUnauthorized(c)
			return
		}

		// 检查用户是否有任一所需角色
		hasRequiredRole := false
		for _, role := range roles {
			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			control.logger.Warnf("[authz] role denied: user=%s, required_roles=%s", user, strings.Join(requiredRoles, ","))
			handleUnauthorized(c)
			return
		}

		// 角色检查通过，继续处理请求
		c.Next()
	}
}

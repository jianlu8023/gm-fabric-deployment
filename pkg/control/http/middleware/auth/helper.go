package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/session"
)

// 认证错误常量
//
// @description Authenticate 方法返回的统一错误，便于上层判断
var (
	// errUnauthorized 未提供认证信息
	errUnauthorized = errors.New("未提供认证信息")
	// errSessionInvalid 会话已失效
	errSessionInvalid = errors.New("会话已失效")
)

// CurrentSession 从 gin.Context 获取当前会话
//
// @description 读取认证中间件注入的 *session.Session；未认证时返回 nil
// @param ctx *gin.Context Gin上下文
// @return *session.Session 会话信息
func CurrentSession(ctx *gin.Context) *session.Session {
	v, exists := ctx.Get(currentSessionKey)
	if !exists || v == nil {
		return nil
	}
	sess, ok := v.(*session.Session)
	if !ok {
		return nil
	}
	return sess
}

// CurrentUserID 从 gin.Context 获取当前用户ID
//
// @description 读取认证中间件注入的 UserID；未认证时返回空字符串
// @param ctx *gin.Context Gin上下文
// @return string 用户ID
func CurrentUserID(ctx *gin.Context) string {
	sess := CurrentSession(ctx)
	if sess == nil {
		return ""
	}
	return sess.UserID
}

// CurrentRole 从 gin.Context 获取当前用户角色
//
// @description 读取认证中间件注入的 Role；未认证时返回空字符串
// @param ctx *gin.Context Gin上下文
// @return string 用户角色
func CurrentRole(ctx *gin.Context) string {
	sess := CurrentSession(ctx)
	if sess == nil {
		return ""
	}
	return sess.Role
}

// HasPermission 判断当前会话是否持有指定权限
//
// @description 从 *session.Session.PermissionList 判断；未认证时返回 false
// @param ctx *gin.Context Gin上下文
// @param perm string 权限标识
// @return bool 是否持有
func HasPermission(ctx *gin.Context, perm string) bool {
	sess := CurrentSession(ctx)
	if sess == nil {
		return false
	}
	for _, p := range sess.PermissionList {
		if p == perm {
			return true
		}
	}
	return false
}

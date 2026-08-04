package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/jwt"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/session"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"
)

// Middleware 必选认证中间件
//
// @description 替代原 jwt.EnableJWT；按认证模式分流：jwt_session 走 token+会话双校验，jwt_only 只校验 token，session_only 只校验会话
// @return gin.HandlerFunc Gin中间件函数
func (m *authManagerImpl) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "auth")
		defer span.End()
		// 把 trace ctx 写回 ctx.Request，避免链路断裂（修正原 jwt 包的遗漏）
		ctx.Request = ctx.Request.WithContext(tCtx)

		var sess *session.Session
		var ok bool
		switch m.mode {
		case modeJWTSession:
			sess, ok = m.authJWTSession(ctx)
		case modeJWTOnly:
			sess, ok = m.authJWTOnly(ctx)
		case modeSessionOnly:
			sess, ok = m.authSessionOnly(ctx)
		}
		if !ok {
			// 失败响应已在各 authXxx 内完成，此处仅收尾 span
			span.SetStatus(codes.Error, "认证失败")
			return
		}

		// 注入上下文：当前会话 + 各业务字段（与 session 包常量保持一致）
		ctx.Set(currentSessionKey, sess)
		ctx.Set(session.UserId, sess.UserID)
		ctx.Set(session.UserName, sess.Username)
		ctx.Set(session.UserRole, sess.Role)
		ctx.Set(session.UserType, sess.UserType)
		ctx.Set(session.PermissionList, sess.PermissionList)
		ctx.Set(session.ClientIP, sess.ClientIP)
		ctx.Set(session.UserAgent, sess.UserAgent)
		ctx.Set(session.DeviceID, sess.DeviceID)
		ctx.Set(session.Scope, sess.Scope)
		ctx.Set(session.TenantID, sess.TenantID)
		ctx.Set(session.SessionId, sess.SessionID)
		ctx.Set(jwt.UserClaims, sess) // 兼容旧代码通过 UserClaims 取用户信息

		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}

// OptionalMiddleware 可选认证中间件
//
// @description 无 token 时放行（不注入用户信息），有 token 则解析并注入；适用于匿名也可访问、登录后享特权的接口
// @return gin.HandlerFunc Gin中间件函数
func (m *authManagerImpl) OptionalMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tCtx, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "authOptional")
		defer span.End()
		ctx.Request = ctx.Request.WithContext(tCtx)

		// 无 token 直接放行
		if !m.hasToken(ctx) {
			ctx.Next()
			span.SetStatus(codes.Ok, "anonymous")
			return
		}

		// 有 token 则走完整认证，失败时返回 401（不允许匿名放行带无效 token）
		m.Middleware()(ctx)
	}
}

// hasToken 判断请求是否携带 token/sessionID
//
// @description 用于 OptionalMiddleware 判断是否放行匿名请求
// @param ctx *gin.Context Gin上下文
// @return bool 是否携带 token
func (m *authManagerImpl) hasToken(ctx *gin.Context) bool {
	header := m.resolveTokenHeader()
	if tokenString := ctx.GetHeader(header); !stringer.IsBlank(tokenString) {
		return true
	}
	// session_only 模式下可能从 cookie/query 取
	if m.mode == modeSessionOnly {
		if cookieName := m.resolveCookieName(); cookieName != "" {
			if _, err := ctx.Cookie(cookieName); err == nil {
				return true
			}
		}
		if sid := ctx.Query(m.resolveSessionIDQueryParam()); !stringer.IsBlank(sid) {
			return true
		}
	}
	return false
}

// authJWTSession jwt_session 模式认证：token 签名校验 + 会话有效性校验 + 续期
//
// @description 解析 Bearer token → jwtManager.ParseToken → sessionStore.Validate → 更新 LastActivityTime
// @param ctx *gin.Context Gin上下文
// @return *session.Session 会话信息
// @return bool 是否认证通过
func (m *authManagerImpl) authJWTSession(ctx *gin.Context) (*session.Session, bool) {
	tokenString, errMsg, ok := m.extractBearerToken(ctx)
	if !ok {
		m.unauthorized(ctx, errMsg)
		return nil, false
	}

	claims, err := m.jwtManager.ParseToken(tokenString)
	if err != nil {
		m.logger.Errorf("[auth] jwt parse failed: %v", err)
		m.unauthorized(ctx, "认证信息无效或已过期")
		return nil, false
	}

	if !m.sessionStore.Validate(claims.SessionID) {
		m.logger.Errorf("[auth] session invalid: %s", claims.SessionID)
		m.unauthorized(ctx, "会话已失效，请重新登录")
		return nil, false
	}

	// 续期：更新最后活动时间，滑动过期模式下更新 ExpiresAt
	now := time.Now().Unix()
	sess := claimsToSession(claims)
	sess.LastActivityTime = now
	if m.cfg != nil && m.cfg.SlidingExpiration {
		sess.ExpiresAt = now + int64(m.getSessionTTL().Seconds())
	}
	if err = m.sessionStore.Set(sess.SessionID, sess); err != nil {
		// Set 失败仅记日志，不阻断请求（会话已 Validate 通过）
		m.logger.Errorf("[auth] session set failed: %v", err)
	}
	return sess, true
}

// authJWTOnly jwt_only 模式认证：仅校验 token 签名
//
// @description 解析 Bearer token → jwtManager.ParseToken；无服务端会话，无法主动失效
// @param ctx *gin.Context Gin上下文
// @return *session.Session 会话信息（由 claims 映射）
// @return bool 是否认证通过
func (m *authManagerImpl) authJWTOnly(ctx *gin.Context) (*session.Session, bool) {
	tokenString, errMsg, ok := m.extractBearerToken(ctx)
	if !ok {
		m.unauthorized(ctx, errMsg)
		return nil, false
	}

	claims, err := m.jwtManager.ParseToken(tokenString)
	if err != nil {
		m.logger.Errorf("[auth] jwt parse failed: %v", err)
		m.unauthorized(ctx, "认证信息无效或已过期")
		return nil, false
	}
	return claimsToSession(claims), true
}

// authSessionOnly session_only 模式认证：仅校验会话
//
// @description 按 token_source 从 header/cookie/query 取 sessionID → sessionStore.Get
// @param ctx *gin.Context Gin上下文
// @return *session.Session 会话信息
// @return bool 是否认证通过
func (m *authManagerImpl) authSessionOnly(ctx *gin.Context) (*session.Session, bool) {
	sessionID := m.extractSessionID(ctx)
	if stringer.IsBlank(sessionID) {
		m.unauthorized(ctx, "未提供认证信息")
		return nil, false
	}

	sess, err := m.sessionStore.Get(sessionID)
	if err != nil {
		m.logger.Errorf("[auth] session get failed: %v", err)
		m.unauthorized(ctx, "会话已失效，请重新登录")
		return nil, false
	}

	// 续期：LastActivityTime 始终更新；ExpiresAt 仅在滑动过期模式下更新
	now := time.Now().Unix()
	sess.LastActivityTime = now
	if m.cfg != nil && m.cfg.SlidingExpiration {
		sess.ExpiresAt = now + int64(m.getSessionTTL().Seconds())
	}
	// 无论是否滑动过期，都需将 LastActivityTime 更新写回存储，
	// 保持 sessionStore 与返回给调用方的 sess 一致（与 authJWTSession 行为对齐）
	if err = m.sessionStore.Set(sess.SessionID, sess); err != nil {
		m.logger.Errorf("[auth] session set failed on session_only: %v", err)
	}
	return sess, true
}

// extractBearerToken 从 Authorization 头提取 Bearer token
//
// @description 校验 "Bearer <token>" 格式，返回 token 与错误信息；不直接写响应，由调用方根据 errMsg 统一响应，避免重复写响应
// @param ctx *gin.Context Gin上下文
// @return string token 字符串
// @return string 错误信息（空表示无错误）
// @return bool 格式是否合法
func (m *authManagerImpl) extractBearerToken(ctx *gin.Context) (string, string, bool) {
	header := m.resolveTokenHeader()
	tokenString := ctx.GetHeader(header)
	if stringer.IsBlank(tokenString) {
		return "", "未提供认证信息", false
	}
	prefix := m.resolveTokenPrefix()
	if prefix == "" {
		return tokenString, "", true
	}
	parts := strings.SplitN(tokenString, " ", 2)
	if len(parts) != 2 || parts[0] != prefix {
		return "", "认证信息格式错误", false
	}
	return strings.TrimSpace(parts[1]), "", true
}

// extractSessionID 从 header/cookie/query 提取 sessionID（session_only 模式用）
//
// @description 按 token_source 配置决定取值位置：header 取 TokenHeader，cookie 取 CookieName，query 取 session_id
// @param ctx *gin.Context Gin上下文
// @return string sessionID
func (m *authManagerImpl) extractSessionID(ctx *gin.Context) string {
	source := m.resolveTokenSource()
	switch source {
	case "cookie":
		cookieName := m.resolveCookieName()
		if cookieName == "" {
			return ""
		}
		sid, err := ctx.Cookie(cookieName)
		if err != nil {
			return ""
		}
		return sid
	case "query":
		return ctx.Query(m.resolveSessionIDQueryParam())
	default: // header 或 auto
		return ctx.GetHeader(m.resolveTokenHeader())
	}
}

// unauthorized 返回 401 认证失败响应
//
// @description 统一 401 响应封装；AbortWithStatusJSON 后不再调用 ctx.Abort()（修正原 jwt 包冗余）
// @param ctx *gin.Context Gin上下文
// @param data 错误描述
func (m *authManagerImpl) unauthorized(ctx *gin.Context, data string) {
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, commonhttp.BaseResponse{
		Code:    http.StatusUnauthorized,
		Message: "业务处理失败",
		Success: false,
		Data:    data,
	})
}

// claimsToSession 将 jwt.Claims 映射为 session.Session
//
// @description 在 auth 包内部完成 jwt 与 session 两个模型的转换，避免 jwt 包反向依赖 session 包
// @param claims *jwt.Claims JWT 声明
// @return *session.Session 会话信息
func claimsToSession(claims *jwt.Claims) *session.Session {
	return &session.Session{
		SessionID:        claims.SessionID,
		UserID:           claims.UserID,
		Username:         claims.Username,
		Role:             claims.Role,
		UserType:         claims.UserType,
		PermissionList:   claims.PermissionList,
		ClientIP:         claims.ClientIP,
		UserAgent:        claims.UserAgent,
		DeviceID:         claims.DeviceID,
		Scope:            claims.Scope,
		TenantID:         claims.TenantID,
		LoginTime:        claims.LoginTime,
		LastActivityTime: claims.LastActivityTime,
		ExpiresAt:        claims.SessionExpires,
	}
}

// resolveTokenHeader 解析 token 头名称
//
// @description 优先用配置，否则默认 Authorization
// @return string token 头名称
func (m *authManagerImpl) resolveTokenHeader() string {
	if m.cfg != nil && m.cfg.TokenHeader != "" {
		return m.cfg.TokenHeader
	}
	return defaultTokenHeader
}

// resolveTokenPrefix 解析 token 前缀
//
// @description 优先用配置，否则默认 Bearer；空字符串表示无前缀
// @return string token 前缀
func (m *authManagerImpl) resolveTokenPrefix() string {
	if m.cfg != nil && m.cfg.TokenPrefix != "" {
		return m.cfg.TokenPrefix
	}
	return defaultTokenPrefix
}

// resolveTokenSource 解析 token 来源
//
// @description 优先用配置，否则默认 auto（jwt 模式下等价于 header）
// @return string token 来源
func (m *authManagerImpl) resolveTokenSource() string {
	if m.cfg != nil && m.cfg.TokenSource != "" {
		return m.cfg.TokenSource
	}
	return "auto"
}

// resolveCookieName 解析 cookie 名称
//
// @description 从 cfg.Session.CookieName 取，session_only + cookie 模式下使用
// @return string cookie 名称
func (m *authManagerImpl) resolveCookieName() string {
	if m.cfg != nil && m.cfg.Session != nil {
		return m.cfg.Session.CookieName
	}
	return ""
}

// resolveSessionIDQueryParam 解析 query 参数名
//
// @description session_only + query 模式下取 sessionID 的参数名，默认 session_id
// @return string query 参数名
func (m *authManagerImpl) resolveSessionIDQueryParam() string {
	return "session_id"
}

// RequireRole 角色门禁中间件
//
// @description 要求当前会话角色命中 roles 之一，否则 403；须在认证中间件之后使用
// @param roles ...string 允许的角色列表
// @return gin.HandlerFunc Gin中间件函数
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(ctx *gin.Context) {
		role := CurrentRole(ctx)
		if _, ok := allowed[role]; !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, commonhttp.BaseResponse{
				Code:    http.StatusForbidden,
				Message: "业务处理失败",
				Success: false,
				Data:    "权限不足",
			})
			return
		}
		ctx.Next()
	}
}

// RequirePermission 权限门禁中间件
//
// @description 要求当前会话权限列表包含 perms 之一，否则 403；须在认证中间件之后使用
// @param perms ...string 所需权限列表
// @return gin.HandlerFunc Gin中间件函数
func RequirePermission(perms ...string) gin.HandlerFunc {
	required := make(map[string]struct{}, len(perms))
	for _, p := range perms {
		required[p] = struct{}{}
	}
	return func(ctx *gin.Context) {
		sess := CurrentSession(ctx)
		if sess == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, commonhttp.BaseResponse{
				Code:    http.StatusUnauthorized,
				Message: "业务处理失败",
				Success: false,
				Data:    "未认证",
			})
			return
		}
		for _, p := range sess.PermissionList {
			if _, ok := required[p]; ok {
				ctx.Next()
				return
			}
		}
		ctx.AbortWithStatusJSON(http.StatusForbidden, commonhttp.BaseResponse{
			Code:    http.StatusForbidden,
			Message: "业务处理失败",
			Success: false,
			Data:    "权限不足",
		})
	}
}

// RequireTenant 多租户隔离中间件
//
// @description 要求当前会话 TenantID 等于指定租户，否则 403
// @param tenantID string 允许的租户ID
// @return gin.HandlerFunc Gin中间件函数
func RequireTenant(tenantID string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sess := CurrentSession(ctx)
		if sess == nil || sess.TenantID != tenantID {
			ctx.AbortWithStatusJSON(http.StatusForbidden, commonhttp.BaseResponse{
				Code:    http.StatusForbidden,
				Message: "业务处理失败",
				Success: false,
				Data:    "跨租户访问被拒绝",
			})
			return
		}
		ctx.Next()
	}
}

// RequireFresh 敏感操作二次校验中间件
//
// @description 要求会话最近 windowSeconds 秒内有活动，否则 401 "请重新认证"；须在认证中间件之后使用
// @param windowSeconds int 敏感操作窗口（秒），<=0 时使用默认 300
// @return gin.HandlerFunc Gin中间件函数
func RequireFresh(windowSeconds int) gin.HandlerFunc {
	if windowSeconds <= 0 {
		windowSeconds = defaultFreshWindow
	}
	return func(ctx *gin.Context) {
		sess := CurrentSession(ctx)
		if sess == nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, commonhttp.BaseResponse{
				Code:    http.StatusUnauthorized,
				Message: "业务处理失败",
				Success: false,
				Data:    "未认证",
			})
			return
		}
		if time.Now().Unix()-sess.LastActivityTime > int64(windowSeconds) {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, commonhttp.BaseResponse{
				Code:    http.StatusUnauthorized,
				Message: "业务处理失败",
				Success: false,
				Data:    "请重新认证",
			})
			return
		}
		ctx.Next()
	}
}

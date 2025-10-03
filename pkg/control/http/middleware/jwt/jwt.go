package jwt

import (
	"net/http"

	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/codes"

	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// EnableJWT 启用JWT认证中间件
// @description 用于验证JWT令牌和会话有效性
// @param logger Logger实例，用于记录日志
// @param sessionManager 会话管理器实例
// @return gin.HandlerFunc Gin中间件函数
func EnableJWT(logger *zap.SugaredLogger, sessionManager SessionManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, span := tracer.StartSpan(ctx.Request.Context(), "ginMiddleware", "jwt")
		defer span.End()
		// 从Authorization头中获取token
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			logger.Errorf("JWT认证失败：未提供token...")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,
				map[string]interface{}{
					"code":    http.StatusUnauthorized,
					"message": "未提供认证信息",
					"success": false,
					"data":    nil,
				},
			)
			ctx.Abort()
			span.SetStatus(codes.Error, "未提供认证信息")
			return
		}

		// 解析token前缀
		bearerToken := strings.Split(tokenString, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			logger.Errorf("JWT认证失败：token格式错误...")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,
				map[string]interface{}{
					"code":    http.StatusUnauthorized,
					"message": "认证信息格式错误",
					"success": false,
					"data":    nil,
				},
			)
			ctx.Abort()
			span.SetStatus(codes.Error, "认证信息格式错误")
			return
		}

		// 验证token
		claims, err := ParseToken(bearerToken[1])
		if err != nil {
			logger.Errorf("JWT认证失败：token解析错误: %v", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized,
				map[string]interface{}{
					"code":    http.StatusUnauthorized,
					"message": "认证信息无效或已过期",
					"success": false,
					"data":    nil,
				},
			)
			ctx.Abort()
			span.SetStatus(codes.Error, "认证信息无效或已过期")
			return
		}

		// 验证会话
		if sessionManager != nil {
			if !sessionManager.ValidateSession(claims.SessionID) {
				logger.Errorf("JWT认证失败：会话已失效: %v", claims.SessionID)
				ctx.AbortWithStatusJSON(http.StatusUnauthorized,
					map[string]interface{}{
						"code":    http.StatusUnauthorized,
						"message": "会话已失效，请重新登录",
						"success": false,
						"data":    nil,
					},
				)
				ctx.Abort()
				span.SetStatus(codes.Error, "会话已失效，请重新登录")
				return
			}

			// 更新会话最后活动时间
			claims.LastActivityTime = time.Now().Unix()
			if err = sessionManager.SetSession(claims.SessionID, claims); err != nil {
				logger.Errorf("设置 Session 失败: %v", err)
				span.SetStatus(codes.Error, err.Error())
				span.RecordError(err)
			}
		}

		// 将用户信息存储在上下文
		ctx.Set(UserId, claims.UserID)
		ctx.Set(UserName, claims.Username)
		ctx.Set(UserRole, claims.Role)
		ctx.Set(SessionId, claims.SessionID)
		ctx.Set(UserClaims, claims)

		ctx.Next()
		span.SetStatus(codes.Ok, "success")
	}
}

// ParseToken 解析JWT令牌
// @param tokenString JWT令牌字符串
// @return *Claims 解析后的声明信息
// @return error 解析过程中的错误
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// GenerateToken 生成JWT令牌
// @param userID 用户ID
// @param username 用户名
// @param role 用户角色
// @param sessionID 会话ID
// @param expireTime 过期时间（秒）
// @return string 生成的令牌
// @return error 生成过程中的错误
func GenerateToken(userID, username, role, sessionID string, expireTime int64) (string, *Claims, error) {
	now := time.Now()
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "gm-fabric-deployment",
			Subject:   userID,
			ID:        sessionID,
		},
		UserID:           userID,
		Username:         username,
		Role:             role,
		SessionID:        sessionID,
		SessionExpires:   now.Add(sessionTTL).Unix(),
		LoginTime:        now.Unix(),
		LastActivityTime: now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, claims, nil
}

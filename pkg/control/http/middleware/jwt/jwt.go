package jwt

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	webhttp "github.com/jianlu8023/gm-fabric-deployment/internal/web/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	// jwtSecret JWT签名密钥
	jwtSecret  = []byte("gm-fabric-deployment-secret")
	sessionTTL = 24 * time.Hour // 会话有效期
)

// Claims 自定义JWT声明结构体
type Claims struct {
	jwt.RegisteredClaims
	UserID           string   `json:"user_id"`
	Username         string   `json:"username"`
	Role             string   `json:"role"`
	SessionID        string   `json:"session_id"`
	UserType         string   `json:"user_type"`
	PermissionList   []string `json:"permission_list"`
	ClientIP         string   `json:"client_ip"`
	UserAgent        string   `json:"user_agent"`
	DeviceID         string   `json:"device_id"`
	SessionExpires   int64    `json:"session_expires"`
	LoginTime        int64    `json:"login_time"`
	LastActivityTime int64    `json:"last_activity_time"`
	Scope            string   `json:"scope"`
	TenantID         string   `json:"tenant_id"`
}

// SessionManager 会话管理器
type SessionManager interface {
	// GetSession 获取会话信息
	GetSession(sessionID string) (*Claims, error)
	// SetSession 设置会话信息
	SetSession(sessionID string, claims *Claims) error
	// DeleteSession 删除会话信息
	DeleteSession(sessionID string) error
	// ValidateSession 验证会话是否有效
	ValidateSession(sessionID string) bool
}

// EnableJWT 启用JWT认证中间件
// @description 用于验证JWT令牌和会话有效性
// @param logger Logger实例，用于记录日志
// @param sessionManager 会话管理器实例
// @return gin.HandlerFunc Gin中间件函数
func EnableJWT(logger *zap.SugaredLogger, sessionManager SessionManager) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 从Authorization头中获取token
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			logger.Errorf("JWT认证失败：未提供token...")
			webhttp.FailedResponseWithMessage(ctx, webhttp.SessionExpired, "未提供认证信息")
			// ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未提供认证信息"})
			ctx.Abort()
			return
		}

		// 解析token前缀
		bearerToken := strings.Split(tokenString, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			logger.Errorf("JWT认证失败：token格式错误...")
			webhttp.FailedResponseWithMessage(ctx, webhttp.SessionExpired, "认证信息格式错误")
			// ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证信息格式错误"})
			ctx.Abort()
			return
		}

		// 验证token
		claims, err := ParseToken(bearerToken[1])
		if err != nil {
			logger.Errorf("JWT认证失败：token解析错误: %v", err)
			webhttp.FailedResponseWithMessage(ctx, webhttp.SessionExpired, "认证信息无效或已过期")
			// ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "认证信息无效或已过期"})
			ctx.Abort()
			return
		}

		// 验证会话
		if sessionManager != nil {
			if !sessionManager.ValidateSession(claims.SessionID) {
				logger.Errorf("JWT认证失败：会话已失效: %v", claims.SessionID)
				webhttp.FailedResponseWithMessage(ctx, webhttp.SessionExpired, "会话已失效，请重新登录")
				// ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "会话已失效，请重新登录"})
				ctx.Abort()
				return
			}

			// 更新会话最后活动时间
			claims.LastActivityTime = time.Now().Unix()
			if err := sessionManager.SetSession(claims.SessionID, claims); err != nil {
				logger.Errorf("设置 Session 失败: %v", err)
			}
		}

		// 将用户信息存储在上下文
		ctx.Set("user_id", claims.UserID)
		ctx.Set("username", claims.Username)
		ctx.Set("role", claims.Role)
		ctx.Set("session_id", claims.SessionID)
		ctx.Set("user_info", claims)

		ctx.Next()
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

// MemorySessionManager 基于内存的会话管理器实现
type MemorySessionManager struct {
	sessions map[string]*Claims
	mutex    sync.RWMutex
	logger   *zap.SugaredLogger
	ctx      context.Context
}

// NewMemorySessionManager 创建内存会话管理器
// @param logger Logger实例，用于记录日志
// @return *MemorySessionManager 内存会话管理器实例
func NewMemorySessionManager(logger *zap.SugaredLogger, ctx context.Context) *MemorySessionManager {
	manager := &MemorySessionManager{
		sessions: make(map[string]*Claims),
		logger:   logger,
		ctx:      ctx,
	}

	// 启动会话清理协程
	go manager.cleanupLoop()
	return manager
}

// GetSession 获取会话信息
func (m *MemorySessionManager) GetSession(sessionID string) (*Claims, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	claims, exists := m.sessions[sessionID]
	if !exists {
		return nil, errors.New("会话不存在")
	}

	// 检查会话是否过期
	if claims.SessionExpires < time.Now().Unix() {
		// 异步删除过期会话
		go func() {
			m.DeleteSession(sessionID)
		}()
		return nil, errors.New("会话已过期")
	}

	return claims, nil
}

// SetSession 设置会话信息
func (m *MemorySessionManager) SetSession(sessionID string, claims *Claims) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessions[sessionID] = claims
	return nil
}

// DeleteSession 删除会话信息
func (m *MemorySessionManager) DeleteSession(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.sessions, sessionID)
	m.logger.Info("会话已删除", zap.String("sessionID", sessionID))
	return nil
}

// ValidateSession 验证会话是否有效
func (m *MemorySessionManager) ValidateSession(sessionID string) bool {
	claims, err := m.GetSession(sessionID)
	if err != nil {
		m.logger.Debug("会话验证失败", zap.String("sessionID", sessionID), zap.Error(err))
		return false
	}

	// 检查会话是否过期
	return claims.SessionExpires >= time.Now().Unix()
}

// cleanupLoop 定期清理过期会话
func (m *MemorySessionManager) cleanupLoop() {
	cleanupTicker := time.NewTicker(1 * time.Hour)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-cleanupTicker.C:
			now := time.Now().Unix()
			expiredCount := 0

			m.mutex.Lock()
			for sessionID, claims := range m.sessions {
				if claims.SessionExpires < now {
					delete(m.sessions, sessionID)
					expiredCount++
				}
			}
			m.mutex.Unlock()

			if expiredCount > 0 {
				m.logger.Debugf("清理过期会话: %v", expiredCount)
			}
		case <-m.ctx.Done():
			m.logger.Infof("会话管理器已停止...")
			return
		}
	}
}

// GetSessionCount 获取当前会话数量
func (m *MemorySessionManager) GetSessionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.sessions)
}

// GetAllSessions 获取所有会话信息
func (m *MemorySessionManager) GetAllSessions() []*Claims {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	sessions := make([]*Claims, 0, len(m.sessions))
	for _, claims := range m.sessions {
		sessions = append(sessions, claims)
	}

	return sessions
}

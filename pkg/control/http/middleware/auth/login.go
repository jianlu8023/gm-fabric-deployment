package auth

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/golang-example/pkg/control/http/middleware/session"
)

// Login 登录
//
// @description 生成 sessionID、创建会话（若启用）、签发 token（若启用 JWT）；session_only 模式返回 sessionID 作为 token
// @param userID string 用户ID
// @param username string 用户名
// @param role string 用户角色
// @return string JWT token 或 sessionID
// @return error 错误信息
func (m *authManagerImpl) Login(userID, username, role string) (string, error) {
	sessionID := uuid.GetUUID()
	now := time.Now().Unix()
	ttl := m.getSessionTTL()
	expiresAt := now + int64(ttl.Seconds())

	// 创建服务端会话（jwt_session / session_only 模式）
	if m.sessionStore != nil {
		sess := &session.Session{
			SessionID:        sessionID,
			UserID:           userID,
			Username:         username,
			Role:             role,
			LoginTime:        now,
			LastActivityTime: now,
			ExpiresAt:        expiresAt,
		}
		if err := m.sessionStore.Set(sessionID, sess); err != nil {
			m.logger.Errorf("[http/auth] session set failed on login: %v", err)
			return "", err
		}
	}

	// 签发 JWT（jwt_session / jwt_only 模式）
	if m.jwtManager != nil {
		token, _, err := m.jwtManager.GenerateToken(userID, username, role, sessionID, int64(ttl.Seconds()))
		if err != nil {
			m.logger.Errorf("[http/auth] jwt generate failed on login: %v", err)
			return "", err
		}
		return token, nil
	}

	// session_only 模式：返回 sessionID 作为 "token"
	return sessionID, nil
}

// Logout 登出
//
// @description 删除服务端会话，使 token 即使签名合法也无法再通过 Validate；jwt_only 模式无操作（无服务端状态可删）
// @param sessionID string 会话唯一标识
// @return error 错误信息
func (m *authManagerImpl) Logout(sessionID string) error {
	if m.sessionStore == nil {
		// jwt_only 模式：无服务端状态，token 自然过期
		return nil
	}
	if err := m.sessionStore.Delete(sessionID); err != nil {
		m.logger.Errorf("[http/auth] session delete failed on logout: %v", err)
		return err
	}
	return nil
}

// Refresh 续期会话
//
// @description 校验会话存在后更新 ExpiresAt 与 LastActivityTime；jwt 模式下若 token 即将过期则重签
// @param sessionID string 会话唯一标识
// @return string 新的 token（若重签），否则空字符串
// @return error 错误信息
func (m *authManagerImpl) Refresh(sessionID string) (string, error) {
	if m.sessionStore == nil {
		// jwt_only 模式：无服务端状态可续期，返回空 token
		return "", nil
	}

	sess, err := m.sessionStore.Get(sessionID)
	if err != nil {
		m.logger.Errorf("[http/auth] session get failed on refresh: %v", err)
		return "", err
	}

	now := time.Now().Unix()
	ttl := m.getSessionTTL()
	sess.LastActivityTime = now
	sess.ExpiresAt = now + int64(ttl.Seconds())
	if err = m.sessionStore.Set(sessionID, sess); err != nil {
		m.logger.Errorf("[http/auth] session set failed on refresh: %v", err)
		return "", err
	}

	// jwt 模式下重签 token
	if m.jwtManager != nil {
		token, _, err := m.jwtManager.GenerateToken(sess.UserID, sess.Username, sess.Role, sessionID, int64(ttl.Seconds()))
		if err != nil {
			m.logger.Errorf("[http/auth] jwt generate failed on refresh: %v", err)
			return "", err
		}
		return token, nil
	}
	return "", nil
}

// Authenticate 从请求中认证并返回会话
//
// @description 供非中间件场景（如 WebSocket 握手、gRPC 桥接）复用认证逻辑；不写 gin.Context，仅返回认证结果
// @param ctx *gin.Context Gin上下文
// @return *session.Session 会话信息
// @return error 认证错误
func (m *authManagerImpl) Authenticate(ctx *gin.Context) (*session.Session, error) {
	switch m.mode {
	case modeJWTSession:
		tokenString, errMsg, ok := m.extractBearerToken(ctx)
		if !ok {
			return nil, errors.New(errMsg)
		}
		claims, err := m.jwtManager.ParseToken(tokenString)
		if err != nil {
			return nil, err
		}
		if !m.sessionStore.Validate(claims.SessionID) {
			return nil, errSessionInvalid
		}
		return claimsToSession(claims), nil
	case modeJWTOnly:
		tokenString, errMsg, ok := m.extractBearerToken(ctx)
		if !ok {
			return nil, errors.New(errMsg)
		}
		claims, err := m.jwtManager.ParseToken(tokenString)
		if err != nil {
			return nil, err
		}
		return claimsToSession(claims), nil
	case modeSessionOnly:
		sessionID := m.extractSessionID(ctx)
		sess, err := m.sessionStore.Get(sessionID)
		if err != nil {
			return nil, err
		}
		return sess, nil
	default:
		return nil, errUnauthorized
	}
}

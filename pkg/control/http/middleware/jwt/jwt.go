package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JwtManager JWT管理器接口
//
// @description 定义JWT令牌的生成、解析及会话有效期获取能力，供认证中间件与会话管理器依赖使用
// @interface
type JwtManager interface {
	// GetSessionTTL 获取会话有效期
	//
	// @description 返回当前管理器配置的会话有效期，供会话管理器等外部组件使用
	// @return time.Duration 会话有效期
	GetSessionTTL() time.Duration
	// ParseToken 解析JWT令牌
	//
	// @description 使用管理器持有的 secret 校验签名并解析令牌；签名算法仅接受 HMAC 系列（与 GenerateToken 的 HS256 对齐）
	// @param tokenString JWT令牌字符串
	// @return *Claims 解析后的声明信息
	// @return error 解析过程中的错误
	ParseToken(tokenString string) (*Claims, error)
	// GenerateToken 生成JWT令牌
	//
	// @description 使用管理器持有的 secret 与 sessionTTL 生成JWT令牌
	// @param userID 用户ID
	// @param username 用户名
	// @param role 用户角色
	// @param sessionID 会话ID
	// @param expireTime 过期时间（秒）
	// @return string 生成的令牌
	// @return *Claims 令牌对应的声明信息
	// @return error 生成过程中的错误
	GenerateToken(userID, username, role, sessionID string, expireTime int64) (string, *Claims, error)
}

// jwtManagerImpl JWT管理器
//
// @description 持有JWT签名密钥与会话有效期，提供令牌生成与解析能力；认证中间件已迁移至 auth 包
// @struct
type jwtManagerImpl struct {
	// secret JWT签名密钥
	secret []byte
	// sessionTTL 会话有效期
	sessionTTL time.Duration
}

// NewManager 创建JWT管理器
//
// @description 根据传入的密钥与会话有效期创建JWT管理器；当 secret 为空时回退到默认值（仅用于开发环境），当 sessionTTLSeconds<=0 时使用默认24小时
// @param secret JWT签名密钥，应从配置文件读取
// @param sessionTTLSeconds 会话有效期（秒），<=0 时使用默认值86400
// @return *jwtManagerImpl JWT管理器实例
func NewManager(secret string, sessionTTLSeconds int) JwtManager {
	if secret == "" {
		secret = defaultSecret
	}
	if sessionTTLSeconds <= 0 {
		sessionTTLSeconds = defaultSessionTTLSeconds
	}
	return &jwtManagerImpl{
		secret:     []byte(secret),
		sessionTTL: time.Duration(sessionTTLSeconds) * time.Second,
	}
}

// GetSessionTTL 获取会话有效期
//
// @description 返回当前管理器配置的会话有效期，供会话管理器等外部组件使用
// @return time.Duration 会话有效期
func (m *jwtManagerImpl) GetSessionTTL() time.Duration {
	return m.sessionTTL
}

// ParseToken 解析JWT令牌
//
// @description 使用管理器持有的 secret 校验签名并解析令牌；签名算法仅接受 HMAC 系列（与 GenerateToken 的 HS256 对齐）
// @param tokenString JWT令牌字符串
// @return *Claims 解析后的声明信息
// @return error 解析过程中的错误
func (m *jwtManagerImpl) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 校验签名算法：仅接受 HMAC 系列（HS256/HS384/HS512），拒绝 RS256/ES256 等非 HMAC 算法，
		// 防止算法混淆攻击（如把公钥当作 HMAC 密钥使用）
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenSignatureInvalid
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}

	// ParseWithClaims 返回 nil 错误时 token.Valid 必为 true（见 jwt/v5 parser.go:128），
	// 此处 fallback 仅在 Claims 类型断言失败时触发（理论上不应发生）。
	// 返回 ErrTokenInvalidClaims 而非 ErrInvalidKey：此时签名已验证通过，密钥没有问题，
	// 问题出在 claims 类型上（"token has invalid claims" 语义更准确）
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

// GenerateToken 生成JWT令牌
//
// @description 使用管理器持有的 secret 与 sessionTTL 生成JWT令牌
// @param userID 用户ID
// @param username 用户名
// @param role 用户角色
// @param sessionID 会话ID
// @param expireTime 过期时间（秒）
// @return string 生成的令牌
// @return *Claims 令牌对应的声明信息
// @return error 生成过程中的错误
func (m *jwtManagerImpl) GenerateToken(userID, username, role, sessionID string, expireTime int64) (string, *Claims, error) {
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
		SessionExpires:   now.Add(m.sessionTTL).Unix(),
		LoginTime:        now.Unix(),
		LastActivityTime: now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.secret)
	if err != nil {
		return "", nil, err
	}

	return tokenString, claims, nil
}

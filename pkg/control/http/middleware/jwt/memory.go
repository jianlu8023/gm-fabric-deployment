package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"go.uber.org/zap"
)

// MemorySessionManager 基于内存的会话管理器实现
type MemorySessionManager struct {
	// sessions map[string]*Claims
	// mutex    sync.RWMutex
	sessions concurrent.Map[string, *Claims]
	logger   *zap.SugaredLogger
	ctx      context.Context
}

// NewMemorySessionManager 创建内存会话管理器
// @param logger Logger实例，用于记录日志
// @return *MemorySessionManager 内存会话管理器实例
func NewMemorySessionManager(logger *zap.SugaredLogger, ctx context.Context) *MemorySessionManager {
	manager := &MemorySessionManager{
		// sessions: make(map[string]*Claims),
		sessions: concurrentmap.NewRWMap[string, *Claims](),
		logger:   logger,
		ctx:      ctx,
	}

	// 启动会话清理协程
	go manager.cleanupLoop()
	return manager
}

// GetSession 获取会话信息
func (m *MemorySessionManager) GetSession(sessionID string) (*Claims, error) {
	// m.mutex.RLock()
	// defer m.mutex.RUnlock()

	// claims, exists := m.sessions[sessionID]
	claims, exists := m.sessions.Get(sessionID)
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
	// m.mutex.Lock()
	// defer m.mutex.Unlock()

	// m.sessions[sessionID] = claims
	m.sessions.Put(sessionID, claims)
	return nil
}

// DeleteSession 删除会话信息
func (m *MemorySessionManager) DeleteSession(sessionID string) error {
	// m.mutex.Lock()
	// defer m.mutex.Unlock()

	// delete(m.sessions, sessionID)
	m.sessions.Del(sessionID)
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

			// m.mutex.Lock()
			// for sessionID, claims := range m.sessions {
			// 	if claims.SessionExpires < now {
			// 		delete(m.sessions, sessionID)
			// 		expiredCount++
			// 	}
			// }
			// m.mutex.Unlock()
			iterator := m.sessions.Iterator()
			for iterator.HasNext() {
				item := iterator.Value()
				if item.Value.SessionExpires < now {
					m.sessions.Del(item.Key)
					expiredCount++
				}
			}

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
	// m.mutex.RLock()
	// defer m.mutex.RUnlock()
	// return len(m.sessions)
	return m.sessions.Len()
}

// GetAllSessions 获取所有会话信息
func (m *MemorySessionManager) GetAllSessions() []*Claims {
	// m.mutex.RLock()
	// defer m.mutex.RUnlock()

	// sessions := make([]*Claims, 0, len(m.sessions))
	// for _, claims := range m.sessions {
	// 	sessions = append(sessions, claims)
	// }
	// return sessions
	claims := m.sessions.Values()
	return claims
}

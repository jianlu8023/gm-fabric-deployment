package session

import (
	"context"
	"errors"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"go.uber.org/zap"
)

// memoryStoreImpl 内存会话存储实现
//
// @description 基于 concurrent.RWMap 的单机会话存储，启动后台协程定期清理过期会话；单机部署使用，多实例部署需换 RedisStore（未来扩展）
// @struct
type memoryStoreImpl struct {
	// sessions 会话表，并发安全
	sessions concurrent.Map[string, *Session]
	// logger 日志实例
	logger *zap.SugaredLogger
	// ctx 生命周期上下文，用于优雅停止清理协程
	ctx context.Context
	// ttl 会话有效期，供 Login 时计算 ExpiresAt 使用
	ttl time.Duration
	// cleanupInterval 过期清理间隔
	cleanupInterval time.Duration
}

// NewMemoryStore 创建内存会话存储
//
// @description 创建 memoryStoreImpl 并启动后台清理协程；ttlSeconds<=0 时使用默认 86400s，cleanupIntervalSeconds<=0 时使用默认 3600s
// @param logger *zap.SugaredLogger 日志实例
// @param ctx context.Context 生命周期上下文
// @param ttlSeconds int 会话有效期（秒）
// @param cleanupIntervalSeconds int 过期清理间隔（秒）
// @return *memoryStoreImpl 内存会话存储实例
func NewMemoryStore(logger *zap.SugaredLogger, ctx context.Context, ttlSeconds int, cleanupIntervalSeconds int) SessionStore {
	if ttlSeconds <= 0 {
		ttlSeconds = defaultSessionTTLSeconds
	}
	if cleanupIntervalSeconds <= 0 {
		cleanupIntervalSeconds = defaultCleanupIntervalSeconds
	}
	store := &memoryStoreImpl{
		sessions:        concurrentmap.NewRWMap[string, *Session](),
		logger:          logger,
		ctx:             ctx,
		ttl:             time.Duration(ttlSeconds) * time.Second,
		cleanupInterval: time.Duration(cleanupIntervalSeconds) * time.Second,
	}
	go store.cleanupLoop()
	return store
}

// GetTTL 获取会话有效期
//
// @description 返回当前存储配置的会话有效期，供 auth 包 Login 时计算 ExpiresAt 使用
// @return time.Duration 会话有效期
func (m *memoryStoreImpl) GetTTL() time.Duration {
	return m.ttl
}

// Get 获取会话信息
//
// @description 从 concurrentmap 读取会话；若 ExpiresAt < now 则异步删除并返回"会话已过期"错误
// @param sessionID string 会话唯一标识
// @return *Session 会话数据
// @return error 错误信息
func (m *memoryStoreImpl) Get(sessionID string) (*Session, error) {
	sess, exists := m.sessions.Get(sessionID)
	if !exists {
		return nil, errors.New("会话不存在")
	}
	if sess.ExpiresAt < time.Now().Unix() {
		// 异步删除过期会话，避免阻塞读路径
		go func() {
			_ = m.Delete(sessionID)
		}()
		return nil, errors.New("会话已过期")
	}
	return sess, nil
}

// Set 设置会话信息
//
// @description 写入会话到 concurrentmap，由调用方保证 ExpiresAt 正确
// @param sessionID string 会话唯一标识
// @param sess *Session 会话数据
// @return error 错误信息
func (m *memoryStoreImpl) Set(sessionID string, sess *Session) error {
	m.sessions.Put(sessionID, sess)
	return nil
}

// Delete 删除会话信息
//
// @description 从 concurrentmap 删除指定会话，用于登出/强制下线
// @param sessionID string 会话唯一标识
// @return error 错误信息
func (m *memoryStoreImpl) Delete(sessionID string) error {
	m.sessions.Del(sessionID)
	m.logger.Info("会话已删除", zap.String("sessionID", sessionID))
	return nil
}

// Validate 验证会话是否有效
//
// @description 调用 Get 并判断错误，直接返回 err == nil；Get 内部已检查过期，此处不重复比较 ExpiresAt
// @param sessionID string 会话唯一标识
// @return bool 是否有效
func (m *memoryStoreImpl) Validate(sessionID string) bool {
	_, err := m.Get(sessionID)
	if err != nil {
		m.logger.Debug("会话验证失败", zap.String("sessionID", sessionID), zap.Error(err))
		return false
	}
	return true
}

// Count 获取当前会话数量
//
// @description 返回 concurrentmap.Len()，用于监控/运营统计
// @return int 会话数量
func (m *memoryStoreImpl) Count() int {
	return m.sessions.Len()
}

// GetAll 获取所有会话
//
// @description 返回 concurrentmap.Values() 的快照，数据量大时慎用
// @return []*Session 会话列表
func (m *memoryStoreImpl) GetAll() []*Session {
	return m.sessions.Values()
}

// cleanupLoop 定期清理过期会话
//
// @description 按 cleanupInterval 周期遍历会话表，删除 ExpiresAt < now 的会话；通过 ctx.Done() 优雅退出
// @return
func (m *memoryStoreImpl) cleanupLoop() {
	ticker := time.NewTicker(m.cleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now().Unix()
			expiredCount := 0
			iterator := m.sessions.Iterator()
			for iterator.HasNext() {
				item := iterator.Value()
				if item.Value.ExpiresAt < now {
					m.sessions.Del(item.Key)
					expiredCount++
				}
			}
			if expiredCount > 0 {
				m.logger.Debugf("清理过期会话: %d", expiredCount)
			}
		case <-m.ctx.Done():
			m.logger.Infof("会话存储清理协程已停止...")
			return
		}
	}
}

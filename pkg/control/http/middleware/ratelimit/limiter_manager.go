package ratelimit

import (
	"container/list"
	"github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent"
	concurrentmap "github.com/jianlu8023/go-tools/v2/pkg/collections/concurrent/map"
	"sync"
	"time"
)

// LimiterEntry 限流器条目，包含限流器实例和元数据
type LimiterEntry struct {
	Limiter    interface{}   // 限流器实例（*rate.Limiter、*CustomRateLimiter 等）
	LastAccess time.Time     // 最后访问时间
	Element    *list.Element // LRU列表中的元素
}

// IPRateLimiterManager IP限流器管理器
//
// @description 管理IP到限流器的映射，支持以下特性：
//  1. 最大条目数限制：防止恶意IP造成内存泄漏
//  2. 过期清理：定期清理长时间未访问的限流器
//  3. LRU淘汰：超过最大条目数时，淘汰最久未使用的限流器
//  4. 并发安全：支持多goroutine并发访问
type IPRateLimiterManager struct {
	mu sync.Mutex // 互斥锁，保护map和lru的并发访问
	// limiters   map[string]*LimiterEntry // IP到限流器条目的映射
	limiters   concurrent.Map[string, *LimiterEntry] // IP到限流器条目的映射
	lru        *list.List                            // LRU链表，用于淘汰最久未使用的条目
	maxSize    int                                   // 最大条目数，超过后使用LRU淘汰
	expireTime time.Duration                         // 过期时间，超过此时间未访问的条目将被清理
	stopCh     chan struct{}                         // 停止清理协程的通道
}

// NewIPRateLimiterManager 创建IP限流器管理器
//
// @param maxSize 最大条目数，0表示不限制
// @param expireTime 过期时间，0表示不过期清理
// @return *IPRateLimiterManager 管理器实例
func NewIPRateLimiterManager(maxSize int, expireTime time.Duration) *IPRateLimiterManager {
	m := &IPRateLimiterManager{
		// limiters:   make(map[string]*LimiterEntry),
		limiters:   concurrentmap.NewRWMap[string, *LimiterEntry](),
		lru:        list.New(),
		maxSize:    maxSize,
		expireTime: expireTime,
		stopCh:     make(chan struct{}),
	}

	// 如果设置了过期时间，启动后台清理协程
	if expireTime > 0 {
		go m.cleanupLoop()
	}

	return m
}

// GetOrCreate 获取或创建指定IP的限流器
//
// @param ip 客户端IP地址
// @param factory 创建限流器的工厂函数
// @return interface{} 限流器实例
func (m *IPRateLimiterManager) GetOrCreate(ip string, factory func() interface{}) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 尝试获取已存在的限流器
	// if entry, exists := m.limiters[ip]; exists {
	// 	// 更新最后访问时间
	// 	entry.LastAccess = time.Now()
	// 	// 移动到LRU链表头部（表示最近使用）
	// 	m.lru.MoveToFront(entry.Element)
	// 	return entry.Limiter
	// }
	if enrty, exists := m.limiters.Get(ip); exists {
		// 更新最后访问时间
		enrty.LastAccess = time.Now()
		// 移动到LRU链表头部（表示最近使用）
		m.lru.MoveToFront(enrty.Element)
		return enrty.Limiter
	}

	// 创建新的限流器
	limiter := factory()

	// 创建新的条目
	entry := &LimiterEntry{
		Limiter:    limiter,
		LastAccess: time.Now(),
	}

	// 添加到LRU链表头部
	entry.Element = m.lru.PushFront(entry)
	// 添加到map
	// m.limiters[ip] = entry
	m.limiters.Put(ip, entry)

	// 检查是否超过最大条目数
	// if m.maxSize > 0 && len(m.limiters) > m.maxSize {
	// 	m.evictLRU()
	// }
	if m.maxSize > 0 && m.limiters.Len() > m.maxSize {
		m.evictLRU()
	}

	return limiter
}

// Get 获取指定IP的限流器（不创建）
//
// @param ip 客户端IP地址
// @return interface{} 限流器实例
// @return bool 是否存在
func (m *IPRateLimiterManager) Get(ip string) (interface{}, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// entry, exists := m.limiters[ip]
	entry, exists := m.limiters.Get(ip)
	if exists {
		// 更新最后访问时间
		entry.LastAccess = time.Now()
		// 移动到LRU链表头部
		m.lru.MoveToFront(entry.Element)
		return entry.Limiter, true
	}
	return nil, false
}

// Delete 删除指定IP的限流器
//
// @param ip 客户端IP地址
func (m *IPRateLimiterManager) Delete(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// if entry, exists := m.limiters[ip]; exists {
	// 	// 从LRU链表中移除
	// 	m.lru.Remove(entry.Element)
	// 	// 从map中删除
	// 	delete(m.limiters, ip)
	// }
	if entry, exists := m.limiters.Get(ip); exists {
		// 从LRU链表中移除
		m.lru.Remove(entry.Element)
		// 从map中删除
		m.limiters.Del(ip)
	}
}

// Size 获取当前限流器数量
//
// @return int 限流器数量
func (m *IPRateLimiterManager) Size() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	// return len(m.limiters)
	return m.limiters.Len()
}

// Close 关闭管理器，停止后台清理协程
func (m *IPRateLimiterManager) Close() {
	select {
	case m.stopCh <- struct{}{}:
	default:
	}
}

// evictLRU 淘汰最久未使用的条目（调用方需持有锁）
func (m *IPRateLimiterManager) evictLRU() {
	// 获取LRU链表尾部的元素（最久未使用）
	back := m.lru.Back()
	if back != nil {
		entry := back.Value.(*LimiterEntry)
		// 从链表中移除
		m.lru.Remove(back)
		// 从map中删除
		// for ip, e := range m.limiters {
		// 	if e == entry {
		// 		delete(m.limiters, ip)
		// 		break
		// 	}
		// }
		iter := m.limiters.Iterator()
		if iter.HasNext() {
			node := iter.Value()
			ip, e := node.Key, node.Value
			if e == entry {
				m.limiters.Del(ip)
			}
		}

	}
}

// cleanupLoop 后台清理协程，定期清理过期的限流器
func (m *IPRateLimiterManager) cleanupLoop() {
	// 清理间隔为过期时间的1/4，最小为10秒
	interval := m.expireTime / 4
	if interval < 10*time.Second {
		interval = 10 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpired()
		case <-m.stopCh:
			return
		}
	}
}

// cleanupExpired 清理过期的限流器
func (m *IPRateLimiterManager) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expiredIPs := make([]string, 0)

	// 收集过期的IP
	// for ip, entry := range m.limiters {
	// 	if now.Sub(entry.LastAccess) > m.expireTime {
	// 		expiredIPs = append(expiredIPs, ip)
	// 	}
	// }

	iter := m.limiters.Iterator()
	for iter.HasNext() {
		node := iter.Value()
		entry := node.Value
		if now.Sub(entry.LastAccess) > m.expireTime {
			expiredIPs = append(expiredIPs, node.Key)
		}
	}

	// 删除过期的条目
	for _, ip := range expiredIPs {
		// if entry, exists := m.limiters[ip]; exists {
		// 	m.lru.Remove(entry.Element)
		// 	delete(m.limiters, ip)
		// }
		if entry, exists := m.limiters.Get(ip); exists {
			m.lru.Remove(entry.Element)
			m.limiters.Del(ip)
		}
	}
}

// GetManagerConfig 获取管理器默认配置
//
// @description 返回推荐的管理器配置，生产环境可根据实际情况调整
// @return maxSize 最大条目数
// @return expireTime 过期时间
func GetManagerConfig() (maxSize int, expireTime time.Duration) {
	return 10000, 30 * time.Minute
}

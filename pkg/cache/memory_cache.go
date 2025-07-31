package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Value      string
	Expiration time.Time
}

// IsExpired 检查是否过期
func (item CacheItem) IsExpired() bool {
	return time.Now().After(item.Expiration)
}

// MemoryCache 内存缓存实现
type MemoryCache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
}

// NewMemoryCache 创建内存缓存
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]CacheItem),
	}

	// 启动清理协程
	go cache.cleanup()

	return cache
}

// Get 获取缓存值
func (c *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists || item.IsExpired() {
		return "", nil // 返回空字符串表示缓存未命中
	}

	return item.Value, nil
}

// Set 设置缓存值
func (c *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var valueStr string
	switch v := value.(type) {
	case string:
		valueStr = v
	default:
		// 尝试JSON序列化
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		valueStr = string(bytes)
	}

	c.items[key] = CacheItem{
		Value:      valueStr,
		Expiration: time.Now().Add(expiration),
	}

	return nil
}

// cleanup 清理过期缓存项
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		for key, item := range c.items {
			if item.IsExpired() {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Clear 清空所有缓存
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items = make(map[string]CacheItem)
}

// Size 获取缓存项数量
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.items)
}

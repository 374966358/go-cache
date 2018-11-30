package cache

import (
	"sync"
	"time"
)

type ItemsCache struct {
	// ItemsCache 读写锁
	sync.RWMutex
	// 缓存名
	key interface{}
	// 缓存值
	value interface{}
	// 缓存生命周期
	lifeCycle time.Duration
	// 缓存创建时间
	createTime time.Time
	// 最后访问时间
	accessTime time.Time
	// 访问次数
	accessCount int
}

// 返回ItemsCache创建的实例
// key：缓存名
// value：缓存值
// lifeCycle：缓存生命周期
func NewItemsCache(key interface{}, value interface{}, lifeCycle time.Duration) *ItemsCache {
	t := time.Now()
	return &ItemsCache{
		key:         key,
		value:       value,
		lifeCycle:   lifeCycle,
		createTime:  t,
		accessTime:  t,
		accessCount: 0,
	}
}

// 触犯获取次数
func (item *ItemsCache) TriggerObtain() {
	item.Lock()
	defer item.Unlock()
	item.accessTime = time.Now()
	item.accessCount++
}

// 返回ItemsCache中数据key（缓存名）
func (item *ItemsCache) Key() interface{} {
	return item.key
}

// 返回ItemsCache中数据value（缓存值）
func (item *ItemsCache) Value() interface{} {
	return item.value
}

// 返回ItemsCache中数据lifeCycle（缓存生命周期）
func (item *ItemsCache) LifeCycle() time.Duration {
	return item.lifeCycle
}

// 返回ItemsCache中数据createTime（缓存创建时间）
func (item *ItemsCache) CreateTime() time.Time {
	return item.createTime
}

// 返回ItemsCache中数据accessTime（最后访问时间）
func (item *ItemsCache) AccessTime() time.Time {
	return item.accessTime
}

// 返回ItemsCache中数据accessCount（访问次数）
func (item *ItemsCache) AccessCount() int {
	return item.accessCount
}

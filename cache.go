package cache

import (
	"sync"
)

var (
	mutex sync.RWMutex
	cache = make(map[string]*CacheTable)
)

func Cache(name string) *CacheTable {
	/* 开启读写锁用于检查当前缓存是否存在 */
	mutex.RLock()

	/* 根据name查找是否存在 */
	v, ok := cache[name]

	/* 关闭读写锁 */
	mutex.RUnlock()

	/* 判断是否有name如果没有的话 */
	if !ok {
		/* 再次开启锁进行验证 */
		mutex.Lock()

		/* 再确认下确实是否有无存储 */
		v, ok = cache[name]

		/* 确认的确没有 */
		if !ok {
			/* 产生新的数据 */
			v = &CacheTable{
				name:  name,
				items: make(map[interface{}]*ItemsCache),
			}

			cache[name] = v
		}

		/* 关闭锁 */
		mutex.Unlock()
	}

	return v
}

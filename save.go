package cache

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type CacheTable struct {
	// 结构体锁
	sync.RWMutex
	// 仓库名称
	name string
	// 存储的键值对集合体
	items map[interface{}]*ItemsCache
	// 用于打印日志
	logger *log.Logger
	// 用于检查过期时间
	deadline time.Duration
	// 用于存储批量删除时将Afeter回收
	timeAfter *time.Timer
}

// 根据键值对存储缓存
// key：缓存名
// value：缓存值
// lifeCycle：缓存生命周期(秒)，如果想要不被销毁则填写-1
func (table *CacheTable) Set(key interface{}, value interface{}, lifeCycle time.Duration) {
	// 判断是否声明长期有效的生命周期
	if lifeCycle <= -1 {
		lifeCycle = 3155760000
	}

	lifeCycle *= time.Second

	// 创建并返回ItemsCache
	item := NewItemsCache(key, value, lifeCycle)

	// 开启CacheTable写锁
	table.Lock()

	// 存储数据
	table.setInside(item)
}

// 执行插入到缓存操作
// item：ItemsCache结构体实例
func (table *CacheTable) setInside(item *ItemsCache) {
	// 打印缓存日志并将缓存存入对应的名中
	fmt.Println("缓存名：", item.key, "缓存值：", item.value, "缓存生命周期：", item.lifeCycle)
	table.items[item.key] = item

	// 获取最小检查时间
	deadline := table.deadline

	// 关闭CacheTable写锁
	table.Unlock()

	// 判断过期时间是否符合并且设置过期时间是否小于最小检查时间
	if item.lifeCycle > 0 && (deadline == 0 || item.lifeCycle < deadline) {
		table.expirationCheck()
	}
}

// 检查是否过期如果过期进行清除
func (table *CacheTable) expirationCheck() {
	// 开启CacheTable写锁
	table.Lock()

	// 记录检查日志
	if table.deadline > 0 {
		fmt.Println("数据存在进行过期检查：", table.deadline, "，表名称：", table.name)
	} else {
		fmt.Println("数据创建时进行检查：", table.name)
	}

	// 定义用于检查的时间
	checkTime := time.Now()

	// 定义用于存储最小时间变量
	smallDuration := 0 * time.Second

	// 循环结果集开启检查机制
	for key, value := range table.items {
		// 开启items中读锁
		value.RLock()

		// 获取用于检查结果的缓存生命周期
		lifeCycle := value.lifeCycle

		// 获取用于检查结果的最后访问时间
		accessTime := value.accessTime

		// 关闭items中读锁
		value.RUnlock()

		// 如果缓存生命周期未设置跳到下次检查
		if lifeCycle == 0 {
			continue
		}

		// 定义时间计算
		computingTime := checkTime.Sub(accessTime)

		// 判断是否进行删除还是执行最小时间设置
		if computingTime >= lifeCycle {
			// 执行删除缓存操作
			table.deleteInside(key)
		} else {
			// 获取时间差
			timeDifference := lifeCycle - computingTime

			// 用于存储下次检查时间防止过量检查
			if smallDuration == 0 || timeDifference < smallDuration {
				smallDuration = timeDifference
			}
		}
	}

	// 存储最小检查时间防止多次消耗
	table.deadline = smallDuration

	// 判断是否有需要消耗时间多少后进行销毁
	if smallDuration > 0 {
		// 使用time函数AfterFunc在多少秒之后执行
		table.timeAfter = time.AfterFunc(smallDuration, func() {
			go table.expirationCheck()
		})
	}

	// 关闭CacheTable写锁
	table.Unlock()
}

// 删除单个缓存
// key：缓存名称
func (table *CacheTable) Delete(key interface{}) (*ItemsCache, error) {
	// 开启CacheTable写锁
	table.Lock()

	// 关闭CacheTable写锁
	defer table.Unlock()

	// 执行删除操作并返回
	return table.deleteInside(key)
}

// 删除全部缓存
func (table *CacheTable) DeleteAll() {
	// 开启CacheTable写锁
	table.Lock()

	// 关闭CacheTable写锁
	defer table.Unlock()

	// 还原至初始状态
	table.items = make(map[interface{}]*ItemsCache)
	table.deadline = 0

	// 判断是否有未结束的time.After如果有强制结束
	if table.timeAfter != nil {
		table.timeAfter.Stop()
	}
}

// 执行删除操作
// key：缓存名称
// 返回已删除的数据，错误信息
func (table *CacheTable) deleteInside(key interface{}) (*ItemsCache, error) {
	v, ok := table.items[key]

	// 检查结果集中是否存在当前要删除的缓存名
	if !ok {
		// 返回错误提示信息
		return nil, ErrKeyNotFound
	}

	// 关闭CacheTable写锁，用于清理过期时防止死锁发生
	table.Unlock()

	// 开启CecheTable写锁
	table.Lock()

	// 执行删除操作
	delete(table.items, key)

	// 记录日志
	fmt.Println("删除键：", key, "，创建时间：", v.createTime, "，访问次数：", v.accessCount)

	return v, nil
}

// 检查缓存是否存在
func (table *CacheTable) Exists(key interface{}) bool {
	// 开启CacheTable读锁
	table.RLock()

	// 关闭CacheTable读锁
	defer table.RUnlock()

	_, ok := table.items[key]

	return ok
}

// 获取缓存值
// Key：存储的缓存名
func (table *CacheTable) Get(key interface{}) (*ItemsCache, error) {
	// 开启CacheTable读锁
	table.RLock()

	// 获取数据
	v, ok := table.items[key]

	// 关闭CacheTable读锁
	table.RUnlock()

	// 根据数据进行判断
	if ok {
		v.TriggerObtain()
		return v, nil
	}

	return nil, ErrKeyNotFound
}

// 返回缓存库总长度
func (table *CacheTable) Count() int {
	// 开启CacheTable读锁
	table.RLock()

	// 关闭CacheTable读锁
	defer table.RUnlock()

	// 返回缓存库长度
	return len(table.items)
}

// 循环遍历
// f：回调函数
func (table *CacheTable) Foreach(f func(key interface{}, value *ItemsCache)) {
	// 开启CacheTable读锁
	table.RLock()

	// 关闭CacheTable读锁
	defer table.RUnlock()

	// 循环将数据插入到回调函数中
	for key, item := range table.items {
		f(key, item)
	}
}

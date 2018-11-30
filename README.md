go-cache
========

## 安装

请求包你有一个GO的工作环境
请参考安装说明
[国内](https://studygolang.com/dl)
[国外](http://golang.org/doc/install.html)

要安装cache只需要运行：

    go get github.com/374966358/go-cache

## 例子
```go
package main

import (
	"github.com/muesli/go-cache"
	"fmt"
	"time"
)

func main() {
	// 创建一个缓存表存储
	cache := cache.Cache("testCache")

	// 我们在缓存表中插入一个缓存项并将缓存时间设置为5s
	cache.Set("test", "测试插入", 4)

	// 让我们从项目中检索这个缓存
	res, err := cache.Get("test")
	if err == nil {
		fmt.Println("缓存中的值:", res.Value())
	} else {
		fmt.Println("获取缓存值时发生错误:", err)
	}

	// 等待缓存过期
	time.Sleep(5 * time.Second)
	_, err = cache.Get("test")
	if err != nil {
		fmt.Println("缓存已不再")
	}

	// 添加一个用不过期的项目
	cache.Set("test", "测试插入", -1)

	// 从缓存中删除test
	cache.Delete("test")

	// 删除全部
	cache.DeleteAll()
}
```

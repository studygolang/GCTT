首发于：https://studygolang.com/articles/25283

# Gocache：一个功能齐全且易于扩展的 Go 缓存库

在先前几周的时候，我完成了 [Gocache](https://github.com/eko/gocache)，对于 Go 开发者而言，它是功能齐全且易于扩展的。

这个库的设计目的是为了解决在使用缓存或者使用多种（多级）缓存时所遇到的问题，它为缓存方案制定了一个标准。

## 背景

当我一开始在为 GraphQL 的 Go 项目构建缓存时，该项目已经包含了一套有简单 API 的内存缓存，还使用了另外一套有不同 API 的内存缓存和加载缓存数据的代码，它们实际上都是在只做了同一件事：缓存。

后来，我们又有了另一个需求：除了内存缓存外，我们还想添加一套基于 Redis 集群的分布式缓存，其主要目的是为了在新版本上线时，避免 Kubernetes 的 Pods 使用的缓存为空。

于是创造 Gocache 的契机出现了，是时候用一套统一的规则来管理多种缓存方案了，不管是内存、Redis、Memcache 或者其他任何形式的缓存。

哦，还不止这些，我们还希望缓存的数据可以被 Metrics 监控（后来被我们的 Prometheus 替代）。（译者注：Metrics 和 Prometheus 均是监控工具。）

Gocache 项目诞生了：https://github.com/eko/gocache 。

## 存储接口

首先，当你准备缓存一些数据时，你必须选择缓存的存储方式：简单的直接放进内存？使用 Redis 或者 Memcache？或者其它某种形式的存储。

目前，Gocache 已经实现了以下存储方案：

* **Bigcache**: 简单的内存存储。
* **Ristretto**: 由 DGraph 提供的内存存储。
* **Memcache**: 基于 bradfitz/gomemcache 的 Memcache 存储。
* **Redis**: 基于 go-redis/redis 的 Redis 存储。

所有的存储方案都实现了一个非常简单的接口：

``` go
type StoreInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key interface{}, value interface{}, options *Options) error
	Delete(key interface{}) error
	Invalidate(options InvalidateOptions) error
	Clear() error
	GetType() string
}
```

这个接口展示了可以对存储器执行的所有操作，每个操作只调用了存储器客户端的必要方法。

这些存储器都有不同的配置，具体配置取决于实现存储器的客户端，举个例子，以下为初始化 Memcache 存储器的示例：

``` go
store := store.NewMemcache(
	memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11212"),
	&store.Options{
		Expiration: 10*time.Second,
	},
)
```

然后，必须将初始化存储器的代码放进缓存的构造函数中。

## 缓存接口

以下为缓存接口，缓存接口和存储接口是一样的，毕竟，缓存就是对存储器做一些操作。

``` go
type CacheInterface interface {
	Get(key interface{}) (interface{}, error)
	Set(key, object interface{}, options *store.Options) error
	Delete(key interface{}) error
	Invalidate(options store.InvalidateOptions) error
	Clear() error
	GetType() string
}
```

该接口包含了需要对缓存数据进行的所有必要操作：Set，Get，Delete，使某条缓存失效，清空缓存。如果需要的话，还可以使用 GetType 方法获取缓存类型。

缓存接口已有以下实现：

* **Cache**: 基础版，直接操作存储器。
* **Chain**: 链式缓存，它允许使用多级缓存（项目中可能同时存在内存缓存，Redis 缓存等等）。
* **Loadable**: 可自动加载数据的缓存，它可以指定回调函数，在缓存过期或失效的时候，会自动通过回调函数将数据加载进缓存中。
* **Metric**: 内嵌监控的缓存，它会收集缓存的一些指标，比如 Set、Get、失效和成功的缓存数量。

最棒的是：这些缓存器都实现了相同的接口，所以它们可以很容易地相互组合。你的缓存可以同时具有链式、可自动加载数据、包含监控等特性。

还记得吗？我们想要简单的 API，以下为使用 Memcache 的示例：

``` go
memcacheStore := store.NewMemcache(
	memcache.New("10.0.0.1:11211", "10.0.0.2:11211", "10.0.0.3:11212"),
	&store.Options{
		Expiration: 10*time.Second,
	},
)

cacheManager := cache.New(memcacheStore)
err := cacheManager.Set("my-key", []byte("my-value"), &cache.Options{
	Expiration: 15*time.Second, // 设置过期时间
})
if err != nil {
    panic(err)
}

value := cacheManager.Get("my-key")

cacheManager.Delete("my-key")

cacheManager.Clear() // 清空缓存
```

现在，假设你想要将已有的缓存修改为一个链式缓存，该缓存包含 Ristretto（内存型）和 Redis 集群，并且具备缓存数据序列化和监控特性：

``` go
// 初始化 Ristretto 和 Redis 客户端
ristrettoCache, err := ristretto.NewCache(&ristretto.Config{NumCounters: 1000, MaxCost: 100, BufferItems: 64})
if err != nil {
    panic(err)
}

redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})

// 初始化存储器
ristrettoStore := store.NewRistretto(ristrettoCache, nil)
redisStore := store.NewRedis(redisClient, &cache.Options{Expiration: 5*time.Second})

// 初始化 Prometheus 监控
promMetrics := metrics.NewPrometheus("my-amazing-app")

// 初始化链式缓存
cacheManager := cache.NewMetric(promMetrics, cache.NewChain(
    cache.New(ristrettoStore),
    cache.New(redisStore),
))

// 初始化序列化工具
marshal := marshaler.New(cacheManager)

key := BookQuery{Slug: "my-test-amazing-book"}
value := Book{ID: 1, Name: "My test amazing book", Slug: "my-test-amazing-book"}

// 插入缓存
err = marshal.Set(key, value)
if err != nil {
    panic(err)
}

returnedValue, err := marshal.Get(key, new(Book))
if err != nil {
    panic(err)
}

// Then, do what you want with the value
```

我们不对序列化做过多的讨论，因为这个是 Gocache 的另外一个特性：提供一套在存储和取出缓存对象时可以自动序列化和反序列化缓存对象的工具。

该特性在使用对象作为缓存 Key 时会很有用，它省去了在代码中手动转换对象的操作。

所有的这些特性：包含内存型和 Redis 的链式缓存、包含 Prometheus 监控功能和自动序列化功能，都可以在 20 行左右的代码里完成。

## 定制你自己的缓存

如果你想定制自己的缓存也很容易。

以下示例展示了如何给对缓存的每个操作添加日志（这不是一个好主意，只是作为示例）：

``` go
package customcache

import (
	"log"

	"github.com/eko/gocache/cache"
	"github.com/eko/gocache/store"
)

const LoggableType = "loggable"

type LoggableCache struct {
	cache cache.CacheInterface
}

func NewLoggable(cache cache.CacheInterface) *LoggableCache {
	return &LoggableCache{
		cache: cache,
	}
}

func (c *LoggableCache) Get(key interface{}) (interface{}, error) {
	log.Print("Get some data...")
	return c.cache.Get(key)
}

func (c *LoggableCache) Set(key, object interface{}, options *store.Options) error {
	log.Print("Set some data...")
	return c.cache.Set(key, object, options)
}

func (c *LoggableCache) Delete(key interface{}) error {
	log.Print("Delete some data...")
	return c.cache.Delete(key)
}

func (c *LoggableCache) Invalidate(options store.InvalidateOptions) error {
	log.Print("Invalidate some data...")
	return c.cache.Invalidate(options)
}

func (c *LoggableCache) Clear() error {
	log.Print("Clear some data...")
	return c.cache.Clear()
}

func (c *LoggableCache) GetType() string {
	return LoggableType
}
```

通过同样的方法，你也可以自己实现存储接口。

如果你认为其他人也可以从你的缓存实现中获得启发，请不要犹豫，直接在项目中发起合并请求。通过共同讨论你的想法，我们会提供一个功能更加强大的缓存库。

## 结论

在构建这个库的过程中，我还尝试改进了 Go 的社区工具。

希望你喜欢这篇博客，如果有需要的话，我非常乐意和你讨论需求或者你的想法。

最后，如果你在缓存方面需要帮助，你可以随时通过 Twitter 或者邮件联系我。

---

via: https://vincent.composieux.fr/article/i-wrote-gocache-a-complete-and-extensible-go-cache-library

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[beiping96](https://github.com/beiping96)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

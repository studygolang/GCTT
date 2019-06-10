首发于：https://studygolang.com/articles/21026

# 如何在 Golang 中为 RESTful 微服务创建健康检查

想象一下，您最近发布并部署了一段很酷的 RESTful 微服务，您已经使用了一段时间。您松了一口气却听到 Ops 团队说您的服务不稳定。您真的很确定服务应该没问题，可能是它依赖的服务有问题。那该怎么办？

健康检查将来拯救你。它是您的服务返回状态的端点，包括您的服务直接依赖的所有外部服务的连接状态。在这篇文章中，我将展示如何为在多个节点上运行的微服务创建健康检查，该服务将其状态存储在 MongoDB 中并调用 Elasticsearch

> 对你的服务应该监控外部服务，会感觉到非常诧异 ... 你是对的，必须独立监控外部服务。但实际上，某些检查可能会暂时停止。没有什么比临时的更永久。因此，将您的直接依赖项包含在服务状态中是一种很好的做法，因此您（和 Ops）始终知道什么出现问题。

## 设计

正如我前面提到的，假设你有一个在多个节点上运行的微服务，在 MongoDB 中保持状态并调用 Elasticsearch。这样的服务应该是什么样的健康检查？

让我们从不同方面解决问题。

## 端点

一个简单的，让我们遵循行业命名约定并调用端点 `/health`。

## 格式

对于 RESTful 服务，您应始终以 JSON 格式返回 HTTP 状态代码 200 和状态作为内容。

## 内容

这是一个有趣的方面。响应内容必须反映服务的所有关键部分的健康状况。在我们的例子中，它们是节点，与 MongoDB 的连接以及与 Elasticsearch 的连接。代表 Golang 结构，健康状态可能如下所示。

```golang
type HealthStatus struct {
    Nodes   map[string]string `json:"nodes"`
    Mongo   string `json:"mongo"`
    Elastic string `json:"elastic"`
}
```

## 实现

描述健康检查如何适应微服务的描述性方法是将其与其协作的其他模块一起展示。我的示例的框架将具有以下模块：

- main
- mongo
- elastic
- health

### main 模块

`main` 模块负责设置和启动服务：

```golang
package main

import (
    "encoding/json"
    "github.com/upitau/goinbigdata/examples/healthcheck/elastic"
    "github.com/upitau/goinbigdata/examples/healthcheck/health"
    "github.com/upitau/goinbigdata/examples/healthcheck/mongo"
    "net/http"
)

func main() {
    healthService := health.New([]string{"node1", "node2", "node3"}, mongo.New(), elastic.New())
    http.HandleFunc("/health", statusHandler(healthService))
    http.ListenAndServe("localhost:8080", nil)
}

func statusHandler(healthService health.Service) func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        bytes, err := JSON.MarshalIndent(healthService.Health(), "", "\t")
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        w.Write(bytes)
    }
}
```

请注意，健康检查服务需要访问 mongo 和 elastic 模块。

### mongo 和 elastic 模块

我将使用 `rand` 包来模拟 `MongoDB`，`Elasticsearch` 和节点中发生的随机错误。下面是一个简单的模拟 `mongo` 的模块。`elastic` 模块类似。

```golang
package mongo

import (
    "math/rand"
    "errors"
)

type Service interface {
    Health() error
    // Business methods Go here
}

func New() Service {
    return &service{}
}

type service struct {
    // Some fields
}

func (s *service) Health() error {
    if rand.Intn(2) > 0 {
        return errors.New("Service unavailable")
    }
    return nil
}
```

### health 模块

最后 `health` 模块本身：

```golang
package health

import (
    "github.com/upitau/goinbigdata/examples/healthcheck/mongo"
    "github.com/upitau/goinbigdata/examples/healthcheck/elastic"
    "math/rand"
    "fmt"
)

type HealthStatus struct {
    Nodes   map[string]string `json:"nodes"`
    Mongo   string `json:"mongo"`
    Elastic string `json:"elastic"`
}

type Service interface {
    Health() HealthStatus
}

type service struct {
    nodes   []string
    mongo   mongo.Service
    elastic elastic.Service
}

func New(nodes []string, mongo mongo.Service, elastic elastic.Service) Service {
    return &service{
        nodes: nodes,
        mongo: mongo,
        elastic: elastic,
    }
}

func (s *service) Health() HealthStatus {
    nodesStatus := make(map[string]string)
    for _, n := range s.nodes {
        if rand.Intn(10) > 7 {
            nodesStatus[n] = "Node ERROR: Node not responding"
        } else {
            nodesStatus[n] = "OK"
        }
    }

    mongoStatus := "OK"
    if err := s.mongo.Health(); err != nil {
        mongoStatus = fmt.Sprintf("Mongo ERROR: %s", err)
    }

    elasticStatus := "OK"
    if err := s.elastic.Health(); err != nil {
        elasticStatus = fmt.Sprintf("Elastic ERROR: %s", err)
    }

    return HealthStatus{
        Nodes: nodesStatus,
        Mongo: mongoStatus,
        Elastic: elasticStatus,
    }
}
```

请注意，错误消息遵循模式 `<service> ERROR: <detail>`。这很重要，因为健康状态消息旨在被监控系统（例如 Sensu）使用，并且应该易于解析。

该示例的完整代码在[GitHub](https://github.com/upitau/goinbigdata/tree/master/examples/healthcheck) 上。

### 测试

通过 `curl` 调用健康检查服务

```bash
curl localhost:8080/health
```

输出

```bash
{
    "nodes": {
        "node1": "OK",
        "node2": "OK",
        "node3": "OK"
    },
    "mongo": "Mongo ERROR: Service unavailable",
    "elastic": "OK"
}
```

每次运行 curl 命令都可能导致输出不同，因为错误是随机的。

via: http://goinbigdata.com/how-to-create-health-check-for-restful-microservice-in-golang/

作者：[Yury Pitsishin](http://goinbigdata.com/about/)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉

首发于：https://studygolang.com/articles/21028

# 对 Golang 代码调用 Elasticsearch 进行单元测试

[Elastic client](https://github.com/olivere/elastic) 是一款很不错的针对 Go 语言的 Elasticsearch 客户端，在[Working With Elasticsearch](http://goinbigdata.com/working-with-elasticsearch-in-go/) 一文中，我用它举例解释了如何对文档建立索引并搜索文档。你如果希望代码能正常执行，不会被重构或者修改所影响，那么你必须要有一个能覆盖所有代码的测试用例。

在本文中我将教你如何用 Go 语言与 Elasticsearch 做单元测试。并且，此方法也适用于几乎所有语言调用外部 RESTful API.

## 服务调用 Elasticsearch

假设你有一个日志服务，能够获取获取某个应用，最近 n 条日志。例如下面代码中的 `GetLog` 方法！我提供的是我们生产环境已经再用的代码，方便你了解实际的应用场景。

```go
package logging

import (
	"gopkg.in/olivere/elastic.v3"
	"reflect"
)

type Service interface {
	GetLog(app string, lines int) ([]string, error)
}

func NewService(url string) (Service, error) {
	client, err := elastic.NewSimpleClient(elastic.SetURL(url))
	if err != nil {
		return nil, err
	}
	return &service{elasticClient: client}, nil
}

type service struct {
	elasticClient *elastic.Client
}

type Log struct {
	Message string `json:"message"`
}

// GetLog returns limited tail of log sorted by time in ascending order
func (s *service) GetLog(app string, limit int) ([]string, error) {
	termQuery := elastic.NewTermQuery("app", app)

	res, err := s.elasticClient.Search("_all").
		Query(termQuery).
		Sort("@timestamp", false).
		Size(limit).
		Do()

	if err != nil {
		return nil, err
	}

	msgNum := len(res.Hits.Hits)
	if msgNum == 0 {
		return []string{}, nil
	}

	messages := make([]string, msgNum, msgNum)

	var l Log
	for i, item := range res.Each(reflect.TypeOf(l)) {
		l := item.(Log)
		messages[i] = l.Message
	}

	// Reversing messages
	for i := 0; i < msgNum/2; i++ {
		messages[i], messages[msgNum-(i+1)] = messages[msgNum-(i+1)], messages[i]
	}

	return messages, nil
}
```
日志是首先通过 Elasticsearch 倒序获取过来的，转换一下格式之后，在将结果返回给调用端。

## 对服务进行单元测试

一般来讲，我们可以通过 mock 客户端的方式来进行单元测试。不过， `elastic.Client` 是用结构体实现的，所以想要 mock 它的话很麻烦。

更深层次的解决方式应该是 mock Elasticsearch 的 API，这种方式就简单多了。一个解决办法是通过 `httptest.Server` 访问一个预制好的服务接口，这里只返回一些预定义好的 Elasticsearch 查询结果。

```go
package logging

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/olivere/elastic.v3"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLog(t *testing.T) {
	handler := http.NotFound
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	}))
	defer ts.Close()

	handler = func(w http.ResponseWriter, r *http.Request) {
		resp := `{
			  "took" : 122,
			  "timed_out" : false,
			  "_shards" : {
			    "total" : 6,
			    "successful" : 5,
			    "failed" : 1,
			    "failures" : [ {
			      "shard" : 0,
			      "index" : ".kibana",
			      "node" : "jucBX9QkQIini9dLG9tZIw",
			      "reason" : {
				"type" : "search_parse_exception",
				"reason" : "No mapping found for [offset] in order to sort on"
			      }
			    } ]
			  },
			  "hits" : {
			    "total" : 10,
			    "max_score" : null,
			    "hits" : [ {
			      "_index" : "logstash-2016.07.25",
			      "_type" : "log",
			      "_id" : "AVYkNv542Gim_t2htKPU",
			      "_score" : null,
			      "_source" : {
				"message" : "Alice message 10",
				"@version" : "1",
				"@timestamp" : "2016-07-25T22:39:55.760Z",
				"source" : "/Users/yury/logs/alice.log",
				"offset" : 144,
				"type" : "log",
				"input_type" : "log",
				"count" : 1,
				"fields" : null,
				"beat" : {
				  "hostname" : "Yurys-MacBook-Pro.local",
				  "name" : "Yurys-MacBook-Pro.local"
				},
				"host" : "Yurys-MacBook-Pro.local",
				"tags" : [ "beats_input_codec_plain_applied" ],
				"app" : "alice"
			      },
			      "sort" : [ 144 ]
			    }, {
			      "_index" : "logstash-2016.07.25",
			      "_type" : "log",
			      "_id" : "AVYkNv542Gim_t2htKPT",
			      "_score" : null,
			      "_source" : {
				"message" : "Alice message 9",
				"@version" : "1",
				"@timestamp" : "2016-07-25T22:39:55.760Z",
				"source" : "/Users/yury/logs/alice.log",
				"offset" : 128,
				"input_type" : "log",
				"count" : 1,
				"beat" : {
				  "hostname" : "Yurys-MacBook-Pro.local",
				  "name" : "Yurys-MacBook-Pro.local"
				},
				"type" : "log",
				"fields" : null,
				"host" : "Yurys-MacBook-Pro.local",
				"tags" : [ "beats_input_codec_plain_applied" ],
				"app" : "alice"
			      },
			      "sort" : [ 128 ]
			    }, {
			      "_index" : "logstash-2016.07.25",
			      "_type" : "log",
			      "_id" : "AVYkNv542Gim_t2htKPR",
			      "_score" : null,
			      "_source" : {
				"message" : "Alice message 8",
				"@version" : "1",
				"@timestamp" : "2016-07-25T22:39:55.760Z",
				"type" : "log",
				"input_type" : "log",
				"source" : "/Users/yury/logs/alice.log",
				"count" : 1,
				"fields" : null,
				"beat" : {
				  "hostname" : "Yurys-MacBook-Pro.local",
				  "name" : "Yurys-MacBook-Pro.local"
				},
				"offset" : 112,
				"host" : "Yurys-MacBook-Pro.local",
				"tags" : [ "beats_input_codec_plain_applied" ],
				"app" : "alice"
			      },
			      "sort" : [ 112 ]
			    } ]
			  }
			}`

		w.Write([]byte(resp))
	}

	s, err := MockService(ts.URL)
	assert.NoError(t, err)

	expectedMessages := []string{
		"Alice message 8",
		"Alice message 9",
		"Alice message 10",
	}
	actualMessages, err := s.GetLog("app", 3)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessages, actualMessages)
}

func MockService(url string) (Service, error) {
	client, err := elastic.NewSimpleClient(elastic.SetURL(url))
	if err != nil {
		return nil, err
	}
	return &service{elasticClient: client}, nil
}
```

预制的结果可以提前写入到一个文件里面，在代码里读取就可以了。源代码可以访问[GitHub](https://github.com/upitau/goinbigdata/tree/master/examples/elastictest) 获取

> 如果你不清楚 `w.Write([]byte(resp))` 中的 `res` 为什么需要被转换成 []byte 的话 , 可以看一下这篇文章： [How To Correctly Serialize JSON String In Golang](http://goinbigdata.com/how-to-correctly-serialize-json-string-in-golang/)。

## 关于测试的一些注意点

尽管本文主要介绍的是如何通过 Go 语言编写外部调用的测试代码，但不得不说集成测试更佳。集成测试是基于整个系统的各个组件共同运行，测试的结果更接近于生产环境，能够提供更高的质量保障。

但是，集成测试一般来说更难实现，而且需要花费更多的时间。因此很少有人编写集成测试的代码。

## 最后

当测试 Go 客户端访问外服 API 的代码的时候，最好的方式就是 mock 外部服务，如果外服服务是通过结构体实现的时候，可以直接 mock 外部 API，返回一些预制的数据方便我们完成真实情况的测试。

---

via: http://goinbigdata.com/unit-testing-golang-code-calling-elasticsearch/

作者：[Yury Pitsishin](http://goinbigdata.com/about/)
译者：[JYSDeveloper](https://github.com/JYSDeveloper)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

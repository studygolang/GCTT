首发于：https://studygolang.com/articles/19579

# 如何在 Golang 中使用 MongoDB 构建微服务

如今，Golang 越来越流行于编写 RESTful 微服务。这些服务常常使用 MongoDB 作为持久性存储。在本文中，我们将使用 Go 和 MongoDB 构建一个简单的 *书店* 微服务。我们将使用 *mgo* 驱动程序连接 MongoDB，并使用 *curl* 测试微服务。

## MongoDB

MongoDB 以其简单、高可用性和面向文档的特性风靡市场。与关系型相比，文档化的优点是：

+ 文档化在许多编程语言中可以对应原生数据类型。
+ 嵌入式文档和数组减少了对高昂的连接需求。
+ 动态模型支持连贯多态性。

### 什么是文档

文档只是由字段和值对组成的数据结构。字段的值可以包含其他文档，数组和文档数组。MongoDB 文档和 JSON 对象类似，每个文档都作为记录存储在 MongoDB 集合中。

例如，一本书可以被如下文档（json）表示：

```json
{
	"isbn":    "0134190440",
	"title":   "The Go Programming Language",
	"authors": ["Alan A. A. Donovan", "Brian W. Kernighan"],
	"price":   "$34.57"
}
```

### 集合

MongoDB 存储相似的文档在同一个集合。例如，我们将存储 books 在 *books 集合*。如果您有关系型背景知识，则集合类似于表。集合的不同之处是不强制要求任何结构，尽管它意味着存储在同一个集合中的文档是相似的。

### 查询

如果您想从 MongoDB 获取数据，你首先要查询它。 *查询* 是 MongoDB 的一个概念，用于指定请求哪些数据的一组筛选器参数。MongoDB 使用 *json* 和 *bson* ( 二进制 JSON) 编写查询。获取具有指定 isbn 图书的查询示例看起来如下：

```json
{
    "isbn": "1234567"
}
```

## Go 的 MongoDB 驱动

mgo ( 发音是 mango) 是一个针对 Golang 开发的 MongoDB 驱动。它的 [API](https://godoc.org/gopkg.in/mgo.v2) 非常简单，并且遵循标准的 Go 风格。稍后我们将看到它是如何帮助构建微服务的 CRUD ( 创建、读取、更新、删除 ) 操作，但首先让我们来熟悉一下会话管理。

### 会话管理

获得一个会话：

```go
session, err := mgo.Dial("localhost")
```

单个 session 不允许并发处理，因此通常需要多个 session。获取另一个 session 最快的办法是 复制一个已经存在的。确保您使用后关闭会话：

```go
anotherSession := session.Copy()
defer anotherSession.Close()
```

### 搜索文档

mgo 和 [bson](https://godoc.org/gopkg.in/mgo.v2/bson) 包一起使用，简化了查询的编写。

从集合中获取所有文档：

```go
c := session.DB("store").C("books")

var books []Book
err := c.Find(bson.M{}).All(&books)
```

从集合中搜索一个文档：

```go
c := session.DB("store").C("books")

isbn := ...
var book Book
err := c.Find(bson.M{"isbn": isbn}).One(&book)
```

### 创建新文档

```go
c := session.DB("store").C("books")
err = c.Insert(book)
```

### 更新文档

```go
c := session.DB("store").C("books")
err = c.Update(bson.M{"isbn": isbn}, &book)
```

### 删除文档

```go
c := session.DB("store").C("books")
err := c.Remove(bson.M{"isbn": isbn})
```

## 用 Go 编写使用 MongoDB 的微服务

下面是一个用 Go 编写的完整的书店微服务示例，并用 MongoDB 做支持。您可以从 [Github](https://github.com/upitau/goinbigdata/tree/master/examples/mongorest) 下载这个例子。

> 这个服务使用 *Goji* 做路由。如果您以前没有使用过 Goji，可以看一下 [怎样用 Goji 写 RESTful 服务](http://goinbigdata.com/restful-web-service-in-go-using-goji/)。

```go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"goji.io"
	"goji.io/pat"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}

func ResponseWithJSON(w http.ResponseWriter, JSON []byte, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(json)
}

type Book struct {
	ISBN    string   `json:"isbn"`
	Title   string   `json:"title"`
	Authors []string `json:"authors"`
	Price   string   `json:"price"`
}

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)

	mux := Goji.NewMux()
	mux.HandleFunc(pat.Get("/books"), allBooks(session))
	mux.HandleFunc(pat.Post("/books"), addBook(session))
	mux.HandleFunc(pat.Get("/books/:isbn"), bookByISBN(session))
	mux.HandleFunc(pat.Put("/books/:isbn"), updateBook(session))
	mux.HandleFunc(pat.Delete("/books/:isbn"), deleteBook(session))
	http.ListenAndServe("localhost:8080", mux)
}

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB("store").C("books")

	index := mgo.Index{
		Key:        []string{"isbn"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func allBooks(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		c := session.DB("store").C("books")

		var books []Book
		err := c.Find(bson.M{}).All(&books)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed get all books: ", err)
			return
		}

		respBody, err := JSON.MarshalIndent(books, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func addBook(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		var book Book
		decoder := JSON.NewDecoder(r.Body)
		err := decoder.Decode(&book)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB("store").C("books")

		err = c.Insert(book)
		if err != nil {
			if mgo.IsDup(err) {
				ErrorWithJSON(w, "Book with this ISBN already exists", http.StatusBadRequest)
				return
			}

			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed insert book: ", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Location", r.URL.Path+"/"+book.ISBN)
		w.WriteHeader(http.StatusCreated)
	}
}

func bookByISBN(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		isbn := pat.Param(r, "isbn")

		c := session.DB("store").C("books")

		var book Book
		err := c.Find(bson.M{"isbn": isbn}).One(&book)
		if err != nil {
			ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
			log.Println("Failed find book: ", err)
			return
		}

		if book.ISBN == "" {
			ErrorWithJSON(w, "Book not found", http.StatusNotFound)
			return
		}

		respBody, err := JSON.MarshalIndent(book, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		ResponseWithJSON(w, respBody, http.StatusOK)
	}
}

func updateBook(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		isbn := pat.Param(r, "isbn")

		var book Book
		decoder := JSON.NewDecoder(r.Body)
		err := decoder.Decode(&book)
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			return
		}

		c := session.DB("store").C("books")

		err = c.Update(bson.M{"isbn": isbn}, &book)
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed update book: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "Book not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func deleteBook(s *mgo.Session) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		session := s.Copy()
		defer session.Close()

		isbn := pat.Param(r, "isbn")

		c := session.DB("store").C("books")

		err := c.Remove(bson.M{"isbn": isbn})
		if err != nil {
			switch err {
			default:
				ErrorWithJSON(w, "Database error", http.StatusInternalServerError)
				log.Println("Failed delete book: ", err)
				return
			case mgo.ErrNotFound:
				ErrorWithJSON(w, "Book not found", http.StatusNotFound)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
```

## 用 curl 测试

[curl](https://curl.haxx.se/) 是一个构建和测试 RESTful 微服务不可或缺的工具。而且在 RESTful API 文档中经常使用 curl 命令来提供 API 调用的示例。

### 添加新书

请求示例：

```shell
curl -X POST -H "Content-Type: application/json" -d @body.json http://localhost:8080/books

body.json:
{
    "isbn":    "0134190440",
    "title":   "The Go Programming Language",
    "authors": ["Alan A. A. Donovan", "Brian W. Kernighan"],
    "price":   "$34.57"
}
```

响应示例：

```shell
201 Created
```

### 获取所有书

请求示例：

```shell
curl -H "Content-Type: application/json" http://localhost:8080/books
```

响应示例：

```shell
200 OK
[
  {
    "ISBN": "0134190440",
    "Title": "The Go Programming Language",
    "Authors": [
      "Alan A. A. Donovan",
      "Brian W. Kernighan"
    ],
    "Price": "$34.57"
  },
  {
    "ISBN": "0321774639",
    "Title": "Programming in Go: Creating Applications for the 21st Century (Developer's Library)",
    "Authors": [
      "Mark Summerfield"
    ],
    "Price": "$31.20"
  }
]
```

### 获取一本书：

请求示例：

```shell
curl -H "Content-Type: application/json" http://localhost:8080/books/0134190440
```

响应示例：

```shell
200 OK
{
  "ISBN": "0134190440",
  "Title": "The Go Programming Language",
  "Authors": [
    "Alan A. A. Donovan",
    "Brian W. Kernighan"
  ],
  "Price": "$34.57"
}
```

### 更新书

请求示例：

```shell
curl -X PUT -H "Content-Type: application/json" -d @body.json http://localhost:8080/books/0134190440

body.json:
{
    "isbn":    "0134190440",
    "title":   "The Go Programming Language",
    "authors": ["Alan A. A. Donovan", "Brian W. Kernighan"],
    "price":   "$20.00"
}
```

响应示例：

```shell
204 No Content
```

### 删除书

请求示例：

```shell
curl -X DELETE -H "Content-Type: application/json" -d @body.json http://localhost:8080/books/0134190440
```

响应示例：

```shell
204 No Content
```

## 最后

MongoDB 是用 Go 编写微服务的一个非常流行的后端。Go 的 MongoDB 驱动程序 (mgo) 是惯用的并非常容易使用。如果您正在构建、测试或记录 RESTful 服务，不要忽略 *curl*。

---

via: http://goinbigdata.com/how-to-build-microservice-with-mongodb-in-golang/

作者：[Yury Pitsishin](http://goinbigdata.com/about/)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉出品

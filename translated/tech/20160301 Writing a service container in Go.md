#用 GO 写一个服务容器（ Service Container ）

我最近一直在做一个相当大 API 项目，里面包括很多路由规则（ routes ），服务接口（ services ），和处理函数（ handlers ）等。首先，我注意到 `main.go` 文件的启动过程开始越来越臃肿。

为了避免设置全局的服务接口，我使用共享结构体（ struct ）将服务接口与处理函数绑定在一起。举个例子：  

 main.go
```go
package main

func main() {
    r := gin.Default()

    userRepo := models.NewUserRepo(
        drives.DataStore().C("users"),
    )

    userHandler := handlers.NewUserHandler(userRepo)
    r.GET("/api/v1/users", userHandler.FindAll)
    r.Run(":8080")
}
```
user_handler.go
```go
type UserHandler struct {  
    userRepo *models.UserRepo
}

func NewUserHandler(userRepo *models.UserRepo) *UserHandler {  
    return &UserHandler{
        userRepo,
    }
}

func (userHandler *UserHandler) FindAll(c *gin.Context) {  
    users, err := userHandler.userRepo.FindAll()
    if err != nil {
        c.JSON(404, nil)
        return
    }
    c.JSON(200, users)
    return
}
```
这些代码工作的很好，但是，你会发现 main.go 中，我只是写了很少的启动过程，仅仅包含一个处理函数和一个持久化数据（ repository ）。（译注：用这种方式写代码比较麻烦且臃肿）。  

于是我想用 GO 写一个容器。我找不到一个喜欢的第三方库来解决这个事情。所以，想出了下面这段代码。
```go
import(  
    "sync"
) 

type Container struct{  
    mux sync.RWMutex
    m map[string]interface{}
}

// Add service
func (c *Container) Add(name string, object interface{}) {  
    if c.m == nil {
        c.m = make(map[string]interface{})
    }    
    c.mux.Lock()
    c.m[name] = object
    c.mux.Unlock()
}

// Remove service
func (c *Container) Remove(name string) Container {  
    c.mux.Lock()
    delete(c.m, name)
    c.mux.Unlock()
}

// Get a service
func (c *Container) Get(name string) (object interface{}, bool) {  
    c.mux.RLock()
    object, ok = c.m[name]
    c.mux.RUnlock()
    return object, ok
}
```
注意这段代码的每个方法都用了 `mutex lock` 来避免容器的并发问题。

现在代码可以这样写。。。
```go
func GetContainer() Container {  
    c := new(container.Container)
    c.Add("user.handler", handlers.UserHandler)
    return c 
}
```

现在的 main.go
```go
func main() {  
    container := container.GetContainer()

    userHandler, ok := container.Get("user.handler")

    if !ok {
        log.Fatal("Service not found")    
    }

    r.GET(
       "/api/v1/users", 
       userHandler.(*handlers.UserHandler).FindAll(),
    )
}
```
_同步方面的代码参考自 itsmontoya ，向他致敬_

现在我已经将启动过程很简洁地封装成了包。我觉得一个有 PHP 语言背景的人在考虑这段语法实现时候会参考 [pimple](http://pimple.sensiolabs.org/) 框架。

我已经把这个抽象成了自己的库，实现在[这里](https://github.com/EwanValentine/Vertebrae) 。

----------------

via: https://ewanvalentine.io/writing-a-service-container-in-go/

作者：[Ewan Valentine](https://ewanvalentine.io/author/ewan/)
译者：[j.zhongming](https://github.com/j.zhongming)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推

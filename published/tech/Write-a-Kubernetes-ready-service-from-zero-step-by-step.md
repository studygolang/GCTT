已发布：https://studygolang.com/articles/12156

# 从零开始一步步构建运行在 Kubernetes 上的服务

如果你用 Go 写过程序，就会发现用 Go 来写服务是很简单的事。比如说，只要几行代码就可以跑起来一个 HTTP 服务。但是如果我们想让服务在生产环境运行，我们还需要添加什么呢？本文将通过写一个能在 Kubernetes 上运行的服务的例子，来讨论上述问题。

文中所有的例子可以在 [这里（按标签分类）](https://github.com/rumyantseva/advent-2017/tree/all-steps) ，或者 [这里（按 commit 分类）](https://github.com/rumyantseva/advent-2017/commits/master) 找到。

## 第一步 最简单的服务

从一个最简单的应用开始：
`main.go`

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/home", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Hello! Your request was processed.")
	})
	http.ListenAndServe(":8000", nil)
}
```
执行 `go run main.go` 即可运行程序。用 curl 命令 `curl -i http://127.0.0.1:8000/home` 可以看到程序返回值。不过目前在终端并没有多少**状态信息**。

## 第二步 添加日志

添加一个 logger 便于查看执行到哪一行、记录错误信息以及其他重要状态。本例中简便起见，会使用 Go 标准库中的 log，而线上生产环境你或许会使用到更强大的日志系统，例如： [glog](https://github.com/golang/glog) 或者 [logrus](https://github.com/sirupsen/logrus) 。

代码中有三个地方需要添加日志：服务开始时候、服务准备好可以接受请求时以及当 `http.ListenAndServe` 返回错误时。具体代码如下：

```go
func main() {
	log.Print("Starting the service...")

	http.HandleFunc("/home", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "Hello! Your request was processed.")
	})

	log.Print("The service is ready to listen and serve.")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
```
一步步趋向完美。

## 第三步 添加路由器

为了让应用更加可用，需要添加一个路由器（router），路由器能够以一种简单的方式处理各种不同的 URI 和 HTTP 方法，以及匹配一些其他的规则。Go 标准库中没有包含路由器（router），本文使用 [gorilla/mux](https://github.com/gorilla/mux) 库，该库提供的路由器能够很好地和标准库 `net/http` 兼容。

服务中如果包含了一定数量的不同路由规则，那么最好是把路由相关的代码单独封装到几个独立的 function 或者是一个 package 中。本文中，会把规则定义和初始化路由器的代码放到 `handlers` package 中（[这里](https://github.com/rumyantseva/advent-2017/commit/1a61e7952e227e33eaab81404d7bff9278244080) 可以看到完整的改动）。

我们添加一个 `Router` 方法，该方法返回一个配置好的路由器变量，其中 `home` 方法处理 `/home` 路径的请求。个人建议处理方法和路由分开写：

`handlers/handlers.go`

```go
package handlers

import (
	"github.com/gorilla/mux"
)

// Router register necessary routes and returns an instance of a router.
func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/home", home).Methods("GET")
	return r
}
```

`handlers/home.go`

```go
package handlers

import (
	"fmt"
	"net/http"
)

// home is a simple HTTP handler function which writes a response.
func home(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "Hello! Your request was processed.")
}
```

然后`main.go`中做点小改动：

```go
package main

import (
	"log"
	"net/http"

	"github.com/rumyantseva/advent-2017/handlers"
)

// How to try it: go run main.go
func main() {
	log.Print("Starting the service...")
	router := handlers.Router()
	log.Print("The service is ready to listen and serve.")
	log.Fatal(http.ListenAndServe(":8000", router))
}
```

## 第四步 添加测试

这一步要开始加点测试了。我们用到了 `httptest` 包。`Router` 方法的测试代码如下：

`handlers/handlers_test.go`：

```go
package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	r := Router()
	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/home")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Status code for /home is wrong. Have: %d, want: %d.", res.StatusCode, http.StatusOK)
	}

	res, err = http.Post(ts.URL+"/home", "text/plain", nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Status code for /home is wrong. Have: %d, want: %d.", res.StatusCode, http.StatusMethodNotAllowed)
	}

	res, err = http.Get(ts.URL + "/not-exists")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("Status code for /home is wrong. Have: %d, want: %d.", res.StatusCode, http.StatusNotFound)
	}
}
```

检查了 `GET` 请求 `/home` 路径是否返回 `200`，而 `POST` 请求该路径应该要返回 `405`。请求不存在的路由期望返回`404`。实际上，这样子测有点太冗余了，`gorilla/mux` 中已经包含类似的测试，所以测试代码可以简化下。

对于 `home` 来说，检查其返回得 code 和 body 值即可。

`handlers/home_test.go`

```go
package handlers

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHome(t *testing.T) {
	w := httptest.NewRecorder()
	home(w, nil)

	resp := w.Result()
	if have, want := resp.StatusCode, http.StatusOK; have != want {
		t.Errorf("Status code is wrong. Have: %d, want: %d.", have, want)
	}

	greeting, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if have, want := string(greeting), "Hello! Your request was processed."; have != want {
		t.Errorf("The greeting is wrong. Have: %s, want: %s.", have, want)
	}
}
```

运行`go test`开始测试。

```
$ go test -v ./...
?       github.com/rumyantseva/advent-2017      [no test files]
=== RUN   TestRouter
--- PASS: TestRouter (0.00s)
=== RUN   TestHome
--- PASS: TestHome (0.00s)
PASS
ok      github.com/rumyantseva/advent-2017/handlers     0.018s
```

## 第五步 添加配置

下一个比较重要的问题是：服务需要能够是可配置的。目前的代码中，写死了监听 `8000` 端口，可能把端口值改成可配置的，会更有用一些。[The Twelve-Factor App manifesto](https://12factor.net/)，这篇文章详尽阐述了如何去写好服务，文中提倡在环境变量中存放配置。后面代码展示本例如何利用上环境变量：

`main.go`

```go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/rumyantseva/advent-2017/handlers"
)

// How to try it: PORT=8000 go run main.go
func main() {
	log.Print("Starting the service...")

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("Port is not set.")
	}

	r := handlers.Router()
	log.Print("The service is ready to listen and serve.")
	log.Fatal(http.ListenAndServe(":"+port, r))
}
```

上例中，若没有设置 port 值，会返回错误。如果配置错误的话，没有必要继续执行后面代码。

## 第六步 添加 Makefile

前几天看过一篇关于 `make` 的 [文章](https://blog.gopheracademy.com/advent-2017/make/)，如果想要把一些重复性高的常用的东西做成自动化，推荐看看该文。让我们看下怎么用上 `make` ，目前有两个操作：运行测试、编译并运行服务，把这两个操作加到 Makefile 文件中。这里我们用到 `go build` 命令，后面会运行编译好的二进制文件，这种方式更符合"在生产环境上运行"的目标，所以就不会用到 `go run` 命令了。

`Makefile`

```makefile
APP?=advent
PORT?=8000

clean:
	rm -f ${APP}

build: clean
	go build -o ${APP}

run: build
	PORT=${PORT} ./${APP}

test:
	go test -v -race ./...
```

上例中把二进制文件名单独放到变量 `APP` 中，减少重复定义名称次数。

运行程序前，先删除旧的二进制文件（存在的话），然后编译代码、设置正确的环境变量并运行新生成的二进制文件，这些操作可以通过执行 `make run` 命令完成。

## 第七步 添加版本控制

这一步要添加到服务中的技巧是版本控制功能。某些场景下，知道生产环境中所使用的具体是哪个构建和 commit 以及什么时间构建的这类信息是非常有用的。

添加一个新的包 `version` 来保存这些信息。

`version/version.go`

```go
package version

var (
	// BuildTime is a time label of the moment when the binary was built
	BuildTime = "unset"
	// Commit is a last commit hash at the moment when the binary was built
	Commit = "unset"
	// Release is a semantic version of current build
	Release = "unset"
)
```

程序启动时，会将这些变量打到日志中。

`main.go`

```go
...
func main() {
	log.Printf(
		"Starting the service...\ncommit: %s, build time: %s, release: %s",
		version.Commit, version.BuildTime, version.Release,
	)
...
}
```

也可以把这些信息添加到 `home` handler 中（别忘了更新对应的测试方法）：

`handlers/home.go`

```go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rumyantseva/advent-2017/version"
)

// home is a simple HTTP handler function which writes a response.
func home(w http.ResponseWriter, _ *http.Request) {
	info := struct {
		BuildTime string `json:"buildTime"`
		Commit    string `json:"commit"`
		Release   string `json:"release"`
	}{
		version.BuildTime, version.Commit, version.Release,
	}

	body, err := json.Marshal(info)
	if err != nil {
		log.Printf("Could not encode info data: %v", err)
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
```

通过 Go 链接器在编译时设置 `BuildTime` ，`Commit` 以及 `Release` 变量。

先在 Makefile 中添加新变量：

`Makefile`

```
RELEASE?=0.0.1
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
```

`COMMIT` 和 `BUILD_TIME`（译者注：原文这里写 `RELEASE` ，可能有误）通过已有的命令获取，`RELEASE` 的赋值按照 [语义化版本控制规范](https://dave.cheney.net/2016/06/24/gophers-please-tag-your-releases) 来。

好，现在重写 Makefile 的 `build` 目标，用上上面定义的变量：

`Makefile`

```
build: clean
	go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o ${APP}
```

将 `PROJECT` 变量添加到 `Makefile` 开头地方（减少多处定义）。

`Makefile`

```
PROJECT?=github.com/rumyantseva/advent-2017
```

本步所有代码变更记录可以在 [这里](https://github.com/rumyantseva/advent-2017/commit/eaa4ff224b32fb343f5eac2a1204cc3806a22efd) 找到。可以多动手尝试运行下 `make run` 命令，看看具体是怎么工作的。

## 第八步 减少依赖

之前代码有个不尽如人意的点：`handler` 包依赖 `version` 包。做个简单的改动，让 `home` 处理器变成可配置的，减少依赖：

`handlers/home.go`

```go
// home returns a simple HTTP handler function which writes a response.
func home(buildTime, commit, release string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		...
	}
}
```
同样，别忘了 [改](https://github.com/rumyantseva/advent-2017/commit/e73b996f8522b736c150e53db059cf041c7c3e64) 测试代码。

## 第九步 添加“健康”检查功能（health checks）

某些情况下，想在 kubernetes 上跑服务，需要添加“健康”检查功能：[ 存活探针（liveness probe） 及就绪探针（readiness probe）](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/)。存活探针（liveness probe）目的是测试程序是否还在跑。如果存活探针（liveness probe）检测失败，服务会被重启。就绪探针（readiness probe）目的是测试程序是否准备好可以接受请求。如果就绪探针（readiness probe）检测失败，该容器会从服务负载均衡器中移除。

实现存活探针（liveness probe）的方式，可以简单写一个 handler 返回 `200`:

`handlers/healthz.go`

```go
// healthz is a liveness probe.
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
```

就绪探针（readiness probe）实现方式类似，不同的就是可能要等待某事件完成（例如：数据库已起来）：

`handlers/readyz.go`

```go
// readyz is a readiness probe.
func readyz(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

```

当 `isReady` 有值并且为 `true`，返回 `200`。

下面是如何使用它的例子：

`handlers.go`

```go
func Router(buildTime, commit, release string) *mux.Router {
	isReady := &atomic.Value{}
	isReady.Store(false)
	go func() {
		log.Printf("Readyz probe is negative by default...")
		time.Sleep(10 * time.Second)
		isReady.Store(true)
		log.Printf("Readyz probe is positive.")
	}()

	r := mux.NewRouter()
	r.HandleFunc("/home", home(buildTime, commit, release)).Methods("GET")
	r.HandleFunc("/healthz", healthz)
	r.HandleFunc("/readyz", readyz(isReady))
	return r
}
```

设置等待 10 s 后服务可以处理请求。当然，实际业务代码不会有空等 10s 的情况，这里是模拟 cache warming（如果有用 cache）或者其他情况。

代码改动 [GitHub](https://github.com/rumyantseva/advent-2017/commit/e73b996f8522b736c150e53db059cf041c7c3e64) 上可以找到。

注意：如果流量过大，服务节点的响应会不稳定。例如，存活探针（liveness probe）检测会因为超时失败。这就是为什么一些工程师不用存活探针（liveness probe）的原因。个人认为，当发现请求越来越多时候，最好是去扩容；例如可以参考 [scale pods with HPA](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)。

## 第十步 添加平滑关闭功能

关闭服务时，最好是不要立即中断连接、请求或者其他一些操作，而应该平滑关闭。Go 从 1.8 版本支持平滑关闭`http.Server`。下面看看怎么用：

`main.go`

```go
func main() {
    ...
	r := handlers.Router(version.BuildTime, version.Commit, version.Release)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()
	log.Print("The service is ready to listen and serve.")

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	srv.Shutdown(context.Background())
	log.Print("Done")
}
```
收到 `SIGINT` 或 `SIGTERM` 任意一个系统信号，服务平滑关闭。

注意：当我在写这段代码的时候，我（作者）尝试去捕获 `SIGKILL` 信号。之前在不同的库中有看到过这种用法，我确认这样是行的通的。但是后来 Sandor Szücs [指出](https://twitter.com/sszuecs/status/941582509565005824) ，不可能获取到 `SIGKILL` 信号。发出 `SIGKILL` 信号后，程序会直接结束。

## 第十一步 添加 Dockerfile

程序基本上可以在 Kubernetes 上跑了。这一步进行 Docker 化。

先添加一个简单的 `Dockerfile`，如下：

`Dockerfile`：

```
FROM scratch

ENV PORT 8000
EXPOSE $PORT

COPY advent /
CMD ["/advent"]
```

创建了一个最小的容器，复制二进制到容器内然后运行（别忘了 `PORT` 配置变量）。

扩展 `Makefile`，使其能够构建镜像以及运行容器。同时添加 `GOOS` 和 `GOARCH` 变量，在 `build` 的目标中交叉编译要用到。

`Makefile`

```
...

GOOS?=linux
GOARCH?=amd64

...

build: clean
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
		-ldflags "-s -w -X ${PROJECT}/version.Release=${RELEASE} \
		-X ${PROJECT}/version.Commit=${COMMIT} -X ${PROJECT}/version.BuildTime=${BUILD_TIME}" \
		-o ${APP}

container: build
	docker build -t $(APP):$(RELEASE) .

run: container
	docker stop $(APP):$(RELEASE) || true && docker rm $(APP):$(RELEASE) || true
	docker run --name ${APP} -p ${PORT}:${PORT} --rm \
		-e "PORT=${PORT}" \
		$(APP):$(RELEASE)

...
```

添加了 `container` 和 `run` goal，前者构建镜像，后者从容器启动程序。所有改动 [这里](https://github.com/rumyantseva/advent-2017/commit/909fef6d585c85c5e16b5b0e4fdbdf080893b679) 可以找到。

请尝试运行 `make run`命令，检查所有过程是否正确。

## 第十二步 添加 vendor

项目中依赖了外部代码（[github.com/gorilla/mux](https://github.com/gorilla/mux)），所以肯定要 [加入依赖管理](https://github.com/rumyantseva/advent-2017/commit/7ffa56a78400367e5d633521dee816b767d7d05d)。如果引入 [dep](https://github.com/golang/dep) 的话，就只要执行 `dep init`：

```
$ dep init
  Using ^1.6.0 as constraint for direct dep github.com/gorilla/mux
  Locking in v1.6.0 (7f08801) for direct dep github.com/gorilla/mux
  Locking in v1.1 (1ea2538) for transitive dep github.com/gorilla/context
```

会创建 `Gopkg.toml` 和 `Gopkg.lock` 文件以及 `vendor` 目录。个人观点，推荐 push `vendor` 到 git，重要的项目尤其应该 push 上去。

## 第十三步 Kubernetes

[最后一步](https://github.com/rumyantseva/advent-2017/commit/27b256191dc8d4530c895091c49b8a9293932e0f)，将程序部署到 Kubernets 上运行。本地环境最简单方式就是安装、配置 [minikube](https://github.com/kubernetes/minikube)。

Kubernetes 从 Docker registry 拉取镜像。本文中，使用公共 Docker registry--[Docker Hub](https://hub.docker.com/)。`Makefile` 中还要添加一个变量和命令：

```
CONTAINER_IMAGE?=docker.io/webdeva/${APP}

...

container: build
	docker build -t $(CONTAINER_IMAGE):$(RELEASE) .

...

push: container
	docker push $(CONTAINER_IMAGE):$(RELEASE)
```

`CONTAINER_IMAGE` 变量定义了 push、pull 镜像的 Docker registry repo，路径中包含了用户名（webdeva）。如果没有 [hub.docker.com](hub.docker.com) 账户，请创建账户并通过 `docker login` 登录。这样就可以 push 镜像了。

运行 `make push`：

```
$ make push
...
docker build -t docker.io/webdeva/advent:0.0.1 .
Sending build context to Docker daemon   5.25MB
...
Successfully built d3cc8f4121fe
Successfully tagged webdeva/advent:0.0.1
docker push docker.io/webdeva/advent:0.0.1
The push refers to a repository [docker.io/webdeva/advent]
ee1f0f98199f: Pushed
0.0.1: digest: sha256:fb3a25b19946787e291f32f45931ffd95a933100c7e55ab975e523a02810b04c size: 528
```

成功了~！然后可以在[这里找到镜像](https://hub.docker.com/r/webdeva/advent/tags/)。

接下来，定义必要的 Kubernetes 配置（manifest）。通常，一个服务至少需要设置 deployment、service 和 ingress 配置。默认情况，manifest 都是静态的，即其中不能使用任何变量。不过可以通过 [helm 工具](https://github.com/kubernetes/helm) 创建更灵活的配置。

本例中，我们没有用 `helm`，但如果能定义两个变量：`ServiceName` 和 `Release` 会更加实用。后面通过 `sed` 命令替换“变量”为实际值。

先看下 deployment 配置：

`deployment.yaml`

```
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: {{ .ServiceName }}
  labels:
    app: {{ .ServiceName }}
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 50%
      maxSurge: 1
  template:
    metadata:
      labels:
        app: {{ .ServiceName }}
    spec:
      containers:
      - name: {{ .ServiceName }}
        image: docker.io/webdeva/{{ .ServiceName }}:{{ .Release }}
        imagePullPolicy: Always
        ports:
        - containerPort: 8000
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8000
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8000
        resources:
          limits:
            cpu: 10m
            memory: 30Mi
          requests:
            cpu: 10m
            memory: 30Mi
      terminationGracePeriodSeconds: 30

```
Kubernetes 的配置要讲清楚可以单独写一篇文章了，这里用到了容器镜像和存活探针（liveness probe）、就绪探针（readiness probe）检测功能，去哪里找镜像，以及检测模块的路径前文都有阐述。

一个经典的服务更简单：

`service.yaml`

```
apiVersion: v1
kind: Service
metadata:
  name: {{ .ServiceName }}
  labels:
    app: {{ .ServiceName }}
spec:
  ports:
  - port: 80
    targetPort: 8000
    protocol: TCP
    name: http
  selector:
    app: {{ .ServiceName }}
```

最后，定义下 ingress。定义从外部访问访问 Kubernetes 中服务的规则。这里假定把服务部署到到 `advent.test` 域上（实际不是）。

`ingress.yaml`

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: nginx
    ingress.kubernetes.io/rewrite-target: /
  labels:
    app: {{ .ServiceName }}
  name: {{ .ServiceName }}
spec:
  backend:
    serviceName: {{ .ServiceName }}
    servicePort: 80
  rules:
  - host: advent.test
    http:
      paths:
      - path: /
        backend:
          serviceName: {{ .ServiceName }}
          servicePort: 80
```

验证配置是否正确，需要安装、运行 `minikube`，官方文档在[这里](https://github.com/kubernetes/minikube#installation)，还需要安装 [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) 工具，用来提供配置和检验服务。

启动 `minikube`，启动 `ingress`，准备 `kubectl`，我们需要运行如下命令：

```
minikube start
minikube addons enable ingress
kubectl config use-context minikube
```

接下来，给 `Makefile` 添加新目标：安装服务到 `minikube` 上。

`Makefile`

```
minikube: push
	for t in $(shell find ./kubernetes/advent -type f -name "*.yaml"); do \
        cat $$t | \
        	gsed -E "s/\{\{(\s*)\.Release(\s*)\}\}/$(RELEASE)/g" | \
        	gsed -E "s/\{\{(\s*)\.ServiceName(\s*)\}\}/$(APP)/g"; \
        echo ---; \
    done > tmp.yaml
	kubectl apply -f tmp.yaml
```
上面命令“编译”所有 `*.yaml` 配置到一个文件。用实际值替换 `Release` 和 `ServiceName` 变量（注意，这里用了 `gsed` 而非标准 `sed`）, 最后运行 `kubectl apply` 命令安装应用到 Kubernetes 上。

验证配置是否正确：
```
$ kubectl get deployment
NAME      DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
advent    3         3         3            3           1d

$ kubectl get service
NAME         CLUSTER-IP       EXTERNAL-IP   PORT(S)   AGE
advent       10.109.133.147   <none>        80/TCP    1d

$ kubectl get ingress
NAME      HOSTS         ADDRESS        PORTS     AGE
advent    advent.test   192.168.64.2   80        1d
```

先在 `/etc/host` 文件添加模拟域名 `advent.test`，然后可以发请求测试服务了。

```
echo "$(minikube ip) advent.test" | sudo tee -a /etc/hosts
```

```
curl -i http://advent.test/home
HTTP/1.1 200 OK
Server: nginx/1.13.6
Date: Sun, 10 Dec 2017 20:40:37 GMT
Content-Type: application/json
Content-Length: 72
Connection: keep-alive
Vary: Accept-Encoding

{"buildTime":"2017-12-10_11:29:59","commit":"020a181","release":"0.0.5"}%
```
成功~！

所有步骤的代码在 [这里](https://github.com/rumyantseva/advent-2017) ，两个版本：[按 commit 划分](https://github.com/rumyantseva/advent-2017/commits/master) 以及 [按步骤划分](https://github.com/rumyantseva/advent-2017/tree/all-steps)。如有任何疑问，请 [提 issue](https://github.com/rumyantseva/advent-2017/issues/new)，或者 tweet[@webdeva](https://twitter.com/webdeva)，或者在评论区留评论。

真实生产环境上的服务其实有更大的灵活性，想知道是代码“长”啥样的么 ^_^？可以参考 [takama/k8sapp](https://github.com/takama/k8sapp) ，是一个 Go 应用模板，满足了 Kubernetes 需求。

---

via: https://blog.gopheracademy.com/advent-2017/kubernetes-ready-service/

作者：[Elena Grahovac](https://github.com/jeffallen)
译者：[dongfengkuayue](https://github.com/dongfengkuayue)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

首发于：https://studygolang.com/articles/25295

# Go 语言 Protobuf 教程之 Message

在序列化结构数据的各种方式中，protocol buffer（或 protobuf）是资源开销最小的一种。protobuf 需要客户端和服务端都知道数据结构而且兼容，不像 JSON 那样结构本身就是编码的一部分。

## Protobuf 和 Go

最基本的 protobuf message 定义像下面这样：

```protobuf
message ListThreadRequest {
	// session info
	string sessionID = 1;

	// pagination
	uint32 pageNumber = 2;
	uint32 pageSize = 3;
}
```

上面的 message 结构定义了字段名字、类型和它编码后的二进制结构中的顺序。管理 protobuf 和 JSON 编码的数据稍微有些不同，protobuf 有一些依赖。

例如，下面是 `protoc` 生成的上面 message 对应的代码：

```go
type ListThreadRequest struct {
	// session info
	SessionID string `protobuf:"bytes,1,opt,name=sessionID,proto3" json:"sessionID,omitempty"`
	// pagination
	PageNumber           uint32   `protobuf:"varint,2,opt,name=pageNumber,proto3" json:"pageNumber,omitempty"`
	PageSize             uint32   `protobuf:"varint,3,opt,name=pageSize,proto3" json:"pageSize,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}
```

开发过程中，请以向前兼容的方式管理 protobuf message，需要遵循以下几条规则：

- 增加新的字段不是破坏性改动
- 删除字段不是破坏性改动
- 不要复用一个字段的序号（破坏已有的 protobuf 客户端）
- 不要改变已有的字段序号（破坏已有的 protobuf 客户端）
- 对字段重命名在 protobuf 中不是破坏性改动
- 对字段重命名在 JSON 中是破坏性改动
- 修改字段的类型在 protobuf 中是破坏性改动，在大部分 JSON 中也是
- JSON 中修改数字类型如 uint16 改为 uint32 通常是安全的
- 如果你调用 JavaScript 的 API，JSON 和 unit64 不建议使用

因此，如果你开发过程中需要修改 protobuf message 的定义，请保持客户端和服务端的同步。如果你用移动客户端（Android 和 iPhone）调用 protobuf API，那么这（保持同步）将尤其重要，因为做了重大变更后是有破坏性影响的。增加新的字段或删除字段是修改 API 最安全的方式，这样 protobuf 定义能保持兼容。

## 生成 Protobuf Go 代码

在本系列的文章中，我将搭建真实的微服务，追踪和聚合多个服务的可视化数据。它不需要认证，仅需要一个 API endpoint，因此从定义上来说是一个真正的微服务。

首先创建一个 `rpc/stats/stats.proto`，然后创建我们的 `stats` 微服务。

```protobuf
syntax = "proto3";

package stats;

option go_package = "github.com/titpetric/microservice/rpc/stats";

message PushRequest {
	string property = 1;
	uint32 section = 2;
	uint32 id = 3;
}

message PushResponse {}
```

该实例用 `proto3` 声明了 protobuf 的版本。最重要的部分是 `go_package` option：为我们的服务定义了一个重要的路径，如果其他的服务想要导入和使用这里的 message 定义就需要用到这个路径。可复用性是 protobuf 是自带的属性。

由于我们不想要半途而废，因此我们用 CI-first 的方式实现我们的微服务。在一开始使用 Drone CI 是用 CI 的一个伟大选项，因为它的 [drone/drone-cli](https://github.com/drone/drone-cli) 不需要搭建 CI 服务，你仅需要在本地通过运行 `drone exec` 来执行 CI 步骤。

为了搭建我们的微服务框架，我们需要：

1. 安装 Drone CI drone-cli
2. 安装 `protoc` 和 `protoc-gen-go` 的 docker 环境
3. 长远着想，加一个 `Makefile` 文件
4. 加一个 `.drone.yml` 配置文件，写明生成 Go 代码的构建步骤

### 安装 Drone CI

安装 drone-cli 很简单。如果你使用的是 amd64 Linux 机器，执行下面的命令。或者从 [drone/drone-cli](https://github.com/drone/drone-cli) 发布页拉取你需要的版本解包到你机器的 `/usr/local/bin` 或通用的执行路径。

```bash
cd /usr/local/bin
wget https://github.com/drone/drone-cli/releases/download/v1.2.0/drone_linux_amd64.tar.gz
tar -zxvf drone*.tar.gz && rm drone*.tar.gz
```

### 创建一个构建环境

Drone CI 通过运行你提供的 Docker 环境中声明的 `.drone.yml` 文件里的 CI 步骤来工作。在我们的构建环境中，我已经创建了 `docker/build/` ，下面有一个 `Dockerfile` 和 `Makefile` ，以此来协助构建和发布我们的案例中需要的构建镜像。

```dockerfile
FROM golang:1.13

# install protobuf
ENV PB_VER 3.10.1
ENV PB_URL https://github.com/google/protobuf/releases/download/v${PB_VER}/protoc-${PB_VER}-linux-x86_64.zip

RUN apt-get -qq update && apt-get -qqy install curl Git make unzip gettext rsync

RUN mkdir -p /tmp/protoc && \
    curl -L ${PB_URL} > /tmp/protoc/protoc.zip && \
    cd /tmp/protoc && \
    unzip protoc.zip && \
    cp /tmp/protoc/bin/protoc /usr/local/bin && \
    cp -R /tmp/protoc/include/* /usr/local/include && \
    chmod go+rx /usr/local/bin/protoc && \
    cd /tmp && \
    rm -r /tmp/protoc

# Get the source from GitHub
RUN Go get -u google.golang.org/grpc

# Install protoc-gen-go
RUN Go get -u github.com/golang/protobuf/protoc-gen-go
```

在 `Makefile` 中实现 `make && make push` ，这样就可以快速构建和发布我们的镜像到 docker 注册中心。本例中的镜像发布在 `titpetric/microservice-build`，但我建议这里你用自己的镜像来操作。

```makefile
.PHONY: all docker push test

IMAGE := titpetric/microservice-build

all: docker

docker:
	docker build --rm -t $(IMAGE) .

push:
	docker push $(IMAGE)

test:
	docker run -it --rm $(IMAGE) sh
```

### 创建 Makefile helper

运行 `drone exec` 很简单，但是我们的需要会随着时间推移越来越多，Drone CI 步骤也会变得越来越复杂和难以管理。使用 Makefile 可以让我们添加更复杂的在 Drone 运行的目标。目前我们先以一个最简单的 Makefile 起步，这个 Makefile 仅包含对 `drone exec` 的一次调用：

```makefile
.PHONY: all

all:
	drone exec
```

这是一个简单的 Makefile，意味着我们仅通过运行 `make` 就可以构建我们的 Drone CI 工程。以后我们会为了支持新的需要去扩展它，但是目前我们仅需要确保它可用就行了。

### 创建 Drone CI 配置

通过下面的代码，我们可以定义构建我们的 protobuf 结构定义的初始 `.drone.yml` 文件，也可以对基本代码做一些修改：

```yaml
workspace:
  base: /microservice

kind: pipeline
name: build

steps:
- name: test
  image: titpetric/microservice-build
  pull: always
  commands:
    - protoc --proto_path=$GOPATH/src:. -Irpc/stats --go_out=paths=source_relative:. rpc/stats/stats.proto
    - Go mod tidy > /dev/null 2>&1
    - Go mod download > /dev/null 2>&1
    - Go fmt ./... > /dev/null 2>&1
```

这几步操作是为了处理我们的 go.mod/go.sum 文件，在我们的基础代码上运行 `go fmt` 也是。

`commands:` 下面定义的第一步是可以生成我们声明的 message 对应的 Go 定义的 `protoc` 命令。我们的 `stats.proto` 文件所在的文件夹中，会创建一个 `stats.pb.go` 文件，每个结构都声明了 `message {}`。

## 结语

所以，这里我们做了什么来实现上面的结果：

- 我们用我们的 `protoc` 代码生成环境创建的我们的 CI 构建镜像
- 我们使用 Drone CI 作为我们的本地构建服务，未来可以迁移到宿主 CI
- 我们为微服务 message 结构创建了 protobuf 定义
- 我们为编码/解码 protobuf message 生成了合适的 Go 代码

从现在起，我们将尝试去实现一个 RPC 服务。

本文是 [Go 微服务驾到](https://leanpub.com/go-microservices) 书的一部分。在圣诞节之前我们会每天发布一篇文章。请考虑购买电子书支持我们的创作，反馈给我，这会让以后的文章对你更有用处。

本系列的所有文章都在这里 [the advent2019 tag](https://scene-si.org/tags/advent2019/)。

## 如果你读到了这里

推荐购买我的书：

- [Advent of Go Microservices](https://leanpub.com/go-microservices)
- [API Foundations in Go](https://leanpub.com/api-foundations)
- [12 Factor Apps with Docker and Go](https://leanpub.com/12fa-docker-golang)

---

via: https://scene-si.org/2019/12/01/introduction-to-protobuf-messages/

作者：[Tit Petric](http://github.com/titpetric)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

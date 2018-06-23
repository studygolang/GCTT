已发布：https://studygolang.com/articles/12094

# Golang 中的微服务 - 第二部分 - Docker 和 go-micro

## 简介: Docker 和 go-micro

**[在上篇文章中](https://studygolang.com/articles/12060)**，我们大致介绍了如何编写一个基于 `gRPC` 的微服务。在这个部分，我们将涵盖 `Docker` 服务的基础知识，我们也将使用 [go-micro](https://github.com/micro/go-micro) 更新我们的服务，并在文本末尾引入第二个服务。

## Docker简介

随着云计算的到来和微服务的诞生，服务在部署的时候有更多的压力，但是一次一小段代码就产生了一些有趣的新思想和新技术，其中之一就是[容器](https://en.wikipedia.org/wiki/Operating-system-level_virtualization)的概念。

在早些的时候，团队部署一个庞大的服务到静态服务器，运行一套操作系统，需要使用一组预定义的依赖来跟踪。例如，可能是由管理员提供的 `VM` 虚拟机或者 `Puppet` 。伸缩是昂贵的并且不一定有效，最常见的是垂直缩放，例如在静态服务器上投入越来越多资源。

针对虚拟机的配置，伴随着像 [vagrant](https://www.vagrantup.com/) 这样的工具越来越常使用。但是运行一个虚拟机仍然是一个相当大的操作。它相当于在你的主机上运行一个完整的操作系统（包括内核，各种应用等）。在资源方面，这是相当昂贵的。所以当微服务出现时，让每个微服务独立跑在自己的虚拟机中变得不可行了。

## 容器的诞生

[容器](https://en.wikipedia.org/wiki/Operating-system-level_virtualization)是精简版的操作系统。容器不包含内核、用户操作系统或通常构成操作系统的较低级别组件。

容器只包含顶层库及其运行组件，内核在主机上共享。所以主机运行一个 `Unix` 内核，然后由 `n` 个容器共享，运行非常不同的运行时集合。

在底层，容器使用各种内核工具。以便跨容器空间共享资源和网络功能。

[进一步阅读](https://www.redhat.com/en/topics/containers/whats-a-linux-container)

这意味着您可以运行代码所需的运行时和依赖关系，无需启动几个完整的操作系统。这改变了游戏规划，因为一个容器和虚拟机比较体积是比较小的。例如 `Ubuntu`，它通常 `1GB` 小一点，而 `Docker` 镜像只有 `188M`。

你会注意到我在这个介绍中更广泛的谈到容器，而不是 `Docker` 容器。尽管人们通常认为 [Docker](https://www.docker.com/) 和容器是一回事。但是，容器在 Linux 中更多是一个概念或一组功能。
[Docker](https://www.docker.com/) 只是其中的一种，只是因为好用而变得流行，还有其他的。但是我们会专注于 `Docker`，因为在我看来它是目前支持度最高且对新手最友好的容器技术。。

所以希望你看到容器的价值。我们开始使用 `Docker` 来运行我们的第一个服务。我们来创建一个 `Dockerfile`。（译注: Docker 容器的创建一般都使用 Dockerfile，容器会根据这个文件创建相对应的运行环境）

```
touch consignment-service/Dockerfile
```

在该文件中添加以下内容:

```
FROM alpine:latest

RUN mkdir /app
WORKDIR /app
ADD consignment-service /app/consignment-service

CMD ["./consignment-service"]
```

如果你在 Linux 上运行，你可能会遇到使用 Alpine 的问题。所以，如果你在 Linux 机器上关注这篇文章，只需用 `debian` 替换 `alpine` 即可，你应该能够正常运行。 稍后我们将会介绍一个更好的方法来构建我们的二进制文件。

首先，我们下载最新的 [Linux Alpine](https://alpinelinux.org/) 镜像。[Linux Alpine](https://alpinelinux.org/) 是一个轻量级 Linux 发行版，为运行 Dockerised Web 应用程序而开发和优化。换一种说法，[Linux Alpine](https://alpinelinux.org/) 具有足够的依赖和运行时功能来运行大多数应用程序。这意味着它的镜像大小只有 8M 左右。与之相比，大约 1GB 的 Ubuntu 虚拟机，你可以开始看到为什么 Docker 镜像更适合微服务和云计算。

接下来我们创建一个新的目录来存放我们的应用程序，并将上下文目录设置到我们的新目录中。这时我们的应用程序目录是默认的目录。然后，我们将编译后的二进制文件添加到我们的 Docker 容器中，并运行它。

现在我们来更新 `Makefile` 文件来构建我们的 Docker 镜像。

```
build:
    ...
    GOOS=linux GOARCH=amd64 go build
    docker build -t consignment-service .
```

我们在这里增加了两个步骤，我想详细解释一下。首先，我们正在构建我们的二进制文件。你会注意到在运行命令 `$ go build` 之前，设置了两个环境变量。GOOS 和 GOARCH 允许您为另一个操作系统交叉编译您的二进制文件，由于我在 Macbook上开发，所以无法编译出二进制文件，让它在 Docker 容器中运行它，而该容器使用的是 Linux。这个二进制在你的 Docker 容器中将是完全没有意义的，它会抛出一个错误。第二步是添加 Docker 构建过程。这将读取你的 Dockerfile 文件，并通过一个名称 `consignment-service` 构建镜像。句号表示一个目录路径，在这里我们只是希望构建过程在当前目录中查找。

我将在我们的 Makefile 中添加一个新条目：

```
run:
    docker run -p 50051:50051 consignment-service
```

在这里，我们运行 `consignment-service` Docker 镜像，并暴露 50051 端口。由于 Docker 在单独的网络层上运行，因此您需要将 Docker 容器中使用的端口转发给主机。您可以通过更改第一个段将内部端口转发到主机上的新端口。例如，如果要在端口 8080 上运行此服务，则需要将 -p 参数更改为 `8080:50051`。您也可以通过包含 `-d` 标志在后台运行容器。例如，`docker run -d -p 50051:50051 consignment-service`。

[您可以阅读更多关于 Docker 网络如何工作的信息。](https://docs.docker.com/engine/userguide/networking/)

当您运行 `$ docker build` 时，您正在将代码和运行时环境构建到镜像中。Docker 镜像是您的环境及其依赖关系的可移植快照。你可以将它分享到 Docker Hub 来共享你的 Docker 镜像。Docker 镜像就像一个 npm 或 yum repo。当你在你的 Dockerfile 里面定义了 `FROM`，你就告诉了 docker 从 docker hub 下载哪个镜像来作为运行环境。然后，您可以扩展并覆盖该基本文件的某些部分，方法是自行重新定义它们。我们不会公开我们的 Docker 镜像，但是可以随时仔细阅读 Docker hub，并且注意到有多少功能被容器化。一些非常[显著的事情](https://www.youtube.com/watch?v=GsLZz8cZCzc)已经被 Docker 化了。

Dockerfile 中的每个声明在第一次构建时都被缓存。这样可以节省每次更改时重新构建整个运行时的时间。 Docker 非常聪明，可以确定哪些部分发生了变化，哪些部分需要重新构建。这使得构建过程非常快速。

我们已经介绍了很多容器的部分了。让我们回到我们的代码。

在创建 gRPC 服务时，创建连接的代码有很多，并且必须将服务地址的位置硬编码到客户端或其他服务中，以便连接到它。这很棘手，因为当您在云中运行服务时，它们可能不共享相同的主机，或者重新部署服务后地址或 IP 可能会更改。

这是服务发现的起点。服务发现保持所有服务及其位置的最新目录。每个服务在运行时注册自己，并在关闭时自行注销。 每个服务都有一个名字或编号分配给它。 因此，即使可能有新的 IP 地址或主机地址，只要服务名称保持不变，您就不需要从其他服务更新对此服务的调用。

通常情况下，解决这个问题的方法有很多，但是像编程中的大多数情况一样，如果已经有人解决了这个问题，那么重新发明轮子就没有意义了。 [Go-micro](https://github.com/micro/go-micro) 的创始人是 @chuhnk（Asim Aslam），他以一种非常清晰和易用的方式解决了这些问题。

### Go-micro

Go-micro 是一个用 Go 编写的强大的微服务框架，大部分用于 Go。但是，您也可以使用 [Sidecar](https://github.com/micro/micro/tree/master/car) 以便与其他语言进行交互。

Go-micro 有一些有用的功能，可以用来制作微型服务。但是，我们将从可能解决的最常见问题开始，那就是服务发现。

我们需要对我们的服务进行一些更新，让它与 go-micro 一起工作。Go-micro 作为 protoc 插件集成，在这种情况下，替换我们当前使用的标准 gRPC 插件。所以让我们开始在我们的 Makefile 中替换它。

确保安装 go-micro 依赖:

```
go get -u github.com/micro/protobuf/{proto,protoc-gen-go}
```

```
build:
    protoc -I. --go_out=plugins=micro:$(GOPATH)/src/github.com/EwanValentine/shippy/consignment-service \
        proto/consignment/consignment.proto
    ...

...
```

我们已经更新了我们的 Makefile 来使用 go-micro 插件，而不是 gRPC 插件。现在需要更新我们的 `consignment-service/main.go` 文件来使用 go-micro。这将抽象我们以前的 gRPC 代码，它将处理注册和轻松启动我们的服务

```go
// consignment-service/main.go
package main

import (

    // Import the generated protobuf code
    "fmt"

    pb "github.com/EwanValentine/shippy/consignment-service/proto/consignment"
    micro "github.com/micro/go-micro"
    "golang.org/x/net/context"
)

type IRepository interface {
    Create(*pb.Consignment) (*pb.Consignment, error)
    GetAll() []*pb.Consignment
}

// Repository - Dummy repository, this simulates the use of a datastore
// of some kind. We'll replace this with a real implementation later on.
type Repository struct {
    consignments []*pb.Consignment
}

func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
    updated := append(repo.consignments, consignment)
    repo.consignments = updated
    return consignment, nil
}

func (repo *Repository) GetAll() []*pb.Consignment {
    return repo.consignments
}

// Service should implement all of the methods to satisfy the service
// we defined in our protobuf definition. You can check the interface
// in the generated code itself for the exact method signatures etc
// to give you a better idea.
type service struct {
    repo IRepository
}

// CreateConsignment - we created just one method on our service,
// which is a create method, which takes a context and a request as an
// argument, these are handled by the gRPC server.
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {

    // Save our consignment
    consignment, err := s.repo.Create(req)
    if err != nil {
        return err
    }

    // Return matching the `Response` message we created in our
    // protobuf definition.
    res.Created = true
    res.Consignment = consignment
    return nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
    consignments := s.repo.GetAll()
    res.Consignments = consignments
    return nil
}

func main() {

    repo := &Repository{}

    // Create a new service. Optionally include some options here.
    srv := micro.NewService(

        // This name must match the package name given in your protobuf definition
        micro.Name("go.micro.srv.consignment"),
        micro.Version("latest"),
    )

    // Init will parse the command line flags.
    srv.Init()

    // Register handler
    pb.RegisterShippingServiceHandler(srv.Server(), &service{repo})

    // Run the server
    if err := srv.Run(); err != nil {
        fmt.Println(err)
    }
}
```

这里的主要变化是我们实例化我们的 gRPC 服务器的方式，它处理注册我们的服务，已经被整齐地抽象到一个 `mico.NewService()` 方法中。最后，处理连接本身的 `service.Run()` 函数。和以前一样，我们注册了我们的实现，但这次使用了一个稍微不同的方法。第二个最大的变化是服务方法本身，参数和响应类型略有变化，把请求和响应结构作为参数，现在只返回一个错误。在我们的方法中，设置了由 `go-micro` 处理响应。

最后，我们不再对端口进行硬编码。go-micro 应该使用环境变量或命令行参数进行配置。设置地址, 使用 `MICRO_SERVER_ADDRESS=:50051`。我们还需要告诉我们的服务使用 [mdns](https://en.wikipedia.org/wiki/Multicast_DNS)（多播DNS）作为我们本地使用的服务代理。
您通常不会在生产环境中使用 [mdns](https://en.wikipedia.org/wiki/Multicast_DNS) 进行服务发现，但我们希望避免在本地运行诸如 Consul 或 etcd 这样的测试。更多我们将在后面介绍。

让我们更新我们的 Makefile 来实现这一点。

```
run:
    docker run -p 50051:50051 \
        -e MICRO_SERVER_ADDRESS=:50051 \
        -e MICRO_REGISTRY=mdns consignment-service
```

`-e` 是一个环境变量标志，它允许你将环境变量传递到你的 Docker 容器中。
每个变量必须有一个标志，例如 `-e ENV=staging -e DB_HOST=localhost` 等。

现在如果你运行 `make run`，您将拥有一个 Dockerised 服务，并具有服务发现功能。所以让我们更新我们的 cli 工具来利用它。

```go
import (
    ...
    "github.com/micro/go-micro/cmd"
    microclient "github.com/micro/go-micro/client"

)

func main() {
    cmd.Init()

    // Create new greeter client
    client := pb.NewShippingServiceClient("go.micro.srv.consignment", microclient.DefaultClient)
    ...
}
```

[完整文件看这里](https://github.com/EwanValentine/shippy/blob/tutorial-2/consignment-cli/cli.go)

在这里，我们导入了用于创建客户端的 go-micro 库，并用 go-micro 客户端代码取代了现有的连接代码，该客户端代码使用服务解析而不是直接连接到地址。

但是，如果你运行它，这是行不通的。这是因为我们现在正在 Docker 容器中运行我们的服务，它有自己的 [mdns](https://en.wikipedia.org/wiki/Multicast_DNS)，独立于我们使用中的主机 [mdns](https://en.wikipedia.org/wiki/Multicast_DNS) 。解决这个问题的最简单的方法是确保服务和客户端都在 “dockerland” 中运行，以便它们都在相同的主机上运行，并使用相同的网络层。让我们创建一个 Makefile `consignment-cli/Makefile`，并创建一些条目。

```
build:
    GOOS=linux GOARCH=amd64 go build
    docker build -t consignment-cli .

run:
    docker run -e MICRO_REGISTRY=mdns consignment-cli
```

与之前类似，我们要为 Linux 构建我们的二进制文件。 当我们运行我们的 docker 镜像时，我们想传递一个环境变量来指示 go-micro 使用 mdns。

现在让我们为我们的 `CLI` 工具创建一个 Dockerfile ：

```
FROM alpine:latest

RUN mkdir -p /app
WORKDIR /app

ADD consignment.json /app/consignment.json
ADD consignment-cli /app/consignment-cli

CMD ["./consignment-cli"]
```

它除了引入了我们的 json 数据文件，其余与之前服务的 Dockerfile 非常相似。如果你在你的 `consignment-cli` 目录，运行 `$ make run` 命令，你应该和以前一样，看见 `Created: true`。

之前，我提到那些使用 Linux 的人应该切换到使用 Debian 作为基础映像。现在看起来是一个很好的时机来看看 Docker 的一个新功能：多阶段构建。这使我们可以在一个 Dockerfile 中使用多个 Docker 镜像。

这在我们的例子中尤其有用，因为我们可以使用一个镜像来构建我们的二进制文件，具有所有正确的依赖关系等，然后使用第二个镜像来运行它。让我们试试看，我会在代码中留下详细的注释：

```
# consignment-service/Dockerfile

# We use the official golang image, which contains all the
# correct build tools and libraries. Notice `as builder`,
# this gives this container a name that we can reference later on.
FROM golang:1.9.0 as builder

# Set our workdir to our current service in the gopath
WORKDIR /go/src/github.com/EwanValentine/shippy/consignment-service

# Copy the current code into our workdir
COPY . .

# Here we're pulling in godep, which is a dependency manager tool,
# we're going to use dep instead of go get, to get around a few
# quirks in how go get works with sub-packages.
RUN go get -u github.com/golang/dep/cmd/dep

# Create a dep project, and run `ensure`, which will pull in all
# of the dependencies within this directory.
RUN dep init && dep ensure

# Build the binary, with a few flags which will allow
# us to run this binary in Alpine.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .

# Here we're using a second FROM statement, which is strange,
# but this tells Docker to start a new build process with this
# image.
FROM alpine:latest

# Security related package, good to have.
RUN apk --no-cache add ca-certificates

# Same as before, create a directory for our app.
RUN mkdir /app
WORKDIR /app

# Here, instead of copying the binary from our host machine,
# we pull the binary from the container named `builder`, within
# this build context. This reaches into our previous image, finds
# the binary we built, and pulls it into this container. Amazing!
COPY --from=builder /go/src/github.com/EwanValentine/shippy/consignment-service/consignment-service .

# Run the binary as per usual! This time with a binary build in a
# separate container, with all of the correct dependencies and
# run time libraries.
CMD ["./consignment-service"]
```

这种方法的唯一问题，我想回来并在某些时候改善这一点，是 Docker 不能从父目录中读取文件。它只能读取 Dockerfile 所在目录或子目录中的文件。

这意味着为了运行 `$ dep ensure` 或 `$ go get`，你需要确保你的代码被推到 Git 上，这样它就可以提取 vessel-service。就像其他 Go 包一样。这种方法不理想，但足够满足我们现在的需求。

现在我将会在其他 Docker 文件中应用这种新方法。

噢，记住要记得从 Makefiles 中删除 `$ go build`。

[更多的多阶段编译在这里](https://docs.docker.com/engine/userguide/eng-image/multistage-build/#name-your-build-stages)

### Vessel 服务

让我们创建第二个服务。我们有一个代销服务，这将处理将一批货物与一批最适合该批货物的 vessel 进行匹配。为了配合我们的货物，我们需要将集装箱的重量和数量发送到我们的新 vessel 服务，然后将找到一个能够处理该货物的船只。

在你的根目录创建一个新的目录 `$ mkdir vessel-service`。现在为我们的新服务 [protobuf](https://github.com/google/protobuf) 定义创建了一个子目录， `$ mkdir -p vessel-service/proto/vessel`。现在让我们来创建新的 `protobuf` 文件， `$ touch vessel-service/proto/vessel/vessel.proto`。

由于 protobuf 的定义确实是我们领域设计的核心，所以我们从这里开始。

```protobuf
// vessel-service/proto/vessel/vessel.proto
syntax = "proto3";

package go.micro.srv.vessel;

service VesselService {
  rpc FindAvailable(Specification) returns (Response) {}
}

message Vessel {
  string id = 1;
  int32 capacity = 2;
  int32 max_weight = 3;
  string name = 4;
  bool available = 5;
  string owner_id = 6;
}

message Specification {
  int32 capacity = 1;
  int32 max_weight = 2;
}

message Response {
  Vessel vessel = 1;
  repeated Vessel vessels = 2;
}
```

正如你所看到的，这和我们的第一个服务非常相似。我们用一个 `rpc` 方法 `FindAvailable` 创建了一个服务。。这需要一个 Specification 类型并返回一个 Response 类型。Response 类型使用重复字段返回 Vessel 类型或多个 Vesse。

现在我们需要创建一个 Makefile 来处理我们的构建逻辑和运行脚本。 `$ touch vessel-service/Makefile`。打开该文件并添加以下内容：

```makefile
// vessel-service/Makefile
build:
    protoc -I. --go_out=plugins=micro:$(GOPATH)/src/github.com/EwanValentine/shippy/vessel-service \
    proto/vessel/vessel.proto
    docker build -t vessel-service .

run:
    docker run -p 50052:50051 -e MICRO_SERVER_ADDRESS=:50051 -e MICRO_REGISTRY=mdns vessel-service
```

这与我们为托管服务创建的第一个 Makefile 几乎相同，但注意服务名称和端口已经改变了一点。我们不能在同一个端口上运行两个 Docker 容器，所以我们在这里利用 Docker 端口转发来确保服务上的 50051 端口映射到主机网络上的 50052 端口。

现在我们需要一个 Dockerfile，使用我们新的多阶段格式：

```
# vessel-service/Dockerfile
FROM golang:1.9.0 as builder

WORKDIR /go/src/github.com/EwanValentine/shippy/vessel-service

COPY . .

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep init && dep ensure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo .


FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir /app
WORKDIR /app
COPY --from=builder /go/src/github.com/EwanValentine/shippy/vessel-service/vessel-service .

CMD ["./vessel-service"]
```

最后，我们可以开始我们的实现。

```go
// vessel-service/main.go
package main

import (
    "context"
    "errors"
    "fmt"

    pb "github.com/EwanValentine/shippy/vessel-service/proto/vessel"
    "github.com/micro/go-micro"
)

type Repository interface {
    FindAvailable(*pb.Specification) (*pb.Vessel, error)
}

type VesselRepository struct {
    vessels []*pb.Vessel
}

// FindAvailable - checks a specification against a map of vessels,
// if capacity and max weight are below a vessels capacity and max weight,
// then return that vessel.
func (repo *VesselRepository) FindAvailable(spec *pb.Specification) (*pb.Vessel, error) {
    for _, vessel := range repo.vessels {
        if spec.Capacity <= vessel.Capacity && spec.MaxWeight <= vessel.MaxWeight {
            return vessel, nil
        }
    }
    return nil, errors.New("No vessel found by that spec")
}

// Our grpc service handler
type service struct {
    repo Repository
}

func (s *service) FindAvailable(ctx context.Context, req *pb.Specification, res *pb.Response) error {

    // Find the next available vessel
    vessel, err := s.repo.FindAvailable(req)
    if err != nil {
        return err
    }

    // Set the vessel as part of the response message type
    res.Vessel = vessel
    return nil
}

func main() {
    vessels := []*pb.Vessel{
        &pb.Vessel{Id: "vessel001", Name: "Kane's Salty Secret", MaxWeight: 200000, Capacity: 500},
    }
    repo := &VesselRepository{vessels}

    srv := micro.NewService(
        micro.Name("go.micro.srv.vessel"),
        micro.Version("latest"),
    )

    srv.Init()

    // Register our implementation with
    pb.RegisterVesselServiceHandler(srv.Server(), &service{repo})

    if err := srv.Run(); err != nil {
        fmt.Println(err)
    }
}
```

我留下了一些注释，但是非常简单。另外，我想提下，一个 Reddit 用户 /r/jerky_lodash46 指出，我曾经使用 IRepository 作为我的接口名称。我想在这里纠正一下，在我的接口名前面加上 Java 和 C＃ 等语言的约定，但 Go 并没有真正鼓励这一点，因为 Go 把接口当作一等公民。所以我把 IRepository 更名为 Repository ，并且把我的具体结构重命名为 ConsignmentRepository。

这个系列中，我会留下任何错误，并在以后的文章中予以纠正，以便我能够解释这些改进。我们可以更多地学习。

现在让我们来看看有趣的部分。当我们创建一个托运货物时，我们需要改变我们的托运服务来呼叫我们的新 vessel 服务，找到一艘船，并更新创建的托运中的 vessel_id：

```go
// consignment-service/main.go
package main

import (

    // Import the generated protobuf code
    "fmt"
    "log"

    pb "github.com/EwanValentine/shippy/consignment-service/proto/consignment"
    vesselProto "github.com/EwanValentine/shippy/vessel-service/proto/vessel"
    micro "github.com/micro/go-micro"
    "golang.org/x/net/context"
)

type Repository interface {
    Create(*pb.Consignment) (*pb.Consignment, error)
    GetAll() []*pb.Consignment
}

// Repository - Dummy repository, this simulates the use of a datastore
// of some kind. We'll replace this with a real implementation later on.
type ConsignmentRepository struct {
    consignments []*pb.Consignment
}

func (repo *ConsignmentRepository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
    updated := append(repo.consignments, consignment)
    repo.consignments = updated
    return consignment, nil
}

func (repo *ConsignmentRepository) GetAll() []*pb.Consignment {
    return repo.consignments
}

// Service should implement all of the methods to satisfy the service
// we defined in our protobuf definition. You can check the interface
// in the generated code itself for the exact method signatures etc
// to give you a better idea.
type service struct {
    repo Repository
    vesselClient vesselProto.VesselServiceClient
}

// CreateConsignment - we created just one method on our service,
// which is a create method, which takes a context and a request as an
// argument, these are handled by the gRPC server.
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {

    // Here we call a client instance of our vessel service with our consignment weight,
    // and the amount of containers as the capacity value
    vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
        MaxWeight: req.Weight,
        Capacity: int32(len(req.Containers)),
    })
    log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)
    if err != nil {
        return err
    }

    // We set the VesselId as the vessel we got back from our
    // vessel service
    req.VesselId = vesselResponse.Vessel.Id

    // Save our consignment
    consignment, err := s.repo.Create(req)
    if err != nil {
        return err
    }

    // Return matching the `Response` message we created in our
    // protobuf definition.
    res.Created = true
    res.Consignment = consignment
    return nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
    consignments := s.repo.GetAll()
    res.Consignments = consignments
    return nil
}

func main() {

    repo := &ConsignmentRepository{}

    // Create a new service. Optionally include some options here.
    srv := micro.NewService(

        // This name must match the package name given in your protobuf definition
        micro.Name("go.micro.srv.consignment"),
        micro.Version("latest"),
    )

    vesselClient := vesselProto.NewVesselServiceClient("go.micro.srv.vessel", srv.Client())

    // Init will parse the command line flags.
    srv.Init()

    // Register handler
    pb.RegisterShippingServiceHandler(srv.Server(), &service{repo, vesselClient})

    // Run the server
    if err := srv.Run(); err != nil {
        fmt.Println(err)
    }
}
```

在这里，我们为我们的 vessel 服务创建了一个客户端实例，这允许我们使用服务名称，即 go.micro.srv.vessel 将船舶服务作为客户端调用，并与其方法交互。在这种情况下，只有一个方法 (`FindAvailable`)。我们把我们的寄售重量，以及我们想要作为规格的集装箱的数量发送到 vessel 服务。然后返回一个适当的 vessel。

更新 `consignment-cli/consignment.json` 文件，删除硬编码的 vessel_id ，我们要确认我们自己正在工作。让我们再添加一些容器，增加权重。例如

```json
{
  "description": "This is a test consignment",
  "weight": 55000,
  "containers": [
    { "customer_id": "cust001", "user_id": "user001", "origin": "Manchester, United Kingdom" },
    { "customer_id": "cust002", "user_id": "user001", "origin": "Derby, United Kingdom" },
    { "customer_id": "cust005", "user_id": "user001", "origin": "Sheffield, United Kingdom" }
  ]
}
```

现在在 `consignment-cli` 中运行 `$ make build && make`。您应该看到一个响应​​，并创建货物清单。在您的货物中，您现在应该看到 vessel_id 已经设置好了。所以现在有了我们两个互联的微服务和一个命令行界面！在这个系列的下一部分，我们将看看使用 [MongoDB](https://www.mongodb.com/what-is-mongodb) 来保存这些数据。我们还将添加第三个服务，并使用 `docker-compose` 来管理我们在本地不断增长的容器生态系统。

查看完整示例的[代码](https://github.com/EwanValentine/shippy/tree/tutorial-2)。如有任何反馈，请将其发送至（mailto：ewan.valentine89@gmail.com）。非常感激！

如果你发现这个系列很有用，而且你使用了一个广告拦截器（谁可以责怪你）。请考虑一下我的时间和精力。干杯! [https://monzo.me/ewanvalentine](https://monzo.me/ewanvalentine)

Docker 简报（2017 年 11 月 22 日）

---

via: https://ewanvalentine.io/microservices-in-golang-part-2/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[guoxiaopang](https://github.com/guoxiaopang)
校对：[QueShengyao](https://github.com/QueShengyao) [polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

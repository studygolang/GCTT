已发布：https://studygolang.com/articles/12452

# golang 中的微服务 - 第 3 部分 - docker compose 和 datastores

在[之前的文章](https://studygolang.com/articles/12094)中，我们介绍了 [go-micro](https://github.com/micro/go-micro) 和 [Docker](https://www.docker.com/) 的一些基础知识。在推出了这两项服务之后我们将在本文介绍 [docker-compose](https://docs.docker.com/compose/)、教大家如何更便捷地在本地运行多个服务，还会列述一些在本系列微服务教程中可以使用的数据库类型，最后引出本系列的第三项服务 —— User service。

(译者注：阅读本文之前建议先下载作者[源码](https://github.com/EwanValentine/shippy.git)配合理解)

## 准备工作

安装 docker-compose: https://docs.docker.com/compose/install/

docker-compose 安装完毕之后，我们来介绍一些可用的数据库以及他们之间的区别。

## 数据库选择

在前两篇文章，我们的数据并不会持久化的存储到某地，它只会存储在我们服务的内存中，当容器重新启动时，这些数据会丢失。所以需要选择一种数据库来持久化的存储和查询我们的数据。

微服务的优点是，你可以为每个服务选择一个不同的数据库。当然许多情况下我们不必这样做。例如生产环境中的小团队完全不必选择多个数据库，这样会增加维护成本。但在某些情况下，一个服务的数据可能并不兼容其他服务的数据库，这时不得不增加一个数据库。微服务使得数据兼容这件事变得更加简单，你完全不必操心不同服务的数据使用不同的数据库带来的额外维护成本。

本文不会解释如何为你的服务选择“正确的”数据库，这是一个值得深入探究的话题，详情可以借鉴[如何为你的服务选择“正确的”数据库](https://www.infoworld.com/article/3236291/database/how-to-choose-a-database-for-your-microservices.html)。本文示例的数据持久化选择 NoSQL 文档存储解决方案，NoSQL 更适用于处理大量松散且不一致的数据集，例如 json 存储数据更加灵活，我会选择效果良好并且社区服务更加完善的 MongoDB 作为我们的 NoSQL。

如果需要处理的数据是被严格定义并且联系紧密，那么可以使用传统的关系型数据库（RDBMS），但实际上并非一定要这么做。在选择之前一定考虑我们服务的数据结构，它是做的读操作更多还是写操作更多？查询的复杂程度如何？等等。这些才是我们选择使用何种数据库的出发点。由于个人原因，关系数据库我更喜欢使用 Postgres，当然，你也可以使用 MySQL 或者 MariaDB 等等。

如果你不想亲自管理自己的数据库（通常是可取的），你可以选择使用亚马逊或者谷歌提供的完全成熟的 NoSQL 和 RDBMS 解决方案。除此之外 **compose** 则是另外一个非常棒的选择，你可以将自己的服务完全托管在 compose 上，它可以提供类似于亚马逊和谷歌的云服务，并且能够更加便捷的扩展各种数据库实例，同时还具备更低的延迟。

* 亚马逊：

	RDBMS：https://aws.amazon.com/rds/
	NoSQL：https：//aws.amazon.com/dynamodb/

* 谷歌：

	RDBMS：https：//cloud.google.com/spanner/
	NoSQL：https：//cloud.google.com/datastore/
 
数据库相关知识讨论完毕之后我们就可以写一些代码了！

## docker-compose

上一篇文章我们介绍了 [Docker](https://docker.com/) ，它可以用轻量级的容器运行我们的服务，并拥有自己独立的运行时间和依赖关系。但是服务数量较多的情景下使用单独的 Makefile 运行和管理每个服务太麻烦了。[docker-compose](https://docs.docker.com/compose/) 应运而生，它很好的帮我们解决了这一问题。 Docker-compose 允许我们在 yaml 文件中定义 Docker 容器列表，并指定关于其运行时间的元数据。我们可以从 Docker-compose 中看到一些熟悉的 docker 命令的影子。例如：

```
$ docker run -p 50052：50051 -e MICRO_SERVER_ADDRESS =：50051 -e MICRO_REGISTRY = mdns vessel-service
```

在 docker-compose 里可以写为：

```
version: '3.1'

services:
  vessel-service:
    build: ./vessel-service
    ports:
      - 50052:50051
    environment:
      MICRO_REGISTRY: "mdns"
      MICRO_SERVER_ADDRESS: ":50051"
```

太简单了，不是吗？

接下来我们可以在根目录下创建一个 docker-compose 文件

```
$ touch docker-compose.yml
```

然后在这个 yml 文件里添加我们的服务：

```
# docker-compose.yml
version: '3.1'

services:

  consignment-cli:
    build: ./consignment-cli
    environment:
      MICRO_REGISTRY: "mdns"

  consignment-service:
    build: ./consignment-service
    ports:
      - 50051:50051
    environment:
      MICRO_ADDRESS: ":50051"
      MICRO_REGISTRY: "mdns"
      DB_HOST: "datastore:27017"

  vessel-service:
    build: ./vessel-service
    ports:
      - 50052:50051
    environment:
      MICRO_ADDRESS: ":50051"
      MICRO_REGISTRY: "mdns"

```

这个 yml 文件首先定义了要使用的 docker-compose 的版本号`3.1`，然后定义一个 service 列表。还有其他的根级定义，比如网络和卷。

先关注 service ，每个 service 都由其名称定义，然后我们加入了一个 build 路径，这里要存放我们的 Dockerfile 文件，docker-compose 会从这个路径下寻找 Dockerfile 来构建镜像，后文我们会演示如何用 `image` 字段来引用一个构建好的镜像。然后我们还可以定义映射端口和环境变量。

这些是 docker-compose 的基本命令：

* 构建你的docker-compose集： `$ docker-compose build && docker-compose run`
* 在后台运行你的容器集： `$ docker-compose up -d`
* 查看当前正在运行的容器的列表： `$ docker ps`
* 停止正在运行的所有容器： `$ docker stop $（docker ps -qa）`

接下来就可以运行我们的容器集了。你能够看到很多 dockerfile 正在 build 的输出信息。也可能会从我们的 CLI 中看到少量 error ，不用太在意，这一般都是一些其他服务在处理彼此之间的依赖关系。

命令 `$ docker-compose run consignment-cli` 可以帮助我们使用 CLI 工具来测试所有服务是否正常工作，一旦所有容器都显示 running 就意味着我们的 docker-compose 成功了。和我们以前不借助 docker-compose 直接使用 dockerfile 和 docker 命令取得的效果一样。

## 容器实体和 protobufs

本系列前面的文章已经讲过使用 [protobufs](https://www.ibm.com/developerworks/cn/linux/l-cn-gpb/) 作为数据模型的模板。我们用它来定义我们的服务结构和功能函数。因为 protobuf 生成的结构基本上都是正确的数据类型，我们也可以将这些结构重复利用为底层数据库模型。这实际上是相当令人兴奋的。protobufs保证了数据的来源单一性和一致性。

当然这种方法确有其不足之处。有时候，将 protobuf 生成的代码封装到一个有效的数据库实体是非常棘手的。有时数据库很难利用 protobuf 生成的自定义本地数据类型。例如，如何将一个 Mongodb 实体里的 `bson.ObjectId` 类型的 ID 转换成 `string` 类型的 ID 就困扰了我很久。后来通过实验论证，无论如何 bson.ObjectId 实际上只是一个字符串，所以你可以把它们封装在一起。此外，mongodb 的 id 索引在内部被存储为 `_id`，该 `_Id` 字段是无法被执行的，所以必须要将它与 Id 字符串字段绑定在一起。这意味着你要为你的 protobuf 文件定义自定义标签。稍后我们讨论如何做到这一点。

与此同时，使用 protobuf 会导致服务间通信与数据库代码的耦合度更高，这也成为了许多人反对使用 protobuf 定义的数据作为数据库实体的一大原因。

一般建议在 protobuf 定义代码和你的数据库实体之间进行相互转换。但是，这两种类型的代码之间相互转换需要大量的编码工作：

```go
func (service *Service) (ctx context.Context, req *proto.User, res *proto.Response) error {
	entity := &models.User{
		Name: req.Name.
		Email: req.Email,
		Password: req.Password,
	}
	err := service.repo.Create(entity)
	...
}
```

表面上看起来不是那么糟糕，但是当你有大量嵌套的结构体和类型的时候。它就会变得非常繁琐，并且可能涉及很多在嵌套的结构之间进行转换的迭代等。

这并不是一种优雅的处理方式，就像编程中的许多选择一样，可以实现需求也不会报错，但感觉并不完美。我个人的观点是选择使用 protobuf 作为我们的数据存储方式，它可以更加优雅的定义我们数据的基本类型。不用 protobufs 简直就是对 protobufs 作为模板定义代码格式所获得的好处的一种浪费。当然，这并不意味着我就是完全正确的，对此[我很想听听你的不同意见，你可以点击此链接与我联系](ewan.valentine89@gmail.com)。

接下来可以开始连接我们的第一个服务，consignment（委托）服务。

先整理一下代码文件： main.go 文件已经包含了我们所有的逻辑代码。为了使我们的微服务示例代码更加清晰，我在 consignment 服务的根目录下创建了另外几个文件：`handler.go`、`datastore.go`和`respository.go`，而不是将这些代码文件创建为一个新的目录或包。这种方式对于一个小型的微服务来说是完全可行的。插一句题外话，对于开发人员来说，可能会非常喜欢使用这样的目录结构来存放自己的代码文件：

```
main.go
models/
  user.go
handlers/
  auth.go
  user.go
services/
  auth.go
```

这是MVC的常见目录结构，但 Golang 并不建议这样做。 就目录结构而言，无论是简单的小项目还是需要处理多个复杂关系的大项目，Golang 的建议是这样的：

```
main.go
users/
  services/
    auth.go
  handlers/
    auth.go
    user.go
  users/
    user.go
containers/
  services/
    manage.go
  models/
    container.go

```
这里是按函数的定义域分组代码，而不是随意地将代码按其功能分组。

但是，由于我们正在处理的是一个只需要关注单一简单问题的微服务，所以我们不需要将目录结构考虑得太复杂。实际上，Go的精神就是鼓励简单。 所以我们将从简单的一步开始，把所有简易命名的代码文件都放在我们的服务的根目录中。

一方面，我们要修改 Dockerfile 文件的内容，因为我们没有将新代码分离出来作为包导入，我们需要在 Dockerfile 文件里告诉go编译器来引入这些新文件并更新构建函数：

```
RUN CGO_ENABLED=0 GOOS=linux go build  -o consignment-service -a -installsuffix cgo main.go repository.go handler.go datastore.go

```
这条命令会将我们刚刚创建的新文件导入。

[Golang 编写的 MongoDB driver 就是这种简单性的一个很好的例子](https://github.com/go-mgo/mgo)。最后，这里有[一篇关于组织Go代码库的文章](https://studygolang.com/articles/11823)推荐给大家学习。

我们可以先删除 main.go 中已经导入的所有仓库代码然后使用 golang 的 mongodb 库 mgo 来重新实现它。我在代码里进行了详细的标注以解释每部分的作用，因此请仔细阅读代码和注释。 尤其是mgo如何处理 sessions 的部分：

```go
// consignment-service/repository.go
package main

import (
	pb "github.com/EwanValentine/shippy/consignment-service/proto/consignment"
	"gopkg.in/mgo.v2"
)

const (
	dbName = "shippy"
	consignmentCollection = "consignments"
)

type Repository interface {
	Create(*pb.Consignment) error
	GetAll() ([]*pb.Consignment, error)
	Close()
}

type ConsignmentRepository struct {
	session *mgo.Session
}

// 创建一个新的 consignment（委托）
func (repo *ConsignmentRepository) Create(consignment *pb.Consignment) error {
	return repo.collection().Insert(consignment)
}

// 获取所有的consignments
func (repo *ConsignmentRepository) GetAll() ([]*pb.Consignment, error) {

	var consignments []*pb.Consignment

    //Find()通常需要一个参数，但如果我们想要返回所有的结果就将参数设为 nil
    //然后将所有的 consignments 作为参数传递给.All（）函数，
    //.All（）函数将所有的 consignments 作为查询的结果返回
    //在这里还可以调用 One（）方法来返回一个单一的consignment

	err := repo.collection().Find(nil).All(&consignments)
	return consignments, err
}

// Close() 负责在每个查询运行结束后关闭数据库session。
// Mgo 在启动时创建一个主 session ，主 session 会为每个请求创建一个新的 session 。 这意味着每个请求都有自己的数据库 session，这样的机制会使得会话更安全、高效。更底层来讲，每个 session 中都有自己独立的数据库 socket 和错误处理机制。
//使用一个主数据库 socket 意味着其余请求必须等待主 session 优先使用 cpu 资源。
// I.e方法使得我们拒绝锁机制而允许多个请求同时处理。这一点很棒！但是...这意味着我们必须确保每个 session 在完成时关闭掉，于此同时你可能会建立大量的连接，以至于达到连接限制。这一点尤其需要注意！！

func (repo *ConsignmentRepository) Close() {
	repo.session.Close()
}

func (repo *ConsignmentRepository) collection() *mgo.Collection {
	return repo.session.DB(dbName).C(consignmentCollection)
}

```

接下来需要编写与 Mongodb 数据库交互的代码，创建主会话/连接。 按照如下所示修改`consignment-service/datastore.go`：

```go
// consignment-service/datastore.go
package main

import (
	"gopkg.in/mgo.v2"
)

// CreateSession() 创建了连接到 mongodb 的主 session
func CreateSession(host string) (*mgo.Session, error) {
	session, err := mgo.Dial(host)
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)

	return session, nil
}
```

就是这样，非常简单。下一步修改`main.go`文件用以连接 repository，它将一个主机字符串作为参数，返回了一个连接到数据库的 session 和一个可能出现的 error，以便程序启动时可以处理这个 error。

```go
// consignment-service/main.go
package main

import (

	// 导入 protobuf
	"fmt"
	"log"

	pb "github.com/EwanValentine/shippy/consignment-service/proto/consignment"
	vesselProto "github.com/EwanValentine/shippy/vessel-service/proto/vessel"
	"github.com/micro/go-micro"
	"os"
)

const (
	defaultHost = "localhost:27017"
)

func main() {

	// 从环境变量中导入 Database host
	host := os.Getenv("DB_HOST")

	if host == "" {
		host = defaultHost
	}

	session, err := CreateSession(host)


	// Mgo 创建的主 session必须于 main() 函数结束之前关闭
	defer session.Close()

	if err != nil {
		//为 CreateSession() 所抛出的 error 添加注释
		log.Panicf("Could not connect to datastore with host %s - %v", host, err)
	}


	//创建一个具有多个选项的新 service
	srv := micro.NewService(

		//这个 name 必须和你 protobuf 定义的包名匹配
		micro.Name("go.micro.srv.consignment"),
		micro.Version("latest"),
	)

	vesselClient := vesselProto.NewVesselServiceClient("go.micro.srv.vessel", srv.Client())

	// Init()用于解析命令行参数
	//
	srv.Init()

	// 注册调度器Register handler
	pb.RegisterShippingServiceHandler(srv.Server(), &service{session, vesselClient})

	// 启动server
	if err := srv.Run(); err != nil {
		fmt.Println(err)
	}
}
```

## Copy vs Clone

你可能已经注意到使用 mgo Mongodb 库时。我们会创建一个被传递给 handlers 的数据库 session，每出现一个请求，就会调用一个该session的克隆方法。

实际上，除了建立与数据库的第一次连接之外，我们从不调用“主会话”，每次我们要访问数据存储时都调用 session.Clone() 方法。如代码注释所言，若使用主会话，则必须再次使用相同的套接字。这意味着您的查询可能会被其他查询阻塞，必须等待此套接字的占用被锁释放。在支持并发的语言中这是绝对无法容忍的。

所以为了避免阻塞请求，mgo 允许你 Copy() 或者 Clone() 一个会话，这样你就可以为每个请求建立一个并发连接。你会注意到我提到了 Copy 和 Clone 方法，这些方法非常相似，但有一个微妙但重要的区别：Clone重新使用和主会话相同的套接字，减少了产生一个新的套接字的开销，十分适用于对写入性能要求更高的代码。但是，在耗时较长的读操作（例如十分复杂的查询或大数据作业等）中其他 Go 协程使用这个相同套接字时可能会引发阻塞。

像我们公司这样写入操作更多的业务，使用 Clone() 更合适一些。

最后一步是将我们的 gRPC 处理程序代码移出到我们新的 handler.go 文件中，代码如下：

```go
// consignment-service.go

package main

import (
	"log"
	"golang.org/x/net/context"
	pb "github.com/EwanValentine/shippy/consignment-service/proto/consignment"
	vesselProto "github.com/EwanValentine/shippy/vessel-service/proto/vessel"
)


// Service 应该实现我们在 protobuf 定义中定义的所有方法，检查生成的代码本身中是否有确切的签名方法等可以帮助你确认该 Service 是否实现了 protobuf 的所有定义。
type service struct {
	vesselClient vesselProto.VesselServiceClient
}

func (s *service) GetRepo() Repository {
	return &ConsignmentRepository{s.session.Clone()}
}

// CreateConsignment 是 service 的一个方法，该方法将 gRPC server 控制的 context 和 request 作为参数
func (s *service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {

    repo := s.GetRepo()
    defer repo.Close()

	// 使用 consignment weight 和 容器数量 作为 capacity value 生成一个客户端实例
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity: int32(len(req.Containers)),
	})

	log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)
	if err != nil {
		return err
	}


	//将从 vessel service 获得的 Id 设为 VesselId
	req.VesselId = vesselResponse.Vessel.Id

	// 保存 consignment
	err = repo.Create(req)
	if err != nil {
		return err
	}

	//返回匹配到我们在 protobuf 里定义的 `Response` message
	res.Created = true
	res.Consignment = req
	return nil
}

func (s *service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	repo := s.GetRepo()
	defer repo.Close()
	consignments, err := repo.GetAll()
	if err != nil {
		return err
	}
	res.Consignments = consignments
	return nil
}

```

我将上一篇教程中 repo 的返回值做了轻微的修改，修改如下：
旧：

```go
type Repository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}
```

新：

```go
type Repository interface {
	Create(*pb.Consignment) error
	GetAll() ([]*pb.Consignment, error)
	Close()
}

```

改成这样的原因是我认为我们在创建 Consignment 之后不需要再返回一个相同的 Consignment。为了防止可能出现的错误，get 查询还会返回一个 error，最后还要添加一个Close()方法。

请对 vessel-service 做相同的修改。这篇文章不再赘述，你应该可以参考我的[代码库](https://github.com/EwanValentine/shippy/tree/tutorial-3)自行完成。

我们也可以修改 protobuf 来在 vessel-service 里添加一个新的方法，这个方法要负责创建新 vessels ：

```go
syntax = "proto3";

package vessel;

service VesselService {
  rpc FindAvailable(Specification) returns (Response) {}
  rpc Create(Vessel) returns (Response) {}
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
	bool created = 3;
}

```

我们在gRPC服务下创建了一个新的Create方法，该方法需要一个 vessel 并返回 generic response。 我已经在 response message  中添加了一个 bool 类型的新字段：`created`。 这是你需要运行`$ make build`来更新这个服务。 现在我们将在`vessel-service / handler.go`中添加一个新的处理程序，并添加一个新的 repository 方法：

```go
// vessel-service/handler.go

func (s *service) GetRepo() Repository {
	return &VesselRepository{s.session.Clone()}
}

func (s *service) Create(ctx context.Context, req *pb.Vessel, res *pb.Response) error {
	repo := s.GetRepo()
	defer repo.Close()
	if err := repo.Create(req); err != nil {
		return err
	}
	res.Vessel = req
	res.Created = true
	return nil
}

```

```go
// vessel-service/repository.go
func (repo *VesselRepository) Create(vessel *pb.Vessel) error {
	return repo.collection().Insert(vessel)
}

```

终于可以创造 vessels 了！ 为了使用新的 Create 方法来存储虚拟数据,我已经更新了 main.go ，[请看这里](https://github.com/EwanValentine/shippy/blob/master/vessel-service/main.go)。

做完上述内容之后。 我们已经使用Mongodb更新了我们的服务。在尝试运行之前，需要更新我们的 docker-compose 文件来启动一个 Mongodb 容器：

```
services:
  ...
  datastore:
    image: mongo
    ports:
      - 27017:27017

```

更新两个服务中的环境变量：`DB_HOST：“datastore：27017”`。

请注意，由于 docker-compose 为我们做了一些内部DNS处理， hostname 被命名为 datastore 而不是示例中的 localhost。

所以最终的 docker-compose 文件应该更新为：

```
version: '3.1'

services:

  consignment-cli:
    build: ./consignment-cli
    environment:
      MICRO_REGISTRY: "mdns"

  consignment-service:
    build: ./consignment-service
    ports:
      - 50051:50051
    environment:
      MICRO_ADDRESS: ":50051"
      MICRO_REGISTRY: "mdns"
      DB_HOST: "datastore:27017"

  vessel-service:
    build: ./vessel-service
    ports:
      - 50052:50051
    environment:
      MICRO_ADDRESS: ":50051"
      MICRO_REGISTRY: "mdns"
      DB_HOST: "datastore:27017"

  datastore:
    image: mongo
    ports:
      - 27017:27017

```
重新 build 一下，运行`$ docker-compose build`然后运行`$ docker-compose up`。 请注意，有时候由于 Dockers 的缓存机制，在 build 时需要加一个`--no-cache` 参数来取消缓存。

## User service

User service 是我们创建的第三个服务。首先修改`docker-compose.yml`，微服务的概念之一就是将所有的东西全部集中起来服务化，所以我们将 Postgres 添加到 docker 容器集里面用来为我们的 User service 服务。

```
  ...
  user-service:
    build: ./user-service
    ports:
      - 50053:50051
    environment:
      MICRO_ADDRESS: ":50051"
      MICRO_REGISTRY: "mdns"

  ...
  database:
    image: postgres
    ports:
      - 5432:5432

```
现在在根目录下创建一个 user-service 目录。根据创建前几个 service 的经验，我们还需要创建以下文件：`handler.go`，`main.go`，`repository.go`，`database.go`，`Dockerfile`，`Makefile`，以及一个存放 proto 文件的子目录，以及 proto 文件本身：
`proto/user/user.proto`

将以下内容添加到 user.proto：

```
syntax = "proto3";

package go.micro.srv.user;

service UserService {
    rpc Create(User) returns (Response) {}
    rpc Get(User) returns (Response) {}
    rpc GetAll(Request) returns (Response) {}
    rpc Auth(User) returns (Token) {}
    rpc ValidateToken(Token) returns (Token) {}
}

message User {
    string id = 1;
    string name = 2;
    string company = 3;
    string email = 4;
    string password = 5;
}

message Request {}

message Response {
    User user = 1;
    repeated User users = 2;
    repeated Error errors = 3;
}

message Token {
    string token = 1;
    bool valid = 2;
    repeated Error errors = 3;
}

message Error {
    int32 code = 1;
    string description = 2;
}

```

现在，确保你已经在根目录下创建了 Makefile 文件，按照前几个服务的 Makefile 文件照猫画虎写一个即可，接下来运行`$ make build`来生成 gRPC 代码。据以往经验，已经自动生成了一些连接的 gRPC 方法的代码。 本文只会讲解其中一小部分 service 的运行，其余 service 的运行详解会在本系列的其余文章给出。在本文我们只介绍 User service 如何创建和获取用户。 在本系列接下来的文章中，我们将讨论认证和 [JWT](https://www.jianshu.com/p/576dbf44b2ae) 的具体实现。阅读过程中请做好相关标记。

【译者注：本系列是 [Ewan Valentine](http://ewanvalentine.io/author/ewan) 编写的关于 golang 微服务的[长文教程系列](https://ewanvalentine.io/tag/golang/)第三篇，每一篇的讲解都很细致，建议大家结合作者[源码](https://github.com/EwanValentine/shippy.git)仔细将每一篇阅读完毕】

你的处理程序应该是这样的：

```go
// user-service/handler.go
package main

import (
	"golang.org/x/net/context"
	pb "github.com/EwanValentine/shippy/user-service/proto/user"
)

type service struct {
	repo Repository
	tokenService Authable
}

func (srv *service) Get(ctx context.Context, req *pb.User, res *pb.Response) error {
	user, err := srv.repo.Get(req.Id)
	if err != nil {
		return err
	}
	res.User = user
	return nil
}

func (srv *service) GetAll(ctx context.Context, req *pb.Request, res *pb.Response) error {
	users, err := srv.repo.GetAll()
	if err != nil {
		return err
	}
	res.Users = users
	return nil
}

func (srv *service) Auth(ctx context.Context, req *pb.User, res *pb.Token) error {
	user, err := srv.repo.GetByEmailAndPassword(req)
	if err != nil {
		return err
	}
	res.Token = "testingabc"
	return nil
}

func (srv *service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {
	if err := srv.repo.Create(req); err != nil {
		return err
	}
	res.User = req
	return nil
}

func (srv *service) ValidateToken(ctx context.Context, req *pb.Token, res *pb.Token) error {
	return nil
}

```

现在让我们添加我们的 repository 代码：

```go
// user-service/repository.go
package main

import (
	pb "github.com/EwanValentine/shippy/user-service/proto/user"
	"github.com/jinzhu/gorm"
)

type Repository interface {
	GetAll() ([]*pb.User, error)
	Get(id string) (*pb.User, error)
	Create(user *pb.User) error
	GetByEmailAndPassword(user *pb.User) (*pb.User, error)
}

type UserRepository struct {
	db *gorm.DB
}

func (repo *UserRepository) GetAll() ([]*pb.User, error) {
	var users []*pb.User
	if err := repo.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (repo *UserRepository) Get(id string) (*pb.User, error) {
	var user *pb.User
	user.Id = id
	if err := repo.db.First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) GetByEmailAndPassword(user *pb.User) (*pb.User, error) {
	if err := repo.db.First(&user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (repo *UserRepository) Create(user *pb.User) error {
	if err := repo.db.Create(user).Error; err != nil {
		return err
	}
}

```

为了避免使用整数 ID ，我们还需要更改 ORM 行为，以便在创建时生成一个 [UUID](https://baike.baidu.com/item/UUID/5921266?fr=aladdin)。如果您不知道，UUID是一组随机生成的带有连字符的字符串，用作ID或主键。这比使用自动递增的整数 ID 更安全，因为它可以阻止人们猜测或遍历 API 节点。MongoDB 已经使用了 UUID，但我们需要告诉 Postgres 模型使用UUID。 因此，我们需要在`user-service/proto/user`中创建一个名为extensions.go的新文件，编码如下：

```go
package go_micro_srv_user

import (
	"github.com/jinzhu/gorm"
	"github.com/satori/go.uuid"
)

func (model *User) BeforeCreate(scope *gorm.Scope) error {
	uuid := uuid.NewV4()
	return scope.SetColumn("Id", uuid.String())
}

```
这段代码会关联到 GORM 的[事件生命周期](http://jinzhu.me/gorm/callbacks.html)中，以便在实体保存之前为我们的 Id 队列生成一个 UUID。

与 Mongodb 服务不同的是我们没有进行任何连接处理，这与原生 SQL/postgres 驱动程序的工作方式稍有不同，但是这次我们不需要为连接问题担心，因为我们正在使用一个名为'gorm'的软件包，我们来简单介绍一下。

## Gorm - Go + ORM

[Gorm](http://gorm.book.jasperxu.com/) 是一个相当轻量级的对象关系映射器，它可以很好地与 Postgres，MySQL，Sqlite 等配合使用。它可以非常容易地自动设置、使用和管理你的数据库模式。

这就是说，使用微服务，您的数据结构要小得多，耦合度更低，整体复杂性更小。

现在尝试一下创建用户，所以我们可以先创建一个 cli 工具。 这个 `user-cli` 位于这个项目的根目录下。类似于 `consignment-cli`，但稍有不同，代码如下：

```go
package main

import (
	"log"
	"os"

	pb "github.com/EwanValentine/shippy/user-service/proto/user"
	microclient "github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	"golang.org/x/net/context"
	"github.com/micro/cli"
	"github.com/micro/go-micro"
)


func main() {

	cmd.Init()

	// 创建一个新的 greeter client
	client := pb.NewUserServiceClient("go.micro.srv.user", microclient.DefaultClient)

	// 定义可用参数
	service := micro.NewService(
		micro.Flags(
			cli.StringFlag{
				Name:  "name",
				Usage: "You full name",
			},
			cli.StringFlag{
				Name:  "email",
				Usage: "Your email",
			},
			cli.StringFlag{
				Name:  "password",
				Usage: "Your password",
			},
			cli.StringFlag{
				Name: "company",
				Usage: "Your company",
			},
		),
	)

	// 作为一个 service 初始化启动
	service.Init(

		micro.Action(func(c *cli.Context) {

			name := c.String("name")
			email := c.String("email")
			password := c.String("password")
			company := c.String("company")

            // 调用 user service
			r, err := client.Create(context.TODO(), &pb.User{
				Name: name,
				Email: email,
				Password: password,
				Company: company,
			})
			if err != nil {
				log.Fatalf("Could not create: %v", err)
			}
			log.Printf("Created: %s", r.User.Id)

			getAll, err := client.GetAll(context.Background(), &pb.Request{})
			if err != nil {
				log.Fatalf("Could not list users: %v", err)
			}
			for _, v := range getAll.Users {
				log.Println(v)
			}

			os.Exit(0)
		}),
	)

	// 运行
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}
```
在这里，我们使用了 go-micro 的命令行助手，这非常简洁。

运行并创建一个用户：

```
$ docker-compose run user-cli command \
  --name="Ewan Valentine" \
  --email="ewan.valentine89@gmail.com" \
  --password="Testing123" \
  --company="BBC"

```
你应该在列表中看到创建完成的用户！

由于使用纯文本存储密码，安全性太低，在本系列的下一部分文章，我会把身份验证和 JWT 添加进来。

所以目前为止，代码雏形已经建立出来了，我们创建了一个额外的命令行工具，并且我们已经开始使用两种不同的数据库来保存我们的数据。

为了避免给读者造成太大压力，本文到此结束，若读者对本系列文章有所建议，请[尽可能给我反馈](ewan.valentine89@gmail.com)！

如果你发现这个系列有用，请[打赏作者](https://monzo.me/ewanvalentine)。

---

via: https://ewanvalentine.io/microservices-in-golang-part-3/

作者：[Ewan Valentine](http://ewanvalentine.io/author/ewan)
译者：[张阳](https://github.com/zhangyang9)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](Go语言中文网) 荣誉推出

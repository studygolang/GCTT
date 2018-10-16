已发布：https://studygolang.com/articles/12485

# Golang 中的微服务 - 第 4 部分 - 使用 JWT 做认证

在之前的文章里，我们创建了一个用户服务，保存了一些用户。现在我们看下如何在用户服务中安全的保存用户密码，同时，通过微服务写几个功能，来验证用户，发布安全令牌。

注意，我现在已经把我们的服务拆分到几个不同的仓库里。 我觉得这样部署起来容易些。 最开始我打算做成一个单独的仓库，但是发现用 Go 的 dep 管理有点麻烦，有很多冲突。我也会说明下如何独立地运行和测试微服务。

遗憾的是，用这种方法我们就不能用 docker-compose 了。 不过目前用起来还不错。如果你在这方面有什么建议，可以[给我发邮件](ewan.valentine89@gmail.com)！

现在你要手动启动数据库：

```
$ docker run -d -p 5432:5432 postgres
$ docker run -d -p 27017:27017 mongo
```

最新的仓库地址在下面：

- [https://github.com/EwanValentine/shippy-consignment-service]( https://github.com/EwanValentine/shippy-consignment-service)
- [https://github.com/EwanValentine/shippy-user-service](https://github.com/EwanValentine/shippy-user-service)
- [https://github.com/EwanValentine/shippy-vessel-service](https://github.com/EwanValentine/shippy-vessel-service)
- [https://github.com/EwanValentine/shippy-user-cli](https://github.com/EwanValentine/shippy-user-cli)
- [https://github.com/EwanValentine/shippy-consignment-cli](https://github.com/EwanValentine/shippy-consignment-cli)

首先，我们要更新下 handler 文件，做密码哈希，这是非常必要的。绝对不能，也坚决不要使用明文密码。你们可能会想 ‘废话，那还用说么’ ，但我还要坚持强调下！

```go
// shippy-user-service/handler.go
...
func (srv *service) Auth(ctx context.Context, req *pb.User, res *pb.Token) error {
	log.Println("Logging in with:", req.Email, req.Password)
	user, err := srv.repo.GetByEmail(req.Email)
	log.Println(user)
	if err != nil {
		return err
	}

	// Compares our given password against the hashed password
	// stored in the database
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return err
	}

	token, err := srv.tokenService.Encode(user)
	if err != nil {
		return err
	}
	res.Token = token
	return nil
}

func (srv *service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {

	// Generates a hashed version of our password
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hashedPass)
	if err := srv.repo.Create(req); err != nil {
		return err
	}
	res.User = req
	return nil
}
```

这里没做太多改动，除了增加了密码哈希功能，我们在保存新用户之前把哈希后的内容作为新的密码。同样的，在认证部分，我们会校验密码的哈希值。

现在我们能够安全的认证数据库里的用户信息，我们需要一个机制，能使用接口和分布式服务来做认证。实现这样的功能有许多方法，但是我发现，能通过服务和 web 使用的最简单的认证方法是用 [JWT](https://jwt.io/)。

不过在我们继续下面的内容之前，请查看下我在每个服务中的 Dockfiles 和 Makefiles 做的修改。为了匹配最新的 git 仓库，我也修改了 imports 。

## JWT

[JWT](https://jwt.io/) 是 JSON web tokens 的缩写，是一个分布式的安全协议。类似 OAuth。 概念很简单，用算法给用户生成一个唯一的哈希，这个哈希值可以用来比较和校验。不仅如此，token 自身也会包含用户的元数据 (metadata)。也就是说，用户数据本身也可以是 token 的一部分。 我们看一个 JWT 的实例：

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ
```

token 被 '.' 分成三部分。每个部分都有各自的含义。第一部分是 token 自身的元数据，像 token 的类型，创建 token 使用的算法等。让客户端能知道怎么给 token 解码。第二部分是用户自定义的元数据。可以是你的用户详情，过期时间，任何你想填的内容。最后一个部分是认证签名，包括用什么方法，什么数据给 token 做哈希这类信息。

当然用 JWT 也有缺点和使用风险，这些缺点在 [这篇文章](http://cryto.net/~joepie91/blog/2016/06/13/stop-using-jwt-for-sessions/) 中总结的很好。同时，我推荐大家读 [这篇文章](https://www.owasp.org/index.php/JSON_Web_Token_(JWT)_Cheat_Sheet_for_Java)，这里有安全层面上的最佳实践方法。

有一点我想要大家特别注意下，就是获取用户的源 IP，并用它作为令牌声明的一部分。这样可以确保没有人能盗取你的令牌，并在另一台机器上伪装成你。确保使用 https，减少这种类型的攻击，因为使用 https 能保护你的 token，免受中间人攻击。

用来做 JWT 哈希的算法有很多，大体分为两类。对称加密和非对称加密。对称加密类似我们现在用的方式，使用共享盐值 (salt)。非对称加密会在客户端和服务端分别使用公钥和私钥。用来在多个服务之间做认证是极好的。

更多资源：

- [Auth0](https://auth0.com/blog/json-web-token-signing-algorithms-overview/)
- [RFC spec for algorithms](https://tools.ietf.org/html/rfc7518#section-3)

现在我们知道 JWT 的基本概念了。我们更新下 `token_service.go` 实现这些操作。我们可以用这个 Go 的库 `github.com/dgrijalva/jwt-go` ，这个库写的很棒，有很多不错的实例。

```go
// shippy-user-service/token_service.go
package main

import (
	"time"

	pb "github.com/EwanValentine/shippy-user-service/proto/user"
	"github.com/dgrijalva/jwt-go"
)

var (

	// Define a secure key string used
	// as a salt when hashing our tokens.
	// Please make your own way more secure than this,
	// use a randomly generated md5 hash or something.
	key = []byte("mySuperSecretKeyLol")
)

// CustomClaims is our custom metadata, which will be hashed
// and sent as the second segment in our JWT
type CustomClaims struct {
	User *pb.User
	jwt.StandardClaims
}

type Authable interface {
	Decode(token string) (*CustomClaims, error)
	Encode(user *pb.User) (string, error)
}

type TokenService struct {
	repo Repository
}

// Decode a token string into a token object
func (srv *TokenService) Decode(tokenString string) (*CustomClaims, error) {

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return key, nil
	})

	// Validate the token and return the custom claims
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}

// Encode a claim into a JWT
func (srv *TokenService) Encode(user *pb.User) (string, error) {

	expireToken := time.Now().Add(time.Hour * 72).Unix()

	// Create the Claims
	claims := CustomClaims{
		user,
		jwt.StandardClaims{
			ExpiresAt: expireToken,
			Issuer:    "go.micro.srv.user",
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token and return
	return token.SignedString(key)
}
```

照常，我写了一些注释解释了部分细节，不过这里的导言写的比较简单。Decode 方法，是将一个字符串 token 解析成一个 token 对象，验证下，如果有效，则返回一个 CustomClaims 对象。这样我们能通过这个 CustomClaims 对象获取到用户的元数据来确认该用户是否有效。

Encode 方法，做的正好相反，将你的自定义的元数据，哈希成一个新的 JWT ，并返回。

注意我们在上面设置了一个 'key' 变量，这个 key 是一个安全盐值（salt），在生产环境下请用一个比这个更安全的盐值。

现在我们有了一个验证 token 的服务。我们更新下 user-cli，我现在已经把这部分简化成一个脚本，因为这之前的 cli 部分的代码有点问题，我以后再来说这个问题，这个工具就是为了测试用：

```go
// shippy-user-cli/cli.go
package main

import (
	"log"
	"os"

	pb "github.com/EwanValentine/shippy-user-service/proto/user"
	micro "github.com/micro/go-micro"
	microclient "github.com/micro/go-micro/client"
	"golang.org/x/net/context"
)

func main() {

	srv := micro.NewService(

		micro.Name("go.micro.srv.user-cli"),
		micro.Version("latest"),
	)

	// Init will parse the command line flags.
	srv.Init()

	client := pb.NewUserServiceClient("go.micro.srv.user", microclient.DefaultClient)

	name := "Ewan Valentine"
	email := "ewan.valentine89@gmail.com"
	password := "test123"
	company := "BBC"

	r, err := client.Create(context.TODO(), &pb.User{
		Name:     name,
		Email:    email,
		Password: password,
		Company:  company,
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

	authResponse, err := client.Auth(context.TODO(), &pb.User{
		Email:    email,
		Password: password,
	})

	if err != nil {
		log.Fatalf("Could not authenticate user: %s error: %v\n", email, err)
	}

	log.Printf("Your access token is: %s \n", authResponse.Token)

	// let's just exit because
	os.Exit(0)
}
```

目前我们有一些硬编码的值，替换下这些值，然后用 `$make build && make run` 执行脚本。你可以看到返回了一个 token。把这个长长的 token 值拷贝下，你马上就会用到！

现在我们更新下我们的 consignment-cli，生成一个 token 字符串，传给 consignment-service：

```go
// shippy-consignment-cli/cli.go
...
func main() {

	cmd.Init()

	// Create new greeter client
	client := pb.NewShippingServiceClient("go.micro.srv.consignment", microclient.DefaultClient)

	// Contact the server and print out its response.
	file := defaultFilename
	var token string
	log.Println(os.Args)

	if len(os.Args) < 3 {
		log.Fatal(errors.New("Not enough arguments, expecing file and token."))
	}

	file = os.Args[1]
	token = os.Args[2]

	consignment, err := parseFile(file)

	if err != nil {
		log.Fatalf("Could not parse file: %v", err)
	}

	// Create a new context which contains our given token.
	// This same context will be passed into both the calls we make
	// to our consignment-service.
	ctx := metadata.NewContext(context.Background(), map[string]string{
		"token": token,
	})

	// First call using our tokenised context
	r, err := client.CreateConsignment(ctx, consignment)
	if err != nil {
		log.Fatalf("Could not create: %v", err)
	}
	log.Printf("Created: %t", r.Created)

	// Second call
	getAll, err := client.GetConsignments(ctx, &pb.GetRequest{})
	if err != nil {
		log.Fatalf("Could not list consignments: %v", err)
	}
	for _, v := range getAll.Consignments {
		log.Println(v)
	}
}
```

现在我们更新 consignment-service 来核对一下获取 token 的请求，并传给 user-service ：

```go
// shippy-consignment-service/main.go
func main() {
	...
	// Create a new service. Optionally include some options here.
	srv := micro.NewService(

		// This name must match the package name given in your protobuf definition
		micro.Name("go.micro.srv.consignment"),
		micro.Version("latest"),
		// Our auth middleware
		micro.WrapHandler(AuthWrapper),
	)
	...
}

...

// AuthWrapper is a high-order function which takes a HandlerFunc
// and returns a function, which takes a context, request and response interface.
// The token is extracted from the context set in our consignment-cli, that
// token is then sent over to the user service to be validated.
// If valid, the call is passed along to the handler. If not,
// an error is returned.
func AuthWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, resp interface{}) error {
		meta, ok := metadata.FromContext(ctx)
		if !ok {
			return errors.New("no auth meta-data found in request")
		}

		// Note this is now uppercase (not entirely sure why this is...)
		token := meta["Token"]
		log.Println("Authenticating with token: ", token)

		// Auth here
		authClient := userService.NewUserServiceClient("go.micro.srv.user", client.DefaultClient)
		_, err := authClient.ValidateToken(context.Background(), &userService.Token{
			Token: token,
		})
		if err != nil {
			return err
		}
		err = fn(ctx, req, resp)
		return err
	}
}
```

然后我们执行 consignment-cli工具，cd 到新的 shippy-consignment-cli 仓库，执行 `$make build` 来创建新的 docker 镜像， 现在运行：

```
$ make build
$ docker run --net="host" \
	  -e MICRO_REGISTRY=mdns \
	  consignment-cli consignment.json \
	  <TOKEN_HERE>
```

注意下我们在运行 docker 容器的时候使用了 `--net="host"` 标记。用来告诉 Docker 在本地局域网来运行我们的容器，如127.0.0.1 或 localhost，而不是 Docker 的内网。注意，用这个方法不需要使用端口转发。因此只需要用 -p 8080 代替 -p 8080:8080 就可以了。 [更多关于 Docker 网络的内容请看这里](https://docs.docker.com/engine/userguide/networking/)。

到这一步，你就能看到新的 consignment 被创建了。试着从 token 里删除几个字母，token 就会变无效。会出现错误。

好了，现在我们创建了一个 JWT token 的服务，和一个用来验证 JWT token 的中间件，而 JWT token 用来验证用户。

如果你不想用 go-micro，而是 vanilla grpc，你的中间件是类似这样的：

```go
func main() {
	myServer := grpc.NewServer(
	grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(AuthInterceptor),
	)
}

func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	// Set up a connection to the server.
	conn, err := grpc.Dial(authAddress, grpc.WithInsecure())
	if err != nil {
	log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewAuthClient(conn)
	r, err := c.ValidateToken(ctx, &pb.ValidateToken{Token: token})

	if err != nil {
		log.Fatalf("could not authenticate: %v", err)
	}

	return handler(ctx, req)
}

```

这个设置在本地运行有点不灵活。但是我们并不总在本地运行每一个服务。我们创建的服务应该互相独立，能单独测试。现在的情况下，如果我们只需要测试 consignment-service，就不需要启动 auth-service。因此我在这加了一个开关，打开或关闭其他的服务。

我更新 consignment-service 里的 auth 封装：

```go
// shippy-user-service/main.go
func AuthWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, resp interface{}) error {
	// This skips our auth check if DISABLE_AUTH is set to true
		if os.Getenv("DISABLE_AUTH") == "true" {
			return fn(ctx, req, resp)
		}
	}
}
```

然后在 Makefile 里加个开关：

```
// shippy-user-service/Makefile
run:
	docker run -d --net="host" \
		-p 50052 \
		-e MICRO_SERVER_ADDRESS=:50052 \
		-e MICRO_REGISTRY=mdns \
		-e DISABLE_AUTH=true \
		consignment-service

```

用这个方式，在本地运行微服务中的某个部分，会更简单，解决这个问题还有几个不同的方法，但是我认为这个最简单。希望你觉得有点帮助，尽管在方向上有小改动。同时，如果你对使用单个仓库运行微服务这方面有什么建议，非常欢迎告知我，那么这部分内容就简单多了。

如果你发现这篇文章里有任何 bug，错误，或者你有对这篇文章有任何反馈或者有帮助的建议，[请给我发邮件](ewan.valentine89@gmail.com)。

如果你觉得这个系列文章有用，而且你用了广告过滤（俺没怪你）。请考虑打赏下我的辛苦劳动噢。非常感谢！。
[https://monzo.me/ewanvalentine](https://monzo.me/ewanvalentine)

---

via: https://ewanvalentine.io/microservices-in-golang-part-4/

作者：[André Carvalho](https://ewanvalentine.io/microservices-in-golang-part-4/)
译者：[ArisAries](https://github.com/ArisAries)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

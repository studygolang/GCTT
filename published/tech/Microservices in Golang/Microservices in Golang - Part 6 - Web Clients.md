已发布：https://studygolang.com/articles/12798

# Microservices in Golang - Part 6 - Web Clients

在之前的文章中，我们看了一些使用 go-micro 和 go 语言的生成的各种事件驱动的方法。 在本篇文章，我们将深入到客户端，探究一下如何创建一个能够与我们之前创建的平台交互的 Web 客户端。

这篇文章会介绍如何使用 [micro](https://github.com/micro/micro) 工具包生成 web 客户端从外部代理内部 rpc 方法。

我们会创建一个 user 接口用于生成平台的登录界面、还会创建一个接口用于使用我们的 consignments。该界面包含了创建用户、登录、和创建 consignments 等功能。 本系列的前几篇文章已经介绍过其中的部分代码了，在这篇文章我会带大家深入了解一下。

所以让我们开始吧！

## RPC 复兴

REST 已经在网络上服务了很多年了，并且迅速成为管理客户端和服务器之间资源的途径。REST 正在逐渐取代已经过时的 RPC 和 SOAP。曾经必须写一个 wsdl 文件的时代已经过去了。

REST 向我们承诺了一种实用，简单和标准化的资源管理方法。 REST 使用 http 协议明确了正在执行的具体 web 动作类型。REST 鼓励我们使用 http 错误响应码来更精确地描述服务器的响应状态。而且大多数情况下，这种方法运行良好，并没有问题。但是像所有好东西一样，REST有许多不足和缺点，我不打算在这里详细介绍。大家有兴趣可以参考[这篇文章](https://medium.freecodecamp.org/rest-is-the-new-soap-97ff6c09896d)。

但是！随着**微服务**的出现，RPC 正在卷土重来。

REST 对于管理不同的资源非常有用，但微服务通常只处理单一资源，这一性质导致我们不需要在微服务的上下文中使用 RESTful 术语。相反，我们可以专注于每个服务的单一的具体操作和交互。

## Micro

我们已经在本系列教程中广泛使用了 go-micro，现在我们将介绍 micro cli 这个工具包。这个 micro 工具包提供了的功能包括 API 网关、 sidecar、Web 代理以及其他一些很酷的功能。但是这篇文章我们使用到的功能主要是 API 网关。

API 网关将允许我们将 rpc 调用代理为 Web 友好的 javascriptON rpc 调用，然后将客户端应用程序中使用的 url 暴露出来。

那么以上这些炫酷功能是如何工作的？ 

首先要确保安装了 micro 工具包：

```
$ go get -u github.com/micro/micro
```

Docker环境下使用 Micro 更好的方法还是建议大家使用Docker镜像：

```
$ docker pull microhq/micro
```
接下来可以看一下 user 服务的代码，我对 user 服务的代码做了一些错误处理和命名约定方面的修改：

```go
// shippy-user-service/main.go
package main

import (
	"log"

	pb "github.com/EwanValentine/shippy-user-service/proto/auth"
	"github.com/micro/go-micro"
	_ "github.com/micro/go-plugins/registry/mdns"
)

func main() {

	// 创建了一个数据库 connection 
	// main 方法结束之前要关闭数据库连接
	db, err := CreateConnection()
	defer db.Close()

	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}

	// 将 user 结构类型自动移植到数据库类型中。此操作在服务每一次重启时都会做一次检测
	db.AutoMigrate(&pb.User{})

	repo := &UserRepository{db}

	tokenService := &TokenService{repo}

	// 创建一个新的服务
	srv := micro.NewService(

		// 这个名字必须于你在protobuf definition定义的包名匹配
		micro.Name("shippy.auth"),
	)

	// Init 用于初始化命令行参数
	srv.Init()
	 
		// Will comment this out for now to save having to run this locally... 
	// publisher := micro.NewPublisher("user.created", srv.Client())

	// 注册 handler
	pb.RegisterAuthHandler(srv.Server(), &service{repo, tokenService, publisher})

	// 启动 server
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

```

```c
// shippy-user-service/proto/auth/auth.proto
syntax = "proto3";

package auth;

service Auth {
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

```go
// shippy-user-service/handler.go
package main

import (
	"errors"
	"fmt"
	"log"

	pb "github.com/EwanValentine/shippy-user-service/proto/auth"
	micro "github.com/micro/go-micro"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

const topic = "user.created"

type service struct {
	repo         Repository
	tokenService Authable
	Publisher    micro.Publisher
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
	log.Println("Logging in with:", req.Email, req.Password)
	user, err := srv.repo.GetByEmail(req.Email)
	log.Println(user, err)
	if err != nil {
		return err
	}

	// 比较输入的密码与存储在数据库里的哈希密码
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

	log.Println("Creating user: ", req)

	// 为我们的密码生成一个哈希值
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New(fmt.Sprintf("error hashing password: %v", err))
	}

	req.Password = string(hashedPass)
	if err := srv.repo.Create(req); err != nil {
		return errors.New(fmt.Sprintf("error creating user: %v", err))
	}

	res.User = req
	if err := srv.Publisher.Publish(ctx, req); err != nil {
		return errors.New(fmt.Sprintf("error publishing event: %v", err))
	}

	return nil
}

func (srv *service) ValidateToken(ctx context.Context, req *pb.Token, res *pb.Token) error {

	// Decode token
	claims, err := srv.tokenService.Decode(req.Token)

	if err != nil {
		return err
	}

	if claims.User.Id == "" {
		return errors.New("invalid user")
	}

	res.Valid = true

	return nil
}

```

现在运行 `$ make build && make run`。 然后转到 shippy-email-service 运行`$ make build && make run`。 一旦这两个服务都运行，运行：

```shell
$ docker run -p 8080:8080 \ 
		-e MICRO_REGISTRY=mdns \
		microhq/micro api \
		--handler=rpc \
		--address=:8080 \
		--namespace=shippy 
```

这将在 Docker 容器中开一个 8080 端口上，该端口将 micro api-gateway 作为 rpc 处理程序暴露出来，使用 mdns 作为本地的注册表，使用命名空间 shippy，shippy 是我们所有服务名称的第一部分。例如 shippy.auth 或 shippy.email。设置它是很重要的，因为它默认为 go.micro.api，在默认情况下，go.micro.api 是无法找到我们需要的特定服务来进行代理的。

我们现在可以使用以下方式调用我们的 user 服务方法：

创建一个 user：

```shell
curl -XPOST -H 'Content-Type: application/javascripton' \
	-d '{ "service": "shippy.auth", "method": "Auth.Create", "request": { "user": { "email": "ewan.valentine89@gmail.com", "password": "testing123", "name": "Ewan Valentine", "company": "BBC" } } }' \ 
	http://localhost:8080/rpc
```
这个请求中包含了我们想要传送给的服务名、要使用的服务方法、以及服务数据。

验证用户：

```shell
$ curl -XPOST -H 'Content-Type: application/javascripton' \ 
	-d '{ "service": "shippy.auth", "method": "Auth.Auth", "request":  { "email": "your@email.com", "password": "SomePass" } }' \
	http://localhost:8080/rpc
```

## Consignment service

现在再次启动我们的 consignment 服务，`$ make build && make run`。 我们不需要在这里改变任何东西，但是，运行 rpc 代理的话我们还应该
创建一个 consignment：

```shell
$ curl -XPOST -H 'Content-Type: application/javascripton' \ 
	-d '{
		"service": "shippy.consignment",
		"method": "ConsignmentService.Create",
		"request": {
			"description": "This is a test",
			"weight": "500",
			"containers": []
		}
	}' --url http://localhost:8080/rpc
		
```
## Vessel service

最后为了测试用户接口界面，我们需要运行 vessel 服务，这里没有对代码有什么修改，直接运行 `$ make build && make run` 即可。

## User interface

现在可以使用我们的刚刚创建的新 rpc 节点创建一个用户界面。本文使用了 React，当然如果你喜欢的话可以使用其余的架构。请求都是一样的。本文使用来自 Facebook 的 react-create-app 库: 

`$ npm install -g react-create-app` 

安装完成后，执行 `$ react-create-app shippy-ui`。 这将为您创建一个 React 应用程序的框架。

```javascript
// shippy-ui/src/App.javascript
import React, { Component } from 'react';
import './App.css';
import CreateConsignment from './CreateConsignment';
import Authenticate from './Authenticate';

class App extends Component {

	state = {
		err: null,
		authenticated: false,
	}

	onAuth = (token) => {
		this.setState({
			authenticated: true,
		});
	}

	renderLogin = () => {
		return (
			<Authenticate onAuth={this.onAuth} />
		);
	}

	renderAuthenticated = () => {
		return (
			<CreateConsignment />
		);
	}

	getToken = () => {
		return localStorage.getItem('token') || false;
	}

	isAuthenticated = () => {
		return this.state.authenticated || this.getToken() || false;
	}

	render() {
		const authenticated = this.isAuthenticated();
		return (
			<div className="App">
				<div className="App-header">
					<h2>Shippy</h2>
				</div>
				<div className='App-intro container'>
					{(authenticated ? this.renderAuthenticated() : this.renderLogin())}
				</div>
			</div>
		);
	}
}

export default App;
```

现在让添加我们的两个主要组件，Authenticate 和 CreateConsignment：

```javascript
// shippy-ui/src/Authenticate.javascript
import React from 'react';

class Authenticate extends React.Component {

	constructor(props) {
		super(props);
	}

	state = {
		authenticated: false,
		email: '',
		password: '',
		err: '',
	}

	login = () => {
		fetch(`http://localhost:8080/rpc`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/javascripton',
			},
			body: javascriptON.stringify({
				request: {
					email: this.state.email,
					password: this.state.password,
				},
				service: 'shippy.auth',
				method: 'Auth.Auth',
			}),
		})
		.then(res => res.javascripton())
		.then(res => {
			this.props.onAuth(res.token);
			this.setState({
				token: res.token,
				authenticated: true,
			});
		})
		.catch(err => this.setState({ err, authenticated: false, }));
	}

	signup = () => {
		fetch(`http://localhost:8080/rpc`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/javascripton',
			},
			body: javascriptON.stringify({
				request: {
					email: this.state.email,
					password: this.state.password,
					name: this.state.name,
				},
				method: 'Auth.Create',
				service: 'shippy.auth',
			}),
		})
		.then((res) => res.javascripton())
		.then((res) => {
			this.props.onAuth(res.token.token);
			this.setState({
				token: res.token.token,
				authenticated: true,
			});
			localStorage.setItem('token', res.token.token);
		})
		.catch(err => this.setState({ err, authenticated: false, }));
	}

	setEmail = e => {
		this.setState({
			email: e.target.value,
		});
	}

	setPassword = e => {
		this.setState({
			password: e.target.value,
		});
	}

	setName = e => {
		this.setState({
			name: e.target.value,
		});
	}

	render() {
		return (
			<div className='Authenticate'>
				<div className='Login'>
					<div className='form-group'>
						<input
							type="email"
							onChange={this.setEmail}
							placeholder='E-Mail'
							className='form-control' />
					</div>
					<div className='form-group'>
						<input
							type="password"
							onChange={this.setPassword}
							placeholder='Password'
							className='form-control' />
					</div>
					<button className='btn btn-primary' onClick={this.login}>Login</button>
					<br /><br />
				</div>
				<div className='Sign-up'>
					<div className='form-group'>
						<input
							type='input'
							onChange={this.setName}
							placeholder='Name'
							className='form-control' />
					</div>
					<div className='form-group'>
						<input
							type='email'
							onChange={this.setEmail}
							placeholder='E-Mail'
							className='form-control' />
					</div>
					<div className='form-group'>
						<input
							type='password'
							onChange={this.setPassword}
							placeholder='Password'
							className='form-control' />
					</div>
					<button className='btn btn-primary' onClick={this.signup}>Sign-up</button>
				</div>
			</div>
		);
	}
}

export default Authenticate;

```

and...

```javascript
// shippy-ui/src/CreateConsignment.javascript
import React from 'react';
import _ from 'lodash';

class CreateConsignment extends React.Component {

	constructor(props) {
		super(props);
	}

	state = {
		created: false,
		description: '',
		weight: 0,
		containers: [],
		consignments: [],
	}

	componentWillMount() {
		fetch(`http://localhost:8080/rpc`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/javascripton',
			},
			body: javascriptON.stringify({
				service: 'shippy.consignment',
				method: 'ConsignmentService.Get',
				request: {},
			})
		})
		.then(req => req.javascripton())
		.then((res) => {
			this.setState({
				consignments: res.consignments,
			});
		});
	}

	create = () => {
		const consignment = this.state;
		fetch(`http://localhost:8080/rpc`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/javascripton',
			},
			body: javascriptON.stringify({
				service: 'shippy.consignment',
				method: 'ConsignmentService.Create',
				request: _.omit(consignment, 'created', 'consignments'),
			}),
		})
		.then((res) => res.javascripton())
		.then((res) => {
			this.setState({
				created: res.created,
				consignments: [...this.state.consignments, consignment],
			});
		});
	}

	addContainer = e => {
		this.setState({
			containers: [...this.state.containers, e.target.value],
		});
	}

	setDescription = e => {
		this.setState({
			description: e.target.value,
		});
	}

	setWeight = e => {
		this.setState({
			weight: Number(e.target.value),
		});
	}

	render() {
		const { consignments, } = this.state;
		return (
			<div className='consignment-screen'>
				<div className='consignment-form container'>
					<br />
					<div className='form-group'>
						<textarea onChange={this.setDescription} className='form-control' placeholder='Description'></textarea>
					</div>
					<div className='form-group'>
						<input onChange={this.setWeight} type='number' placeholder='Weight' className='form-control' />
					</div>
					<div className='form-control'>
						Add containers...
					</div>
					<br />
					<button onClick={this.create} className='btn btn-primary'>Create</button>
					<br />
					<hr />
				</div>
				{(consignments && consignments.length > 0
					? <div className='consignment-list'>
							<h2>Consignments</h2>
							{consignments.map((item) => (
								<div>
									<p>Vessel id: {item.vessel_id}</p>
									<p>Consignment id: {item.id}</p>
									<p>Description: {item.description}</p>
									<p>Weight: {item.weight}</p>
									<hr />
								</div>
							))}
						</div>
					: false)}
			</div>
		);
	}
}

export default CreateConsignment;

```

**注意**：我还将 Twitter Bootstrap 添加到 /public/index.html 并更改了一些CSS。

现在运行用户界面 `$ npm start`。 之后应该浏览器会自动打开一个界面。您现在应该可以注册并登录并查看 consignment 表单，您可以在其中创建新 consignments。看看你的开发工具中的 network 选项，然后看看 rpc 方法从我们的不同微服务中触发和获取我们的数据。

第6部分到这里就结束了，如果您有任何反馈，请给我发一封[电子邮件](ewan.valentine89@gmail.com)，我会尽快答复（有可能不会很及时，敬请见谅）。

如果你发现这个系列有用，并且你使用了一个广告拦截器。 请考虑为我的时间和努力赞助几块钱。十分感谢！！ [https://monzo.me/ewanvalentine](https://monzo.me/ewanvalentine)

----------------

via: https://ewanvalentine.io/microservices-in-golang-part-6/

作者：[André Carvalho](https://ewanvalentine.io/microservices-in-golang-part-6/)
译者：[zhangyang9](https://github.com/zhangyang9)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
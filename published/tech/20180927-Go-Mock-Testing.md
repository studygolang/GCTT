首发于：https://studygolang.com/articles/16801

# Go Mock 测试

(在开发的过程中) 你应该一直都写测试。不管使用什么语言，为了完全的理解如何能写出生产环境级别的应用，你应该拥有写单元测试的能力。有些人把测试驱动 (TDD) 做到极致。TDD 提倡你在写功能之前写测试。当你尝试写一些容易测试的代码的时候，这是个好方法。直到现在，我发现 TDD 很像 agile ( 敏捷软件开发 ); 很多人说他们有做测试驱动，但是实际中，他们只在一些混合的版本中做。到最后，它归根到底像，在吃牛排之前吃蔬菜的问题。( 西餐中 ) 比较健康的吃法是先吃蔬菜，然后吃牛排。就像在编程中，比较好的做法是，先写单元测试，然后写功能。但是大多数人，做不到健康的吃法，他们只吃牛排，不吃蔬菜，也就是只写功能，不写单元测试，只要功能可以正常跑起来工作就感到满意了。

我对于测试的观点是，如果不去测试我就感到不适，但是就像我的儿子会说：我并不喜欢他们。你会发现很少人认为单元测试是没有价值的，但是大多数人都不能写好的单元测试。很明显这只是 (我的一个) 观点，但是好的测试，在实际测试代码的时候能跑的很快，也应该能使代码变得更强健。看一眼这本书：*Ye Olde Testing Pyramid* 你就会明白，你应该写大量的单元测试，小量的集成测试，以及很小量的端到端测试。( 如下图所示 ) 当你把处于尾部的测试往金字塔上方提升时，从本质上说，因为不断增加的依赖数量，测试变得脆弱。越往金字塔的上方提升，你就会发现时间变得越长。

![testing img](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-mock-testing/testing-pyramid.png)

这就是为啥你的大部分努力应该花在单元测试。比较好的做法是，为了方便的对失败进行精确定位，你应该把你的大部分努力花在测试独立的组件上。现在我的文章中提到了 Hexagonal Architecture，其中提到的写 " 端口式的 " 代码。每一个端口使用一个接口来连接域名逻辑，域名逻辑也使用接口来和端口交互。当他们被组合的时候，我们就会看到你可以将不同的接口实现互换，并且不需要去重构一群代码。

在很多关于测试的书上你有时会发现这样一段话：

> 写能测试的代码
> -- by 不同测试书的作者

这段话看起来很显而易见，但是如果你不去关注测试优先 (就像那些狂热的 TDD 说教者) 的话，通常你会发现，当你写代码的时候，去写测试是一件很难得事情。Hexagonal 架构通常会使你拥有可以测试的代码，因为你可以插入一些模拟依赖，这些依赖允许我们去测试当业务逻辑在没有去配置资源或者启动测试数据的时候也能工作。

## 启动 Mockgen

Go 是一门特殊的语言，它内置框架里将测试放在一等公民的位置，所以支持很多类型的测试是 Go 工程的一部分。所以 Go 中存在很多模拟框架，但是我们将会使用其中一个作为主程序的一部分，那就是 [Mockgen](https://github.com/golang/mock)。

安装 Mockgen 只需要运行以下命令 (确保你的 `$GOPATH/bin` 是源路径的一部分)

```
go get github.com/golang/mock/gomock
go install github.com/golang/mock/mockgen
```

现在我们已经安装了 Mockgen，我们将会使用它来生成 mocks。

## 关于 Mocking

Mocks 和 stubs 不同。一个 stub，在调用的时候，将会允许你去生成期望的响应，但是不会提供一个简单的机制去校验已被调用。Mocks 允许你去模拟和特殊服务或者方法的交互，以及常常提供一个允许测试者去核实交互的核实方法。

Mockgen 将会为我们做很多笨重的提升去生成 mocks。Mockgen 有很多特性，这些特性我这里不讲，但是你有空的时候就去看看它的文档吧，去找找一些测试的高级方法。我们将会为上个例子中写的所有接口生成 mocks。

    mockgen -package mocks -destination mocks/ticket.go hex-example/ticket TicketRepository,TicketService,TicketHandler

在这里你可以看到我们想要把我们的输出放在一个模拟的包和目录中，名字和我们模拟的域一样。如果你需要迁移代码，这将会变得有用。然后我们定义输入包和最后我们想 mock 的接口。瞧，我们有 mocks 了。

## 测试程序组

下一步，我们将会安装一个程序包，这个程序包将会允许我们去减少启动时候的重复代码，以及保持测试简洁。我们将会使用一个程序包叫做：testify。程序组允许你将测试组织起来，以及分享一个通用的启动。测试，像代码一样，应该是可读的。测试也应该是简洁和定向的。程序组允许你移动一些逻辑，移动到启动和分解方法的前面或者后面。在这个例子里，我们想要在每一个测试之前启动我们的 mocks，而不想周而复始的使用一些代码。

所以，我们会在 ticket 目录里创建一个新的测试文件，叫做：service_test.go。根据 IDE 的情况，你可能会在文件的最上方预定义这个包。请确保你正在使用一个单独的 `_test` 包，在这个例子里它叫做：ticket_test。这很重要，因为你想要你的测试代码在你测试的包外面。我们做个因为我们想要把我们正在测试的包当做一个黑盒，同时，使用 `*_test` 包我们可以保证我们只可以接触到 exported identifiers。

下一步我们将会创建一个测试程序组的框架，以及一个初始运行的脚本。我们的框架会把我们需要用到的任何服务以及我们将测试的服务存储起来。这个案例的支持服务将会被模仿，同时通常用在测试下创建服务。个人认为我喜欢将需要测试的服务命名为 underTest，来确保我们在测试中正在做的事。最后，我们需要运行一个方法，将 T 插入测试组中的测试对象中。这个依附于 Go 测试的标准。

```go
func TestTicketServiceSuite(t *testing.T) {
    suite.Run(t, new(TicketServiceTestSuite))
}

type TicketServiceTestSuite struct {
    suite.Suite
    ticketRepo *mocks.MockTicketRepository
    underTest  ticket.TicketService
}
```

现在我们将会创建一个 启动方法，这个方法将允许我们启动 mocks。这将会举例说明我们使用 Mockgen 定义的模拟的仓库，以及创建一个新的服务对象来测试。

```go
func (suite *TicketServiceTestSuite) SetupTest() {
    mockCtrl := Gomock.NewController(suite.T())
    defer mockCtrl.Finish()

    suite.ticketRepo = mocks.NewMockTicketRepository(mockCtrl)
    suite.underTest = ticket.NewTicketService(suite.ticketRepo)
}
```

最后，我们准备写测试。

## 写测试

有很多方法来写测试，但是按照早期设定，你想保证它是清晰和简洁的。我发现 Arrage, Act, Assert 方法帮助保持我的测试更有组织性一些。这种方式来定义你的测试，你可以更容易经过一个测试的流程，以及找到它是如何工作的。看 Arrange 章节，这会允许你测试中需要运行的状态。Act 是一个正在执行服务。然后 Assert 验证测试正在支持做什么。典型的，如果一个测试失败了，那么它会是一个非预期的状态，进而促使断言失败或者会改变逻辑以至于改变结果。

之前规定的模仿类给你一个能力去做一些关于被模仿的对象本身的断言。在这个方案中，我们将会验证一个库方法被调用然后它将会返回一个期望的值。当我们安排一个测试时，我们将会启动一个模仿，以及设置一个支持如何演示的方案。

```go
t := &ticket.Ticket{
    Creator: "Joel",
}
suite.ticketRepo.EXPECT().Create(gomock.AssignableToTypeOf(&ticket.Ticket{})).Return(nil)
```

在他们模仿测试的最后，会跑完一个 Expect 的字句，然后做验证。我们也可以做一些额外的验证，这样我们会知道一些关于具体服务方法的事情发生。以创建为例，我们想要赋值一个 id 来创建、更新票的状态。所以我们应该声明这些从结果中被算出的值。

```go
func (suite *TicketServiceTestSuite) TestCreate() {
    //Arrange
    t := &ticket.Ticket{
        Creator: "Joel",
    }
    suite.ticketRepo.EXPECT().Create(gomock.AssignableToTypeOf(&ticket.Ticket{})).Return(nil)

    //Act
    err := suite.underTest.CreateTicket(t)

    //Assert
    suite.NoError(err, "Shouldn't error")
    suite.NotNil(t.ID, "should not be null")
    suite.NotNil(t.Created, "should not be null")
    suite.NotNil(t.Updated, "should not be null")

}
```

剩下的方法我们可以做这个，如果我们想的话，甚至也可以去测试错误处理方法。测试处理代码是非常符合期待的，我们将会模仿服务，而不是库。而且我们也会有一个稍微不同的方法去测试处理。

Go 有一个测试包，用来记录 HTTP 发给处理的请求，它可以帮助我们捕捉响应。所以这里我们将会启动一个新的请求，这个请求带着一个 id，用来测试 FindById 处理。我们将会模仿服务调用，但是确保他仅仅想要我们发送给路径的 id，我们有一个结构体，也会被返回。

```go
func (suite *TicketHandlerTestSuite) TestFindTicketById() {
    t := &ticket.Ticket{
        Creator: "Joel",
    }
    suite.ticketService.EXPECT().FindTicketById("test").Return(t, nil)

    vars := map[string]string{
        "id": "test",
    }

    r, _ := http.NewRequest("GET", "/tickets/test", nil)
    r = mux.SetURLVars(r, vars)

    w := httptest.NewRecorder()
    suite.underTest.GetById(w, r)

    response := w.Result()
    suite.Equal("200 OK", response.Status)

    defer response.Body.Close()
    result := new(ticket.Ticket)
    json.NewDecoder(response.Body).Decode(result)

    suite.Equal("Joel", result.Creator)
}
```

再次，我们可以跑通其他的测试案例，达到完整覆盖。
使用以下命令运行你的测试：

```
go test ./..
```

## 结论

模仿测试仅仅是测试的一种方法，但是也是最重要的。它会测试你代码的业务逻辑，以及测试各种元素是如何交互的。最后，你将会发现，这是你拥有的最健壮的测试代码，而且它只有在业务代码被改变或者是 bug 被发现的时候才会改变。

源代码可以在[这里](https://github.com/Holmes89/hex-example/tree/testing) 找到

---

via: http://www.joeldholmes.com/post/go-mock-testing/

作者：[Joel Holmes](http://www.joeldholmes.com/)
译者：[yangshuting](https://github.com/yangshuting)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

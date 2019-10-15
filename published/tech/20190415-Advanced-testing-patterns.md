首发于：https://studygolang.com/articles/23980

# Go 中的进阶测试模式

![image](https://raw.githubusercontent.com/studygolang/gctt-images/master/advanced-testing-patterns/heading.jpg)

Go 使编写测试非常简单。实际上，测试工具是内置在标准工具链里的，你可以简单地运行 `go test` 来运行你的测试，无需安装任何额外的依赖或任何别的东西。测试包是标准库的一部分，我很高兴地看到它的使用范围非常广泛。

当你在使用 Go 编写服务实现时，希望你的测试覆盖率随着时间的推移而增长。随着测试范围的扩大，测试运行时间也会变长。你希望用服务集成及集成测试来测试服务的重要部分。你发现在某些情况下，集成测试和各种公共服务的耦合对 CI 和开发产生限制。

## 集成测试

我是集成测试的忠实信徒。有人可能无法直接看到它的好处，但对于 LTS（长期支持）的版本，进行集成测试是一个很好的主意，因为你显然想要随着时间的推移升级你的服务。如果你要从 MySQL 5.7 切换到 8.0 （甚至是换成 PostgreSQL），你需要合理地确保你的服务依然正常工作，然后你可以检测问题并根据需要对实现进行更新。

集成测试最近对我有用的一个例子是检测到 MySQL 保留字增加的情况：我有一个数据库部署里用到了 `rank` 字段。这个词在 MySQL 5.7 及之前是可以使用的，但在 MySQL 8.0 里它变成了一个保留字。集成测试捕获了这个问题，而模拟（mock）则无法做到这一点。

> RANK ®; 在 8.0.2 版增加 (为保留字) 见 [MySQL 8.0 的关键词和保留字](https://dev.mysql.com/doc/refman/8.0/en/keywords.html)

模拟是单元测试的一种扩展，而由于集成测试可能意味着高昂的成本，今天做集成测试比过去容易得多。随着 Docker 的不断发展，并有了像 [Drone CI](https://drone.io/) 这样 Docker-first 的 CI，我们可以在 CI 测试套件里声明我们的服务。让我们看一下我定义的 MySQL 服务：

```yml
services:
- name: crust-db
  image: percona:8.0
  ports:
	- 3306
  environment:
	MYSQL_ROOT_PASSWORD: bRxJ37sJ6Qu4
	MYSQL_DATABASE: crust
	MYSQL_USER: crust
	MYSQL_PASSWORD: crust
```

这基本上就是随我们的测试和构建一起开启数据库所需的全部。虽然在过去，这可能意味着你需要一个一直在线的数据库实例，你需要在某处进行管理，而今天大门已经打开，基本上你可以在你所用的 CI 框架里声明服务的一切所需。

> ["Go 以及集成测试: 使用 Drone CI #golang" via @TitPetric](http://twitter.com/intent/tweet?url=https%3a%2f%2fscene-si.org%2f2019%2f04%2f15%2fnext-level-go-testing%2f&text=%22Go%20and%20integration%20tests%3a%20simple%20with%20Drone%20CI%20%23golang%22&via=TitPetric)

我有点跑题了，但这里的学问是 - 如果你可以避免模拟一些东西，尤其是在你掌控下的服务，一定要考虑编写集成测试。你无需借助使用 go-get 获取的像 [gomock](https://github.com/golang/mock) 或 [moq](https://github.com/matryer/moq) 这样的项目。模拟一切是不明智的（例如，`net.Conn` 不需要模拟，它足够简单，可以在你的测试中创建你自己的 client/server，它将存在于内存中）。

实际上，在集成测试和模拟之间也有中间立场，你可以编写像 Redis 这样的简单外部服务的 fake 实现，但你仍然不能捕捉到真实服务的所有细微之处。基本上，只满足你用到的简单接口大大降低了实现面（implementation surface），这就只需实现你用到的 API 子集的行为。

## 测试范围（testing surface area）

我正在开发一个项目，目前有 53 个测试文件，其中 28 个是需要外部服务（例如上述的数据库）的集成测试。你可能并不总是处理完整的环境，或者可能只对在项目中分散的一小部分测试感兴趣，并且你希望能够运行这些（且只运行这些）。

查看 `testing` 包的 API 面（API surface），我们注意到有一个 [Short()](https://golang.org/pkg/testing/#Short) 函数可用，它在运行 Go test 时对 `-test.short` 起作用。这使得我们在想运行测试的某个子集时可以跳过一些测试：

```go
func TestTimeConsuming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	...
}
```

从纸面上看，这意味着你在以 short 模式运行时可以跳过集成测试。但即使从上面的例子也可以看出，当测试持续时间是个重要因素时可以用来跳过某些测试，这才是动机 —— 实际上，这个**应该仅适用于基准测试**。

那么，当考虑到你需要显式地以 `-bench` 参数启用基准测试时，你可能会琢磨一个基准与另一个基准测试能否比较快慢。Go 已经很聪明，它默认限制了每个基准运行的时间，而是否要修改这个配置，以及是否想同时使用 short 模式和基准，都由你来决定 - 对我来说，两个选项同时使用毫无意义。

事实上，short 测试标记不应该用来跳过集成测试。它的目的是加速构建过程，但是代码判断或人为地判断哪个测试应该是 short 或是 long 让人望而却步。强调：要么运行所有的基准测试要么不运行。随着测试集的增长，short 测试标记无法给我们所需的灵活性，所以我们需要一种更具声明性的方式来界定我们需要运行哪些类型的测试。

## 更好的方式

现在，传统观点会说“运行所有的测试”。作为真正了解人们如何处理问题，提出问题以及确立实践准则的工程师中的一个 —— 现在，这可以帮助你找到一个更好的解决办法，解决并不只是你遇到的问题。

在 2016 年，以及 2016 年晚些时候，Peter Bourgon 写了两篇极好的长篇幅文章，这些文章对需要实现实际服务并超出基本实现的人来说，是参考书一样的存在：

- [Go best practices for production environments (2014)](https://peter.bourgon.org/go-in-production/)
- [Go best practices, six years in (2016)](https://peter.bourgon.org/go-best-practices-2016/)

在 2014 年的文章里，Peter 建议使用构建标记来引入有价值的测试习惯：

> 包测试主要针对单元测试，但对集成测试来说，事情有点棘手。启动外部服务的过程通常依赖于你的集成环境，不过我们确实找到了一个针对它们进行集成测试的好习惯。写一个 integration_test.go 并给它一个 integration 的构建标记。为服务地址以及连接字符串等定义（全局的）flag，并在测试中使用它们。

事实上，Peter 建议使用 Go 的构建标记来标识集成测试。如果你需要一个单独的环境来运行这些测试，你只要使用 `-tags=intergration` 作为 Go test 的参数。

这完全合乎情理 —— 尽管我在的这个项目的集成测试需要花费一分钟左右，但我知道在有的项目里需要几个小时。这些项目可能有很特殊的专用测试设置，这样你也可以不测试这些服务的配置 —— 它们只在测试环境中使用。

我很想知道他的观点在 2014 到 2016 年是否发生了什么变化。如果有的话，作者会深入研究各种非标准测试包如何成为他们的 DSL（领域特定语言）。但是经验是一位好老师，他并没有对一个 `http.Client` 进行测试，并指出你不想测试请求进入的 HTTP transport 或正在写文件的磁盘上的路径。

在单元测试中你应该专注于业务逻辑，并且通过集成测试，您将验证集成服务的功能，而不是标准库或第三方软件包如何实现集成。

> ["Go 测试: 哪个适合你 - 单元测试还是集成测试? #golang" via @TitPetric](http://twitter.com/intent/tweet?url=https%3a%2f%2fscene-si.org%2f2019%2f04%2f15%2fnext-level-go-testing%2f&text=%22Go%20testing%3a%20which%20one%20are%20right%20for%20you%20-%20unit%20or%20integration%20tests%3f%20%23golang%22&via=TitPetric)

## 边界情况

将你的应用程序与第三方服务集成是很常见的，由于 API 弃用是可能发生的，所以集成测试可能还需要验证应用程序的响应是否仍然有意义。 因此，Peter 的文章需要一点改进。

你不能总是依赖你正在使用的 API；它会在未来几年都保持原样吗？没有人希望你创建一堆 GitHub 用户和组织来测试你的 webhook 端点和集成，但这并不意味着你不会偶尔需要这样做。

一个最近的例子是 [larger deprecation of Bitbucket APIs due to GDPR](https://developer.atlassian.com/cloud/bitbucket/bbc-gdpr-api-migration-guide/). 这篇弃用通知是在大约一年前宣布的，
从 10 月开始，并计划在 2019 年 4 月底废弃各种 API 及返回的数据，可能会对现有的各种 CI 集成造成严重破坏。

考虑到这一点，我这样扩展了 Peter 的建议：

- `// +build unit` - 不需要任何服务的测试,
- `// +build integration` - 一个强制标记来测试我们自己的服务，
- `// +build external` - 针对第三方和公共服务进行测试，
- `// +build integration,external` - 针对我们自己的服务以及公共服务进行测试,
- `// +build !integration,external` - 专门针对公共服务进行测试,
- `// +build unit integration` - 不依赖服务，提供集成功能

我们的测试通常属于单元测试、集成测试或外部测试的某一类，或者是它们的某种组合。我们肯定希望在 CI 任务中跳过 `external` 测试，原因显而易见，但如果我们正在考虑调试开发中的一些相关问题，它们是非常有价值的。我们经常需要定位到具体包中具体的测试，因此运行类似下面的内容是有意义的：

```sh
go test --tags="integration external" ./messaging/webhooks/...
```

根据你的构建标记，这可能会运行你的代码库某个子集里面的所有集成和外部测试，跳过单元测试，或者它可能只运行那个既是集成测试也是外部测试的测试。 无论哪种方式，你都专注于包实现，尤其是该包中与提供的标记匹配的所有测试的子集。

> ["Go 测试: 按需运行集成测试的实用方法 #golang" via @TitPetric](http://twitter.com/intent/tweet?url=https%3a%2f%2fscene-si.org%2f2019%2f04%2f15%2fnext-level-go-testing%2f&text=%22Go%20testing%3a%20A%20practical%20method%20to%20run%20integration%20tests%20on%20demand%20%23golang%22&via=TitPetric)

对于 CI 任务，范围确定为：

```sh
go test --tags="unit integration" ./...
```

这样，你可以完整地测试所有集成测试，以及完整的包范围。 我们将跳过可能导致我们的 CI 构建失败的 `external` 和 `integration AND external` 测试，不让它们成为构建的问题。可能每月有那么一天，GitHub 或 Bitbucket 是坏的，我们只能一直看着它们的状态页面。

因此，基本上，除了将某些测试标记为 `integration` 之外，我们还希望将其他标记为 `unit` 和 `external` ，以便我们可以根据需要在开发环境中跳过它们。 没有人喜欢运行完整的测试集，并且发现它仅仅因为 GitHub 出问题而失败。 而具有开发和调试目的的选项是非常宝贵的。

## 对测试进行测试

在重构测试时，你经常只会在运行测试时才发现，有些符号或其他东西已经不存在了，导致你的测试无法编译。这个问题的一个好的解决办法是仅仅针对测试文件的编译步骤进行测试。有一些东西可以发挥作用：

1. 可以通过给 `go test` 填写 `-run` 参数来跳过测试。你可以运行 `go test -run=^$ ./...`，它将有效地编译你的完整测试集并跳过所有测试。这对于运行时间较长的 CI 任务非常有用，因为它实际上是一个编译时的检查，确保所有测试都是可运行的。但是，这仍然会运行你的 `TestMain` 函数。

2. Go 1.10 引入了 `-failfast`  标志。如果你的某些测试失败而你有一个非常大的测试集，那么在错误/失败之间会有很多输出，在其他测试完成之前，以及通知你失败之前也会有很多。使用此选项，你可以对这个问题稍做优化，代价是同一测试集中在之后运行的测试中可能还会有失败的。这是测试所有内容和报告所有错误，或仅在发现第一个错误之前进行测试的区别。

3. `-failfast` 标志对 `./ ...` 没有任何作用，例如，如果其中一个包由于编译错误而失败，它将继续针对剩余的检测到的包进行测试。

这些基本上是围绕 [golang/go#15535](https://github.com/golang/go/issues/15513) 的问题，实际上这意味着我们无法像使用 `go build` 一样只针对测试的编译进行测试。

＞ [Go 测试：编译时检查你的测试而无需运行它们 #protip #golang via @TitPetric](http://twitter.com/intent/tweet?url=https%3a%2f%2fscene-si.org%2f2019%2f04%2f15%2fnext-level-go-testing%2f&text=%22Go%20testing%3a%20compile-time%20check%20your%20tests%20without%20running%20them%20%23protip%20%23golang%22&via=TitPetric)

## 公有以及私有测试 API

理想情况下，你将对你的包进行黑盒测试。这意味着若你的包起名为 `store`，你的测试就会在 `store_test` 包中。这可能为你解决了这样一个依赖问题，即 url 包依赖 http 包，而反过来也存在依赖。使用 `url_test` 和 `http_test` 包 [解决了这个问题](https://talks.golang.org/2014/testing.slide#21)。

此外，有一些适用于任何代码库的准则：

- 如果你在做内部测试，给你的文件加 `_internal_test.go` 后缀，
- 如果你是在做黑盒测试，你的文件应该只有 `_test.go` 后缀。

特别地，对于名为 store 的东西，你应该有：

- `store.go` —— 主包（`package store`）
- `store_test.go`—— 黑盒测试（`package store_test`）
- `store_internal_test.go` —— 内部测试（`package store`）

有一些关于如何使用这些准则的例子。在 Michael Hashimoto 的一次题为高级的 Go 测试的演讲里，他主张[测试作为公共 API](https://about.sourcegraph.com/go/advanced-testing-in-go#testing-as-a-public-api)。

- > Hashimoto 公司较新的项目采取了使用 “testing.go” 或 “testing_*.go” 文件的实践。
- > 这些文件本身是包的一部分（与普通的测试文件不同）。这些都是为提供模拟，测试治理，帮助方法等而导出的 API。
- > 允许别的包使用我们的包进行测试，且无需为了在一个有意义的测试中使用我们的包而对所需组件进行彻底改造。

对于此我有一个问题（也许不是一个特别相关的问题），对公共 API 的任何修改都需要有某种兼容性保证。尽管这本身是可以接受的，但它并不是一个明确的规范。在大多数情况下，将这些测试函数限定在当前项目测试的范围内是更容易接受的。

我会只是将这些函数添加在 `store_internal_test.go` 中 —— 在 `*_test.go` 中定义的任意公共标识符在包中依然是可用的，不过只能在测试中访问。当你的应用被编译时，不会去拉取你在测试文件中声明的任何东西。当你改变主意，需要将其中一些变为公共的 —— 你只需将相关代码移动到 `testing.go` 文件中，而不需要修改任意一行测试代码。

> [“Go 测试：你是否应该用公共 API 提供测试所需设施？#golang” via @TitPetric](http://twitter.com/intent/tweet?url=https%3a%2f%2fscene-si.org%2f2019%2f04%2f15%2fnext-level-go-testing%2f&text=%22Go%20testing%3a%20Should%20you%20have%20a%20public%20API%20to%20provide%20testing%20utilities%3f%20%23golang%22&via=TitPetric)

以上建议的原则也适用于从包中对外暴露一些私有符号，以便在黑盒测试中使用。我似乎无法找到一个强有力的例子来证明这种方法的合理性，除开上面讲到的循环引用的问题，从你的包中对外暴露内部的东西然后只用于你的测试是可以做到的。但如果沿这条路走下去，你实际上是将内部测试和黑盒测试混在一起，我建议你不要这么做。内部的东西会变化，导致你的测试也变得更脆弱。

在这片荒野之地很少有例子，不过还是有一些：

- [API Testing - Swagger](https://goswagger.io/faq/faq_testing.html) —— 为黑盒测试暴露/包装私有函数提供的功能，
- [Separate _test package](https://segment.com/blog/5-advanced-testing-techniques-in-go/) —— 通过额外的文件避免导出/模拟（并没有链接示例），
- [Export unexported method for test](https://medium.com/@robiplus/golang-trick-export-for-test-aa16cbd7b8cd) —— 为测试导出私有函数的一个不好的例子，

实际上，在大多数情况下人们可以编写内部测试来实现相同的目标。 我并不是在提倡，尤其是这样的 `_internal_test.go` 文件应该将内部暴露给黑盒测试，但是我看到了使用它们来提供有一天可能成为公共包 API 的效用实体，是有意义的。 这仍然是太大的一步，但这一切都取决于你的需求。 如果你不希望在给定日期或给定版本之前发布某部分 API，可以采取这种方式，对于每个 API，可以将其实现为公共包 API，而无需真正对外发布供测试之外使用。

## 既然我已经把你带到了这里……

如果你买我的一本书会很棒：

- [API Foundations in Go](https://leanpub.com/api-foundations)
- [12 Factor Apps with Docker and Go](https://leanpub.com/12fa-docker-golang)

我保证如果买了你会学到更多。购买副本支持我写更多关于类似主题的内容。 谢谢你，请买我的书。

如果想预定我的顾问/自由职业服务时间，请随时[给我发邮件](black@scene-si.org)。我对 API，Go，Docekr，VueJS 以及扩展服务[等等](https://scene-si.org/about)都很在行。

---

via: https://scene-si.org/2019/04/15/next-level-go-testing/

作者：[Tit Petric](https://scene-si.org/about)
译者：[krystollia](https://github.com/krystollia)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

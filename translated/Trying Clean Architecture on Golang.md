# 在Golang中尝试简洁架构
>（独立性，可测试性的和简洁性）

在阅读了 Bob 叔叔的 Clean Architecture Concept之后，我尝试在 Golang 中实现它。我们公司也有使用相似的架构，[Kurio - App Berita Indonesia](https://kurio.co.id/)， 但是结构有点不同。并不是太不同， 相同的概念，但是文件夹结构不同。

你可以在这里找到一个示例项目[https://github.com/bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch)，这是一个CRUD管理示例文章
![](https://cdn-images-1.medium.com/max/1600/1*CyteJRpIHC-DFE23UtlZfQ.png)

* 免责声明：

  我不推荐这里使用的任何库或框架，你可以使用你自己的或者第三方具有相同功能的任何框架来替换。

## 基础

在设计简洁架构之前我们需要了解如下约束：

1. 独立于框架。该架构不会依赖于某些功能强大的软件库存在。这可以让你使用这样的框架作为工具，而不是让你的系统陷入到框架的限制的约束中。

2. 可测试性。业务规则可以在没有 UI， 数据库，Web 服务或其他外部元素的情况下进行测试。

3. 独立于 UI。在无需改变系统的其他部分情况下， UI 可以轻松的改变。例如，在没有改变业务规则的情况下，Web UI 可以替换为控制台 UI。

4. 独立于数据库。你可以用 Mongo， BigTable， CouchDB 或者其他数据库来替换 Oracle 或 SQL Server，你的业务规则不要绑定到数据库。
5. 独立于外部媒介。 实际上，你的业务规则可以简单到根本不去了解外部世界。

更多详见：[ https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html]( https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html)

所以， 基于这个约束，每一层都必须是独立的和可测试的。

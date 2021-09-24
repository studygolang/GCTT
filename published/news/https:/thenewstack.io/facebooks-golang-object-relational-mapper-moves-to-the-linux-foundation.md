首发于：https://studygolang.com/articles/35233

# Facebook的 Go ORM：ent 移动到了 Linux 基金会

![](https://cdn.thenewstack.io/media/2021/09/7e01e106-go-ent.png)

"[Ent](https://entgo.io/)" 是最初由 Facebook 创建并于 2019 年开源的 Go 实体框架，现已加入 [Linux基金会](https://training.linuxfoundation.org/training/course-catalog/?utm_content=inline-mention)。 Ent 帮助开发人员处理复杂的后端应用程序，在这些应用程序中他们可能需要处理大量实体类型以及它们之间的关系。

[Ariel Mashraki](https://il.linkedin.com/in/ariel-mashraki-435a1250)，Ent 的创建者和主要维护者说，当他在在致力于为世界带来互联网连接的 [Facebook Connectivity](https://www.facebook.com/connectivity) 团队中，该团队需要一个对象关系映射 (ORM) 工具，该工具可以处理映射他们正在处理的网络拓扑，但找不到合适的工具，于是他们创造了 Ent。经过几年的开源并看到社区越来越多的采用和参与，Facebook 决定将 Ent 项目转移到 Linux 基金会的管理之下，在那里它将为未来找到一个供应商中立的家。

通过转向 Linux 基金会，Mashraki 说这让他可以离开 Facebook 共同创立数据图初创公司 [Ariga](https://ariga.io/)，同时仍在 Ent 工作，该项目正在寻求更多希望参与的其他公司的参与。

传统的 ORM 将面向对象编程中的对象映射到数据库。 Mashraki 解释说，Ent 提供了这种基本的 ORM 功能，同时还对一些附加功能进行了分层，这就是使其成为“实体框架”的原因。根据一篇[博客文章](https://www.linuxfoundation.org/press-release/ent-joins-the-linux-foundation/)，Ent“使用图概念对应用程序的模式进行建模，并采用先进的代码生成技术来创建类型安全、高效的代码，与其他方法相比，这大大简化了对数据库的处理。”

将其分解， Mashraki 进一步解释说，Ent 的创建考虑了三个特定的设计原则。首先，它使用图概念，例如节点和边，对数据进行建模和查询，这意味着数据库可以是关系型的，也可以是基于图的。其次，Ent 的代码生成引擎会分析应用程序架构并生成类型安全的显式 API，供开发人员与数据库（例如 MySQL、PostgreSQL 或 AWS Neptune）交互。最后，Ent 在 Go 代码中表达了与实体相关的所有逻辑（包括授权规则和副作用），这提供了 Ent 的内置支持，可以直接从模式定义中自动生成 GraphQL、REST 或 gRPC 服务器。 Mashraki 说，所有这些不仅有助于处理这些大型数据集，而且还提供了更好的开发人员体验。

“这意味着，对于开发人员在其 Ent 架构中定义的每个实体，显式、类型-为开发人员生成安全代码以有效地与其数据交互。类型安全代码提供了卓越的开发体验，因为 IDE 非常了解 API 并且可以提供非常准确的代码完成建议，此外，通过这种方法，在编译过程中可以捕获许多类别的错误，这意味着更快的反馈循环和更高质量的软件，” Mashraki 在一封电子邮件中写道。 “此外，网络拓扑在图概念中更容易建模和查询。尝试维护以关系术语遍历数百种实体的代码和查询太容易出错且速度缓慢，而 Ent 正是从这种痛苦中直接创建的。”

Ent 都用 Go 编写并用 Go 生成代码， Mashraki 表示该框架在云原生社区找到了一个天然的家，许多云原生计算基金会 (CNCF) 项目都在使用该框架。虽然 Go 是一个最初的选择，但 Mashraki 说他们也在考虑在不久的将来添加其他语言，例如 TypeScript，因为它在前端开发人员中很受欢迎。

至于 Go，Mashrak 谈到即将[添加的泛型](https://go.dev/blog/generics-proposal)此举是“将减少生成代码的数量，并使创建通用扩展而不使用代码生成成为可能。我们已经在试验它。”

展望未来，Mashraki 说 Ent 项目目前有两个主要计划正在进行中。首先是一个迁移 API，它“旨在与 Kubernetes 和 Terraform 等云原生技术无缝集成”。接下来，Ent 将获得一个新的查询引擎，该引擎将允许在同一个 Ent 架构上定义多种存储类型，例如，这将允许开发人员使用相同的 Ent 客户端查询 SQL、blob 和文档数据库。

“我们邀请更多公司加入这项努力并成为其中的一部分，” Mashraki 补充道。

---

via: https://thenewstack.io/facebooks-golang-object-relational-mapper-moves-to-the-linux-foundation/

作者：[Mike Melanson](https://thenewstack.io/author/mike-melanson/)
译者：[lavaicer](https://github.com/lavaicer)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

# Go 2 范型：目前的合约很无聊（不过这很棒）

> August 19, 2019

7月底，Go 团队推出了 [Go 2 范型设计的修订版](https://go.googlesource.com/proposal/+/master/design/go2draft-contracts.md)，完全重写了合约，其中明确规定了范型方法等接受的约束（令我惊讶的是，修订版设计尚未链接到 [wiki 反馈页面](https://github.com/golang/go/wiki/Go2GenericsFeedback)，但据我所知，该设计完全是官方的）。自从修订版设计问世，我一直在考虑对它的看法，简而言之，我并不想说什么，换一种说法，是因为新的合约设计很无聊。

你会对[初版设计](https://go.googlesource.com/proposal/+/master/design/go2draft.md)中的合约有很多评论，但是他们确实并不无聊。尽管合约复用了现有的 Go 语法（用于方法体），它们与 Go 中其他任何内容都存在明显差异，[一种我觉得非常聪明的差异](https://utcc.utoronto.ca/~cks/space/blog/programming/Go2ContractsTooClever)。为了尽量减少新语法，初版合约设计通常[间接的描述事情](https://utcc.utoronto.ca/~cks/space/blog/programming/Go2ContractsMoreReadable)，导致了一些问题，例如有许多方式可以表示同一种约束（尽管 [Go 可能要求最少的合约](https://utcc.utoronto.ca/~cks/space/blog/programming/Go2RequireMinimalContracts)）。

新版合约设计消除了这些问题。合约使用了新语法，尽管它复用了普通 Go 语言中的许多语法元素，新语法基本上是最少的，并且直接表述了它的约束。之前只能通过暗示来表达的一些棘手的事情，现在可以通过字面含义表达，例如，将一个类型限定为基于整型的某种类型。到目前为止，那种情况下的类型约束仍然很难表示，Go 团队引入了一个预先声明的称作 *comparable* 的合约，而非尝试变得更聪明。正如提案本身所说：

> 总是会有数量有限的预声明类型，以及这些类型支持的数量有限的运算符。将来语言的修改不会从根本上改变这些要素，因此这种方式将持续有效。

对于这个问题，这是一种标准的 「Go 语言」 方法，这让我印象深刻。它既不聪明，也不让人兴奋，但它有效（并且清晰明了）。

另一个决定 - 范型函数中的所有方法调用都将是指针类型的方法调用，并不优雅，但却解决了一个真实潜在的混乱。关于可寻址的值和指针值的方法的规则[有些令人困惑且晦涩难懂](https://utcc.utoronto.ca/~cks/space/blog/programming/GoAddressableValues)，因此 Go 团队决定让它们变得无聊，而非尝试变得聪明。无论这让人觉得有多么不优雅，当我需要处理范型时，它可能会使我的生活更轻松。

（该设计还允许你指定特定的方法必须是指针方法，尽管哪种类型满足生成的合约可能有点太聪明，也太晦涩了。）

修订版的合约设计几乎保留了初版设计中[我喜欢的所有内容](https://utcc.utoronto.ca/~cks/space/blog/programming/Go2ContractsLike)。它放弃了一件事，我并不希望如此，即你不再能够要求类型具有某些特定字段（隐式要求它们是结构类型）。在实践中，内联的 getter 和 setter 方法也可能同样有效，[即便我不喜欢它们](https://utcc.utoronto.ca/~cks/space/blog/programming/GettersSettersDislike)，并且总有一些使用场景是不清晰的。

我不知道目前的 Go 2 范型以及合约设计是否绝对正确，但我现在确实觉得，和初版方案给人的感觉相对比，这并不是个错误。我确实希望人们尝试使用这种设计来写范型代码，如果它以某种试验形式在 [Go 的某些版本](https://utcc.utoronto.ca/~cks/space/blog/programming/GoAppearanceOfChanges)中提供，或者以其他方式可用，因为我怀疑这是我们找到任何剩余的大致的边界和痛点的主要方式。

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/Go2ContractsNowBoring

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

首发于：https://studygolang.com/articles/22243

# 致 Go 团队的一封公开信 —— 关于 try

> polaris 注：目前关于 try 的提案被否决了，具体见：https://studygolang.com/articles/22043

*“一旦语言变得足够复杂，在其中编程更像是从无限多的特性海洋中划出一个子集，其中大部分都是我们永远不会学到的。一旦语言像是有无限多的特性，为其添加更多特性的成本就不再明显。”* - *[Mark Miler](https://medium.com/@erights/the-tragedy-of-the-common-lisp-why-large-languages-explode-4e83096239b9)*

新的关于 `try` 的提议是对语言的补充，它引入了第二种错误处理机制。它是根据 [2018 年 Go 语言调查](https://blog.golang.org/survey2018-results) 收集的数据和对 [Go 2 提案流程](https://blog.golang.org/go2-here-we-come) 中提交的提案的审查而引入的。Go 团队从这些收集到的数据中得出的结论是，Go 开发人员希望更好地错误处理机制。

如果你看一下 Go 调查提供的这个图表，你会看到，5% 的开发人员将错误处理作为 Go 开发中的最大挑战。

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/an-open-letter-to-the-go-team-about-try/102_figure1.png)

基于此，我认为对错误处理的抱怨可能被夸大了，这些改变并不是大部分 Go 开发人员想要或者需要的。偏见可能源于这样的事实，对于 Go 的抱怨实在太少了，因此错误处理会出现在任何调查中。如果你仔细观察，错误处理甚至不是开发人员面临的前 3 大挑战，它是第 5 位。

我认为 *Go 2 调查* 的数据是有偏差的，因为只有那些有问题的人提交了提案，这遗漏了其他所有觉得 Go 的错误处理并不需要改进的开发人员。尽管调查数据是一个很广泛的数据集，但在消除偏差方面做的并不好。

Rob Pike 去年在 Go 悉尼大会上发表了演讲，他谈到了 [Go 2 的变化](https://www.youtube.com/watch?v=RIvL2ONhFBI&feature=youtu.be&t=440)。在演讲中，他表态认为在代码中对 `if err != nil` 的使用远不如少数声音声称的那样普遍。Marcel van Lohuizen 做了类似的研究，发现对于栈底部的代码（更靠近 `main` 方法）,`if err != nil` 并不常见，但是随着方法更靠近栈顶（更靠近网络或操作系统），它确实变得更常见。“管道代码”往往需要更多的检查，但是[良好的设计](https://blog.golang.org/errors-are-values) 是可以消除这些检查的，我想很多使用 Go 编写过一段时间代码的人都会同意这个评价。

我相信，通过调查和提案，Go 社区的部分开发者确实想要改进版的错误处理，但是数据并不支持 `try` 作为 `if err != nil` 的替代方案。我不相信提案的 `try` 解决方案是明显正确的设计，因为它引入了两种方法去做同一件事，就是关于仅仅将错误返回给调用者这样的简单情况。由于这项新机制将会导致代码库中的严重不一致和团队的分歧，并且给推行一致性规范的产品负责人创造了一个不可能完成的任务，因此，需要放慢速度并收集更多的数据。

这是一个严肃的改变，貌似还没一致努力去理解这 5% 的开发人员表示想要改进版的错误处理时所表达的意思，就一直被推进。我恳请 Go 团队在将 `try` 错误处理机制在任何版本中进行实验前，重新评估正在使用的数据集。在 Go 的历史中，一旦将特性实验性地引入，它们都从未被回滚。我请求用更多的时间来收集数据，依此判断这项更改会对代码库造成的实际影响。最后，Go 2 特性的优先级应该被重新评估，Go 2 提案数据不应该用来作为判断优先级的依据。

Sincerely,
William kennedy

---

via: https://www.ardanlabs.com/blog/2019/07/an-open-letter-to-the-go-team-about-try.html

作者：[William Kennedy](https://www.ardanlabs.com/training)
译者：[DoubleLuck](https://github.com/DoubleLuck)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

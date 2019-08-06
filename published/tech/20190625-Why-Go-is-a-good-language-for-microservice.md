首发于：https://studygolang.com/articles/22110

# 为什么 Go 适合微服务

![Go Crew at SafetyCulture](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-is-good-microservice/1_K95qdWukzkEadEpxqiNdEQ.jpeg)

去年早些时候，我们决定改用 Go(Golang) 作为我们（[SafetyCulture](https://safetyculture.com/)）开发微服务的选择。在这之前，我们的微服务使用 Node.js(CoffeeScript, Javascript 和 TypeScript 的混合 ) 编写。下来我将分享我们更改的原因。

## 1. 静态类型

Go 是一种静态类型语言，这意味着编译器可以为您做更多的工作。人们倾向于强调这一点的重要性。

### 生产事故的故事

去年，在为我们的一个核心微服务修复 bug 时，我造成了一个生产事故（用 Node 编写），因为我在函数中添加了一个额外的参数，忘记在调用函数时传递正确的参数。

```typescript
// 函数定义
function saveDocument({id, oldDocument, newDocument}) {}

// 函数调用
saveDocument({
	id: "xyz",
	oldDoc: "blah blah",
	newDocument: "new doc"
})
```

我的函数期望传入 oldDocument 类型的参数，但是我传了 oldDoc 类型。最终， oldDocument 应该需要写入 Kafaka, 以供下游其他微服务使用。所有测试都通过了，我将其作为产品发布，仅发布 3 天，我们就意识到问题。这个问题花费了我们两个全职工程师工作了整整 3 天才修复。

当然，您可以使用 TypeScript 并解决这个问题（希望如此），但为什么不选择一种语言来帮助您在编译时捕获问题？在当今的世界，团队们正在开发 20 种不同的微服务，你需要记忆大量内容，编译过程提供一些帮助是有好处的。我在生产 ROR 应用程序之前也看到了类似的问题，我可以将 string 类型更改为 int, 数组，甚至任意类型都没有类型，直到出现问题或者伤害了你的用户。[Airbnb 表示，使用类型可以预防 38 ％ 的漏洞](https://www.reddit.com/r/typescript/comments/aofcik/38_of_bugs_at_airbnb_could_have_been_prevented_by/)。

## 2. 可读性

***清晰好过聪明（Clear is better than clever）***

Go 的原作者在设计时非常注重可读性。Go 是一种易于学习的语言，并且针对大多数问题，都有一种惯用方式。而在 NodeJS ，您可以在同一代码库中遇到回调，约定和异步 / 等待（不同方式实现同一件事）。人们抱怨 Go 很冗长，但我们认为这对 SafetyCulture 来说是一件好事。当然，你可以在 Python/Node 中编写复杂的，很少有人能够理解的 1-liners（一行代码实现），但这并不能形成一个可读的代码库，不便他人维护。

## 3. 标准库

Go 的标准库设计的很棒，你可以用它做很多事情。维护巨大的外部依赖树是一项挑战。您添加到应用程序的每个依赖项都会带来许多其他依赖项，需要对其进行性能和安全性审计。在某些情况下，它甚至会妨碍您对应用程序的更改，，比如[leftpad incident](https://www.davidhaney.io/npm-left-pad-have-we-forgotten-how-to-program/)。

***[小复制好于小依赖 (A little copying is better than a little dependency)](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=9m28s)***

Go 的作者和社区希望开发人员注意他们的应用程序的依赖关系。

## 4. 性能

Go 对并发编程有很好的支持，可以很好地利用多个内核。这对于繁重的计算以及网络 I/O 和磁盘搜索等都是有益的。想了解更多关于优化内存和 CPU 使用，可以看看项目：[Why we moved our GraphQL server to Go.](https://medium.com/safetycultureengineering/why-we-moved-our-graphql-server-from-node-js-to-golang-645b00571535)

## 5. 代码格式化

***[Gofmt 的风格没有人喜欢 , 但是 Gofmt 是每个人的最爱 (Gofmt ’ s style is no one ’ s favorite, yet Gofmt is everyone ’ s favorite)](https://www.youtube.com/watch?v=PAAkCSZUG1c&t=8m43s)***

Go 有固定的代码格式和工具（gofmt），帮你自动化格式化代码，因此你不需要花费时间讨论设置 ESLint 或 Prettier。

## 结论

相比其他动态编程语言，Go 提供了很多优良特性。但是，没有一个语言是完美的。比如：Go 中的并发代码很容易编写，但新手却并不能很好上手使用。这是因为，数据经常出现竞争状态，代码不好调试。这一点上，[Rust](https://www.rust-lang.org/) 做的更好一些，因为它提供了权衡机制防止数据出现竞争状态，但你必须了解更复杂的系统。

参考连接：[https://go-proverbs.github.io/](https://go-proverbs.github.io/)

---

via: <https://medium.com/safetycultureengineering/why-go-is-a-good-language-for-microservices-b4fc6a5a532c>

作者：[Pawan Rawal](https://medium.com/@pawan_rawal)
译者：[TomatoAres](https://github.com/TomatoAres)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

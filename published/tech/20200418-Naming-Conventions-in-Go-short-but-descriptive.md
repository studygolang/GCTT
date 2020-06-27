首发于：https://studygolang.com/articles/28975

# Go 语言中命名规范——如何简短却更具描述性

> 在计算机科学与技术中，有两件事情最难，第一是缓存无效，第二就是给一些东西命名 —— Phil Karlton

上面的话可不是一个笑话。写代码很容易，但是阅读起来却很痛苦。你是否有想知道一个变量具体指什么或者某个包的具体含义是什么这种类似的经历？这就是为什么我们需要一些规则和约定。

不过，约定虽然能够让我们的生活变得更轻松，但是也容易被高估和滥用。设置一些合理的命名约定和规则非常重要，不过盲目的遵循它也可能带来很多弊端。

在这篇文章里面，我将介绍在[Go](https://golang.org/)中，一些重要的变量命名约定(官方的以及非官方的规则)以及在什么场景下会存在滥用的情况，特别是那些短变量命名的场景。篇幅有限，包和文件的命名以及项目结构命名有关的内容不在本文讨论范围内，他们应该可以单独再写一篇文章。

## Go 官方的书写规范

与其他编程语言类似，Go 有自己的[命名规范](https://golang.org/doc/effective_go.html#names)。此外，命名也具有一些语义的效果，它决定了包外部对他们的可见性。

### 大小写混合

Go 中的约定是使用 `MixedCaps` 或 `mixedCaps` 这种形式（简称为**驼峰命名**），而不是使用下划线来编写多词名称。 如果需要在包外部可见，则其第一个字符应为大写。 如果您不想将其用于其他包，则可以放心地使用 `mixedCaps`。

```go
package awesome
type Awesomeness struct {
}
//Do 方法是一个外部方法，可以被别的包调用
func (a Awesomeness) Do() string {
 return a.doMagic("Awesome")
}
func (a Awesomeness) doMagic(input string) string {
 return input
}
```

如果你尝试在外部使用 `doMagic` 方法，你会得到一个编译错误。

### 接口名称

> 根据命名规则，一种方法的接口，需要在名称后面加上 `-er` 的后缀，或者通过代理名词的方式来进行修饰：`Reader, Writer,Formatter,CloseNotifier` 等 —— Go 官方文档

按照这个规则，`MethodName + er = InterfaceName`。这里最麻烦的是，你的一个接口有多个方法的时候，按照这个方式命名，就不总是很清晰明了。那是否要把结构拆分，变成一个接口对应一个方法呢？我觉得这个取决具体的使用场景了。

### Getters

[官方文档](https://golang.org/doc/effective_go.html#Getters)中提及了，Go 并没有自动支持 setters 与 getters，不过这里并不禁止它，而是有一些特定的规则：

> 自己实现 getters 和 setters 方法并没有什么问题，而且大部分场景是有用处的。不过没有必要也不需要一直把 `Get` 放在 getter 的方法名字上面。(译者注：编写者希望 Getter 方法前面不需要用 Get 再修饰了)

```go
owner := obj.Owner()
if owner != user {
    obj.SetOwner(user)
}
```

这里需要额外提醒一下，如果你的 setter 方法里面没有任何特殊的逻辑，建议直接导出这个属性，并摆脱掉 setter 和 getter 方法。如果你是一个 oop(面向对象)的死忠粉，可能会听起来比较奇怪，但事实并非如此

## Go 非官方的书写规范

一些规则虽然在官方文档里面没有提及，不过却在社区中被普遍使用。

### 简短的变量名

Go 社区推荐使用简短的变量名，不过我认为这个约定遭到了滥用。有些时候，通常会忘记添加一些描述性的部分。一个描述性的名称，可以帮助读者理解他的实际作用，甚至在使用过程中也能快速理解其含义。

> 编写的程序必须能够让人们阅读它，只不过是顺带让机器执行了一下！—— Harold Abelson

- 单个字母标识符：通常只在范围有限的局部变量里面使用。我们都认可不需要通过 `index` 或者 `idx` 来标识自增变量。单字母标识符只推荐在循环范围内使用。

```go
for i := 0; i < len(pods); i++ {
   //
}
...
for _, p := range pods {
  //
}
```

- 简写名称：只要可能，建议速记名称，只要对于第一次阅读该代码的人来说都易于理解。使用范围越广，就越需要描述。

```go
pid // Bad (does it refer to podID or personID or productID?)
spec // good (refers to Specification)
addr // good (refers to Address)
```

### 唯一名称

这种通常都是一些缩写，例如：API，HTTP 等等，或者类似于 ID，DB 这种。按照惯例，我们保留原样:

- `userID` 来替代 `userId`
- `productAPI` 来替代 `productApi`

### 长度限制

在 Go 中没有长度限制，不过避免特别长的内容也是值得推荐的。

## 结论

我总结了 Go 语言中的一些常用的命名规范，在什么时候以及场景应用他们。我也解释了 Go 语言中简短命名背后主要的思想——在简洁和描述之间找到平衡点。

规定是为了引导你，而不是阻碍你。只要合适，您就应该轻松地打破它们，并且仍然可以满足一般目的。

---

via: https://medium.com/better-programming/naming-conventions-in-go-short-but-descriptive-1fa7c6d2f32a

作者：[Dhia Abbassi](https://medium.com/@dhiatn)
译者：[Alex.Jiang](https://github.com/JYSDeveloper)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

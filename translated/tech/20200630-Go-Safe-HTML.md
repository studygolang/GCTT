# Go语言的HTML安全编码

>免责声明：这不是一个官方的谷歌帖子或公告，本文只是我理解的一些可以用的理论方法。

谷歌信息安全小组谷歌发布了 [Go 语言 “safehtml” 包](https://github.com/google/safehtml) 。如果你有兴趣使你的应用程序能够自适应服务器端XSS，那么你可能希望采用它来替代“html/template”。（将你的应用程序中的HTML类库）迁移到 safehtml 会非常简单，因为它只是原始 html/template 标准包的一个强化分支。如果你的应用程序没有大的缺陷，那么将它转换为使用安全版本应该不会太复杂。

这是谷歌内部用来保护产品免受 XSS 攻击的方法。

如果你只想使用它而不想了解其中的原理，可以跳到[结论列表](#checklist)。

## "html/template" 的问题

“html/template” 没有 ”tainting“（污染）的概念，也没有跟踪像 [template.HTML](https://golang.org/pkg/html/template/#HTML) 这样危险的类型。HTML被构造后，只有一行文档说明：

>使用这种类型会带来安全风险：封装的内容应该来自受信任的源，因为它将被逐字逐句包含在模板的输出中。

这不仅缺乏对“为什么”和“如何”使用该类型的解释，而且它的使用也非常普遍，这使得它与文档中一同被提及的所有其他类型（总共七个）一起成为一个非常危险的隐患。

此外，“html/template”还有一些长期存在的问题，在不破坏向后兼容性的情况下，没有一种很好的方法来正确地修复这些问题。潜在的破坏兼容可能性和安全性之间的权衡还不清楚，因此，如果你想 **选择** 更安全的方式，你可能应该放弃使用 “html/template”。

请注意，我正在维护“html/template”（我在 [这个项目](https://dev.golang.org/owners) 的昵称是 empijei ），所以我告诉你一些被那个包坑过的背景。如果我能很神奇的将所有用户迁移到安全版本，我肯定会。

## 结构

safehtml 由几个包组成。根据你的构建系统或公司的工具栈，你应该约束一些限制条件。你公司的安全团队或安全意识强的人应该为每个人灌输这个观点。

### safehtml

这是基础包，它只是提供结构安全的类型。简而言之，它是如何工作的：

- 这里有三种方式构建 [HTML type](https://godoc.org/github.com/google/safehtml#HTML)

- [ ] [通过转义不受信任的字符串](https://godoc.org/github.com/google/safehtml#HTMLEscaped) ，使其安全使用。
- [ ] [通过提供一个安全模板](https://godoc.org/github.com/google/safehtml/template#Template.ExecuteToHTML) ，它使用上下文自动转义，所以它是安全的（你可以在 [我之前的文章](https://blogtitle.github.io/robn-go-security-pearls-cross-site-scripting-xss/) 中阅读更多关于这部分内容）
- [ ] [通过连接已经被信任的HTML](https://godoc.org/github.com/google/safehtml#HTMLConcat) ，这基本上是安全的，实际上只有心怀不轨的开发者才能制造出错误。

这保证了 HTML 类型的每个实例都是安全的。[脚本类型](https://godoc.org/github.com/google/safehtml#Script) 的行为类似，但它不是模板，而是只能从常量或数据构建。为了表达“编译时常量”的概念，它有[接受不可被包外访问字符串类型的构造函数](https://godoc.org/github.com/google/safehtml#ScriptFromConstant)，因此调用它们的唯一方法是使用 [字符串文本](https://golang.org/ref/spec#String_literals) （我发现这是一个非常巧妙的技巧）。

此包中的所有其他类型都遵循类似的模式。

### safehtml/template

这是真正的“html/template”替代品，也是每个人都应该使用的包。如果不使用“legacyconversions”和“uncheckedconversions”（请参阅下文），并且**你的所有HTML响应都是由这个包生成的**，那么你可以保证在你的程序(products)中不会有任何服务器端 XSS。

我们正在研究确保最后一个条件为真的工具，但这需要一些时间。请 [继续关注](https://blogtitle.github.io/index.xml) 最新消息。

### safehtml/legacyconversions

此包只能用于转换到安全 API 。它允许任何任意字符串都是安全类型，这样转换到 safehtml 可以非常迅速，所有新代码都将是安全的。
++一旦迁移发生，就应该阻止使用这个包。++
顾名思义：这只是针对遗留代码，不应该有新代码使用它，**并且应该逐步重构此包的所有用法，以使用安全构造函数**。

### safehtml/testconversions

此软件包只能在测试目标中使用，并且仅在必要时使用。你应该设置一些 [linters](https://godoc.org/golang.org/x/tools/go/analysis) 来确保这一点。

### safehtml/uncheckedconversions

这是最微妙的问题。有时safehtmlapi太不方便，甚至无法使用。有时你不得不使用不安全的模式，因为你想做一些不能被证明是安全的事情（例如，从数据库中获取一些你信任的HTML并将其提供给客户端）。

对于这些**非常罕见**的情况，您可以使用此软件包。导入它应该限制在一组手工挑选的依赖项，并且每一个新的导入都需要一些安全意识强的人来审查它。

确保使用是**安全的，并将保持安全**，因为 uncheckedconversions  不会增加安全性。它们只是通知编译器您已经检查了代码，并希望它是可信的。遵循以下准则：

- 仅在严格必要的情况下使用（例如，如果使用 safehtml/template 需要更多的工作，但需要做额外的工作）。
- 为将来的审查者和维护者提供安全使用方法的文档。
- 通过减少 uncheckedconversion  对闭包参数、结构不确定类型字段的依赖性，缩小上下文范围。

这个包的用法是您的单点故障，所以请确保您遵循这些。（这句话假设您最终将摆脱 legacyconversions ）。

正确使用这个包的一个例子是（HTML）清理器的输出。（原文 sanitizer 这里译为清理器）如果您需要将用户提供的一些 HTML 嵌入到响应中（例如，因为您呈现 markdown 或网页邮箱），您将清理该 HTML 。一旦它被清理（如果你的清理器程序被正确实现），就可以使用未经检查的转换（unchecked conversion）将其升级为 HTML 类型。

### safehtml/raw

应阻止导入此包。“safehtml/”目录树之外的任何内容都不可见此包。

### safehtml/safehtmlutil

是的，我知道，名字不好。考虑一下，这个包和前一个一样，也不应该在safehtml之外导入，它只是为了减少代码重复和避免循环依赖而创建的。我同意可以用不同的名称或结构来命名它，但是既然你永远不会和这个包交互，就不必太麻烦你了。

## 如何进行重构

### `Printf` 和嵌套模板

您可能拥有的一个代码示例是

``` go
var theLink template.HTML = fmt.Sprintf("<a href=%q>the link</a>", myLink)
myTemplate.Execute(httpResponseWriter, theLink)
```

要重构它，您有多种选择：要么用另一个模板构建字符串（注意这里的“template”变量是“safehtml/template”类型）。

``` go
myLinkTpl := template.Must(template.New("myUrl").Parse("<a href={{.}}>the link</a>"))
theLink, err := myLinkTpl.ExecuteToHtml(myLink)
// handle err
myTemplate.Execute(httpResponseWriter, theLink)
```

或者，对于更复杂的情况，可以使用嵌套模板：

``` go
const outer = `<h1> This is a title <h2> {{ template "inner" .URL }}`
const inner = `<a href="{{.}}">the link</a>`
t := template.Must(template.New("outer").Parse(outer))
t = template.Must(t.New("inner").Parse(inner))
t.ExecuteTemplate(os.Stdout, "outer", map[string]string{"URL": myLink})
```

## 常量

如果代码中有一个HTML `常量`，那么可以将其用作模板并将其执行为HTML。这将检查所有标签是否配对以及其他内容，并返回一个HTML类型的实例。

如下：

``` go
var myHtml template.HTML := `<h1> This is a title </h1>`
```

<span id="checklist">结论列表</span>

1. 组织访问这些包
- 防止“safehtml”目录外的包导入“raw”、“uncheckedconversions”和“safehtmlutil”。
- 只允许测试构建时导入“testconversions”包。
2. 从“html/template”迁移并替换为“safehtml/template”。
- 对于每一个破损或每一个问题，使用“legacyconversions”调用。可能需要一些手动重构，但迁移应该相当简单。
- **运行所有的集成和E2E测试**。这很重要，所以我用 SHIFT 而不是 CAPS 来输入。
- 封锁 legacyconversions 列表：从现在起，禁止新导入“legacyconversions”包。
- 禁止使用“html/template”，这样所有新代码都是安全的。
3. 重构 legacyconversions 以使用安全模式。
- 尽可能以安全的方式构造 HTML 并删除 legacy conversions。
- 如果不可能使用  unchecked conversions。“uncheckedconversions”包的每一个新导入都应该被审查。

## 结论

如果您想确保Go代码中没有服务器端XSS，这可能是最好的方法。如果您有任何问题或需要更多的重构示例，请让我知道，您可以[通过twitter](https://twitter.com/empijei)（直接消息是开放的）或通过[电子邮件](mailto:empijei@gmail.com)与我联系。

------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
via: https://blogtitle.github.io/go-safe-html/

作者：[Rob](https://blogtitle.github.io/authors/rob/)
译者：[lts8989](https://github.com/lts8989)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
首发于：https://studygolang.com/articles/25119

# Go 程序的包含物：Go 程序中的非 Go 后缀文件

静态文件，也有人叫资产或资源，是一些被程序使用、没有代码的文件。在 Go 中，这类文件就是非 `.go` 的文件。它们大部分被用在 Web 内容，像 HTML 、javascript 还有网络服务器处理的图片，然而它们也可以以模板、配置文件、图片等等形式被用在任何程序中。主要问题是这些文件不会随代码一起被编译。开发一个程序时，我们可以通过本地的文件系统访问它们，但是当软件被编译和部署后，这些文件就不再在部署环境中的本地文件系统了，我们必须提供给程序一种访问它们的方式。Go 语言对这个问题并没有提供一种开箱即用的解决方案。本文会讨论这个问题，此问题的通用解决方案，以及 [gitfs](https://github.com/posener/gitfs) 中处理它的方法。[这部分](https://posener.github.io/static-files/#fsutil) 额外赠送，讲述下 `http.FileSystem` 接口一些有趣的方面。

我希望能听到你的想法。请用文末评论平台来讨论。

## 问题

很多情况下，Go 程序需要访问 非 Go 文件。开发过程中，可以从本地文件系统访问它们。例如，用 `os.Open` 函数通过本地文件系统读一个文件。很多标准库的函数就是用本地文件系统。例如用于（静态文件服务）为文件提供服务的 `http.Dir` 函数（译注：此处应为 `http.Dir` 类型，不应为函数），还有全部工作都在本地文件系统完成用于加载模板的 `template.ParseFiles` 和 `template.ParseGlob` 。

开发过程中使用本地文件系统是一种没有任何困难的体验。用 `go run` 启动程序，用 `go test` 进行测试，通过当前工作目录（CWD）的相对路径就可以访问文件。（用 `go buid`）构建完工程部署二进制文件时，问题就来了。（因为）在部署的环境中再用跟开发环境相同的路径不一定能访问到静态文件，也可能任何路径都访问不到。在下一段我们会讨论部署后让程序访问这些文件的不同的解决方案。

## 可选的解决方案

在讨论 Go 之前， 我们先来看一下 Python 是怎么解决的。pip 是 Python 的包管理工具。在众多功能中，它让程序可以定义 [data_files](https://docs.python.org/2/distutils/setupscript.html#installing-additional-files) ，这些文件随程序一起打包，被安装在一个可以在部署的环境中访问到的位置。Python 开发者不需要考虑程序环境的问题。不论是在开发环境还是生产环境，只要配置得当，静态文件就可以访问到。

Go 的 modules 不支持打包静态文件。在 Go 语言中，最常见的解决方案是 **binary-packing** （二进制打包）和 [resource-embedding](https://github.com/avelino/awesome-go#resource-embedding) （资源嵌入），像流行的库 [statik](https://github.com/rakyll/statik) ，Buffalo 写的 [packr](https://github.com/gobuffalo/packr) ，历久弥坚的 [go-bindata](https://github.com/go-bindata/go-bindata) 等等等等。据我所知，（上述）这些实现方式，都是用一个 CLI 工具通过把文件进行编码后存进一个生成的 Go 文件中的方式来把资产文件打包进一个 Go 文件。生成的文件提供一个公开接口来访问资产文件，构建程序时，这些文件被编译进了 Go 二进制文件。通常情况下，CLI 命令会以 `//go:generate` 进行注释，那些（以 `//go:generate` 注释的）文件会在每次 `go generate` 被调用的时候生成。

这种解决方案有一个好处就是安全 -- 无论 Go 代码是运行在开发环境中、测试环境中还是生成环境中，它都会使用静态文件内容新生成的版本 -- 所有环境中的版本和内容都相同。然而，这种方法有几个坏处。第一，（这种方法会让）开发流程繁琐累赘，尤其是在修改那些资产文件时。每次修改后，我们需要花费很长时间去（调用 `go generate`）重新生成那些文件。有一些工具给出了解决这个问题的部分解决方案，（但是）没有一种是小白式、能很容易地跟其他 Go 工具或命令整合。另一个缺点是，我们在提交流程中需要额外再增加一次验证 -- 例如运行 `go generate` 后执行 `git diff` 命令（diff 的结果为空，证明可以提交）。最后一个缺点，修改静态文件的那次提交，生成文件的 diff 通常会很难看，或者需要为这次的 diff 额外增加一次提交。

我个人以为这种方式并不方便，我更喜欢简单的，直接的解决方案：手动把静态内容嵌入到 Go 文件中。通过把依赖的内容添加到 Go 文件中可以实现。例如：

```go
var tmpl = template.Must(template.New("tmpl").Parse(`
…
`))

const htmlContent = `
<html>
…
</html>
`
```

上述方案在规模小的工程中可以使用。但是它也有缺点：嵌入到 Go 文件中的静态内容很难编辑和管理。第一，编辑器/IDE 是解析 Go 代码的，所以这些静态内容不会有语法高亮。其次，提示语法错误的行号是从嵌入的文本里计算的，并不是整个 Go 文件的行号。例如，如果 `template` 有个错误，`template.Must(template.New("tmpl").Parse("..."))` panic 了，提示的错误行号会是 `template` 文本里的而不是 Go 文件的。最后，用这种方式嵌入二进制内容是相当困难的。

另一个可选的方案是（维持）一个外部打包机制。例如，提供一个 docker 容器，这个容器包含需要的静态文件或包含诸如 RPM 等在给定位置保存静态文件的可安装的包。这种方法有几个缺点 -- docker daemon 运行的必要、抑或为不同的操作系统打不同的包的必要。但是最大的缺点是，程序不是自包含的，在开发环境和生产环境运行的方式有很大不同，且很难管理。

## gitfs

gitfs 是一个集上述几个方案众家之所长的库。它的设计目的是实现让开发者开发过程中在本地运行代码，可以快速修改静态内容，无缝衔接到生成环境中运行该份代码，并且可以不用二进制包。

它的设计原则之一就是 **seamless transition** （无损过渡）-- 一个 flag 或环境变量就可以改变程序运行的方式。这是通过使用抽象了底层文件系统的 `http.FileSystem`  来实现的。`http.FileSystem` 的具体实现可以是一个本地的目录、被打包进 Go 文件的文件以及从远程服务器上拉取下来的文件。要使用静态文件，开发者需要调用 `gitfs.new` ，这个函数返回 `http.FileSystem` 。他们使用这个抽象的文件系统去读静态内容，无视底层的实现。

下一个问题是，本地和生产环境中的相同的路径怎么用同一位置表示。Go 引入包的方式能一定程度上解决这个问题。域名加路径的格式，如 `github.com/user/project` ，被广泛地用来在一个工程中表示路径。 `gitfs` 采用了这个命名文件系统的方式，因此 Go 开发者会习惯这种方式。一个工程中的所有路径，特定分支或标签，都可以用相同的规则来确定。例如： `github.com/user/project/path@v1.2.3` 表示 `github.com/user/project` 这个工程下 `path` 这个路径的 `v1.2.3` 标签。

想象一个不用二进制包来访问静态内容的生产环境系统。`gitfs` 可以通过调 Github 的公开接口的方式来从获取文件系统结构和文件内容进而实现这种系统。程序创建文件系统时会通过 Github 的公开接口加载文件结构。文件自己的内容可以通过两个模式来获取：懒加载，仅在被访问时加载；文件系统加载时预获取所有内容。

`gitfs` 也实现了二进制打包，但是它（比一般意义上的二进制打包）体验更流畅平顺。第一，生成 Go 代码包的 CLI 工具会探测所有用 `gitfs.New` 创建文件系统的请求，因此开发者们运行 CLI 时不需要指定特定的文件系统，因为它能自动推断出来。而后，它下载所有依赖的内容并保存在生成的 Go 文件中。 这个 Go 文件在 `init()` 函数中注册有效的内容。当 `gitfs.New` 函数在程序里再次被调用来创建一个文件系统时，它会检查被注册过的内容，如果被已经被注册了，就直接使用注册的内容，而不是从远程仓库里获取。这样做的结果就是无缝衔接 -- 如果二进制中的内容是有效的，就直接使用，否则，从远程服务器上拉取。

前面提到过，生成二进制内容的缺点之一就是静态内容和打包好的内容会有差异的可能性。如果开发者修改了静态文件而没有运行 `go generate`  ，程序就可能不按预期运行。`gitfs` 处理这个问题的方式是，额外生成一个加载和比较生成的内容和静态文件（差异）的 Go 测试文件，如果本地有修改而没有重新运行 `go generate` ， 测试会不通过。

一件有趣的轶事：`gitfs` 工具用它自己来把它的模板文件打包成二进制，使用 gitfs 库加载它们。

## 实例

我们一起来看一个 `gitfs` 库中用 glob 匹配模式来加载模板文件的 [例子](https://github.com/posener/gitfs/blob/master/examples/templates/templates.go)

```go
// Add debug mode environment variable. When running with
// `LOCAL_DEBUG=.`, the local Git repository will be used
// instead of the remote github.
var localDebug = os.Getenv("LOCAL_DEBUG")

func main() {
	ctx := context.Background()
	// Open repository 'github.com/posener/gitfs' at path
	// 'examples/templates' with the local option from
	// environment variable.
	fs, err := gitfs.New(ctx,
		"github.com/posener/gitfs/examples/templates",
		gitfs.OptLocal(localDebug))
	if err != nil {
		log.Fatalf("Failed initializing Git filesystem: %s.", err)
	}
	// Parse templates from the loaded filesystem using a glob
	// pattern.
	tmpls, err := fsutil.TmplParseGlob(fs, nil, "*.gotmpl")
	if err != nil {
		log.Fatalf("Failed parsing templates.")
	}
	// Execute a template according to its file name.
	tmpls.ExecuteTemplate(os.Stdout, "tmpl1.gotmpl", "Foo")
}
```

通过命令 `go run main.go` 来运行这段代码，运行后会从 Github 加载模板。如果执行的命令是 `LOCAL_DEBUG=. Go run main.go` 则会加载本地文件。

## fsutil

`http.FileSystem` 是一个表示抽象文件系统的 interface 。它只有一个方法，`Open` 入参是从文件系统根目录开始的一个相对路径，返回值是一个实现了 `http.File` interface 的对象。`http.File` 是一个表示文件或目录的通用 interface 。因为 `gitfs` 大量使用了 `http.File` ，所以 `gitfs` 模块就包含了提供了使用这个 interface 的很多工具的  [`fsutil`](https://godoc.org/github.com/posener/gitfs/fsutil) 包。

`Walk` 函数，整合了 `http.FileSystem` interface 和可能访问所有文件系统的文件的 [`github.com/kr/fs.Walker`](https://godoc.org/github.com/kr/fs#Walker) 。

Go 语言标准库模板加载函数只能访问本地文件系统。 `fsutil` 里是一个可以使用 `http.FileSystem` 任何实现的移植版本。用  [`fsuitl.TmplParse`](https://godoc.org/github.com/posener/gitfs/fsutil#TmplParse) 代替 [`text/template.ParseFiles`](https://golang.org/pkg/text/template/#ParseFiles) 。[`fsuitl.TmplParseGlob`](https://godoc.org/github.com/posener/gitfs/fsutil#TmplParseGlob) 代替 [`text/template.ParseGlob`](https://golang.org/pkg/text/template/#ParseGlob) 。以此类比到 HTML ，[`fsutil.TmplParseHTML`](https://godoc.org/github.com/posener/gitfs/fsutil#TmplParseHTML) 代替 [`html/template.ParseFiles`](https://golang.org/pkg/html/template/#ParseFiles) ， [`fsutil.TmplParseGlobHTML`](https://godoc.org/github.com/posener/gitfs/fsutil#TmplParseGlobHTML) 代替 [`html/template.ParseGlob`](https://golang.org/pkg/html/template/#ParseGlob) 。

 [`Glob`](https://godoc.org/github.com/posener/gitfs/fsutil#Glob) 函数的入参是 `http.FileSystem` 和 glob 匹配模式的 list ，返回值是仅包含命中给定的 glob 模式的文件的文件系统。

 [`Diff`](https://godoc.org/github.com/posener/gitfs/fsutil#Diff) 函数计算文件系统结构的差异和两个文件系统间文件内容的差异。

如果你有关于此类使用函数的更多想法，请付诸行动， [open an issue](https://github.com/posener/gitfs/issues) 。

## 总结

Go 程序中的非 Go 文件需要特殊对待。文本中我试着阐述现在面临的挑战、现有的有效的解决方案和 `gitfs` 是怎样让使用静态文件变得简单地。我们已经学习过 http.FileSystem interface，及它对文件系统操作进行抽象的强大能力。最后的想法：新的 Go 模块系统是否为对静态文件进行内建处理保留了一席之地。

---

via: https://posener.github.io/static-files/

作者：[Eyal Posener](https://posener.github.io/)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

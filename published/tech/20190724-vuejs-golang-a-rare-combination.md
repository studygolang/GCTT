首发于：https://studygolang.com/articles/23978

# Vuejs + Golang = 一个稀缺的组合

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/0*SJ43Bk4fxc44mgVR.jpg)

时间回到 2018 年，我写了一篇获得 15k 阅读的文章：Django + Angular 4 = A powerful web application。出于好奇心，我尝试了Angular4 和 Django 的组合。接着上个系列，这是一篇使用 Vuejs 和 Golang 来帮助你构建极佳应用的文章。

我知道这两者一起用不是很常见，但是，让我们试一试。

## 背景

首先，让我们来聊聊两者提供的技术特征。

### Golang

1. 二进制 —— 所有需要构建的内置依赖项都是二进制文件。因此，您不需要安装运行时来运行应用程序。

2. 静态类型 —— 编译器不仅在编译过程检查类型，而且在代码编写的过程。处理类型转换和兼容性等问题都有兼顾。

3. 并发性 —— Golang 最好的特性是它对高并发性的优先支持。

4. 标准库 —— 标准库功能足够强大，你基本上不需要第三方库。

### Vue.js

1. 体积 —— 经过 gzip 压缩后，它的大小仅为 18kb ，对比压缩后的 jQuery，gzip 压缩后的大小为 29kb。

2. 可读性 —— Vue.js 的源码和语法非常清晰简单。

3. 文档 —— Vue.js 完善的文档使它能很快上手和学习。

4. 灵活 —— 数据在 HTML 和 JavaScript 间的绑定是非常流畅的。

5. Vue CLI 3 —— cli 提供了一系列的功能让你很快的上手，尝试着使用它你会喜欢上它。

Golang 和 Vuejs 在运行时都很快，所以让他们一起合作构建一个很快的单页面应用。

我们开始构建：

## 初始化文件目录

首先，像下面这样初始化你的文件目录，为 Git 添加 “LICENSE” 、“README.md” 等文件.

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*vtaJKeFNo6dKujZYsEi6hw.png)

在 backend 文件夹创建一个 'server.go' 文件：

```bash
server.go
```

前端部分，在命令行输入以下命令来创建一个新的 app 应用：

```shell
 vue create calculator
```

它会询问你选择 `preset` 预设，选择默认的即可（babel, eslint).

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*tbr9X84OEsJCrWgEOJSHvA.png)

现在文件目录结构如下：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*IUbOGEl5b4ozSYWQmuDM_w.png)

代码的结构准备好了!

现在让我们看下我们将要做开发的应用，它不是一个很复杂但是可以帮助你开始开发复杂应用。我们将要开发一个计算器，我们在前端输入两个数字，通过 POST 请求发送到 Golang 服务端。服务端将会进行加、减、乘、除，然后通过一个 POST 请求返回结果，前端将会渲染结果。

## 构建后端

Golang 很快是因为它的编译器，它不允许你定义多余的变量。如果你定义变量或者引入任何 package 包，那么这些变量、包应该是你需要使用的，Golang 在这方面很严格。另外无论你想要做什么，你需要事先告诉给编译器，例如：如果你想使用 POST 获取数据，你必须事先定义返回数据的 JSON 格式。这有点麻烦，但是为了速度这是值得的。

回到代码上。

因此，我们将会使用 `encoding/json` and `net/http` 包。然后，我们定义 JSON 类型的数据结构。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*bBx8qYZkWpg8R92e8gT_5g.png)

现在，我们将会写一个简单的方法来做数字的运算。这里，我们需要明确指定返回数据的类型。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*AR0TlByqhirRpm8s7R5FUA.png)

我们需要写一个方法，当我们使用 POST 方法发送 JSON 格式的请求时，这个方法能够被执行。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*JaFzDlfYH0LwTwLTOIs2Iw.png)

在代码第 33 行，我们定义了一个 `JSON`的译码器来转译从请求的实体中的传过来的 JSON 数据。

`numsData` 和 `numsResData` 是定义好的数据结构。

接收的数据存储在 `numsData`中 并且在 38 行被转译。

然后我们设置 `ResponseWriter` 的 header 头部，并且在 47 行返回 JSON 格式的响应数据以及检查错误。

最后，在这个主函数上，我们可以定义 HTTP 路由，例如 53 行为每个 URL 请求分别定义响应函数。后端服务将运行在 `8090` 端口上。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*ug8QMqiNpF9QefRl9uuNuQ.png)

完整的代码如下：

## 现在我们来构建前端

首先使用 `cd` 进入前端目录并且使用以下命令安装依赖：

```shell
npm install --save bootstrap-vue bootstrap axios vee-validate
```

我们将使用 `axios` 来处理 POST 请求，使用`vee-validate` 校验表单的输入数据，使用 `bootstrap-vue` 构建优美的界面。

在 `src/Calculator.vue` 文件里编写前端部分代码：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/vuejs-golang/1*13qy_tphvGcHiOM1wR3KIg.png)

从 17 行到 25 行，我们定义 input 输入框和 label 标签来获取数据。在Vue 中使用 `v-model`指令获取数据。

30-35 行完成计算器 UI 层的展示，并且 43 行定义了一个按钮，将会被触发 `postreq` 方法，这个方法会在接下来完成。

我们在 `script` 标签内写 `JavaScript` 代码，首先引入依赖：

![](https://miro.medium.com/max/1400/1*2oy5ZoqYZVh0bF_iml9onw.png)

54-59 行是必须的，用于引入`axios` 和 `vee-validate`。
然后在 64-69 行我们定义一些变量，这些 `data` 变量用来存储计算器组件的变量的值。

所有的函数都会定义在 `methods` 对象里。我们创建 `postreq()` 方法用来向 `http://localhost:8090/calc` 发送 JSON 格式的 POST 请求。还记得之前在 `server.go` 文件创建的 `calc` 方法吗？我们发送 JSON 数据后，后端返回结果后数据会被储存在 `add`, `mul`, `sub` 和 `div`等变量中，这些绑定在 HTML 的变量例如 {{ add }} 的占位符将会显示结果。

很简单是吧？ 是的。

以上就是所有我们需要做的，只需要记住这些要点：

- 使用 Golang 写服务端逻辑并且运行在单独的端口上.
- 构建代码结构来处理 JSON 数据，你不能将它们存储在变量中。
- 前端的 Vuejs 会使用 GET 或 POST 请求来调用服务端的 API 接口。

## 运行应用

启动后端服务可以使用以下命令，它将运行在 `8090` 端口上：

```go
go run server.go
```

运行前端可以使用：

```shell
npm run serve
```

祝贺！你的 App 完成了。

整个代码托管在这个 [Github 仓库](https://github.com/adesgautam/Calculator)上。

并且，如果你想看另一个我使用 Vue.js 和 Golang 构建的应用 Rocket Engine Designer，请点击[这里](https://github.com/adesgautam/Proton)

如果你喜欢这篇文章，请点击 👏 按钮给与你对这篇文章的喜爱。

---

via: https://medium.com/@adeshg7/vuejs-golang-a-rare-combination-53538b6fb918

作者：[Adesh Gautam](https://medium.com/@adeshg7)
译者：[M1seRy](https://github.com/M1seRy)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

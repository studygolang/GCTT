首发于：https://studygolang.com/articles/25281

# 使用 Hugo 和 Github Pages 创建你的开发者作品集

拥有一个作品集网站可以使你在寻找一个开发外包时脱颖而出。作品集网站可以让潜在的客户或雇主了解你是一个专家，了解你过去和正在做的工作。不幸的是，一些常见的困难阻碍了许多人拥有作品集网站，包括最近的我--害怕所有的工作，计划并且从一个草图中构建一个网站，选择一个主机提供商，如果你想域名可用，那些主机和域名会让你破费（特别当你缺钱的时候）等等。

在本指导中，我会带领你快速且免费的建立并且上线你的工作集网站。

## 但是，首先 ... 有一个作品集网站真的那么重要吗？

当然。以下是原因：

1. 它列出了你作为一个开发者的技术技能。你不仅可以列出你在职业生涯中磨练出的技能，还能连接到一切能够具体表现他们的东西。这些东西可以包括你的项目，你参加过的黑客马拉松，你参加过的面试，你发表过的文章，你做过的教程、演讲、开源工作等。
2. 它支持你建立关系网。[关系网就是和兴趣相投的人建立联系](https://podcasts.google.com/?feed=aHR0cHM6Ly9mZWVkcy5mZWVkYnVybmVyLmNvbS9Xb3JrbGlmZVdpdGhBZGFtR3JhbnQ%3D&episode=cHJ4XzEzMV9iMjNiZDRlYi04YWYyLTQ3ZmMtYTYwZi1iOGMwMWVjMTBmYjg%3D)，而且，你的作品集网站可以给 SWE 空间中的其他人展示你的兴趣。
3. 这是一种吸引潜在客户和雇主的方式。
4. 它帮助你建立你的品牌。它允许你来定义那些看到你作品集的陌生人如何看待你。
5. 它允许你来展现你的风格。
6. 作为开发者，它是你在网络上其他部分的门户。你能把人们指向到你的 Dribble, Github, LinkedIn, Codepen, Dev.to, Medium, Youtube, Twitch 和你的其它的开发者账户。

## 前置条件

1. Github 账户：如果你还没有账户的话，通过 [这个链接](https://github.com/join) 创建一个。
2. 安装好的 Hugo：[macOS](https://gohugo.io/getting-started/installing#macos)，[Windows](https://gohugo.io/getting-started/installing#windows) 和 [Linux](https://gohugo.io/getting-started/installing#linux) 的安装说明。

## 第一步：生成你的网站

1. 打开你的终端然后使用 `cd` 命令进入一个你用来安装网站的目录。

2. 为你的网站的文件夹起一个名字。我们使用 `<PORTFOLIO_NAME>` 来作为占位符。

3. 就像下面这样生成你的站点。

   ```shell
   hugo new site <PORTFOLIO_NAME>
   ```

4. 使用 `cd` 命令进入刚新生成的文件夹并且初始化为一个 Git 仓库。

   ```shell
   cd <PORTFOLIO_NAME> && Git init
   ```

## 第二步：选择并添加一个主题

1. 前往 [Hugo 的作品集主题页](https://themes.gohugo.io/tags/portfolio/) 然后选择一个你喜欢的主题。在本篇教程中，我选择了一个叫做 [UILite](https://themes.gohugo.io/tags/portfolio/) 的主题。它简洁，看起来相当酷而且满足一个作品集的基本需求。当然那儿也有很多其他很酷的主题可供选择。
   UILite 主题看起来像这样。
   ![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/create-your-developer-portfolio-using-hugo-and-github-pages-35en/pic1.png)

2. 把这个主题作为一个 Git 子模块添加到 `<PORTFOLIO_NAME>`。这步因主题不同而异，但差不多都鼓励作为一个子模块来使用主题。你可以通过下面的命令添加主题：

   ```shell
   Git submodule add <LINK_TO_THEME_REPO> themes/<THEME_NAME>
   ```

   在我们的例子中需要这样：

   ```shell
   Git submodule add https://github.com/uicardiodev/hugo-uilite themes/hugo-uilite
   ```

3. 在已经自动生成的 **config.toml** 文件中指明你将在你的作品集网站中使用的 `<THEME_NAME>`。这个 **config.toml** 文件允许你对于你的整个站点进行详细设定。

   ```shell
   Echo 'theme = "<THEME_NAME>"' >> config.toml
   ```

   对于我们的主题，我们这样写：

   ```shell
   Echo 'theme = "hugo-uilite"' >> config,toml
   ```

   在这一步的最后，你的 **config.toml** 文件应该看起来像这个样子：

   ```shell
   baseURL = "http://example.org/"
   languageCode = "en-us”
   title = "My New Hugo Site”
   theme = "hugo-uilite”
   ```

   改变作品集网站的 `title` 和 `baseURL` 会是个很好的想法。

## 第三步：测试你的网站

1. 打开 Hugo 服务器。

   ```shell
   hugo server
   ```

2. 在浏览器中打开 [http://localhost:1313](http://localhost:1313)。你的站点现在应该正在工作并且看起来应该像这样：
   ![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/create-your-developer-portfolio-using-hugo-and-github-pages-35en/pic2.png)
   它看起来像是有点儿缺陷，但它是正常的。只是我们还没有在网站中添加内容。这是我们接下来将会做的事情。

## 第四步：调整你的主题

这步是因人而异的，并且取决于你的主题，你想添加的内容和你的设计感。一些常见的作品集包括突出显示的姓名，简介，开发者社交简历的链接，技能，项目，工作经验，成就等等。

以下是调整 **hugo-uilite** 主题的一种方式：

1. 首先把主题恢复到它原来的样子。对于 **hugo-uilite** 主题，我们可以通过从它的 *exampleSite* 文件夹中复制文件来实现。如果刚开始你的主题看起来坏掉了，首先从主题的 README 中寻求解决方法。如果就像我们这个例子一样什么都没有说明，那么寻找你的主题仓库中的 *exampleSite* 文件夹。通过从 *exampleSite* 的 data 文件夹中复制缺失的文件到你的作品集网站的 data 文件夹中，来修复你的站点。

   ```shell
   mkdir data
   cp themes/hugo-uilite/exampleSite/data/* data/
   ```

   以下展示了在我们的 data 文件夹中的文件层级：

   ```shell
   data
   ├── config.json
   ├── experience.json
   ├── services.json
   ├── sidebar.json
   ├── skills.json
   └── social.json
   ```

   Hugo 服务器发现了这些更改然后重新载入，现在我们的网站看就来就像 [主题示例页](https://themes.gohugo.io/theme/hugo-uilite/) 显示的那样了。

2. 下一步，我们将会在网站上添加专业信息。就像上面提到的那样，你应该在网站上添加一些细节。也能做一些对主题样式的改变。以下，我会描述我做了哪些工作来更改主题，使得网站看起来像下面的网站截图那样。变化很大，因为并不是所有的主题都是一样的所以没有详细描述。以下是一些总结：
   **a.  更改主题的色调**

   * 我在 `static`  文件夹中添加了一个额外的 CSS 文件。
   * 任何你在此处添加的样式都会覆盖主题本身的样式。
   * 主题的 `layout` 文件夹是查找你想修改样式的组件的 id 和类的好地方。

   **b. 添加专业信息**

   * 因为我们已经复制了 `data` 文件夹，所以我们需要做的就是修改对应文件中信息，使其能够反映在站点上。举例来说，我们可以通过修改 `experience.json` 文件来修改 `Experience` 。

   **c. 改变站点图标**

   * 如果一个主题没有提供能够改变站点图标的配置设置项，那需要将其添加到站点的头部中。
   * 在本文示例中，头文件存放在主题的 `layouts/partials` 文件夹中。
   * 为了增加图标，复制头文件到作品集的 `layout/partials` 文件夹中并且修改必要项。
   * 另外，如果图标不是一个链接而是文件的话，记得把这个图标文件添加到 `static` 文件夹中。

   **d. 增加项目栏**

   * 这个主题没有提供**项目栏**。
   * 因为项目栏与 **服务栏** 是相似的，所以我复制它的 HTML 文件（ `layouts/partials/project.html` ），修改它使得项目栏可以被链接，为它增加一个数据文件（ `data/projects.json` ）并且把它添加到 `layout/index.html` 文件中（我从主题中复制过来的）。
   * 在这个示例中，你没有提供服务，所有你所做的事情只是更改了 `data/services.json` 文件中的内容来反映到 **项目** 或其他地方，而不是 **服务**。

   **e. 增加社交信息**

   * 已经提供了 `data/socials.json` 和 `data/socialsfas.json` （用于 Awesome-Font 实体风格图标）文件可以用来列举社交信息。

   最后站点看起来像这样：

   ![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/create-your-developer-portfolio-using-hugo-and-github-pages-35en/pic3.png)

## 第五步：在 Github 上创建仓库并把源码推上去

1. 向主分支提交更改。

2. 在 Github 上创建一个仓库来存放你的站点的源码，举例来说叫做 `<PORTFOLIO_NAME>`。你可以把这个仓库设置为私有的，从而它只用来存放你的站点的源码。

3. 把你最近的更改推送到这个仓库中。

4. 创建第二个仓库来部署你的站点，它的名字应该按照 `<USERNAME>.github.io` 这种格式。

5. 在你本地计算机中，使用 `cd` 命令回到 `<PORTFOLIO_NAME>` 文件夹，并且通过运行 Hugo 服务器来检查你是否满意站点的样子。如果你很满意作品集的最后样子，关掉 Hugo 服务器。

6. 然后我们需要把本地编辑站点生成的 `public` 文件夹与我们创建的 `<USERNAME>.github.io` 仓库连接起来。我们会使 `<USERNAME>.github.io` 仓库作为 public 文件夹的远程源，并且使 `public` 文件夹作为我们 `<PORTFOLIO_NAME>` 项目的一个子模块。运行下面这个命令：

   ```shell
   Git submodule add -b master git@github.com:<USERNAME>/<USERNAME>.github.io.git public
   ```

7. 我们可能会经常修改我们的站点，那么就需要能够在更改后能够很容易的进行我们作品集的部署。 Hugo 提供了一个能够把你的更改推送到源（带有可选的提交信息）并且部署作品集网站的脚本。它可以被添加到你的 `<PORTFOLIO_NAME>` 项目中并且当你做出更改后只要简单的运行一下就可以了。脚本叫做 **deploy.sh** ，在 [这里](https://gohugo.io/hosting-and-deployment/hosting-on-github/#put-it-into-a-script) 。如果你已经复制好了这个脚本，在第一次部署的时候你需要做的事情就是：

   ```shell
   # 给脚本执行权限
   chmod +x deploy.sh
   # 部署你的作品集，带有一个可选的信息
   ./deploy.sh "Deploy the first version of my portfolio site”
   ```

## 第六步：在 Github Pages 上面部署你的网站

如果你对 Github 用户页面不熟悉或者不感兴趣，你可以读一下 [这些](https://help.github.com/en/articles/user-organization-and-project-pages#user--organization-pages) 。

我们在 **用户页面** 中部署的作品集可以通过 `<USERNAME>.github.io` URL 进行访问。我们直接从第五步创建的 `<USERNAME.github.io` 仓库的主分支进行部署：

1. 首先，如果你是免费用户，这个仓库必须是公开的（如果你是升级用户，你才可以从私有仓库中部署）。如果你的 `<USERNAME>.github.io` 仓库是私有的，通过 **Settings > Options > Danger Zone** 把它改回公开。
   ![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/create-your-developer-portfolio-using-hugo-and-github-pages-35en/pic4.png)
2. 然后，我们需要设置 GIthub 用户页面的来源。如果你的仓库是以 `<USERNAME>.github.io` 进行命名的，Github 页面自动能够使用，默认是公开仓库，如果你是升级用户，私有仓库也是可以的。在 **Settings > Options > Github Pages** 中进行设置。我们将会选择 `master` 分支作为页面的来源。![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/create-your-developer-portfolio-using-hugo-and-github-pages-35en/pic5.png)
3. 打开你仓库页面的 **Code** 标签页下的 **Environments** 标签页，确保部署已经成功。在这个标签页下面，你会看到你过去部署的日志，最上面高亮的就是你最近的部署。![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/create-your-developer-portfolio-using-hugo-and-github-pages-35en/pic6.png)
4. 最后，在一个新浏览器标签页输入 `<USERNAME>.github.io` ，就可以看到你的作品集了。

## 结语

我希望本篇教程是有用的。尽管在 Github 页面上得到一个静态作品集是很容易的，但我更喜欢 Hugo 的方式，因为它的可定制性，题材广泛的主题和模板内建的东西，如谷歌分析，DIsqus 等等。在此，我也很感激 Github 的用户页面免费静态站点。

---

via: https://dev.to/zaracooper/create-your-developer-portfolio-using-hugo-and-github-pages-35en

作者：[Zara Cooper](http://github.com/zaracooper)
译者：[Ollyder](https://github.com/Ollyder)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

## 如何使用 Go 从 HTML 生成 PDF 报告

![](https://raw.githubusercontent.com/lavaicer/Img/main/image)
作为一名开发人员，我有时需要为我的应用程序创建 PDF 报告。完全通过编程来创建它们可能很麻烦，并且每个库都有些不同。最后，让事物看起来像设计师想要的那样可能具有挑战性。如果我们能在不花大量时间的情况下让它看起来像设计，那不是很好吗？设计师和前端通常会 HTML 和 CSS，所以使用 HTML 是说得通的。但网站通常在打印出来时看起来不太好，而且不是为多页设计的。我们提出了一个解决方案，我们相信它可以解决上述所有问题。

## 认识 UniHTML 与 UniPDF 的结合

[UniHTML](https://github.com/unidoc/unihtml?ref=hackernoon.com) 是  [UniPDF](https://github.com/unidoc/unipdf?ref=hackernoon.com) 的新插件， [UniDoc](https://unidoc.io/?ref=hackernoon.com) 是我们在 UniDoc 的旗舰库之一。

它是基于容器的解决方案，带有 Go 驱动程序，根据原理图：

![](https://raw.githubusercontent.com/lavaicer/Img/main/2.jpg)




[Docker 映像](https://hub.docker.com/repository/docker/unidoccloud/unihtml?ref=hackernoon.com)在 Docker Hub 上公开可用。

UniPDF Creator 软件包可以[创建灵活的 PDF 报告](https://www.unidoc.io/post/creating-pdf-reports-in-golang?ref=hackernoon.com)和[发票](https://www.unidoc.io/post/simple-invoices?ref=hackernoon.com)。 UniHTML 基于容器的模块具有灵活的 Web 渲染引擎，并且与 UniPDF 相结合汇集了为 UniPDF 报告生成添加完整 HTML 支持的功能。

## 试一试

让我们试试看

**Create a free metered API key.**
**第 1 步：创建一个免费的计量的 API 密钥**

这很简单，只需在 https://cloud.unidoc.io 上注册一个帐户并在 UI 中创建一个计量 API 密钥。

有关这方面的分步说明，请参阅：

- [如何注册UniCloud](https://help.unidoc.io/article/142-how-to-sign-up-for-unicloud?ref=hackernoon.com)
- [如何生成计量 API 密钥](https://help.unidoc.io/article/141-metered-license-api-key?ref=hackernoon.com)

**第 2 步：让 UniHTML 容器运行**

``` shell
$ docker run -p 8080:8080 -e UNIDOC_METERED_API_KEY=mymeteredkey unidoccloud/unihtml
Unable to find image 'unidoccloud/unihtml:latest' locally
latest: Pulling from unidoccloud/unihtml
6e640006d1cd: Pull complete
1a3def68b0c4: Pull complete
5b1718db67b4: Pull complete
8d4c41b870b6: Pull complete
b1a4436c2bab: Pull complete
3c3af5a4fff5: Pull complete
29863d0ede88: Pull complete
Digest: sha256:c1c69af194358179d836a648f07f71af07ed0c968938abe3a3e2550e49980728
Status: Downloaded newer image for unidoccloud/unihtml:latest
[INFO]  server.go:173 Listening private API on: :8081
[INFO]  server.go:164 Listening public API on: :8080
```

**第 3 步：运行一个示例**

受博客文章“[使用 CSS 创建漂亮的 HTML 表格](https://dev.to/dcodeyt/creating-beautiful-html-tables-with-css-428l?ref=hackernoon.com)”的启发，我们将以下 HTML 文件放在一起，以说明带有 HTML 表格的 PDF 报告

**sample.html**

``` html
<html>
<head>
<style>

.styled-table {
    border-collapse: collapse;
    margin: 25px 0;
    font-size: 0.9em;
    font-family: sans-serif;
    min-width: 400px;
    box-shadow: 0 0 20px rgba(0, 0, 0, 0.15);
}

.styled-table thead tr {
    background-color: #009879;
    color: #ffffff;
    text-align: left;
}

.styled-table th,
.styled-table td {
    padding: 12px 15px;
}


.styled-table tbody tr {
    border-bottom: 1px solid #dddddd;
}

.styled-table tbody tr:nth-of-type(even) {
    background-color: #f3f3f3;
}

.styled-table tbody tr:last-of-type {
    border-bottom: 2px solid #009879;
}


.styled-table tbody tr.active-row {
    font-weight: bold;
    color: #009879;
}

</style>
</head>

<table class="styled-table">
    <thead>
        <tr>
            <th>Name</th>
            <th>Points</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>Dom</td>
            <td>6000</td>
        </tr>
        <tr class="active-row">
            <td>Melissa</td>
            <td>5150</td>
        </tr>
        <!-- and so on... -->
    </tbody>
</table>

</html>
```

**example.go**

``` go
package main

import (
    "fmt"
    "os"

    "github.com/unidoc/unihtml"
    "github.com/unidoc/unipdf/v3/common/license"
    "github.com/unidoc/unipdf/v3/creator"
)

func main() {
    // Set the UniDoc license.
    if err := license.SetMeteredKey("my api key goes here"); err != nil {
        fmt.Printf("Err: setting metered key failed: %v\n", err)
        os.Exit(1)
    }

    // Establish connection with the UniHTML Server.
    if err := unihtml.Connect(":8080"); err != nil {
        fmt.Printf("Err:  Connect failed: %v\n", err)
        os.Exit(1)
    }

    // Get new PDF Creator.
    c := creator.New()

    // AddTOC enables Table of Contents generation.
    c.AddTOC = true

    chapter := c.NewChapter("Points")

    // Read the content of the sample.html file and load it to the conversion.
    htmlDocument, err := unihtml.NewDocument("sample.html")
    if err != nil {
        fmt.Printf("Err: NewDocument failed: %v\n", err)
        os.Exit(1)
    }

    // Draw the html document file in the context of the creator.
    if err = chapter.Add(htmlDocument); err != nil {
        fmt.Printf("Err: Draw failed: %v\n", err)
        os.Exit(1)
    }

    if err = c.Draw(chapter); err != nil {
        fmt.Printf("Err: Draw failed: %v\n", err)
        os.Exit(1)
    }


    // Write the result file to PDF.
    if err = c.WriteToFile("sample.pdf"); err != nil {
        fmt.Printf("Err: %v\n", err)
        os.Exit(1)
    }
}
```

## 结果

运行结果:

``` go
$ go run example.go
```

创建了一个看起来这样的 sample.pdf ：

![](https://raw.githubusercontent.com/lavaicer/Img/main/3.jpg)



我们注意到我们还有目录，这对于高质量的 PDF 制作至关重要，以及链接到 PDF 中每一章的书签，页眉和页脚也可以轻松创建。

![](https://raw.githubusercontent.com/lavaicer/Img/main/1.jpg)

## 结论

UniPDF 中的 UniHTML 通过一个完整的渲染引擎提供了简单的 HTML 到 PDF 的转换。对于已经拥有 HTML 设计并需要添加专业 PDF 报告的团队而言，这将使 PDF 报告生成过程变得非常容易，而网站的纯 PDF 打印输出是不够的

_以前发表于_ [_https://www.unidoc.io/post/html-for-pdf-reports-in-go_](https://www.unidoc.io/post/html-for-pdf-reports-in-go?ref=hackernoon.com)

---
via: https://hackernoon.com/how-to-generate-a-pdf-report-from-html-with-go-p02f33kx

作者：[gushall](https://hackernoon.com/u/gushall)
译者：[lavaicer](https://github.com/lavaicer)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出 
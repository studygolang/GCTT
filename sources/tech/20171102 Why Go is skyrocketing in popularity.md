[翻译中] by mosliu

#### In only two years, Golang leaped from the 65th most popular programming language to #17\. Here's what's behind its rapid growth.

![](https://opensource.com/sites/default/files/styles/image-full-size/public/lead-images/build_structure_tech_program_code_construction.png?itok=nVsiLuag)  
Image by opensource.com

The [Go programming language,](https://golang.org/) sometimes referred to as Google's golang, is making strong gains in popularity. While languages such as Java and C continue to dominate programming, new models have emerged that are better suited to modern computing, particularly in the cloud. Go's increasing use is due, in part, to the fact that it is a lightweight, open source language suited for today's microservices architectures. Container darling Docker and Google's container orchestration product [Kubernetes](https://opensource.com/sitewide-search?search_api_views_fulltext=Kubernetes) are built using Go. Go is also gaining ground in data science, with strengths that data scientists are looking for in overall performance and the ability to go from "the analyst's laptop to full production."

As an engineered language (rather than something that evolved over time), Go benefits developers in multiple ways, including garbage collection, native concurrency, and many other native capabilities that reduce the need for developers to write code to handle memory leaks or networked apps. Go also provides many other features that fit well with microservices architectures and data science.

Because of this, Go is being adopted by interesting companies and projects. Recently an API for [Tensorflow](https://www.tensorflow.org/) has been added, and products like [Pachyderm](http://www.pachyderm.io/) (next-gen data processing, versioning, and storage) are being built using Go. Heroku's [Force.com](https://github.com/heroku/force) and parts of [Cloud Foundry](https://www.cloudfoundry.org/) were also written in Go. More names are being added regularly.

## Rising popularity and usage

In the September 2017 TIOBE Index for Go, you can clearly see the incredible jump in Go popularity since 2016, including being named TIOBE's Programming Language Hall of Fame winner for 2016, as the programming language with the highest rise in ratings in a year. It currently stands at #17 on the monthly list, up from #19 a year ago, and up from #65 two years ago.

![tiobe_index_for_go.png](https://opensource.com/sites/default/files/u128651/tiobe_index_for_go.png)   
TIOBE Index for Go [TIOBE](https://www.tiobe.com/tiobe-index/go/)

The Stack Overflow Survey 2017 also shows signs of Go's rise in popularity. Stack Overflow's comprehensive survey of 64,000 developers tries to get at developers' preferences by asking about the "Most Loved, Dreaded, and Wanted Languages." This list is dominated by newer languages like Mozilla's Rust, Smalltalk, Typescript, Apple's Swift, and Google's Go. But for the third year in a row Rust, Swift, and Go made the top five "most loved" programming languages.

![stackoverflow_most_loved.png](https://opensource.com/sites/default/files/u128651/stackoverflow_most_loved.png)  
Most Loved, Dreaded, and Wanted Languages, [Stackoverflow.com](https://insights.stackoverflow.com/survey/2017#most-loved-dreaded-and-wanted)

## Go advantages

Some programming languages were hacked together over time, whereas others were created academically. Still others were designed in a different age of computing with different problems, hardware, and needs. Go is an explicitly engineered language intended to solve problems with existing languages and tools while natively taking advantage of modern hardware architectures. It has been designed not only with teams of developers in mind, but also long-term maintainability.

At its core, Go is pragmatic. In the real world of IT, complex, large-scale software is written by large teams of developers. These developers typically have varying skill levels, from juniors up to seniors. Go is easy to become functional with and appropriate for junior developers to work on.

Also, having a language that encourages readability and comprehension is extremely useful. The mixture of duck typing (via interfaces) and convenience features such as "**:=**" for short variable declarations give Go the feel of a dynamically typed language while retaining the positives of a strongly typed one.

Go's native garbage collection removes the need for developers to do their own memory management, which helps negate two common issues:

*   First, many programmers have come to expect that memory management will be done for them.
*   Second, memory management requires different routines for different processing cores. Manually attempting to account for each configuration can significantly increase the risk of introducing memory leaks.

Go's native concurrency is a boon for network applications that live and die on concurrency. From APIs to web servers to web app frameworks, Go projects tend to focus on networking, distributed functions, and/or services for which Go's goroutines and channels are well suited.

## Suited for data science

Extracting business value from large datasets is quickly becoming a competitive advantage for companies and a very active area in programming, encompassing specialties like artificial intelligence, machine learning, and more. Go has multiple strengths in these areas of data science, which is increasing its use and popularity.

*   Superior error handling and easier debugging are helping it gain popularity over Python and R, the two most commonly used data science languages.
*   Data scientists are typically not programmers. Go helps with both prototyping and production, so it ends up being a more robust language for putting data science solutions into production.
*   Performance is fantastic, which is critical given the explosion in big data and the rise of GPU databases. Go does not have to call in C/C++ based optimizations for performance gains, but gives you the ability to do so.

## Seeds of Go's expansion

Software delivery and deployment have changed dramatically. Microservices architectures have become key to unlocking application agility. Modern apps are designed to be cloud-native and to take advantage of loosely coupled cloud services offered by cloud platforms.

Go is an explicitly engineered programming language, specifically designed with these new requirements in mind. Written expressly for the cloud, Go has been growing in popularity because of its mastery of concurrent operations and the beauty of its construction.

Not only is Google supporting Go, but other companies are aiding in market expansion, as well. For example, Go code is supported and expanded with enterprise-level distributions such as [ActiveState's ActiveGo](https://www.activestate.com/activego). As an open source movement, the [golang.org](https://golang.org/) site and annual [GopherCon](https://www.gophercon.com/) conferences form the basis of a strong, modern open source community that allows new ideas and new energy to be brought into Go's development process.

----------------

via: https://opensource.com/article/17/11/why-go-grows

作者：[Jeff Rouse](https://opensource.com/users/jeffr)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出

已发布：https://studygolang.com/articles/12397

# 接受 interface 参数，返回 struct 在 go 中意味着什么

## 注意细节

在之前的文章中，我提到了一个关于 *accept interfaces, return structs* 的参考指南，在查看同事代码的时候经常会被问“为什么”。特别是这不是一个必须遵守的规则。这个想法的关键点以及理解什么时候妥协，在于维护项目灵活性和避免抢占抽象（译者注：“Preemptive abstractions” 并发系统中连续组件的轻量级验证方案的一种抽象技术）之间的平衡。

![c90d04d0.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/accept-interface/c90d04d0.png)

## 抢占抽象让系统变得复杂

> 除了因为太多的迂回方式所造成的问题之外，所有的计算机科学问题都能够通过另一个级别的迂回方式来解决。  
>  - David J. Wheeler

软件工程师喜欢抽象。个人看法，我从未看到过一个同事参与写代码超过他为了某个事务建立抽象多。Go 语言从结构中抽象出接口，这种处理方式会产生嵌入复杂性。遵循[你并不需要它](http://c2.com/xp/YouArentGonnaNeedIt.html)软件设计理念，如果不需要就没有理由增加复杂性。一个常见的返回接口的理由是让用户把注意力放在函数所提供的 API 上。在 Go 中因为隐含实现了接口，所以这并不需要。返回结构的公共函数就成为那个API。

> 永远只有当你真正需要的时候才抽象，不要因为预见可能会需要而抽象

一些语言需要你预见每一个可能从未用过的接口。隐含实现接口一个最大的好处，就是允许你在后面实际需要的时候优雅的抽象事务，而不是需要你预先抽象出来。

## 使用者眼中的需求

> 当你真正需要他们的时候

你怎么知道什么时候需要抽象？对于返回类型来说，比较容易。你是写函数的人，所以你确切的知道什么时候需要抽象返回值。  

对于输入参数来说，是否需要不在你的控制范围之内。你也许认为你的数据模型足够了，但是一个用户可能需要和某些属性封装一下。如果可能的话，可以预想一下每个调用你的函数的情况，但这是比较困难的。这种可以控制输出，但是不能预期用户输入的不平衡的状况产生了一种强烈的偏见，抽象输入而不是输出。

## 去掉无用的代码细节

![0ba28e07.png](https://raw.githubusercontent.com/studygolang/gctt-images/master/accept-interface/0ba28e07.png)
复杂的做鸡蛋的方法

简化的另一个方面是去除不需要的细节。类似菜单函数：给一个输入然后你得到一个蛋糕！不需要列出做蛋糕的材料。同样的，函数也不需要列出不使用的输入参数。下面的函数你会怎么想？  

```go
func addNumbers(a int, b int, s string) int {
	return a + b
}
```

对于大多数程序员来说很明显参数 **s** 是不需要的。当参数是结构的时候就不那么明显了。

```go
type Database struct{ }
func (d *Database) AddUser(s string) {...}
func (d *Database) RemoveUser(s string) {...}
func NewUser(d *Database, firstName string, lastName string) {
	d.AddUser(firstName + lastName)
}
```

就像一个写满了配料的菜单一样，NewUser 输入参数是一个有很多功能的 Database 对象。实际上只需要 AddUser 但是却得到了额外的 RemoveUser。接口允许我们在创建函数的时候只依赖我们需要的功能。

```go
type DatabaseWriter interface {
	AddUser(string)
}
func NewUser(d DatabaseWriter, firstName string, lastName string) {
	d.AddUser(firstName + lastName)
}
```

Dave Cheney 在写[接口隔离原则](https://en.wikipedia.org/wiki/Interface_segregation_principle)的时候也提到了[这一点](https://dave.cheney.net/2016/08/20/solid-go-design)。他还描述了一些关于限制输入的其他好处，值得读一下。概括一下就是：  

> 依照需求描述的结果也就是函数-仅仅是需要可写并且提供相应的功能

我会按照这个思想，重新考虑上面的函数 addNumbers，很明显不需要参数 s 字符串，函数 NewUser 同样也不需要一个包括 RemoveUser 的Database参数。

## 总结理由和审查例外情况

主要的理由如下：

- 移除不需要的抽象
- 在函数参数上用户需求的歧义
- 简化函数参数

这些理由也允许有例外的情况。例如，如果你的函数实际上返回多种类型而不是一个接口。同样地，如果函数是私有的，你能控制输入参数，会偏向于不要做抢占抽象。对于第三条规则，go 没有方式可以抽象出结构成员的值。 所以，如果你的函数需要访问结构成员（并且不只是结构方法），那么你必须接受结构作为参数。

## 问答

提问：如果 Database 有很多方法，比如 20 个，那么如何处理呢？

回答：一个结构可能有 20 个方法，但是一个方法不需要调用那么多方法。 试着把这些方法按组分类，比如读的方法，写的方法，管理的方法等等。这样需要 Database 的函数可以使用方法子集来处理。

----------------

via: https://medium.com/@cep21/what-accept-interfaces-return-structs-means-in-go-2fe879e25ee8

作者：[Jack Lindamood](https://medium.com/@cep21)
译者：[tyler2018](https://github.com/tyler2018)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

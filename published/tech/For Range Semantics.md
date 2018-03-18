已发布：https://studygolang.com/articles/12607

# For Range 的语义

## 前言

为了更好地理解本文中提及的内容，这些是需要首先阅读的好文章：

下面列出 4 篇文章的索引：
1. [Go 语言机制之栈和指针](https://studygolang.com/articles/12443)
2. [Go 语言机制之逃逸分析](https://studygolang.com/articles/12444)
3. [Go 语言机制之内存剖析](https://studygolang.com/articles/12445)
4. [Go 语言机制之数据和语法的设计哲学](https://studygolang.com/articles/12487)

在 Go 编程语言中，值语义和指针语义的思想无处不在。如前面的文章所述，语义一致性对于完整性和可读性至关重要。它允许开发人员在代码持续不断增长时保持强大的代码库[心理模型](https://en.wikipedia.org/wiki/Mental_model)。它还有助于最大限度地减少错误，副作用和未知行为。

## 介绍

在这篇文章中，我将探索 Go 中的 `for range` 语句如何提供值和指针语义形式。我将教授语言机制，并展示这些语义有多深奥。然后，我将展示一个简单的例子，说明混合这些语义和可能导致的问题是多么容易。

## 语言机制

从这段代码开始，它展示了 `for range` 循环的值语义形式。

[play.golang.org](https://play.golang.org/p/_CWCAF6ge3)

**代码清单1**

```go
package main

import "fmt"

type user struct {
	name string
	email string
}

func main() {
	users := []user{
		{"Bill", "bill@email.com"},
		{"Lisa", "lisa@email.com"},
		{"Nancy", "nancy@email.com"},
		{"Paul", "paul@email.com"},
	}

	for i, u := range users {
		fmt.Println(i, u)
	}
}
```
在代码清单1中，程序声明一个名为 `user` 的类型，创建四个用户值，然后显示关于每个用户的信息。第 18 行的范围循环使用值语义。这是因为在每次迭代中都会在循环内部创建并操作来自切片的原始用户值的副本。实际上，对 `Println` 的调用会创建循环副本的第二个副本。如果要为用户值使用值语义，这就是你想要的。

如果你要使用指针语义，`for range` 循环看起来像这样。

**代码清单2**

```go
for i := range users {
	fmt.Println(i, users[i])
}
```

现在该循环已被修改为使用指针语义。循环内的代码不再它的副本上运行，而是在切片内存储的原始 `user` 上运行。但是，对 `Println` 的调用仍然使用值语义，并且传递了一份副本。

要解决这个问题，需要再做一次最后的修改。

**代码清单3**

```go
for i := range users {
	fmt.Println(i, &users[i])
}
```

现在会一直使用 `user` 的指针语义。

作为参考，清单4并排显示了值和指针语义。

**代码清单4**

```go
// Value semantics.           // Pointer semantics.
for i, u := range users {     for i := range users {
	fmt.Println(i, u)             fmt.Println(i, &users[i])
}                             }
```

## 深层机制

语言机制比这更深入。请看代码清单 5 中的这个程序。程序初始化一个字符串数组，对这些字符串进行迭代，并在每次迭代中更改索引为 1 的字符串。

[https://play.golang.org/p/IlAiEkgs4C](https://play.golang.org/p/IlAiEkgs4C)

**代码清单5**

```go
package main

import "fmt"

func main() {
	five := [5]string{"Annie", "Betty", "Charley", "Doug", "Edward"}
	fmt.Printf("Bfr[%s] : ", five[1])

	for i := range five {
		five[1] = "Jack"

		if i == 1 {
		   fmt.Printf("Aft[%s]\n", five[1])
		}
	}
}
```

这个程序的预期输出是什么？

**清单6**
```
Bfr[Betty]
Aft[Jack]
```
正如你所期望的那样，第 10 行的代码已经改变了索引 1 的字符串，你可以在输出中看到。该程序使用 `for range` 循环的指针语义版本。接下来，代码将使用 `for range` 循环的值语义版本。

[https://play.golang.org/p/opSsIGtNU1](https://play.golang.org/p/opSsIGtNU1)

**清单7**

```go
package main

import "fmt"

func main() {
	five := [5]string{"Annie", "Betty", "Charley", "Doug", "Edward"}
	fmt.Printf("Bfr[%s] : ", five[1])

	for i, v := range five {
		five[1] = "Jack"

		if i == 1 {
			fmt.Printf("v[%s]\n", v)
		}
	}
}
```

在循环的每次迭代中，代码再次更改索引 1 处的字符串。此时代码显示索引 1 处的值时，输出不同。

**清单8**
```
Bfr[Betty] : v[Betty]
```
我们可以看到这种形式的 `for range` 真的是使用了值语义。`for ranage` 在数组的拷贝上进行迭代。这就是为什么在输出中并未看到值的改变。

当使用值语义形式覆盖切片时，将采用切片标头的副本。 这就是为什么清单 9 中的代码不必惊慌。

**清单9**

```go
package main

import "fmt"

func main() {
	five := []string{"Annie", "Betty", "Charley", "Doug", "Edward"}

	for _, v := range five {
		five = five[:2]
		fmt.Printf("v[%s]\n", v)
	}
}

Output:
v[Annie]
v[Betty]
v[Charley]
v[Doug]
v[Edward]
```

如果您查看第09行，循环内的切片值会缩减为2，但循环将在切片值的自身副本上进行操作。 这允许循环使用原始长度进行迭代而没有任何问题，因为后备数组仍然是完整的。

如果代码使用 `for range ` 的指针语义形式，程序就会发生混乱。

**清单10**

```go
package main

import "fmt"

func main() {
	five := []string{"Annie", "Betty", "Charley", "Doug", "Edward"}

	for i := range five {
		five = five[:2]
		fmt.Printf("v[%s]\n", five[i])
	}
}

Output:
v[Annie]
v[Betty]
panic: runtime error: index out of range

goroutine 1 [running]:
main.main()
	/tmp/sandbox688667612/main.go:10 +0x140
```

`for range` 在迭代之前获取到切片的长度，但是在循环过程中长度发生了变换。现在在第三次迭代的时候，循环尝试访问不再与切片长度相关联的元素。

## 混合语义

这是一个完全糟糕的例子。该代码混合了用户类型定义的语义，并引发了一个 bug。

**清单11**

```go
package main

import "fmt"

type struct user {
	name string
	likes int
}

func (u *user) notify() {
	fmt.Printf("%s has %d likes\n", u.name, u.likes)
}

func (u *user) addLike() {
	u.likes++
}

func main() {
	users := []user{
		{name: "bill"},
		{name: "lisa"},
	}

	for _, u := range users {
		u.addLike()
	}

	for _, u := range users {
		u.notify()
	}
}
```

这个例子没有那么做作。在第05行，`user` 类型被声明并且选择指针语义来实现为用户类型设置的方法。然后在 `main` 程序中，在 `for range` 循环中使用值语义为每个用户添加一个 like。然后使用第二个循环来再次使用值语义来通知每个 `user`。

**清单12**

```
bill has 0 likes
lisa has 0 likes
```

输出显示并没有增加 like。我无法强调，您应该为给定类型选择语义，并坚持使用该类型的数据。

这是代码应该看起来如何与用户类型的指针语义保持一致。

**清单13**

```go
package main

import "fmt"

type user struct {
	name string
	likes int
}

func (u *user) notify() {
	fmt.Printf("%s has %d likes\n", u.name, u.likes)
}

func (u *user) addLike() {
	u.likes++
}

func main() {
	users := []user{
		{name: "bill"},
		{name: "lisa"},
	}

	for i := range users {
		users[i].addLike()
	}

	for i := range users {
		users[i].notify()
	}
}

// Output:
bill has 1 likes
lisa has 1 likes
```

## 结论

值和指针语义是Go编程语言的重要组成部分，正如我已经展示的那样，集成到了 `for range` 循环中。在使用 `for range` 时，验证你正在迭代的给定类型在使用正确的形式。最后一件事是混合语义，如果你没有注意的话，`for range` 很容易混合使用语义。

语言给了你这种选择语义的力量，并且干净而一致地使用它。这是你想要充分利用的东西。 我想让你决定每种类型使用的语义并保持一致。你对一段数据的语义越一致，您的代码库就越好。如果你有一个很好的理由来改变语义，然后广泛地记录下来。

---
via: https://www.ardanlabs.com/blog/2017/06/for-range-semantics.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[shniu](https://github.com/shniu)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

# 从 JavaScript 到 Go 语言的排序算法

在计算机科学中，排序的意思是获取一个数组，然后重新对他们进行排列，使他们遵循指定的顺序，例如按字母顺序对字符串进行排序、按最小到最大的顺序对数字进行排序，或按结构中的一个字段对结构数组进行排序。您可以使用它（排序）来提高算法的工作效率，或按特定顺序显示数据（例如时间上的从最近到最远）。

对于 Go 中的排序，标准库提供了 sort 包，有意思的是，它使用了 Go 接口来定义对数据进行排序的规则。如果您使用过 JavaScript 的 Array.prototype.sort 方法，（那么，您对此）会很熟悉！

## 字符串排序

让我们从按字母顺序排列一组字符串开始：

```golang
var languages = []string{"Go", "C", "Ruby", "JavaScript", "XML"}
```

在 JavaScript 中，对它们进行排序（代码）就像这样：

```javascript
let languages = ["Go", "C", "Ruby", "JavaScript", "XML"];

languages.sort();
console.log(languages); // ["C", "Go", "JavaScript", "Ruby", "XML"]
```

由于 ```languages``` 是一个数组，我们可以使用 Array.prototype.sort, 对它们进行排序。

由于与 JS 数组不同，Go 切片没有开箱即用的方法，我们没办法直接使用已有的排序算法，我们需要导入 sort 包并使用其 Sort 函数来对数组重新排列。让我们试试吧！（首先，）将此代码放在一个名为 sort-strings.go 的文件中：

```golang
package main

import (
    "fmt"
    "sort"
)

func main() {
    languages := []string{"Go", "C", "Ruby", "JavaScript", "XML"}
    sort.Sort(languages)

    fmt.Println(languages)
}
```

然后，运行 ```go run sort-strings.go```，你应该会得到（如下错误）：

```bash
./sort-strings.go:10:14: cannot use languages (type []string) as type sort.Interface in argument to sort.Sort:
    []string does not implement sort.Interface (missing Len method)
```

编译器错误？之所以会这样，是因为 ```sort.Sort``` 不接受切片类型，它无法对切片类型进行自动转换。它的函数签名实际上是这样的：

```golang
func Sort(data Interface)
```

sort.Interface（带有一个大 I）是一个 Go 接口，表示可以排序的数据集合，如字符串、数字抑或是结构体列表。由于对字符串和整数的切片进行排序很常见，所以 sort 包也提供了一些内置方法，使 sort.Sort 方法与字符串或整数切片可以兼容. 试试这个！

```golang
  func main() {
      languages := []string{"Go", "C", "Ruby", "JavaScript", "XML"}
-     sort.Sort(languages)
+     sort.Sort(sort.StringSlice(languages))

      fmt.Println(languages)
  }
```

sort.StringSlice 是一个字符串切片方法，但它实现了 sort.Interface 接口. 因此，通过将一个 []string 类型转换为 StringSlice 类型，就可以使用 sort.Sort! 现在，如果您（再）执行 ```go run sort-strings.go``` 命令，您应该会看到按字母顺序排列的编程语言列表！

为什么我们需要使用一个特殊的接口来对数据进行排序，而不是让 Go 语言的 sort.Sort 方法直接接受（字符串或整型）切片？原因是因为，我们传入的是一个元素集合，Go 语言需要通过某种方法来知道元素的顺序。为了编写这些规则来对切片进行排序，您需要实现 sort.Interface 方法。正如您看到的，Interface 使我们可以灵活地以任何您喜欢的方式来定义元素的顺序！

## 实现自定义排序类型

假设我们的 languages 切片包含 "fish"（一种 shell 脚本语言）。如果您按字母顺序对 "编程工具" 进行排序，那么像这样有意义的排序是：

```golang
[C, fish, Go, JavaScript, Ruby, XML]
```

但是，即使有 XML，"fish" 也排在最后！（这是因为）使用 sort.StringSlice, 与使用 JS 中的字符串列表排序算法 Array.prototype.sort 相同，默认按照字典顺序排序，而不是字母顺序。在字典顺序中，小写字母（如fish 中的 f）在大写字母（如 XML 中的 X）之后。如果我们想不区大小写，就按照字母的顺序排序，我们需要实现一些自定义行为。那会是什么样子呢？

在实现自定义排序规则之前，我们需要想想排序的作用。在本教程中，我们不会研究不同排序算法（如快速排序、归并排序和冒泡排序）的细节，虽然学习它们在编程中很重要。关于在 Go 和 JS 中编写自定义排序算法，您需要了解的是，它们需具备：

- 查看集合中的元素
- 比较它们，看看哪些元素应该排在前面
- 根据这些比较将元素按顺序排列

在 JavaScript 中，您将传入一个自定义函数来告诉 sort 如何对数组中的元素进行比较，如下所示：

```golang
languages.sort((langA, langB) => {
  langA = langA.toLowerCase();
  langB = langB.toLowerCase();
  if (langA < langB) {
    return -1; // return -1 if langA should go before langB in the array
  } else if (langB > langA) {
    return 1;  // return 1 if langB should go before langA in the array
  }
  return 0;    // return 0 if they can go in either order
})
```

因为我们在比较之前已使用 toLowerCase 方法，（这样可以）使得 fish 语言排在 Go、JavaScript、Ruby 和 XML 语言之前，但在 C 语言之后！

如果我们查看 Go sort.Interface，我们可以看到需要实现的方法如下：

```golang
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}
```

所以要创建一个可以排序的类型，我们需要实现 sort.Interface 接口：

- 告诉 Go sort 包集合的长度
- 取集合中的任意两个元素（元素 i 和 j），并将他们进行交换
- 查看集合中的任意两个元素，看看 Less 方法在对集合进行排序时哪个应该排在前面

让我们以 Len 和 Swap 的实现方法开始。

```golang
type alphabeticalStrings []string

func (a alphabeticalStrings) Len() int { return len(a) }

func (a alphabeticalStrings) Swap(i, j int) {
    placeholder := a[j]
    a[j] = a[i]
    a[i] = placeholder
}
```

首先，我们对字符串切片进行封装，定义一个新的类型，alphabeticalStrings。在 Go 语言中，我们通过定义自己的类型，我们可以为它编写方法。

对于 Len 方法，我们只是使用 Go 的内置 len 函数来获取切片的长度，对于 Swap，我们交换切片中的两个元素。目前为止一切顺利。现在让我们实现 Less 方法。先导入 strings 包，并添加这个函数：

```golang
func (a alphabeticalStrings) Less(i, j int) bool {
    return strings.ToLower(a[i]) < strings.ToLower(a[j])
}
```

注意到关于 Less 方法了吗？它看起来非常像我们在 Array.prototype.sort 函数中定义的比较方法，除了它返回一个 bool 类型而不是 int 类型，并接受的是切片索引而不是元素本身！

现在，让我们来试试！编辑 main 函数，让它像这样：

```golang
  func main() {
      languages := []string{"Go", "C", "fish", "Ruby", "JavaScript", "XML"}
-     sort.Sort(sort.StringSlice(languages))
+     sort.Sort(alphabeticalStrings(languages))

      fmt.Println(languages)
  }
```

如果您执行 ```go run sort-strings.go``` 命令，现在您应该可以看到按预期排序的列表！

```golang
[C, fish, Go, JavaScript, Ruby, XML]
```

你知道 Go 有什么好玩的 sort.Interface 接口吗？我们编写的字母字符串类型和 Go 团队编写的 StringSlice 类型都建立在一个普通的旧类型之上，[]string 并且都可以传递到 sort.Sort. 我们可以通过选择我们将字符串切片转换为哪种类型来选择我们想要的字符串顺序！

## 使用 sort.Slice 简化我们的排序！

JS 和 Go 语言不同版本的 sort 方法之间的一个重要区别是，对 Go 切片进行排序时，除了比较函数之外，我们还需要编写 Len 和 Swap 方法。对于不同的切片类型，Len 和 Swap 看起来都差不多。故，定义一个新的（元素）排序，都必须实现这三种方法感觉有点麻烦。

需要实现这三种方法的原因是，你实现 sort.Interface 接口时，对应的数据，并不一定是数组或切片。我只是使用了切片的 sort 包，但您可以使用其他数据类型实现 sort.Interface 接口，例如链表。

对于切片，在 Len 和 Swap 方法中，我们经常使用的是相同的逻辑，我们能不能只实现 Less，就像在 JavaScript 中一样？（其实）sort 包就有这样的方法，sort.Slice！

```golang
func Slice(
    slice interface{},
    less func(i, j int) bool,
)
```

我们传入我们想要排序的数据切片作为第一个参数，以及一个函数来比较切片的元素作为第二个参数。甚至无需创建新类型，现在我们就可以对数据进行排序！让我们再一次重构我们的 main 函数来试试：

```golang
  func main() {
      languages := []string{"Go", "C", "fish", "Ruby", "JavaScript", "XML"}
-     sort.Sort(alphabeticalStrings(languages))
+     sort.Slice(languages, func(i, j int) bool {
+         return strings.ToLower(languages[i]) < strings.ToLower(languages[j])
+     })

      fmt.Println(languages)
  }
```

完成了！我们已经获取到我们要的排序切片了！

sort 包还有一些很酷的地方，除了我们能够选择排序所依据的顺序外，注意在 sort.Sort 和 sort.Slice 中，我们不需要知道我们正在使用哪种排序算法。sort.Sort 已包含了具体的实现算法，所有需要从我们这边获取到的信息是，如何比较元素，如何交换它们，以及我们有多少元素。这就是接口的作用！

顺便说一下，熟悉排序算法的工作原理仍然是绝对值得的，这会让您知道如何让你的计算机做更少的工作来排列数据，而其（原理）的应用场景非常多。因此，如果您想了解它们是如何工作，sort.Sort 和我们编写的这些函数在背后都做了什么，下面是一些关于算法本身的材料（可供学习）。

- [Free Code Camp - 解释排序算法](https://www.freecodecamp.org/news/sorting-algorithms-explained/)
- [Toptal - 排序算法动画](https://www.toptal.com/developers/sorting-algorithms)
- [在 JavaScript 中实现排序算法](https://medium.com/@rwillt/implementing-sorting-algorithms-in-javascript-b08504cdf4a9)

---
via: https://dev.to/andyhaskell/sorting-in-go-from-javascript-4k8o

作者：[andyhaskell](https://dev.to/andyhaskell)
译者：[译者ID](https://github.com/gogeof)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

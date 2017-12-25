第 13 部分: Maps
====================
29 四月 2017

欢迎访问第13篇，[Golang 系列教程](https://golangbot.com/learn-golang-series/)。


什么是 map？
---------------------
map 是 Go 语言一个内置类型，关联一个 value 和一个 key。通过相应的 key 可以获取到 value。
   
如何创建map?
---------------------
map 可以通过使用 `make` 方法传入 key 和 value 创建。 `make(map[type of key]type of value)` 是创建 map 的语法。

```go
personSalary := make(map[string]int)
```

上面的代码创建了一个名为 `personSalary` 的 map，有一个 string 类型的 key 和 一个 int 类型的 value。

map 的零值是 `nil`。如果你想添加元素到空的（`nil`） map 中，会发生运行时错误。因此 map 必须使用 `make` 方法初始化。

```go
package main
import (
"fmt")
func main() {
var personSalary map[string]int
if personSalary == nil {
fmt.Println("map is nil. Going to make one.")
personSalary = make(map[string]int)
}}
```

[在线运行程序](https://play.golang.org/p/IwJnXMGc1M)

上面的程序中，personSalary 是 nil，因此需要使用 make 方法初始化，程序将输出

```go
map is nil. Going to make one.
```


给 map 添加元素
---------------------

添加 map 元素的语法和 array 相同。下面的程序在 map `personSalary` 中添加了几个新元素。

```go
package main
import (
"fmt")
func main() {
personSalary := make(map[string]int)
personSalary["steve"] = 12000
personSalary["jamie"] = 15000
personSalary["mike"] = 9000
fmt.Println("personSalary map contents:", personSalary)}
```

[在线运行程序](https://play.golang.org/p/V1lnQ4Igw1)

上面的程序输出，`personSalary map contents: map[steve:12000 jamie:15000 mike:9000]`

你也可以在声明的时候初始化 map。

```go
package main
import (
"fmt")
func main() {
personSalary := map[string]int {
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
fmt.Println("personSalary map contents:", personSalary)
}
```   

[在线运行程序](https://play.golang.org/p/nlH_ADhO9f)

上面的程序声明了 personSalary 并在声明同时添加两个元素。之后又添加了 key `mike` 。程序输出，

`personSalary map contents: map[steve:12000 jamie:15000 mike:9000]`

key 不一定只能是 string 类型。所有可比较的类型如 boolean，interger，float，complex，string 等都可以做 key。如果你想了解更多可比较的类型，请访问 [http://golang.org/ref/spec#Comparison_operators](http://golang.org/ref/spec#Comparison_operators)。


获取 map 中的元素
---------------------

现在我们已经添加了几元素到 map，现在学习下如何获取。获取 map 元素的语法是 `map[key]` 。

```go
package main
import (
"fmt")
func main() {
personSalary := map[string]int{
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
employee := "jamie"
fmt.Println("Salary of", employee, "is", personSalary[employee])}
```   

[在线运行程序](https://play.golang.org/p/-TSBac7F1v)

上面的程序很简单。获取并打印 employee `jamie` 的 salary。 程序输出 `Salary of jamie is 15000`.

如果一个元素的不存在怎么办呢？map 会返回该元素类型的零值。在 `personSalary` 这个 map 里，如果我们获取一个不存在的元素，会返回 `int` 类型的零值 `0` 。

```go   
package main
import (
"fmt")
func main() {
personSalary := map[string]int{
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
employee := "jamie"
fmt.Println("Salary of", employee, "is", personSalary[employee])
fmt.Println("Salary of joe is", personSalary["joe"])}
```

[在线运行程序](https://play.golang.org/p/EhUJhIkYJU)

上面程序输出：

```go
Salary of jamie is 15000
Salary of joe is 0
```

上面程序返回 `joe` 的 salary 是 0。`personSalary` 中不包含 `joe` 的情况下我们不会获取到任何运行时错误。

如果我们想知道 map 中到底是不是存在这个 `key` 怎么办呢？
```go
value, ok := map[key]
```

上面就是获取 map 中某个 key 是否存在的语法。如果 `ok` 是 true，表示 key 存在，key 对应的值就是 `value` ，反之表示 key 不存在。

```go
package main
import (
"fmt")
func main() {
personSalary := map[string]int{
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
newEmp := "joe"
value, ok := personSalary[newEmp]
if ok == true {
fmt.Println("Salary of", newEmp, "is", value)
} else {
fmt.Println(newEmp,"not found")
}
}
```

[在线运行程序](https://play.golang.org/p/q8fL6MeVZs)

上面的程序中，第 15 行，`joe` 不存在，所以 `ok` 是 false。程序将输出，

```go
joe not found
```

遍历 map 中所有的元素需要用 `for range` 循环。

```go
package main
import (
"fmt")
func main() {
personSalary := map[string]int{
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
fmt.Println("All items of a map")
for key, value := range personSalary {
fmt.Printf("personSalary[%s] = %d\n", key, value)
}
}
```

[在线运行程序](https://play.golang.org/p/gq9ZOKsI9b)

上面程序输出，

```go
All items of a map
personSalary[mike] = 9000
personSalary[steve] = 12000
personSalary[jamie] = 15000
```   

__有一点很重要，当使用 `for range` 遍历 map 时，不保证每次执行程序获取的元素顺序相同。__


删除 map 中的元素
---------------------

删除 `map` 中 `key` 的语法是 _delete(map, key)_。这个方法没有返回值。

```go
package main
import (
"fmt")
func main() {
personSalary := map[string]int{
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
fmt.Println("map before deletion", personSalary)
delete(personSalary, "steve")
fmt.Println("map after deletion", personSalary)
}
```

[在线运行程序](https://play.golang.org/p/nroJzeF-a7)

上述程序删除了 key "steve"，输出

```go
map before deletion map[steve:12000 jamie:15000 mike:9000]
map after deletion map[mike:9000 jamie:15000]
```

获取 map 的长度
---------------------

获取 map 的长度使用 [len](https://golang.org/pkg/builtin/#len) 方法。

```go
package main
import (
"fmt")
func main() {
personSalary := map[string]int{
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
fmt.Println("length is", len(personSalary))
}
```

[在线运行程序](https://play.golang.org/p/8O1WnKUuDP)

上述程序中的 _len(personSalary)_ 方法表示 map 的长度。程序输出 `length is 3`。

Map是引用类型
---------------------

和 [slices](https://golangbot.com/arrays-and-slices/) 类似，map 也是引用类型。当 map 被赋值为一个新变量的时候，它们指向同一个内部数据结构。因此一个变量中的内容改变也会导致另一个变量的内容修改。

```go
package main
import (
"fmt")
func main() {
personSalary := map[string]int{
"steve": 12000,
"jamie": 15000,
}
personSalary["mike"] = 9000
fmt.Println("Original person salary", personSalary)
newPersonSalary := personSalary
newPersonSalary["mike"] = 18000
fmt.Println("Person salary changed", personSalary)
}
```

[在线运行程序](https://play.golang.org/p/OGFl3addq1)

上面程序中的第 14 行，`personSalary` 被赋值给 `newPersonSalary`。下一行 ，`newPersonSalary` 中 `mike` 的 salary 变成 `18000` 。`personSalary` 中 `Mike` 的 salary 也会变成 `18000`。程序输出，

```go
Original person salary map[steve:12000 jamie:15000 mike:9000]
Person salary changed map[steve:12000 jamie:15000 mike:18000]
```

当 map 作为函数参数传递时也会发生同样情况。函数中对 map 的任何修改，对调用者来说是可见的。


Map 相等判断
---------------------

map 之间不能使用 `==` 操作符判断，`==` 只能用来检查 map 是否为 `nil`。

```go
package main

func main() {
map1 := map[string]int{
  "one": 1,
  "two": 2,
}
   map2 := map1
   if map1 == map2 {
}
}
```

[在线运行程序](https://play.golang.org/p/MALqDyWkcT)

上面程序抛出异常 __invalid operation: map1 == map2 (map can only be compared to nil)__。

判断两个 map 是否相等的方法是遍历比较两个 map 中的每个元素。我建议你写一段这样的程序实现这个功能 :)。


我在一个程序里实现了我们讨论过的所有概念。你可以从 [github](https://github.com/golangbot/maps) 下载代码。

这就是 map 。谢谢你的阅读。祝好。


via: https://golangbot.com/maps/

作者：[Nick Coghlan](https://golangbot.com/about/)
译者：[ArisAries](https://github.com/ArisAries)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
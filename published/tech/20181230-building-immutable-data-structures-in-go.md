首发于：https://studygolang.com/articles/23435

# 用Go构建不可变的数据结构

![Photo by Frantzou Fleurine on Unsplash](https://raw.githubusercontent.com/studygolang/gctt-images/master/building-immutable-data-structures-in-go/cover.jpeg)

[共享状态](https://www.quora.com/What-is-the-meaning-of-shared-state-and-the-side-effects-in-functional-programming)是比较容易理解和使用的，但是可能产生隐晦以至于很难追踪的 bugs。尤其是在我们的数据结构只有部分是通过引用传递的。切片就是这么一个很好的例子。后续我会作出更加详细的讲解。

在处理经过多级变换或*状态*的数据时，不可变数据结构是非常有用的。不可变仅意味着原始结构是不可以被改变的，而每一个新的结构副本都是以新的属性值创建。

让我们看个简单的例子：

```go
type Person struct {
    Name           string
    FavoriteColors []string
}
```

显然，我们可以实例化一个`Person`然后随心所欲地更改它的属性。事实上，这样做并没有任何错。但是，当你处理更加复杂的、传递引用和切片的嵌套式数据结构，或者利用通道传递副本时，以某些姿势更改这些共享的数据副本可能会导致不易察觉的 bugs。

## 为啥我之前就没有遇到过这种问题呢？

如果没有重度使用 channel 或代码基本是串行执行的，由于从定义上讲每次只有一个操作能够作用在数据上，你不大可能会遇见这些不明显的 bugs。

再者，除了避免 bugs外，不可变数据结构还有其他优势：

1. 由于状态绝不会原地更新，这对一般的调试和记录每个变换步骤以用于后续监控是非常有用的
2. 撤销或“时光倒流”的能力不仅是可能的，而且是小菜一碟，只需一个赋值操作即可
3. 由于正确且安全的实现需要损失性能和费尽心思地仔细设置/测试内存锁，共享状态被广泛认为是糟糕的做法

## Getter 和 Wither

*Getter* 返回数据，*setter* 改变数据，*wither* 创建新状态。

基于 getter 和 wither，我们可以精准控制能被改变的属性。这也为我们提供了一种记录变换的有效方式（后续）。

新的代码如下：

```go
type Person struct {
    name           string
    favoriteColors []string
}

func (p Person) WithName(name string) Person {
    p.name = name
    return p
}

func (p Person) Name() string {
    return p.name
}

func (p Person) WithFavoriteColors(favoriteColors []string) Person {
    p.favoriteColors = favoriteColors
    return p
}

func (p Person) FavoriteColors() []string {
    return p.favoriteColors
}
```

需要注意的关键点如下：
1. `Person` 的属性都是私有的，因此外部包无法绕过 `Person` 提供的方法来访问其属性
2. `Person` 的方法接收的不是 `*Person`。这就保证了结构通过值传递，返回的也是值
3. 注意一下：我用了“With”而不是“Set”来表明重要的是返回值且原始对象并没有像调用 setter 那样被更改
4. 对同一个包下的代码来说，所有属性依然是可访问（也就可更改）的。我们绝不应该直接和属性交互，而是在同一个包下也应一直坚持使用方法
5. 每个 wither 返回的都是 `Person`，所以他们是可串联的

    ```go
    me := Person{}.
        WithName("Elliot").
        WithFavoriteColors([]string{"black", "blue"})

    fmt.Printf("%+#v\n", me)
    // main.Person{name:"Elliot", favoriteColors:[]string{"black", "blue"}}
    ```

## 处理切片

目前为止仍然不是完美的，因为对于最爱颜色我们返回的是切片。由于切片通过引用传递，我们来看看这么一个稍不留神就会忽略的 bug：

```go
func updateFavoriteColors(p Person) Person {
    colors := p.FavoriteColors()
    colors[0] = "red"

    return p
}

func main() {
    me := Person{}.
        WithName("Elliot").
        WithFavoriteColors([]string{"black", "blue"})

    me2 := updateFavoriteColors(me)

    fmt.Printf("%+#v\n", me)
    fmt.Printf("%+#v\n", me2)
}

// main.Person{name:"Elliot", favoriteColors:[]string{"red", "blue"}}
// main.Person{name:"Elliot", favoriteColors:[]string{"red", "blue"}}
```

我们想要改变第一种颜色，但是连带地改变了 `me` 变量。因为在复杂应用程序中这不会导致代码无法运行，试图搜寻出这么个变化是相当烦人和耗时的。

解决方法之一是确保我们绝不通过索引赋值，而是永远都是分配一个新的切片：

```go
func updateFavoriteColors(p Person) Person {
    return p.WithFavoriteColors(append([]string{"red"}, p.FavoriteColors()[1:]...))
}

// main.Person{name:"Elliot", favoriteColors:[]string{"black", "blue"}}
// main.Person{name:"Elliot", favoriteColors:[]string{"red", "blue"}}
```

在我看来，这有点拙而且容易出错。更好的方式是一开始就不返回切片。拓展我们的 getter 和 wither 来仅对元素操作（而不是整个切片）：

```go
func (p Person) NumFavoriteColors() int {
    return len(p.favoriteColors)
}

func (p Person) FavoriteColorAt(i int) string {
    return p.favoriteColors[i]
}

func (p Person) WithFavoriteColorAt(i int, favoriteColor string) Person {
    p.favoriteColors = append(p.favoriteColors[:i],
        append([]string{favoriteColor}, p.favoriteColors[i+1:]...)...)


    return p
}
```

> 译者注：上述代码是错误的，如果`p.favoriteColors`的容量大于`i`则会就地改变副本的`favoriteColors`，参见[反例](https://gist.github.com/sammyne/e845c24ad89ef04fd2207cdc6196e29f)，稍作调整即可得到[正确实现](https://gist.github.com/sammyne/d77be41112df33f53ccc40a20e5a605a#file-20181230-building-immutable-data-structures-in-go-ok-example-go-L11)

现在我们就可以放心使用：

```go
func updateFavoriteColors(p Person) Person {
    return p.WithFavoriteColorAt(0, "red")
}
```

想要了解更多切片的妙用参见这篇牛逼的wiki：https://github.com/golang/go/wiki/SliceTricks

## 构造函数

某些情况下，我们会假设结构体的默认值是合理的。但是，强烈建议总是创建构造函数，一旦将来需要改变默认值时，我们只需要改动一个地方：

```go
func NewPerson() Person {
    return Person{}
}
```

你可以随心所欲地实例化 `Person`，但个人偏爱总是通过 setter 来执行状态变换从而保持代码一致性：

```go
func NewPerson() Person {
    return Person{}.
        WithName("No Name")
}
```

## 接口 (Interface)

到现在为止，我们使用的还是公有的结构体。任由这些结构体方法摆布之下，加上创建 mock 可能会引发非预期的副作用，测试起来会很痛苦。

我们可以创建一个同名的接口，并把相应的结构体重命名为 `person` 使之私有化：

```go
type Person interface {
    WithName(name string) Person
    Name() string
    WithFavoriteColors(favoriteColors []string) Person
    NumFavoriteColors() int
    FavoriteColorAt(i int) string
    WithFavoriteColorAt(i int, favoriteColor string) Person
}

type person struct {
    name           string
    favoriteColors []string
}
```

我们现在就可以只重写想要替换的逻辑来创建测试 mock：

```go
type personMock struct {
    Person
    receivedNewColor string
}

func (m personMock) WithFavoriteColorAt(i int, favoriteColor string) Person {
    m.receivedNewColor = favoriteColor
    return m
}
```

测试代码样例如下：

```go
mock := personMock{}
result := updateFavoriteColors(mock)

result.(personMock).receivedNewColor // "red"
```

## 记录变化

如我早前所言，完整的状态转换非常有益于调试，而且我们可以 wither 来挂入钩子的方式捕捉到所有或部分变换过程：

```go
func (p person) nextState() Person {
    fmt.Printf("nextState: %#+v\n", p)
    return p
}

func (p person) WithName(name string) Person {
    p.name = name
    return p.nextState() // <- Use "nextState" whenever you return.
}
```

对于更加复杂的逻辑或个人偏好，你也可以采用 `defer` 的方式：

```go
func (p person) WithFavoriteColors(favoriteColors []string) Person {
    defer func() {
        p.nextState()
    }()

    p.favoriteColors = favoriteColors
    return p
}
```

这样变换就可看到了：

```bash
nextState: main.person{name:"No Name", favoriteColors:[]string(nil)}
nextState: main.person{name:"Elliot", favoriteColors:[]string(nil)}
nextState: main.person{name:"Elliot", favoriteColors:[]string{"black", "blue"}}
```

你可以添加更多诸如此类的信息。例如，时间戳、栈追踪记录和其他自定义的上下文信息来使得调试更加容易。

## 历史及回滚

除了打印变化之外，我们还可以收集这些状态作为历史：

```go
type Person interface {
    // ...
    AtVersion(version int) Person
}

type person struct {
    // ...
    history        []person
}

func (p *person) nextState() Person {
    p.history = append(p.history, *p)
    return *p
}

func (p person) AtVersion(version int) Person {
    return p.history[version]
}

func main() {
    me := NewPerson().
        WithName("Elliot").
        WithFavoriteColors([]string{"black", "blue"})

    // We discard the result, but it will be put into the history.
    updateFavoriteColors(me)

    fmt.Printf("%s\n", me.AtVersion(0).Name())
    fmt.Printf("%s\n", me.AtVersion(1).Name())
}

// No Name
// Elliot
```

这非常利于最后进行审查。记录所有日志打印的历史对处理后续异常的场景也是很有用的，如果不需要的话，让历史随实例消亡即可。

---

via: http://elliot.land/post/building-immutable-data-structures-in-go

作者：[ELLIOT CHANCE](http://elliot.land/)
译者：[sammyne](https://github.com/sammyne)
校对：[zhoudingding](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
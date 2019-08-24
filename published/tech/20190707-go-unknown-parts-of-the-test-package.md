# Go：Test 包不为人知的一面

首发于：https://studygolang.com/articles/23050

![test package](https://github.com/studygolang/gctt-images/blob/master/20190707-go-unknown-parts-of-the-test-package/test-pkg.png?raw=true)

Go 被用得最频繁的命令我想应该是`go test`。然而，这个命令一些有趣的细节和用法可能你还不知道哟。下面让我们从测试本身讲起。

## 规避缓存的习惯用法

如果连续两次运行同一份测试且第一次完全通过的话，会发现测试只真正被运行了一次。事实上，所有测试都采用缓存机制来避免运行没有变化的测试样例。下面看`math`包的一个测试用例：

```bash
root@91bb4e4ab781:/usr/local/go/src# go test ./math/
ok     math   0.007s
root@91bb4e4ab781:/usr/local/go/src# go test ./math/
ok     math   (cached)
```

测试时，Go 不仅会检查测试的内容，还会检查环境变量和命令行参数。更新环境变量或添加标识符都会导致缓存失效：

```bash
go test ./math/ -v
[...]
=== RUN   ExampleRoundToEven
--- PASS: ExampleRoundToEven (0.00s)
PASS
ok   math 0.007s
```

再执行一次的话缓存就会生效了。缓存是测试内容、环境变量和命令行参数的哈希。一旦计算出来，这个缓存会转储到`$GOCACHE`指向的文件夹（Unix 系统下默认是`$XDG_CACHE_HOME`或`$HOME/.cache`）。清空这个文件夹也就会清空缓存。

关于标识符的话，如[文档](https://golang.org/cmd/go/#hdr-Test_packages)所述，并不是所有标识符都是可缓存的：

> 缓存匹配的规则为：测试涉及的二进制可执行文件一样，同时命令行标识符属于'可缓存的'测试标识符限定子集（包括`-cpu`，`-list`，`parallel`，`-run`，`-short`和`-v`等）。使用任何不属于可缓存范围的标识符或参数都会导致缓存失效。显式屏蔽缓存的习惯用法是采用`-count=1`标识符。

因为`count`规定测试必须执行的次数，因此`-count=1`显式地声明测试应该不多不少地只运行一次，使得这个标识符成为规避缓存的最优习惯用法。

再提一下：Go 1.12 之前可通过设置[`GOCACHE`](https://golang.org/doc/go1.12#gocache)环境变量`GOCACHE=off go test math/`的方式绕过缓存。

运行测试时，Go 会逐个包依次运行它们。Go 处理测试包名的方式也给测试提供了更多策略。

## 白盒测试 vs 黑盒测试

黑盒测试不触及内部结构，只能访问到导出的函数和结构，而白盒测试则允许深入到非导出函数的内部实现。两种方式都是 Go 原生支持的。以下是展示黑盒测试和白盒测试各自优势的简单程序：

```go
package deck

import (
	"errors"
	"math/rand"
)

var Empty = errors.New("Empty deck")

type Deck struct {
	cards    []uint8
	shuffled bool
}

func NewDeck(numbers uint8) *Deck {
	cards := make([]uint8, 0, numbers)
	for i := uint8(0); i < numbers; i++ {
		cards = append(cards, i+1)
	}

	d := Deck{cards: cards}

	return &d
}

func (d *Deck) Draw() (card uint8, err error) {
	if !d.shuffled {
		d.shuffle()
	}

	if len(d.cards) == 0 {
		return 0, Empty
	}
	card, d.cards = d.cards[0], d.cards[1:]

	return card, nil
}

func (d *Deck) shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
	d.shuffled = true
}
```

上述代码只是执行洗牌操作后让用户抽牌。黑盒测试确保牌组能够创建并依次抽取直到没牌。

```go
package deck_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"deck"
)

func TestDeckCanDrawCards(t *testing.T) {
	var num uint8 = 10

	d := deck.NewDeck(num)
	for i := uint8(0); i < num; i++ {
		_, err := d.Draw()
		assert.Nil(t, err)
	}
	_, err := d.Draw()
	assert.Equal(t, err, deck.Empty)
}
```

> 译者注：原文给出的`deck`包的`import`路径不对，已修正如上

编写黑盒测试的唯一要求是给包名加上`_test`后缀。这个包被看作不同于`deck`的包，所以无法访问到非导出的函数。Go 原生支持这个方式，编译器不会抱怨同一个文件夹下有两个不同包名。

白盒测试则检验牌组只会在第一次抽取时被洗一次：

```go
package deck

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeckShouldBeShuffledOnce(t *testing.T) {
	var num uint8 = 5

	d := NewDeck(num)
	assert.Equal(t, len(d.cards), int(num))
	assert.Equal(t, d.shuffled, false, "Deck should init as not shuffled")
	orderBefore := fmt.Sprint(d.cards)

	d.shuffle()
	assert.Equal(t, d.shuffled, true, "Deck has not been marked as shuffled")
	orderAfter := fmt.Sprint(d.cards)

	assert.NotEqual(t, orderBefore, orderAfter, "Deck once shuffled should have new card order")
}
```

上述测试采用和`deck`相同的包名，因此能够访问到非导出的函数。

但是，白盒测试有这么一个短板。黑盒测试导出的函数能够保证正确的结果不受包的内部实现影响。因此，我们能够自由地改变和优化内部实现而不破坏现有测试。而对于白盒测试，它是和内部实现绑定的，优化操作有可能会破坏现有测试。

接下来看一下这个测试包的其他特性：性能测试。

## 只执行一次性能测试

[Go 1.12](https://golang.org/doc/go1.12#testing)引入的`-benchtime=1x`、`-benchtime=10x`等允许性能测试只执行我们想要的次数。这个`-benchtime=1x`标识符对测试套件（test suite）是很有用的，它使得我们只需运行至少一次性能测试即可验证上一次变更是否破坏了现有代码。

Go 1.12 之前的`-benchtime=1ns`标识符也能起到相同的效果，它会指示1ns后跳出性能测试的循环。因为 1 ns 是最小的时间单位，所以性能测试只会运行一次。性能测试为我们汇报诸如执行操作的时间、所需内存或堆上的内存分配次数等指标。Go 1.13 更是允许我们获取更多想要的指标。

## 汇报自定义的指标

Go 1.13 引入的`ReportMetric`方法允许我们汇报自定义的指标。复用一下之前牌组的示例，并修改为：抽取第一张牌之前先把牌组随机洗多次，次数在 1 到 20 次。以下是汇报洗牌次数的性能测试：

```go
package deck

import (
	"testing"
)

func BenchmarkGC(b *testing.B) {
	b.ReportAllocs()
	shuffled := 0

	for i := 0; i < b.N; i++ {
		d := NewDeck(100)
		_, _ = d.Draw()
		shuffled += int(d.shuffled)
	}

	b.ReportMetric(float64(shuffled)/float64(b.N), "shuffle/op")
}
```

> 译者注：原文没有贴出修改后的代码，以下是译者理解的实现

```go
package deck

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"math/rand"
)

var Empty = errors.New("Empty deck")

type Deck struct {
	cards    []uint8
	shuffled uint
}

func NewDeck(numbers uint8) *Deck {
	cards := make([]uint8, 0, numbers)
	for i := uint8(0); i < numbers; i++ {
		cards = append(cards, i+1)
	}

	d := Deck{cards: cards}

	var n uint32
	binary.Read(crand.Reader, binary.LittleEndian, &n)
	for n = 1 + n%20; n > 0; n-- {
		d.shuffle()
	}

	return &d
}

func (d *Deck) Draw() (card uint8, err error) {
	if len(d.cards) == 0 {
		return 0, Empty
	}
	card, d.cards = d.cards[0], d.cards[1:]

	return card, nil
}

func (d *Deck) shuffle() {
	rand.Shuffle(len(d.cards), func(i, j int) {
		d.cards[i], d.cards[j] = d.cards[j], d.cards[i]
	})
	d.shuffled++
}
```

所得结果如下：

```bash
BenchmarkDeckWithRandomShuffle-8       88666        12389 ns/op            5.15 shuffle/op        144 B/op         2 allocs/op
PASS
ok     1.529s
```

如结果所示，Go 1.13 还引入另一项变更：用准确的次数取代之前大致的[`b.N`](https://tip.golang.org/pkg/testing/#B.N)。这个[CL](https://go-review.googlesource.com/c/go/+/112155/)降低了诸如 GC 等外界噪声对性能测试的影响，尤其利于时长非常短的测试。测试速度也得到了提升：

```bash
// go 1.13
BenchmarkDeckWithRandomShuffle-8      88666         12389 ns/op
PASS
ok     1.529s
// go 1.12
BenchmarkDeckWithRandomShuffle-8      100000        12765 ns/op
PASS
ok     1.890s
```

---
via: https://medium.com/@blanchon.vincent/go-unknown-parts-of-the-test-package-df8988b2ef7f

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[sammyne](https://github.com/sammyne)
校对：[zhoudingding](https://github.com/dingdingzhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
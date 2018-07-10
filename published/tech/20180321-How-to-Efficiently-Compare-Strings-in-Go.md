首发于：https://studygolang.com/articles/13532

# Go 中如何有效的比较字符串

当优化软件时字符串比较可能和你想的有些不同。特别是包括拆分跨 goroutines 的循环， 找到一个更快的哈希算法，或者一些听起来更科学的方式。当我们做出这样的修改时，会有一种成就感。然而, 字符串比较通常是信息传递中（in a pipeline）的最大瓶颈。下面的代码段经常被使用，但它是最糟糕的解决方案 (参见下面的 benchmarks)，并导致了实际问题。

```go
strings.ToLower(name) == strings.ToLower(othername)
```

这是一种很直接的写法。把字符串转换成小写，然后在比较。要理解为什么这是一个差的解决方案，你需要知道字符串是如何表示的，以及 `ToLower` 是如何工作的。但是首先，让我们讨论一下字符串比较中主要的使用场景，当使用 `==` 操作符时，我们得到最快和最优化的解决方案。通常 APIs 或类似的软件通常都会考虑这些使用场景。我们使用 `ToLower` 称之为 eature-complete。[^注1]。
[^注1] This is when we drop in ToLower and call it feature-complete.

在 Go 中，字符串是一系列*不可变*的 runes。Rune 是 Go 的一个术语，代表一个码点（Code Point）。你可以在 [Go blog](https://blog.golang.org/strings) 获取更多关于 Strings, bytes, runes 和 characters 的信息。 `ToLower` 是一个标准库函数循环处理字符串中的每个 rune 转换成小写，然后返回新的字符串。所以上面的代码在比较之前遍历了整个字符串。这就和字符串的长度十分相关了。下面的伪代码大概的展示了上面代码片段的复杂度。

注意：因为字符串是不可变的，`strings.ToLower` 为两个字符串分配了新的内存空间。这增加了时间复杂度，但是现在这不是我们的关注点。为了简化演示，下面的伪代码认为字符串是可变的。

```go
// Pseudo code
func CompareInsensitive(a, b string) bool {
    // loop over string a and convert every rune to lowercase
    for i := 0; i < len(a); i++ {  a[i] = unicode.ToLower(a[i])  }
    // loop over string b and convert every rune to lowercase
    for i := 0; i < len(b); i++ {  b[i] = unicode.ToLower(b[i])  }
    // loop over both a and b and return false if there is a mismatch
    for i := 0; i < len(a); i++ {
        if a[i] != b[i] {
            return false
        }
    }
    return true
}
```

时间复杂度是 O(n) `n` 是 `len(a) + len(b) + len(a)` 请看下面的例子：

```go
CompareInsensitive("fizzbuzz", "buzzfizz")
```

意味着我们需要循环 24 次来确定两个完全不相同的字符串不匹配。这是非常低效的，我们可以通过比较 `unicode.ToLower(a[0])` 和 `unicode.ToLower(b[0])` (伪代码)来区分这些字符串。因此，需要把这种情况考虑在内。

优化一下，我们可以去掉 `CompareInsensitive` 前面的两个循环，比较相应位置的每个字符。如果 runes 不相等，我们转换成小写再比较。如果仍然不相等，我们结束循环，认为两个字符串不相等。如果他们相等，就继续比较下一个 rune 直到结束或者发现不相等的地方。现在重写一下代码

```go
// Pseudo code
func CompareInsensitive(a, b string) bool {
    // a quick optimization. If the two strings have a different
    // length then they certainly are not the same
    if len(a) != len(b) {
        return false
    }

    for i := 0; i < len(a); i++ {
        // if the characters already match then we don't need to
        // alter their case. We can continue to the next rune
        if a[i] == b[i] {
            continue
        }
        if unicode.ToLower(a[i]) != unicode.ToLower(b[i]) {
            // the lowercase characters do not match so these
            // are considered a mismatch, break and return false
            return false
        }
    }
    // The string length has been traversed without a mismatch
    // therefore the two match
    return true
}
```

新函数效率更高。上限是一个字符串的长度而不是两个字符串的长度和。怎么看我们上面的比较？循环比较最多只有 8 次。甚至，如果第一个字符不同，那就只循环一次。我们优化使得比较操作减少了大约 20 倍！

幸运的是在 `strings` 包里面有这样的函数。叫做 `strings.EqualFold`。

## 性能测试

```
// When both strings are equal
BenchmarkEqualFoldBothEqual-8                   20000000               124 ns/op
BenchmarkToLowerBothEqual-8                     10000000               339 ns/op

// When both strings are equal until the last rune
BenchmarkEqualFoldLastRuneNotEqual-8            20000000               129 ns/op
BenchmarkToLowerLastRuneNotEqual-8              10000000               346 ns/op

// When both strings are distinct
BenchmarkEqualFoldFirstRuneNotEqual-8           300000000             11.2 ns/op
BenchmarkToLowerFirstRuneNotEqual-8             10000000               333 ns/op

// When both strings have a different case at rune 0
BenchmarkEqualFoldFirstRuneDifferentCase-8      20000000               125 ns/op
BenchmarkToLowerFirstRuneDifferentCase-8        10000000               433 ns/op

// When both strings have a different case in the middle
BenchmarkEqualFoldMiddleRuneDifferentCase-8     20000000               123 ns/op
BenchmarkToLowerMiddleRuneDifferentCase-8       10000000               428 ns/op
```
当字符串的第一个字符不同时的差异很惊人 (30x)。因为不需要循环比较两个字符串，而是只循环一次就直接返回 false。在每个情况中 `EqualFold` 都要比开始的比较方式好出几个量级。

## 这很重要么？

你可能认为 400 纳秒并不重要。大多数情况下你可能是对的。不管怎样，一些微小的优化处理像其他的处理一样简单。在这个例子中，要比原来的处理方式更简单。合格的工程师在日常工作中就会使用这些微小的优化处理方式。他们不会等到变成问题的时候才去优化软件，他们从开始的时候就写出优化的软件。就算是最优秀的工程师也不可能从 0 开始写出最优化的软件。不可能凭空想象出每个极端的案例然后优化它。并且，在我们提供给用户软件的时候，我们也无法预知用户的行为。不管怎样，在日常工作中加入这些简单的处理方式有益于延长软件的生命周期，预防将来可能不必要的瓶颈。就算那个瓶颈没什么影响，你也不会浪费你的付出。

---

via: https://www.digitalocean.com/community/questions/how-to-efficiently-compare-strings-in-go

作者：[blockloop](https://www.digitalocean.com/community/users/blockloop)
译者：[tyler2018](https://github.com/tyler2018)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

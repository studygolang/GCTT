已发布：https://studygolang.com/articles/12682

# Go 如何编写简洁测试 -- 表格驱动测试

表格驱动测试是一种编写易于扩展测试用例的测试方法。表格驱动测试在 Go 语言中很常见（并非唯一），以至于很多标准库<sup>[1](#reference)</sup>都有使用。表格驱动测试使用匿名结构体。

在这篇文章中我会告诉你如何编写表格驱动测试。继续使用 [errline repo](https://github.com/virup/errline) 这个项目，现在我们来为 `Wrap()` 函数添加测试。`Wrap()` 函数用于给一个 `error` 在调用位置添加文件名和行数的修饰。我们尤其需要测试其中计算文件的短名称的逻辑（以粗体表示部分）。最初的 `Wrap()` 函数如下：

<pre>
func Wrap(err error) error {
    if err == nil {
        return nil
    }
    // If error already has file line do not add it again.
    if _, ok := err.(*withFileLine); ok {
        return err
    }
    _, file, line, ok := runtime.Caller(calldepth)
    if !ok {
        file = "???"
        line = 0
    }
    <b>short := file
    for i := len(file) - 1; i > 0; i-- {
        if file[i] == '/' {
            short = file[i+1:]
            break
        }
    }
    file = short</b>
    return &withFileLine{err, file, line}
}
</pre>

为了测试短文件名计算的逻辑更加简便，我们将这部分逻辑提取出来作为函数 `getShortFilename()`。代码现在变成这样：

<pre>
func Wrap(err error) error {
    if err == nil {
        return nil
    }
    // If error already has file line do not add it again.
    if _, ok := err.(*withFileLine); ok {
        return err
    }
    _, file, line, ok := runtime.Caller(calldepth)
    if !ok {
        file = "???"
        line = 0
    }
    file = getShortFilename(file)
    return &withFileLine{err, file, line}
}

func <b>getShortFilename(file string)</b> string {
    short := file
    for i := len(file) - 1; i > 0; i-- {
        if file[i] == '/' {
            short = file[i+1:]
            break
        }
    }
    file = short
    return file
}
</pre>

通过重构代码使其便于测试是很常见的做法。

我们现在通过传递多个文件名参数来测试 `getShortFilename()`，验证其输出结果是否符合预期。

我们先从一个空的测试函数开始：

```go
func TestShortFilename(t *testing.T) {
}
```

紧接着，我们引入一个包含字段 `in` 和 `expected` 的匿名结构体（struct）。`in` 表示传递给 `getShortFilename()` 的参数，`expected` 则代表我们预期的返回结果。`tests` 是包含多个这样结构体的一个数组。

```go
func TestShortFilename(t *testing.T) {
    tests := []struct {
        in       string   // input
        expected string   // expected result
    }{
        {"???", "???"},
        {"filename.go", "filename.go"},
        {"hello/filename.go", "filename.go"},
        {"main/hello/filename.go", "filename.go"},
    }
}
```

有了这个，我们就能通过循环来实现我们的测试方法。

```go
func TestShortFilename(t *testing.T) {
    tests := []struct {
        in       string
        expected string
    }{
        {"???", "???"},
        {"filename.go", "filename.go"},
        {"hello/filename.go", "filename.go"},
        {"main/hello/filename.go", "filename.go"},
    }
    
    for _, tt := range tests {
        actual := getShortFilename(tt.in)
        if strings.Compare(actual, tt.expected) != 0 {
            t.Fail()
        }
    }
}
```

可以注意到，添加测试用例极其简单，只需在 `tests` 中添加项目即可。

这个方案可以扩展以适应于测试接受和返回多个参数的方法。

就这样了。

代码可以从 [我的 github](https://github.com/virup/errline/tree/master) 获取。

## <p id="reference">引用</p>

1. 一些 Go 语言标准库的表格驱动测试例子
    * [https://github.com/golang/go/blob/master/src/strconv/ftoa_test.go](https://github.com/golang/go/blob/master/src/strconv/ftoa_test.go)
    * [https://github.com/golang/go/blob/master/src/path/match_test.go](https://github.com/golang/go/blob/master/src/path/match_test.go)
    * [https://github.com/golang/go/blob/master/src/archive/tar/strconv_test.go](https://github.com/golang/go/blob/master/src/archive/tar/strconv_test.go)


----------------

via: https://medium.com/@virup/how-to-write-concise-tests-table-driven-tests-ed672c502ae4

作者：[Viru](https://medium.com/@virup)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
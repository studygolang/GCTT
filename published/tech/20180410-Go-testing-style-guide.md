首发于：https://studygolang.com/articles/14406

# Go 测试风格指南

Go 测试的一个小风格（自认为）指南。关于写好测试的文章比我在这写的要多的多。但我写的主要是关于风格而不是技术。

## 使用 table-drive 测试，并始终使用 tt 作为测试用例

尝试在可行的情况下使用 table-driven 测试，但当不可行时，可以复制一些代码；不要强制使用它（例如，有时候除了一两个案例之外，更容易为这之外的情况编写一个 table-driven 的测试；实际情况就是如此）。

始终为一个测试用例使用相同变量名会使它更容易为大量代码工作。你不必使用 tt，但是在 Go 标准库中它是最常用的（ 564 次对比 tc 用了 116 次）。

可以看看 [TableDrivenTests](https://github.com/golang/go/wiki/TableDrivenTests)。

例如：

```go
tests := []struct {
	// ...
}{}

for _, tt := range tests {
}
```

## 使用子测试

使用子测试可以从 table 中运行一个单独的测试，且可以容易的看出哪个测试完全失败了。由于子测试是比较新的版本（ Go 1.7，2016年10月），所以许多现存的测试不能使用它们（子测试）。

如果测试内容很明显，我倾向于简单地使用测试编号；如果不明显或有很多测试用例，就添加一个测试名。

可以看下[使用子测试和子基准](https://blog.golang.org/subtests)

例如：

```go
tests := []struct {
	// ...
}{}

for i, tt := range tests {
	t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
		got := TestFunction(tt.input)
		if got != tt.want {
			t.Errorf("failed for %v ...", tt.input)
		}
	})
}
```

## 不要忽略错误

我经常看到在测试中有人忽略错误。这是不好的想法并且使失败的测试混乱。

例如：

```go
got, err := Fun()
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
```

或者：

```go
got, err := Fun()
if err != tt.wantErr {
    t.Fatalf("wrong error\ngot:  %v\nwant: %v", err, tt.wantErr)
}
```

我经常使用 [ErroContains](https://github.com/Teamwork/test/blob/859eda3cd87ed7713df79c5bb2b2a90601ad0524/test.go#L13-L26)，它是一个很有用的帮助函数对测试错误信息（避免一些 `if err != nil && [..]`）。

## 检查你的测试作为常规代码

测试代码也会失败，错误，所以需要维护。如果你认为运行 linter 对你的正规代码是值得的，那么对你的测试运行也是一样值得的。（例如：go vet, errcheck 等）。

## 使用 want 和 got

want 比 expected 短，got 比 actual 短。短命名总是有优势的，IMHO，并且特别有利于对齐输出（看下面的例子）。

例如：

```go
cases := []struct {
    want     string
    wantCode int
}{}
```

## 添加有用的，可对齐的信息

当一个测试失败伴随着无用的错误信息，或者是一个混乱的错误信息，使你很难看出准确的错误时是很恼人的。

这不是特别有用：

```go
t.Errorf("wrong output: %v", got)
```

当测试失败时，它告诉我们得到了错误输出，但是我们想要得到的是什么呢？

这个就比较好：

```go
name := "test foo"
got := "this string!"
want := "this string"
t.Errorf("wrong output for %v, want %v; got %v", name, got, want)
```

下面这个很难看到准确的失败：

```
--- FAIL: TestX (0.00s)
		a_test.go:9: wrong output for test foo, want this string!; got this string
```

当把它对齐，就很容易了：

```go
name := "test foo"
want := "this string"
t.Run(name, func(t *testing.T) {
    got := "this string!"
    t.Errorf("wrong output\ngot:  %q\nwant: %q", got, want)
})
```

```
--- FAIL: TestX (0.00s)
	--- FAIL: TestX/test_foo (0.00s)
		a_test.go:10: wrong output
				got:  "this string!"
				want: "this string"
```

注意 `got:` 后面的俩个空格，是为了和 `want` 对齐的。如果我使用 `expected` 就要使用6个空格。

我还倾向于使用 `%q` 或 `%#v`，因为这会很清楚的显示后面的空白或不可打印字符。

使用 diff 比较较大的对象；例如用 [go-cmp](https://github.com/google/go-cmp)：

```go
if d := cmp.Diff(got, tt.want); d != "" {
	t.Errorf("(-got +want)\n:%s", d)
}
```

```
--- FAIL: TestParseFilter (0.00s)
	--- FAIL: TestParseFilter/alias (0.00s)
		query_test.go:717: (-got +want)
			:{jsonapi.Filter}.Alias:
				-: "fail"
				+: "alias"

```

## 搞清楚要测试什么

有时我看到一些测试，我困惑 “这是在测试什么？” 如果测试失败的原因不名这会令人特别困惑。应该改哪？测试正确吗？

例如：

```go
cases := []struct {
	name string
}{
	{
		"space after @",
	},
	{
		"unicode space before @",
	},
	// ...
}
```

如果添加 `name` 到存在的测试用例一定比注释要更有用。

## 反馈

你可以给我发邮件: [martin@arp242.ent](martin@arp242.net) 或者 [提交一个 GitHub 问题](https://github.com/Carpetsmoker/arp242.net/issues/new)反馈，提问等。

---

via: https://arp242.net/weblog/go-testing-style.html

作者：[Martin Tournoij](https://arp242.net/)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

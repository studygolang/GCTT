首发于：https://studygolang.com/articles/25296

# 类似 Go 中的表格驱动测试的步骤驱动评估

如果你听说过表驱动测试，那你就能更容易理解本文所描述的概念，因为它们使用的是相同的技术，只不过本文使用在非测试场景中。

假设你有一个函数，该函数中调用了多个其他函数。那么这个函数很可能主要有两个作用：

1. 检查出现的所有错误返回。
2. 传递一个函数的输出作为另一个函数的输入。

```go
// process is an example pipeline-like function.
func queryFile(filename, queryText string) (string, error) {
	data, err := readData(filename)
	if err != nil {
		return nil, errors.Errorf("read data: %v", err)
	}
	rows, err := splitData(data)
	if err != nil {
		return nil, errors.Errorf("split data: %v", err)
	}
	q, err := compileQuery(queryText)
	if err != nil {
		return nil, errors.Errorf("compile query: %v", err)
	}
	rows, err = filterRows(rows, q)
	if err != nil {
		return nil, errors.Errorf("filter rows: %v", err)
	}
	result, err := rowsToString(rows)
	if err != nil {
		return nil, errors.Errorf("rows to string: %v", err)
	}
	return result, nil
}
```

这个函数包含了 5 个步骤。准确地说是 5 个相关的调用，其他所有的一切都不是重点。在这个算法中，函数的调用是有序的。

让我们使用步骤驱动评估算法重写上面的代码。

```go
func queryFile(filename, queryText string) ([]row, error) {
	var ctx queryFileContext
	steps := []struct {
		name string
		fn   func() error
	}{
		{"read data", ctx.readData},
		{"split data", ctx.splitData},
		{"compile query", ctx.compileQuery},
		{"filter rows", ctx.filterRows},
		{"rows to string", ctx.rowsToString},
	}
	for _, step := range steps {
		if err := step.fn(); err != nil {
			return errors.Errorf("%s: %v", step.name, err)
		}
	}
	return ctx.result
}
```

这种管道式的做法使得代码清晰、明确，也便于调整步骤的顺序、新增或者移除某些步骤。另外，在循环体中增加调试日志也非常的简单，你只需要在循环程序中新加一个声明语句就可以，不需要像一开始那样，在每个函数调用的地方都要增加声明语句。

当引入一个新类型的复杂性低于其带来的收益时，在 4 个或更多步骤的情况下，这种方法表现亮眼。

```go
// queryFileContext might look like the struct below.

type queryFileContext struct {
	data   []byte
	rows   []row
	q      *query
	result string
}
```

诸如方法 queryFileContext.splitData，仅调用了相同的函数并同时更新对象 ctx 的状态。

```go
func (ctx *queryFileContext) splitData() error {
	var err error
	ctx.rows, err = splitData(ctx.data)
	return err
}
```

main 函数特别适合本文这种，能使各个步骤清晰明确的、适合 4+ 个步骤以上的方法。

```go
func main() {
	ctx := &context{}

	steps := []struct {
		name string
		fn   func() error
	}{
		{"parse flags", ctx.parseFlags},
		{"read schema", ctx.readSchema},
		{"dump schema", ctx.dumpSchema}, // Before transformations
		{"remove builtin constructors", ctx.removeBuiltinConstructors},
		{"add adhoc constructors", ctx.addAdhocConstructors},
		{"validate schema", ctx.validateSchema},
		{"decompose arrays", ctx.decomposeArrays},
		{"replace arrays", ctx.replaceArrays},
		{"resolve generics", ctx.resolveGenerics},
		{"dump schema", ctx.dumpSchema}, // After transformations
		{"decode combinators", ctx.decodeCombinators},
		{"dump decoded combinators", ctx.dumpDecodedCombinators},
		{"codegen", ctx.codegen},
	}

	for _, step := range steps {
		ctx.debugf("start %s step", step.name)
		if err := step.fn(); err != nil {
			log.Fatalf("%s: %v", step.name, err)
		}
	}
}
```

另外一个好处就是使得测试更加简单。即使我们要使用到函数 log.Fatalf，虽然这不是一个好的做法，在一个测试方法中，能够很容易的重新开启一个测试流程，同时，能够执行一系列失败的测试案例，而无需调用 os.Exit。

你也可以忽略测试中一些与 CLI 相关的步骤，比如“dump schema” 或者 “codegen”。你也可以在列表中插入测试专用的步骤。

这个方法也有一些缺点，比如：

1. 你必须要定义新的类型和方法。
2. 并不是总能很直接的找到合适的上下文对象，它只能适用于那些不会使得整个流程变得过于复杂的场景。

试着用用这个方法，也许你会喜欢上它的。

---

via: https://quasilyte.dev/blog/post/step-pattern/

作者：[Iskander Sharipov](https://github.com/quasilyte)
译者：[yangzhenxiong](https://github.com/yangzhenxiong)
校对：[DingdingZhou](https://github.com/DingdingZhou)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

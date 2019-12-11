首发于：https://studygolang.com/articles/25100

# Go 标准库 `encoding/json` 真的慢吗？

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-is-the-encoding-json-package-really-slow/A-Journey-With-Go.png)

插图来自于“A Journey With Go”，由 Go Gopher 组织成员 Renee French 创作。

本文基于 Go 1.12。

关于标准库 `encoding/json` 性能差的问题在很多地方被讨论过，也有很多第三方库在尝试解决这个问题，比如[easyjson](https://github.com/mailru/easyjson)，[jsoniter](https://github.com/json-iterator/go)和[ffjson](https://github.com/pquerna/ffjson)。但是标准库 `encoding/json` 真的慢吗？它一直都这么慢吗？

## 标准库 `encoding/json` 的进化之路

首先，通过一个简短的 makefile 文件和一段基准测试代码，我们看下在各个 Go 版本中，标准库 `encoding/json` 的性能表现。以下为基准测试代码：

```go
type JSON struct {
   Foo int
   Bar string
   Baz float64
}

func BenchmarkJsonMarshall(b *testing.B) {
   j := JSON{
      Foo: 123,
      Bar: `benchmark`,
      Baz: 123.456,
   }
   b.ResetTimer()
   for i := 0; i < b.N; i++ {
      _, _ = json.Marshal(&j)
   }
}

func BenchmarkJsonUnmarshal(b *testing.B) {
   bytes := `{"foo": 1, "bar": "my string", bar: 1.123}`
   str := []byte(bytes)
   b.ResetTimer()
   for i := 0; i < b.N; i++ {
      j := JSON{}
      _ = json.Unmarshal(str, &j)
   }
}
```

makefile 文件在不同的文件夹中基于不同版本的 Go 创建 Docker 镜像，在各镜像启动的容器中运行基准测试。将从以下两个维度进行性能对比：

* 比较 Go 各版本与 1.12 版本中标准库 `encoding/json` 的性能差异
* 比较 Go 各版本与其下一个版本中标准库 `encoding/json` 的性能差异

第一个维度的对比可以得到在特定版本的 Go 与 1.12 版本的 Go 中 json 序列化和反序列化的性能差异；第二个维度的对比可以得到在哪次 Go 版本升级中 json 序列化和反序列化发生了最大的性能提升。

测试结果如下：

* Go1.2 至 Go1.3 的版本升级，序列化操作耗时减少了约 28%，反序列化操作耗时减少了约 35%

```bash
name           old time/op    new time/op    delta
JsonMarshall     1.91 µ s ± 2%    1.37 µ s ± 2%  -28.23%
JsonUnmarshal    2.70 µ s ± 2%    1.75 µ s ± 3%  -35.18%
```

* Go1.6 至 Go1.7 的版本升级，序列化操作耗时减少了约 27%，反序列化操作耗时减少了约 40%

```bash
name             old time/op    new time/op    delta
JsonMarshall-4     1.24 µ s ± 1%    0.90 µ s ± 2%  -27.65%
JsonUnmarshal-4    1.52 µ s ± 3%    0.91 µ s ± 2%  -40.05%
```

* Go1.10 至 Go1.11 的版本升级，序列化内存消耗减少了约 60%，反序列化内存消耗减少了约 25%

```bash
name             old alloc/op   new alloc/op   delta
JsonMarshall-4       208B ± 0%       80B ± 0%  -61.54%
JsonUnmarshal-4      496B ± 0%      368B ± 0%  -25.81%
```

* Go1.11 至 Go1.12 的版本升级，序列化操作耗时减少了约 15%，反序列化操作耗时减少了约 6%

```bash
name             old time/op    new time/op    delta
JsonMarshall-4      670ns ± 6%     569ns ± 2%  -15.09%
JsonUnmarshal-4     800ns ± 1%     747ns ± 1%   -6.58%
```

可以在这里看到完整的[测试结果](https://gist.github.com/blanchonvincent/227b6691777a1de254ce75b304a36277)。

如果对比 Go1.2 与 Go1.12，会发现标准库 `encoding/json` 的性能有显著提高，操作耗时减少了约 69%/68%，内存消耗减少了约 74%/29%：

```bash
name           old time/op    new time/op    delta
JsonMarshall     1.72 µ s ± 2%    0.52 µ s ± 2%  -69.68%
JsonUnmarshal    2.72 µ s ± 2%    0.85 µ s ± 5%  -68.70%

name           old alloc/op   new alloc/op   delta
JsonMarshall       188B ± 0%       48B ± 0%  -74.47%
JsonUnmarshal      519B ± 0%      368B ± 0%  -29.09%
```

该基准测试使用了较为简单的 json 结构。使用更加复杂的结构（例如：Map or Array）进行测试会导致各版本之间性能增幅与本文不同。

## 速读源码

想了解标准库性能较差的原因的最好的办法就是读源码，以下为 Go1.12 版本中 `json.Marshal` 函数的执行流程：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/go-is-the-encoding-json-package-really-slow/json-marshal.png)

在了解了 `json.Marshal` 函数的执行流程后，再来比较下在 Go1.10 和 Go1.12 版本中的 `json.Marshal` 函数在实现上有什么变化。通过之前的测试，可以发现从 Go1.10 至 Go1.12 版本中的 `json.Marshal` 函数的内存消耗上有了很大的改善。从源码的变化中可以发现在 Go1.12 版本中的 `json.Marshal` 函数添加了 encoder（编码器）的内存缓存：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/go-is-the-encoding-json-package-really-slow/json-marshal-diff.png)

在使用了 `sync.Pool` 缓存 encoder 后，`json.Marshal` 函数极大地减少了内存分配操作。实际上 `newEncodeState()` 函数在 Go1.10 版本中就[已经存在了](https://github.com/golang/go/commit/c0547476f342665514904cf2581a62135d2366c3#diff-e79d4db81e8544657cb631be813f89b4)，只不过没有被使用。为验证是添加了内存缓存带来了性能提升的猜想，可以在 Go1.10 版本中修改 `json.Marshal` 函数后，再进行测试：

```bash
name           old alloc/op   new alloc/op   delta
CodeMarshal-4    4.59MB ± 0%    1.98MB ± 0%  -56.92%
```

可以直接在[Go 源码](https://github.com/golang/go)中，执行以下命令进行基准测试：

```bash
go test encoding/json -bench=BenchmarkCodeMarshal -benchmem -count=10 -run=^$
```

结果和我们的猜想是一致的。是[sync 包](https://golang.org/pkg/sync/)给 `json.Marshal` 函数带来了性能提升。同样也给我们带来一点启发，当项目也有这种对同一个结构体进行大量的内存分配时，也可以通过添加内存缓存的方式提升性能。

以下为 Go1.12 版本中，`json.Unmarshal` 函数的执行流程：

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/go-is-the-encoding-json-package-really-slow/json-unmarshal.png)

`json.Unmarshal` 函数同样使用 `sync.Pool` 缓存了 decoder。
对于 json 序列化和反序列化而言，其性能瓶颈是迭代、反射 json 结构中每个字段。

## 与第三方库的性能对比

GitHub 上也有很多用于 json 序列化的第三方库，比如[ffjson](https://github.com/pquerna/ffjson)就是其中之一，ffjson 的命令行工具可以为指定的结构生成静态的 MarshalJSON 和 UnmarshalJSON 函数，MarshalJSON 和 UnmarshalJSON 函数在序列化和反序列化操作时会分别被 `ffjson.Marshal` 和 `ffjson.Unmarshal` 函数调用。以下为 ffjson 生成的解析器示例：

```go
func (j *JSONFF) MarshalJSON() ([]byte, error) {
   var buf fflib.Buffer
   if j == nil {
      buf.WriteString("null")
      return buf.Bytes(), nil
   }
   err := j.MarshalJSONBuf(&buf)
   if err != nil {
      return nil, err
   }
   return buf.Bytes(), nil
}

// MarshalJSONBuf marshal buff to JSON - template
func (j *JSONFF) MarshalJSONBuf(buf fflib.EncodingBuffer) error {
   if j == nil {
      buf.WriteString("null")
      return nil
   }
   var err error
   var obj []byte
   _ = obj
   _ = err
   buf.WriteString(`{"Foo":`)
   fflib.FormatBits2(buf, uint64(j.Foo), 10, j.Foo < 0)
   buf.WriteString(`,"Bar":`)
   fflib.WriteJsonString(buf, string(j.Bar))
   buf.WriteString(`,"Baz":`)
   fflib.AppendFloat(buf, float64(j.Baz), 'g', -1, 64)
   buf.WriteByte('}')
   return nil
}
```

现在比较一下标准库和 ffjson（使用了 `ffjson.Pool()`）的性能差异：

```bash
standard lib:
name             time/op
JsonMarshall-4   500ns ± 2%
JsonUnmarshal-4  677ns ± 2%

name             alloc/op
JsonMarshall-4   48.0B ± 0%
JsonUnmarshal-4   320B ± 0%

ffjson:
name               time/op
JsonMarshallFF-4   538ns ± 1%
JsonUnmarshalFF-4  827ns ± 3%

name               alloc/op
JsonMarshallFF-4    176B ± 0%
JsonUnmarshalFF-4   448B ± 0%
```

对于 json 序列化/反序列化，标准库与 ffjson 相比反而更加高效一些。

对于内存使用情况（堆分配），可以通过 `go run -gcflags="-m"` 命令进行测试：

```bash
:46:19: buf escapes to heap
:48:23: buf escapes to heap
:27:26: &buf escapes to heap
:22:6: moved to heap: buf
```

[easyjson](https://github.com/mailru/easyjson)库也使用了和 ffjson 同样的策略，以下为基准测试结果：

```bash
standard lib:
name             time/op
JsonMarshall-4   500ns ± 2%
JsonUnmarshal-4  677ns ± 2%

name             alloc/op
JsonMarshall-4   48.0B ± 0%
JsonUnmarshal-4   320B ± 0%

easyjson:
name               time/op
JsonMarshallEJ-4   349ns ± 1%
JsonUnmarshalEJ-4  341ns ± 5%

name               alloc/op
JsonMarshallEJ-4    240B ± 0%
JsonUnmarshalEJ-4   256B ± 0%
```

这次，easyjson 比标准库更高效些，对于 json 序列化有 30%的性能提升，对于 json 反序列化性能提升接近 2 倍。通过阅读 `easyjson.Marshal` 的源码，可以发现它高效的原因：

```go
func Marshal(v Marshaler) ([]byte, error) {
   w := jwriter.Writer{}
   v.MarshalEasyJSON(&w)
   return w.BuildBytes()
}
```

通过 easyjson 的命令行工具生成的编码器 `MarshalEasyJSON` 方法可用于 json 序列化：

```go
func easyjson42239ddeEncode(out *jwriter.Writer, in JSON) {
   out.RawByte('{')
   first := true
   _ = first
   {
      const prefix string = ",\"Foo\":"
      if first {
         first = false
         out.RawString(prefix[1:])
      } else {
         out.RawString(prefix)
      }
      out.Int(int(in.Foo))
   }
   {
      const prefix string = ",\"Bar\":"
      if first {
         first = false
         out.RawString(prefix[1:])
      } else {
         out.RawString(prefix)
      }
      out.String(string(in.Bar))
   }
   {
      const prefix string = ",\"Baz\":"
      if first {
         first = false
         out.RawString(prefix[1:])
      } else {
         out.RawString(prefix)
      }
      out.Float64(float64(in.Baz))
   }
   out.RawByte('}')
}

func (v JSON) MarshalEasyJSON(w *jwriter.Writer) {
   easyjson42239ddeEncode(w, v)
}
```

正如我们所见，这里没有使用反射。整体流程也很简单。而且，easyjson 也可以兼容标准库：

```go
func (v JSON) MarshalJSON() ([]byte, error) {
   w := jwriter.Writer{}
   easyjson42239ddeEncodeGithubComMyCRMTeamEncodingJsonEasyjson(&w, v)
   return w.Buffer.BuildBytes(), w.Error
}
```

然而，使用这种兼容标准库的方式进行序列化会比直接使用标准库性能更差，因为在进行 json 序列化的过程中，标准库依然会通过反射构造 encoder，且 `MarshalJSON` 中这一段代码也会被执行。

## 结论

无论在标准库上做多少努力，它都不会比通过**对明确的 json 结构生成 encoder/decoder**的方式性能好。而通过结构生成解析器代码的方式需要生成和维护此代码，并且依赖于外部的库。

在做出使用第三方序列化库替换标准库的决定前，最好先测试下 json 序列化和反序列化是否是应用的性能瓶颈点，提高 json 序列化的效率是否能改善应用的性能。如果 json 序列化和反序列化并不是应用的性能瓶颈点，为了极少的性能提升，付出第三方库的维护成本是不值得的。毕竟，在大多数业务场景下，Go 的标准库 `encoding/json` 已经足够高效了。

---

via: https://medium.com/a-journey-with-go/go-is-the-encoding-json-package-really-slow-62b64d54b148

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[beiping96](https://github.com/beiping96)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

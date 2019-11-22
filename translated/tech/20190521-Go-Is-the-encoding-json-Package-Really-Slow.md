# Go标准库`encoding/json`真的慢吗？

![](https://github.com/beiping96/gctt-images2/blob/master/go-is-the-encoding-json-package-really-slow/A-Journey-With-Go.png)
插图来自于“A Journey With Go”，由Go Gopher组织成员Renee French创作。

本文基于Go 1.12。
关于标准库`encoding/json`性能差的问题在很多地方被讨论过，也有很多第三方库在尝试解决这个问题，比如[easyjson](https://github.com/mailru/easyjson)，[jsoniter](https://github.com/json-iterator/go)和[ffjson](https://github.com/pquerna/ffjson)。但是标准库`encoding/json`真的慢吗？它一直都这么慢吗？

## 标准库`encoding/json`的进化之路

首先，通过一个简短的makefile文件和一段基准测试代码，我们看下在各个Go版本中，标准库`encoding/json`的性能表现。以下为基准测试代码：

``` go
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

makefile文件在不同的文件夹中基于不同版本的Go创建Docker镜像，在各镜像启动的容器中运行基准测试。将从以下两个维度进行性能对比：
* 比较Go各版本与1.12版本中标准库`encoding/json`的性能差异
* 比较Go各版本与其下一个版本中标准库`encoding/json`的性能差异

第一个维度的对比可以得到在特定版本的Go与1.12版本的Go中json序列化和反序列化的性能差异；第二个维度的对比可以得到在哪次Go版本升级中json序列化和反序列化发生了最大的性能提升。

测试结果如下：
* Go1.2至Go1.3的版本升级，序列化操作耗时减少了约28%，反序列化操作耗时减少了约35%

``` bash
name           old time/op    new time/op    delta
JsonMarshall     1.91µs ± 2%    1.37µs ± 2%  -28.23%
JsonUnmarshal    2.70µs ± 2%    1.75µs ± 3%  -35.18%
```

* Go1.6至Go1.7的版本升级，序列化操作耗时减少了约27%，反序列化操作耗时减少了约40%

``` bash
name             old time/op    new time/op    delta
JsonMarshall-4     1.24µs ± 1%    0.90µs ± 2%  -27.65%
JsonUnmarshal-4    1.52µs ± 3%    0.91µs ± 2%  -40.05%
```

* Go1.10至Go1.11的版本升级，序列化内存消耗减少了约60%，反序列化内存消耗减少了约25%

``` bash
name             old alloc/op   new alloc/op   delta
JsonMarshall-4       208B ± 0%       80B ± 0%  -61.54%
JsonUnmarshal-4      496B ± 0%      368B ± 0%  -25.81%
```

* Go1.11至Go1.12的版本升级，序列化操作耗时减少了约15%，反序列化操作耗时减少了约6%

``` bash
name             old time/op    new time/op    delta
JsonMarshall-4      670ns ± 6%     569ns ± 2%  -15.09%
JsonUnmarshal-4     800ns ± 1%     747ns ± 1%   -6.58%
```

可以在这里看到完整的[测试结果](https://gist.github.com/blanchonvincent/227b6691777a1de254ce75b304a36277)。

如果对比Go1.2与Go1.12，会发现标准库`encoding/json`的性能有显著提高，操作耗时减少了约69%/68%，内存消耗减少了约74%/29%：

``` bash
name           old time/op    new time/op    delta
JsonMarshall     1.72µs ± 2%    0.52µs ± 2%  -69.68%
JsonUnmarshal    2.72µs ± 2%    0.85µs ± 5%  -68.70%

name           old alloc/op   new alloc/op   delta
JsonMarshall       188B ± 0%       48B ± 0%  -74.47%
JsonUnmarshal      519B ± 0%      368B ± 0%  -29.09%
```

该基准测试使用了较为简单的json结构。使用更加复杂的结构（例如：Map or Array）进行测试会导致各版本之间性能增幅与本文不同。

## 速读源码

想了解标准库性能较差的原因的最好的办法就是读源码，以下为Go1.12版本中`json.Marshal`函数的执行流程：

![](https://github.com/beiping96/gctt-images2/blob/master/go-is-the-encoding-json-package-really-slow/json-marshal.png)

在了解了`json.Marshal`函数的执行流程后，再来比较下在Go1.10和Go1.12版本中的`json.Marshal`函数在实现上有什么变化。通过之前的测试，可以发现从Go1.10至Go1.12版本中的`json.Marshal`函数的内存消耗上有了很大的改善。从源码的变化中可以发现在Go1.12版本中的`json.Marshal`函数添加了encoder（编码器）的内存缓存：

![](https://github.com/beiping96/gctt-images2/blob/master/go-is-the-encoding-json-package-really-slow/json-marshal-diff.png)

在使用了`sync.Pool`缓存encoder后，`json.Marshal`函数极大地减少了内存分配操作。实际上`newEncodeState()`函数在Go1.10版本中就[已经存在了](https://github.com/golang/go/commit/c0547476f342665514904cf2581a62135d2366c3#diff-e79d4db81e8544657cb631be813f89b4)，只不过没有被使用。为验证是添加了内存缓存带来了性能提升的猜想，可以在Go1.10版本中修改`json.Marshal`函数后，再进行测试：

``` bash
name           old alloc/op   new alloc/op   delta
CodeMarshal-4    4.59MB ± 0%    1.98MB ± 0%  -56.92%
```

可以直接在[Go源码](https://github.com/golang/go)中，执行以下命令进行基准测试：

``` bash
go test encoding/json -bench=BenchmarkCodeMarshal -benchmem -count=10 -run=^$
```

结果和我们的猜想是一致的。是[sync包](https://golang.org/pkg/sync/)给`json.Marshal`函数带来了性能提升。同样也给我们带来一点启发，当项目也有这种对同一个结构体进行大量的内存分配时，也可以通过添加内存缓存的方式提升性能。

以下为Go1.12版本中，`json.Unmarshal`函数的执行流程：

![](https://github.com/beiping96/gctt-images2/blob/master/go-is-the-encoding-json-package-really-slow/json-unmarshal.png)

`json.Unmarshal`函数同样使用`sync.Pool`缓存了decoder。
对于json序列化和反序列化而言，其性能瓶颈是迭代、反射json结构中每个字段。

## 与第三方库的性能对比

GitHub上也有很多用于json序列化的第三方库，比如[ffjson](https://github.com/pquerna/ffjson)就是其中之一，ffjson的命令行工具可以为指定的结构生成静态的MarshalJSON和UnmarshalJSON函数，MarshalJSON和UnmarshalJSON函数在序列化和反序列化操作时会分别被`ffjson.Marshal`和`ffjson.Unmarshal`函数调用。以下为ffjson生成的解析器示例：

``` go
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

// MarshalJSONBuf marshal buff to json - template
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

现在比较一下标准库和ffjson（使用了`ffjson.Pool()`）的性能差异：

``` bash
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

对于json序列化/反序列化，标准库与ffjson相比反而更加高效一些。

对于内存使用情况（堆分配），可以通过`go run -gcflags="-m"`命令进行测试：

``` bash
:46:19: buf escapes to heap
:48:23: buf escapes to heap
:27:26: &buf escapes to heap
:22:6: moved to heap: buf
```

[easyjson](https://github.com/mailru/easyjson)库也使用了和ffjson同样的策略，以下为基准测试结果：

``` bash
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

这次，easyjson比标准库更高效些，对于json序列化有30%的性能提升，对于json反序列化性能提升接近2倍。通过阅读`easyjson.Marshal`的源码，可以发现它高效的原因：

``` go
func Marshal(v Marshaler) ([]byte, error) {
   w := jwriter.Writer{}
   v.MarshalEasyJSON(&w)
   return w.BuildBytes()
}
```

通过easyjson的命令行工具生成的编码器`MarshalEasyJSON`方法可用于json序列化：

``` go
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

正如我们所见，这里没有使用反射。整体流程也很简单。而且，easyjson也可以兼容标准库：

``` go
func (v JSON) MarshalJSON() ([]byte, error) {
   w := jwriter.Writer{}
   easyjson42239ddeEncodeGithubComMyCRMTeamEncodingJsonEasyjson(&w, v)
   return w.Buffer.BuildBytes(), w.Error
}
```

然而，使用这种兼容标准库的方式进行序列化会比直接使用标准库性能更差，因为在进行json序列化的过程中，标准库依然会通过反射构造encoder，且`MarshalJSON`中这一段代码也会被执行。

## 结论

无论在标准库上做多少努力，它都不会比通过**对明确的json结构生成encoder/decoder**的方式性能好。而通过结构生成解析器代码的方式需要生成和维护此代码，并且依赖于外部的库。

在做出使用第三方序列化库替换标准库的决定前，最好先测试下json序列化和反序列化是否是应用的性能瓶颈点，提高json序列化的效率是否能改善应用的性能。如果json序列化和反序列化并不是应用的性能瓶颈点，为了极少的性能提升，付出第三方库的维护成本是不值得的。毕竟，在大多数业务场景下，Go的标准库`encoding/json`已经足够高效了。

---

via: https://medium.com/a-journey-with-go/go-is-the-encoding-json-package-really-slow-62b64d54b148

作者：[Vincent Blanchon](https://medium.com/@blanchon.vincent)
译者：[beiping96](https://github.com/beiping96)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

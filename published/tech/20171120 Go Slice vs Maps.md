# Go Slice vs Map

Slice 和 Map 是 Go 中的两种重要的数据类型。本文将记录我关于这两种数据结构性能的一些关键的发现。

在讨论性能方面之前，我们先来简单介绍一下 Slice 和 Map。

**Slice：**

Slice 是构建在数组之上的一种抽象数据结构。Slice 拥有一个指向数组开始位置的指针、数组长度以及 Slice 可以使用该数组的最大容量。Slice 可以按需增长或收缩。Slice 的增长通常包括为底层的数组重新分配内存。像 copy 和 append 这样的函数可以帮助增长数组。

**Map：**

Go 中的 Map 和其他语言类似（内部实现可能会有所不同）。Go 中的 Map 创建 bucket （每个 bucket 可以容纳 8 个键）。

**性能统计：**

我对这两种数据结构都进行了一些基准测试，结果记录如下。

**TEST-1**：查找 Slice 中的一个 **INT** 元素 vs 查找 Map 中的一个元素 -

这里，我们试着在一个长度为 n 的 Slice 中查找一个元素，并与 Map 中的键查找相比较。要查找 Slice 中的一个元素，我们将遍历该 Slice，然后进行简单的 `if` 来检查元素。至于 Map，我们简单查找键。在所有测试中，我都会寻找最坏的情况。

总样本 | len_size ( n ) | f(n)- []int (o(N)) - for 循环和 if（查找最后一个元素）| map[int]int 直接查找 o(1)  
---|---|---|---  
2 million | 5 | 5.42 ns/op | 12.7 ns/op  
2 million | 10 | 8.19 ns/op | 17.8 ns/op  
2 million | 100 | 63.3 ns/op | 16.5 ns/op  
2 million | 200 | 118 ns/op | 16.6 ns/op  
2 million | 400 | 228 ns/op | 18.4 ns/op  
2 million | 1000 | 573 ns/op | 17.0 ns/op  
2 million | 10000 | 5674 ns/op | 17.6 ns/op  
2 million | 100000 | 55141 ns/op | 15.1 ns/op  
  
正如你所见的（不出所料），对于不同的 n，Map 查找是常数复杂度（O(1)）。然而，有趣的是，对于 n 较小的 Slice，简单的 _for 循环和 if_ 比较花费的时间比 Map 少。较大的 n 如期望那样花费更多的时间。

![11.png](https://boltandnuts.files.wordpress.com/2017/11/11.png)

**结论：对于查找一个给定的键，我建议使用 Map。而对于一个较小的大小（n），使用 Slice 仍然是可以的。**

**TEST-2**：查找 Slice 中的一个 **STRING** 元素 vs 查找 Map 中的一个字符串键 -

我们所进行的步骤与 TEST-1 完全相同，这里唯一的不同是，我们使用字符串（String）。

总样本 | len - size ( n ) | f(n) []string [for 循环和 if（查找最后一个元素）] | f(n) map[string]string  
---|---|---|---  
2 million | 5 | 30.4 ns/op | 32.7 ns/op  
2 million | 10 | 56.5 ns/op | 23.5 ns/op  
2 million | 100 | 128 ns/op | 25.7 ns/op  
2 million | 200 | 665 ns/op | 23.6 ns/op  
2 million | 400 | 1766 ns/op | 23.7 ns/op  
2 million | 1000 | 905 ns/op | 25.7 ns/op  
2 million | 10000 | 8488 ns/op | 24.4 ns/op  
2 million | 100000 | 82444 ns/op | 25.9 ns/op  
  
从上面的数据，我们看到，给定一个键，查找一个字符串（使用 Map），具有 O(1) 复杂度。对于字符串比较，Map 击败了 Slice。

![2.png](https://boltandnuts.files.wordpress.com/2017/11/2.png)

**结论：给定一个字符串类型的键进行查找，我推荐使用 Map。即使对于较小的 n ，使用 Map 也是不错的。**

**TEST-3：** 给定索引，查找 Slice 元素。

如果我们知道索引，那么，在 Go 中查找 Slice 类似于在任何语言中查找数组，并且如它一样简单。

总样本 | len - size | **[]int**（直接索引查找） - O(1) | **[]string**（直接索引查找） - O(1) | map[int]int 直接查找 o(1) | map[string]string o(1)  
---|---|---|---|---|---  
2 million | 5 | 0.30 ns/op | 0.29 ns/op | 12.7 | 32.7  
2 million | 10 | 0.29 ns/op | 0.29 ns/op | 17.8 | 23.5  
2 million | 100 | 0.29 ns/op | 0.29 ns/op | 16.5 | 25.7  
2 million | 200 | 0.29 ns/op | 0.29 ns/op | 16.6 | 23.6  
2 million | 400 | 0.29 ns/op | 0.29 ns/op | 18.4 | 23.7  
2 million | 1000 | 0.29 ns/op | 0.29 ns/op | 17 | 25.7  
2 million | 10000 | 0.58 ns/op | 0.57 ns/op | 17.6 | 24.4  
2 million | 100000 | 0.58 ns/op | 0.55 ns/op | 15.1 | 25.9  

如上所示，Slice 的直接查找是 O(1) 固定增长率。

![3.png](https://boltandnuts.files.wordpress.com/2017/11/3.png)

**结论：直接查找都是常数复杂度，所以，假定你知道索引，那么使用哪个都无所谓。但是，假定索引已知，那么 Slice 或者数组查找仍然比 Map 查找快得多。**

**TEST-4**：遍历 Slice vs 遍历 Map

这里，我试着遍历 Map 和 Slice，并且在循环内执行一个常量操作。总复杂度将保持为 O(N)。

总样本 | len - size ( n ) | 遍历 Int slice O(N) | 遍历 int Map [O(N)](https://confluence.walmart.com/pages/createpage.action?spaceKey=SWTF&title=O%28N%29)  
---|---|---|---  
2 million | 5 | 9.02 ns/op | 107 ns/op  
2 million | 10 | 12.5 ns/op | 196 ns/op  
2 million | 100 | 59.2 ns/op | 1717 ns/op  
2 million | 200 | 84.9 ns/op | 3356 ns/op  
2 million | 400 | 155 ns/op | 6677 ns/op  
2 million | 1000 | 315 ns/op | 18906 ns/op  
2 million | 10000 | 2881 ns/op | 178804 ns/op***  
2 million | 100000 | 29012 ns/op | 1802439 ns/op***  
  
如上所示，遍历 Slice 比遍历 Map 快了近 20 倍。原因是，跟 Map 不一样，Slice （通过数组抽象出来）是在一个连续的内存块上创建的。至于 Map，循环必须遍历键空间（在 Go 中称为 bucket），并且内存分配可能并不连续。这就是为什么每次运行的时候，遍历 Map 的结果以不同的顺序显示键值。

![4.png](https://boltandnuts.files.wordpress.com/2017/11/4.png)

**结论：如果要求是在整个元素列表上打印或者执行操作，而不是查找，那么 Slice 是最佳选择。出于上述原因，遍历 Map 会花费更多时间。**

此外，请注意，就像对 Map 进行插入一样，Slice 上的 append 操作保证了 O(1) 复杂度，但是这是一种 **摊销(amortized)** 常量。append 可能偶尔需要为底层数组重新分配内存。

( *** ) - 因为对于大 Map，Go 的基准函数超时，所以样本大小从 200 万缩减至 2000。

有关测试的细节：

系统详情 | go操作系统：darwin | Go-1.9.2  
---|---|---  
 MAC-OSX | go架构：amd64 |  
  

源代码：

```go

// m-c02jn0m1f1g4:slicevsmap user1$ cat slicemap.go 

package slicemap

// 你可以取消打印行前的注释来看看结果（确认代码是否正常）。
// 但是对于基准测试，我们不关心打印的结果。我们关心的是它运行的时间

// import "fmt"

func RangeSliceInt(input []int, find int) (index int) {
        for index,value := range input {
                if (value == find) {
        //			fmt.Println("found at",index)
                        return index
                }
        }

        return -1
}


func RangeSliceIntPrint(input []int) {
        for _,_ = range input {
                continue
        }
}


func MapLookupInt(input map[int]int, find int) (key,value int) {
        if value,ok := input[find];ok {
        //		fmt.Println("found at", find,value)
                return find,value
        }
        return 0,0
}

func MapRangeInt(input map[int]int) {
        for _,_ = range input {
                continue
        }
}

func DirectSliceInt(input []int, index int) int {
        return input[index]
}

// 对于字符串 //

func RangeSliceString(input []string, find string) (index int) {
        for index,value := range input {
                if (value == find) {
                        // fmt.Println("found at",index)
                        return index
                }
        }
        return -1
}

func RangeSliceStringPrint(input []string) {
        for _,_ = range input {
                continue
        }
}

func MapLookupString(input map[string]string, find string) (key,value string) {
        if value,ok := input[find];ok {
//              fmt.Println("found at", find,value)
                return find,value
        }
        return "0", "0"
}


func MapRangeString(input map[string]string) {
        for _,_ = range input {
                continue
        }
}


func DirectSliceString(input []string, index int) string {
        return input[index]
}
```

```go
// m-c02jn0m1f1g4:slicevsmap user1$ cat slicemap_test.go

package slicemap

import (
        "testing"
        "strconv"
)

func Benchmark_TimeRangeSliceInt(b *testing.B) {
        b.StopTimer()
        input := make([]int, 100000, 500000)
        for i := 0; i < 100000; i++ {
                input[i] = i+10
        }

        b.StartTimer()

        b.N = 2000000  // 只是为了避免数百万次fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）

        for i := 0; i < b.N; i++ {
                RangeSliceInt(input, 100009)  // 对于最坏情况，检查最后一个元素
        }
}

func Benchmark_TimeDirectSliceInt(b *testing.B) {
        b.StopTimer()

        input := make([]int, 100000, 500000)
        for i := 0; i < 100000; i++ {
                input[i] = i+10
        }

        b.StartTimer()        

        b.N = 2000000  // 只是为了避免数百万次 fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）

        for i := 0; i < b.N; i++ {
                DirectSliceInt(input, 99999)  // 直接检查索引值。o(1)。发送索引
        }
}

func Benchmark_TimeMapLookupInt(b *testing.B) {

        b.StopTimer()

        input := make(map[int]int)

        // 扔一些值到 Map 中
        for i := 1; i <=100000; i++ {
                input[i] = i+10
        }

        b.StartTimer()

        b.N = 2000000  // 只是为了避免数百万次fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）

        for k := 0; k < b.N; k++ { 
                MapLookupInt(input, 100000)
        }

        /*
        运行命令：
        go test -bench=Benchmark_TimeMapLookup
        */
}

func Benchmark_TimeSliceRangeInt(b *testing.B) {
        b.StopTimer()

        input := make([]int, 5, 10)
        for i := 0; i < 5; i++ {
                input[i] = i+10
        }

        b.StartTimer()

        b.N = 2000000  // 只是为了避免数百万次fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）

        for k := 0; k < b.N; k++ {
                RangeSliceIntPrint(input)
        }
}

func Benchmark_TimeMapRangeInt(b *testing.B) {
        b.StopTimer()

        input := make(map[int]int)
        // 扔一些值到 Map 中
        for i := 1; i <=100000; i++ {
                input[i] = i+10
        }

        b.StartTimer()

        b.N = 2000  // 只是为了避免数百万次fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）

        for k := 0; k < b.N; k++ {
                MapRangeInt(input)
        }
}

// 测试字符串

func Benchmark_TimeRangeSliceString(b *testing.B) {
        b.StopTimer()

        input := make([]string, 100000, 500000)
        for i := 0; i < 100000; i++ {
                input[i] = strconv.FormatInt(int64(i+10),10)

        }

        b.StartTimer()

        b.N = 2000000  // 只是为了避免数百万次fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）

        for i := 0; i < b.N; i++ {
                RangeSliceString(input, "100009")  // 对于最坏情况，检查最后一个元素
        }
}

func Benchmark_TimeDirectSliceString(b *testing.B) {
        b.StopTimer()

        input := make([]string, 100000, 500000)
        for i := 0; i < 100000; i++ {
                input[i] = strconv.FormatInt(int64(i+10),10)
        }

        b.StartTimer()

        b.N = 2000000  // 只是为了避免数百万次fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）
        for i := 0; i < b.N; i++ {
                DirectSliceString(input, 99999)  // 直接检查索引值。o(1)
        }
}

func Benchmark_TimeMapLookupString(b *testing.B) {
        b.StopTimer()

        input := make(map[string]string)

        // 扔一些值到 Map 中
        for i := 1; i <=100000; i++ {
                input[strconv.FormatInt(int64(i),10)] = strconv.FormatInt(int64(i+10),10)
        }

        b.StartTimer()

        b.N = 2000000  // 只是为了避免数百万次fmt.Println（以防你在 slicemap.go 包中进行 fmt.Println）

        for k := 0; k < b.N; k++ {
                MapLookupString(input, "100000")
        }

        /*

        运行：

        go test -bench=Benchmark_TimeMapLookupString

        */
}
```

----------------

via: https://boltandnuts.wordpress.com/2017/11/20/go-slice-vs-maps/

作者：[qwerty.ytrewq86](https://boltandnuts.wordpress.com/about/)
译者：[ictar](https://github.com/ictar)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推

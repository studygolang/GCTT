首发于：https://studygolang.com/articles/21388

# Let US sort

## 在这篇文章里，我尝试发掘 [Go](https://golang.org/) 语言的所有特性，以便用最优的、利用多核处理器的方式来实现 [归并排序](https://en.wikipedia.org/wiki/Merge_sort)。

时间复杂度为 `O(nlogn)` 的最优排序算法中，归并排序是其中之一。它的原理为将数组分为两部分，分别进行排序，最后再归并，这种做法的开销没那么大。

![红色表示分割和排序，绿色表示归并](https://raw.githubusercontent.com/studygolang/gctt-images/master/let-us-sort/1_I9QJGWEgHtoo9H_hgVOg4g.png)
*颜色说明：红色表示分割和排序，绿色表示归并*

让我们通过代码演示最基本的形式：

![图 1：数值表示数组元素形成的时间](https://raw.githubusercontent.com/studygolang/gctt-images/master/let-us-sort/1_ntX4izwyS5AUMLVTJp0QSw.png)
*图 1：数值表示数组元素形成的时间*

```go
func Sort(arr []int) []int {
	if(len(arr) <= 1) {return arr}
	mid := len(arr)/2
	s1 := Sort(arr[:mid])
	s2 := Sort(arr[mid:])
	return merge.Merge(s1, s2)
}
```

```go
func update(final_arr []int, arr []int, index *int, increment_index *int) {
	final_arr[*index] = arr[*increment_index]
	*increment_index++
	*index++
}

func Merge(arr1 []int, arr2 []int) []int {
	size1 := len(arr1); size2 := len(arr2)
	final_arr := make([]int, size1 + size2)
	i := 0; j := 0; index := 0
	for ; i < size1 && j < size2; {
		if arr1[i] < arr2[j] {
			update(final_arr, arr1, &index, &i)
		} else {
			update(final_arr, arr2, &index, &j)
		}
	}
	for ; i < size1; {
		update(final_arr, arr1, &index, &i)
	}
	for ; j < size2; {
		update(final_arr, arr2, &index, &j)
	}
	return final_arr
}
```

> 观察可知，每个分割和归并的函数都是按照顺序执行的。

一个显而易见的优化方法是让 2 部分的排序并发执行，比如，如果长度为 2x 的 A 被划分为长度均为 x 的 X 和 Y，那么对 X 和 Y 的排序就可以并发地进行，因为同一内存地址不会被两个排序的线程访问。

```go
func Sort(arr []int) []int {
	if(len(arr) <= 1) {return arr}
	mid := len(arr)/2
	var s1, s2 []int
	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrency established
	Go func (s *[]int) {
		defer func() {wg.Done()} ()
		*s = Sort(arr[:mid])
	} (&s1)
	Go func (s *[]int) {
		defer func() {wg.Done()} ()
		*s = Sort(arr[mid:])
	} (&s2)
	// The sorting of arr[mid:] & arr[:mid] occurs Concurrently now.

	wg.Wait()
	return merge.Merge(s1, s2)
}
```

![图 2：并发地排序，但是顺序地归并。数值表示元素形成的时间](https://raw.githubusercontent.com/studygolang/gctt-images/master/let-us-sort/1_O9xBGZyf0gW5vQAhjj1sTQ.png)
*图 2：并发地排序，但是顺序地归并。数值表示元素形成的时间*

这使得分割 / 排序的过程更为快速，因为并发地对每个子数组进行排序（希望能在多核之间并行地执行），但是归并的过程仍然被阻塞，等待子数组的排序完成。

> 观察可知，分割是并发进行的，但是归并要顺序地进行。

为了使整个过程都是并发的，我们不再等待每个子数组都完成排序，而是一有子数组完成排序就开始归并。

### 为什么更快？

```go
...

/*
  update(f, a, &index, &i) =>
  f[index] = a[i];
  i++;
  index++;
*/
for ; i < size1 && j < size2; {
	if arr1[i] < arr2[j] {
		update(final_arr, arr1, &index, &i)
	} else {
		update(final_arr, arr2, &index, &j)
	}
}

...
```

回想一下，每次迭代中，归并的数组从每个已经排好序的数组中获取一个元素。如果每次比较的开销是 C，从子数组中重组一个长度为 N 的归并数组的开销将为 `O(C*N)`。因此，如果最终数组的长度为 M，总的开销将为 ` ∑ C*(M+2*(M/2)+4*(M/4)+ … .) = C*M*log(M) is O(M*log(M))`，因为每个归并操作要被阻塞，等到它的每个子数组都完成归并。

另一方面，并发的归并并不等到每一层完成后才进行，因此数组的值一被接收到就会向下传递。

> 注意，分割操作的复杂度为 `O(1)`。

![图 3：并发的归并，方框中的值代表其形成的时间](https://raw.githubusercontent.com/studygolang/gctt-images/master/let-us-sort/1_dSgQ8jHZ7VwG99uvnZ2O4w.png)
*图 3：并发的归并，方框中的值代表其形成的时间*

### 如何确保并发的归并？

Channels 可以用来传递和接收数据，因此一旦接收到初始元素，归并操作就会开始，不会被阻塞，等到之前的归并完成了才能进行。

```go
func Sort(arr []int, ch chan int) {
	defer close(ch)
	if(len(arr) <= 1) {
		if(len(arr)==1) {
			ch <- arr[0]
		}
		return
	}
	mid := len(arr)/2
	s1 := make(chan int, mid)
	s2 := make(chan int, len(arr) - mid)

	// Concurrency established
	Go Sort(arr[:mid], s1)
	Go Sort(arr[mid:], s2)
	// The sorting of arr[mid:] & arr[:mid] occurs Concurrently now.

	// Merging happens simultaneously and is not blocked on individual sorting.
	merge.Merge(s1, s2, ch)
}
```

s1/s2 一接收到数据，它们就会进行处理然后把数据传递给 ch，ch 再将数据向下传递，以构造最终的数组。

```go
func update(s chan int, ch chan int, c *int, ok *bool) {
	ch <- *c
	*c, *ok = <-s
}

func Merge(s1, s2, ch chan int) {
	// v, ok = <-s; ok returns false if there's no more element to be received from s.
	v1, ok1 := <-s1
	v2, ok2 := <-s2
	for ok1 && ok2 {
		if(v1<v2) {
			update(s1, ch, &v1, &ok1)
		} else {
			update(s2, ch, &v2, &ok2)
		}
	}
	for ok1 {
		update(s1, ch, &v1, &ok1)
	}
	for ok2 {
		update(s2, ch, &v2, &ok2)
	}
}
```

上面的过程和使用数组类似，但是现在使用的是 channels。

这个版本的归并排序开发了 Go 的所有特性，以此确保从排序到合并（不会阻塞直到完成）的并发性。正如现在你所猜测的那样，时间复杂度（假设是在多核的条件下进行，即可以有无数多个核来处理负载）从 `O(N*log(N))` 下降到了 `O(log(N)+N)` ，即构造最终数组的时间复杂度是 N + 树的高度 (`logN`)，也就是 `O(N)`。

> 我没法说归并排序的时间复杂度是 `O(N)` ！
>
> 声明：多核处理器上最优化的归并排序可以将运行时间显著降为 `O(N)`。

*希望大家发现 Go 的美妙之处并为之做出贡献！*

---

via: https://medium.com/@jayaganesh1997/let-us-sort-25e41a4ba854

作者：[Jayaganesh Kalyanasundaram](https://medium.com/@jayaganesh1997)
译者：[maxwellhertz](https://github.com/maxwellhertz)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
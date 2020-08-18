首发于：https://studygolang.com/articles/30253

# Golang 有多快？—— 对比 Go 和 Python 实现的 Eratosthenes 筛选法

![Photo by Max Duzij on Unsplash](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200523-How-Fast-Is-Golang/Photo.jpeg)

时间宝贵，所以为什么浪费时间等待程序运行？除非过去几年与世隔绝，否则不会错过 Go 的兴起。由谷歌工程师 Robert Griesemer，Rob Pike 和 Ken Thompson [创造的](https://golang.org/doc/faq#Origins) 新颖的编程语言，Go 被誉为[近乎完美](https://towardsdatascience.com/why-we-deploy-machine-learning-models-with-go-not-python-a4e35ec16deb)，在易用性上可以与 Python 相媲美，而在执行速度上又可以与 C 语言相媲美。真的如传闻所说吗？今天，我们会分别使用 Go 和 Python 实现埃拉托色尼筛选法，并以耗时为结果。最终，得到的问题的结果，也就是“Golang 到底有多快？”

## Eratosthenes 筛选法

Eratosthenes 是古希腊博学家。对很多领域均有涉猎（数学，地理，诗歌，天文学，音乐 —— 还不仅仅只是这些）他是一位著名的学者，[据称他是第一个](https://en.wikipedia.org/wiki/Eratosthenes)测量地球周长及其轴线倾斜度的人。要知道当时是公元前 3 世纪，就会足够让人惊奇。

![Eratosthenes 的蚀刻（维基共享资源/公共领域）](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200523-How-Fast-Is-Golang/An-etching-of-Eratosthenes.jpg)

尽管如此，我们将从他的一项尝试中使用一种技术进入数论世界：Eratosthenes 筛选法。简单来说，这是一个从整数集合中生成（或者说“筛选”）质数列表的相对有效的方法。生成最大为 n 的质数的方法如下：

- 构造一个整数集合，a，例如 a = {2, 3, ..., n}。
- 遍历 a 的元素，每轮遍历移除该原色的倍数（但是不移除这个原始的元素）。
- a 现在是一个包含着最大为 n 的质数集合。

理解该方法的一个奇妙方式是看下面的这个动图：

![动态展示 Eratosthenes 筛选法 的动画](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-How-Fast-Is-Golang/An-animation-showing-the-Sieve-of-Eratosthenes-in-action.gif)

## Python 代码
先从 Python 代码开始，因为它可能是最简单易懂的。下面是实际运行代码的精简版本（如果想看带错误捕获的实际运行版本，从[我的 GitHub 仓库](https://github.com/8F3E/sieve-of-eratosthenes)检出）。

```python
def sieve(primes, factor):
    for p in primes:
        if p != 0 and p != factor:
            if p % factor == 0:
                primes[primes.index(p)] = 0
    return primes


def main(n):
    primes = [i for i in range(2, n + 1)]
    for p in primes:
        if p != 0:
            primes = sieve(primes, p)
    print('\n'.join([str(p) for p in primes if p != 0]))
```

这段代码是如何工作的？

- 主函数接受一个参数，n，这个整型参数限定了我们要生成的质数的大小。
- 之后，创建一个包含 2 到 n 的数字的列表。
- 对于这些数字中的每一个数字，通过向列表中的每一个元素调用 sieve 函数来移除其所有的倍数。(实际上是将倍数设置为 0，因为遍历过程中移除他们会导致错误)
- 最后我们输出一个剩下的所有质数的列表（即，原列表中所有不是 0 的元素）确保另起一行分隔他们。
- 注意第 4 行的 `p % factor`。这是一个获取（在这个例子中）p 除以 factor 的余数的“取模函数”。如果余数是 0，那我们便得到了一个倍数（比如，10 ÷ 5=2，余数是 0）。

## Go 代码

为了保证对比公平，在 Go 的筛选代码中使用完全一样的算法。出来这里是（从 [GitHub](https://github.com/8F3E/sieve-of-eratosthenes)上面的完整代码中剥离出来的）精简代码外，我不会重复深入解释这段代码。来将这段代码与上面的 Python 代码进行对比。
```go
func main() {
	var primes []int

	for i := 2; i <= n; i++ {
		primes = append(primes, i)
	}

	for i := 0; i < len(primes); i++ {
		if primes[i] != 0 {
			sieve(primes, primes[i])
		}
	}

	for i := 0; i < len(primes); i++ {
		if primes[i] != 0 {
			fmt.Println(primes[i])
		}
	}
}

func sieve(primes []int, factor int) {
	for index, value := range(primes) {
		if value != 0 && value != factor {
			if value % factor == 0 {
				primes[index] = 0
			}
		}
	}
}
```

- 第 2 行到第 6 行的代码创建我们的数字列表。
- 第 8 行到第 12 行代码遍历该列表，使用 sieve 函数（工作流程与 Python 代码中的完全一样）移除（每个元素的）倍数。
- 最后输出列表，每个指数另起一行来进行分隔。

## 结果出来了！

我在 bash shell 中使用 `time` 函数（见下面的例子）来测试这些代码[^1]，使用不同的 n （范围从 1 到 100,000）同时使用 [Jupyter Notebook](https://nbviewer.jupyter.org/github/8F3E/sieve-of-eratosthenes/blob/61876c17a697d1c8439e39ab790de15adc678804/testing/Testing.ipynb) 生成下面的结果。图表由 [Plotly](https://medium.com/swlh/forget-matplotlib-you-should-be-using-plotly-ada76b650ff4) 生成

```
$ time bin/sieve 10
2
3
5
7
real    0m0.004s
user    0m0.000s
sys     0m0.005s
```

![结果数据，首先是标准比例，然后是对数比例。使用 Plotly 制作。](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-How-Fast-Is-Golang/The-resulting-data-1.png)

![](https://raw.githubusercontent.com/studygolang/gctt-images2/blob/master/20200523-How-Fast-Is-Golang/The-resulting-data-2.png)

结果十分明显，Go 要比 Python 要快，尤其是在大规模数字计算领域上。当在在较小规模时（n < 1000）他们并没有明显差异（Go 相较于 Python 轻微缓慢的节奏来说，几乎是瞬时的），当达到了 10,000 的规模，Python 落后了很多。

## 总结

> Go 是一种开源编程语言，可轻松构建简单，可靠和高效的软件。—— golang.org

总结如下，Go 明显比 Python 快得多。正如该语言的网站所说，它简单，可靠且*非常*高效。因此，是的，如果您发现 Python 更容易或更简单或者只是更快地编程，请使用 Python。但是对于时间紧迫，计算能力强的软件，Go 可能是正确的选择。

> 更新：如果您需要坚持使用 Python 但又想尽可能地提高它的速度，为什么不看看我的其他文章：[5 种可以立即提高 Python 速度的方法](https://medium.com/@8F3E/how-you-can-improve-pythons-speed-right-now-6a0ec2234618)。

[^1]: 确实，我应该每个 n 次测试一次。但是通过一遍就说明了这种趋势，所以我节省了一些时间，只停留了一轮。

---
via: https://medium.com/swlh/how-fast-is-golang-135c658205eb

作者：[8F3E](https://medium.com/@8F3E)
译者：[dust347](https://github.com/dust347)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

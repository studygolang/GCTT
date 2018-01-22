# Go 运行时 Debug 纪录
## 序言
我是 Prometheus 和 Grafana 的忠实粉丝。作为一个前 Google 公司 SRE, 我一直以来都知道良好的监控的重要性， Prometheus 和 Grafana 的组合是我多年的最爱。我用他们来监控我的个人服务器(黑盒和白盒都有)， Euskal Encounter 内外部事件以及我服务的专业客户。Prometheusa 让编写定制的数据导出器变得非常简单， 而且你能够找到很多现成的满足你要求的导出器。比如说，我们使用 sql_exporter 来为 Encounter 的与会者数据做了一个监控面板。
![Event dashboard for Euskal Encounter (fake staging data)](https://marcan.st/posts/go_debug/euskalstats.png)

既然我们能够很容易地把 node_exporter 部署到任意的机器上并且用 Prometheus 实例去拿到机器的基础数据维度（CPU，内存，网络，磁盘，文件系统使用等), 那我为什么不同时监控我的笔记本呢？我有一个蓝天"游戏"本充当我的首席工作站，主要在家里被当做台式机使用，同时也会被我带去参加一个大型活动比如 Chaos Communication Congress。由于我已经在这台笔记本和一台部署了 Prometheus 的服务器之间有一个 VPN 通道了，我可以使用 `emerge prometheus-node_exporter` 命令来启动服务，让我的 Prometheus 实例指向这个服务。这个命令会为这台笔记本自动设置警报，每当我打开了太多 Chrome 页面用完了 32G 内存的时候，我的手机就会发出很大的声音来提醒我。完美！

## 问题浮现
这一切看起来很完美，但是，在我完成这些设置后仅仅一小时，我的手机接收到一条信息： 我新添加的目标不可访问。但是，我能够通过 SSH 连到我的笔记本，所以它肯定是启动了的，只是 node_exporter 崩溃了。
```
fatal error: unexpected signal during runtime execution
[signal SIGSEGV: segmentation violation code=0x1 addr=0xc41ffc7fff pc=0x41439e]

goroutine 2395 [running]:
runtime.throw(0xae6fb8, 0x2a)
        /usr/lib64/go/src/runtime/panic.go:605 +0x95 fp=0xc4203e8be8 sp=0xc4203e8bc8 pc=0x42c815
runtime.sigpanic()
        /usr/lib64/go/src/runtime/signal_unix.go:351 +0x2b8 fp=0xc4203e8c38 sp=0xc4203e8be8 pc=0x443318
runtime.heapBitsSetType(0xc4204b6fc0, 0x30, 0x30, 0xc420304058)
        /usr/lib64/go/src/runtime/mbitmap.go:1224 +0x26e fp=0xc4203e8c90 sp=0xc4203e8c38 pc=0x41439e
runtime.mallocgc(0x30, 0xc420304058, 0x1, 0x1)
        /usr/lib64/go/src/runtime/malloc.go:741 +0x546 fp=0xc4203e8d38 sp=0xc4203e8c90 pc=0x411876
runtime.newobject(0xa717e0, 0xc42032f430)
        /usr/lib64/go/src/runtime/malloc.go:840 +0x38 fp=0xc4203e8d68 sp=0xc4203e8d38 pc=0x411d68
github.com/prometheus/node_exporter/vendor/github.com/prometheus/client_golang/prometheus.NewConstMetric(0xc42018e460, 0x2, 0x3ff0000000000000, 0xc42032f430, 0x1, 0x1, 0x10, 0x9f9dc0, 0x8a0601, 0xc42032f430)
        /var/tmp/portage/net-analyzer/prometheus-node_exporter-0.15.0/work/prometheus-node_exporter-0.15.0/src/github.com/prometheus/node_exporter/vendor/github.com/prometheus/client_golang/prometheus/value.go:165 +0xd0 fp=0xc4203e8dd0 sp=0xc4203e8d68 pc=0x77a980
```

和很多 `Prometheus` 组件一样， `node_exporter` 是用 `Go` 语言写的。 `Go` 是一个相对安全的语言：尽管如果你愿意的话，你可以用它来“射自己的脚”， 它也没有像 `Rust` 那样强大的安全性保证，但是要想用 `Go` 语言写出一个段错误来仍然不容易。更不用提， `node_exporter` 是一个相对简单的 `Go` 应用， 只用到了纯 Go 的依赖。因此， 这是一个很罕见的崩溃。 尤其是这次崩溃是在 `mallocgc` 中发生的，这也是在正常情况下不会发生的。

当我重启 `node_exporter` 几次之后，更有趣的事情发生了：
```
2017/11/07 06:32:49 http: panic serving 172.20.0.1:38504: runtime error: growslice: cap out of range
goroutine 41 [running]:
net/http.(*conn).serve.func1(0xc4201cdd60)
        /usr/lib64/go/src/net/http/server.go:1697 +0xd0
panic(0xa24f20, 0xb41190)
        /usr/lib64/go/src/runtime/panic.go:491 +0x283
fmt.(*buffer).WriteString(...)
        /usr/lib64/go/src/fmt/print.go:82
fmt.(*fmt).padString(0xc42053a040, 0xc4204e6800, 0xc4204e6850)
        /usr/lib64/go/src/fmt/format.go:110 +0x110
fmt.(*fmt).fmt_s(0xc42053a040, 0xc4204e6800, 0xc4204e6850)
        /usr/lib64/go/src/fmt/format.go:328 +0x61
fmt.(*pp).fmtString(0xc42053a000, 0xc4204e6800, 0xc4204e6850, 0xc400000073)
        /usr/lib64/go/src/fmt/print.go:433 +0x197
fmt.(*pp).printArg(0xc42053a000, 0x9f4700, 0xc42041c290, 0x73)
        /usr/lib64/go/src/fmt/print.go:664 +0x7b5
fmt.(*pp).doPrintf(0xc42053a000, 0xae7c2d, 0x2c, 0xc420475670, 0x2, 0x2)
        /usr/lib64/go/src/fmt/print.go:996 +0x15a
fmt.Sprintf(0xae7c2d, 0x2c, 0xc420475670, 0x2, 0x2, 0x10, 0x9f4700)
        /usr/lib64/go/src/fmt/print.go:196 +0x66
fmt.Errorf(0xae7c2d, 0x2c, 0xc420475670, 0x2, 0x2, 0xc420410301, 0xc420410300)
        /usr/lib64/go/src/fmt/print.go:205 +0x5a

```

发生在 `Sprintf` 中的崩溃，这就更神奇了。
```
runtime: pointer 0xc4203e2fb0 to unallocated span idx=0x1f1 span.base()=0xc4203dc000 span.limit=0xc4203e6000 span.state=3
runtime: found in object at *(0xc420382a80+0x80)
object=0xc420382a80 k=0x62101c1 s.base()=0xc420382000 s.limit=0xc420383f80 s.spanclass=42 s.elemsize=384 s.state=_MSpanInUse
 <snip>
fatal error: found bad pointer in Go heap (incorrect use of unsafe or cgo?)

runtime stack:
runtime.throw(0xaee4fe, 0x3e)
        /usr/lib64/go/src/runtime/panic.go:605 +0x95 fp=0x7f0f19ffab90 sp=0x7f0f19ffab70 pc=0x42c815
runtime.heapBitsForObject(0xc4203e2fb0, 0xc420382a80, 0x80, 0xc41ffd8a33, 0xc400000000, 0x7f0f400ac560, 0xc420031260, 0x11)
        /usr/lib64/go/src/runtime/mbitmap.go:425 +0x489 fp=0x7f0f19ffabe8 sp=0x7f0f19ffab90 pc=0x4137c9
runtime.scanobject(0xc420382a80, 0xc420031260)
        /usr/lib64/go/src/runtime/mgcmark.go:1187 +0x25d fp=0x7f0f19ffac90 sp=0x7f0f19ffabe8 pc=0x41ebed
runtime.gcDrain(0xc420031260, 0x5)
        /usr/lib64/go/src/runtime/mgcmark.go:943 +0x1ea fp=0x7f0f19fface0 sp=0x7f0f19ffac90 pc=0x41e42a
runtime.gcBgMarkWorker.func2()
        /usr/lib64/go/src/runtime/mgc.go:1773 +0x80 fp=0x7f0f19ffad20 sp=0x7f0f19fface0 pc=0x4580b0
runtime.systemstack(0xc420436ab8)
        /usr/lib64/go/src/runtime/asm_amd64.s:344 +0x79 fp=0x7f0f19ffad28 sp=0x7f0f19ffad20 pc=0x45a469
runtime.mstart()
        /usr/lib64/go/src/runtime/proc.go:1125 fp=0x7f0f19ffad30 sp=0x7f0f19ffad28 pc=0x430fe0
```
这次是 `GC` 出现了崩溃，又是一个新的问题。

通常在这时，有两个常见的结论：要么我的笔记本出现了一个严重的硬件问题，要么可执行文件中有严重的内存泄露。刚开始时我认为前者可能性不大因为这台笔记本承载了很大的工作量但是之前并没有出现看起来像是硬件问题的 bug（它曾经出现过软件问题，但是绝不会是随机的）。由于 `node_exporter` 这样的 `Go` 二进制文件是静态链接的且不依赖于其他库，我可以下载官方发布的二进制版本来尝试一下，这可以用来判断是否是我的系统出现了问题。结果是，我替换二进制文件后，仍然出现了崩溃情况。
```
unexpected fault address 0x0
fatal error: fault
[signal SIGSEGV: segmentation violation code=0x80 addr=0x0 pc=0x76b998]

goroutine 13 [running]:
runtime.throw(0xabfb11, 0x5)
        /usr/local/go/src/runtime/panic.go:605 +0x95 fp=0xc420060c40 sp=0xc420060c20 pc=0x42c725
runtime.sigpanic()
        /usr/local/go/src/runtime/signal_unix.go:374 +0x227 fp=0xc420060c90 sp=0xc420060c40 pc=0x443197
github.com/prometheus/node_exporter/vendor/github.com/prometheus/client_model/go.(*LabelPair).GetName(...)
        /go/src/github.com/prometheus/node_exporter/vendor/github.com/prometheus/client_model/go/metrics.pb.go:85
github.com/prometheus/node_exporter/vendor/github.com/prometheus/client_golang/prometheus.(*Desc).String(0xc4203ae010, 0xaea9d0, 0xc42045c000)
        /go/src/github.com/prometheus/node_exporter/vendor/github.com/prometheus/client_golang/prometheus/desc.go:179 +0xc8 fp=0xc420060dc8 sp=0xc420060c90 pc=0x76b998
```
又一个完全不同的崩溃。这时我认为 `node_exporter` 或者它的依赖出问题的可能性变大了，因此我在 Github 上提了一个 issue。也许开发者以前见过这个情况？无论如何，问一下他们是否有解决方案是没错的。

## 不简短的硬件排查
不出所料， 开发者首先想到的是硬件问题。这是合理的，毕竟我只在一台机器上碰到这个问题。我的所有其他机器上的 `node_exporter` 都运行得很好。尽管我没有任何有关硬件不稳定的证据，我也没有任何关于为什么 `node_exporter` 会只在那台机器上崩溃的解释。我尝试使用 `Memtest86+` 看一下内存使用情况，出现了下面的情况：
![Memtest86](https://marcan.st/posts/go_debug/memtest.png)

损坏的内存！更具体的说，一个比特的内存。完整地跑了一遍测试之后，我只发现了这一个坏掉的比特，和一些误报的内存损坏。

之后的测试显示 `Memtest86+ test #5` 在 `SMP` 模式下能够快速地发现这个错误，但通常不是第一次。出错的永远是相同的比特。这提醒我们问题在于一个内存单元的损坏。具体来说，是一个和天气一起变坏的内存比特。这听上去是很合理的：高温导致了内存单元的泄露，也因此让一个比特翻转更有可能发生。

总结来说，274,877,906,944 比特中的一个坏掉了。这算起来其实是一个很优秀的出错率。硬盘和闪存的出错率可比这个数字要高多了 - 只不过这些设备把损坏的存储块做了标记并且在用户不知情的情况下就废弃了这些存储块，它们还可以把新发现的弱存储块标记为坏的且将它们重定向到一个空的区块。然而，内存受限于容量，不会做这么奢侈的操作，所以一个坏掉的比特会一直留在那。
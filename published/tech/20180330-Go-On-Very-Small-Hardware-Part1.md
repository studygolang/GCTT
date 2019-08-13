已发布：https://studygolang.com/articles/22742

# Go 最小硬件编程（第一部分）

我们能够让 Go 在多低的配置下运行并做一些实用的事情呢？

最近我购买了这个特别便宜的开发板：

[![STM32F030F4P6](https://ziutek.github.io/images/mcu/f030-demo-board/board.jpg)](https://ziutek.github.io/2018/03/30/go_on_very_small_hardware.html)

购买它，我基于以下三个理由：第一，我（作为一个程序员）从未搞过 STM32F0 系列的开发板；第二，STM32F10x 系列的板子已经很陈旧了，STM32F0 系列的 MCU 十分便宜，有更新的外设，并有很多改进和 bug 修复；第三，我选择这个系列中最低配的是为了本文，这会让整个事情变得妙趣横生。

## 硬件

[STM32F030F4P6](http://www.st.com/content/st_com/en/products/microcontrollers/stm32-32-bit-arm-cortex-mcus/stm32-mainstream-mcus/stm32f0-series/stm32f0x0-value-line/stm32f030f4.html) 是令人印象深刻的硬件：

- CPU: [Cortex M0](https://en.wikipedia.org/wiki/ARM_Cortex-M#Cortex-M0) 48 MHz (最低配置中，只有 12000 个逻辑门电路),
- RAM: 4 KB,
- Flash: 16 KB,
- ADC、SPI、I2C、USART 和几个定时器,

全部采用 TSSOP20 封装。如你所见，它是非常小的 32 位系统。

## 软件

如果你想知道如何在这块开发板上使用 [Go](https://golang.org/) 进行编程，你需要再阅读一次硬件手册。你必须面临的一个真实情况是：几乎没有人会在 Go 编译器中加入对 Cortex-M0 的支持，这就是一开始需要解决的问题。

我将会使用 [Emgo](https://github.com/ziutek/emgo)，不用担心，你将会看到它会让你能够在如此小的系统上运行 Go。

在这块开发板送达我这里之前，还没有任何对 [stm32/hal](https://github.com/ziutek/emgo/tree/master/egpath/src/stm32/hal) 系列 F0 MCU 的支持。在简单研究 [参考手册](http://www.st.com/resource/en/reference_manual/dm00091010.pdf) 后，STM32F0 系列与 STM32F3 系列似乎是相似的，这就为工作展开找到了一个新的突破口。

如果你想跟上本文后续的步骤，你需要安装 Emgo：

```bash
cd $HOME
git clone https://github.com/ziutek/emgo/
cd emgo/egc
go install
```

同时配置几个环境变量：

```bash
export EGCC=path_to_arm_gcc      # eg. /usr/local/arm/bin/arm-none-eabi-gcc
export EGLD=path_to_arm_linker   # eg. /usr/local/arm/bin/arm-none-eabi-ld
export EGAR=path_to_arm_archiver # eg. /usr/local/arm/bin/arm-none-eabi-ar

export EGROOT=$HOME/emgo/egroot
export EGPATH=$HOME/emgo/egpath

export EGARCH=cortexm0
export EGOS=noos
export EGTARGET=f030x6
```

想了解更多的细节，请访问 [Emgo](https://github.com/ziutek/emgo) 官网。

保证 egc 在你的 PATH 中。你可以使用 `go build` 而不是 `go install`，然后将 egc 复制到你的 *$HOME/bin* 或者 */usr/local/bin* 中。

现在为你的第一个 Emgo 程序创建新的目录，将例子中的连接器脚本复制到如下目录中：

```bash
mkdir $HOME/firstemgo
cd $HOME/firstemgo
cp $EGPATH/src/stm32/examples/f030-demo-board/blinky/script.ld .
```

## 最小程序

在 *main.go* 文件中创建最小程序：

```go
package main

func main() {
}
```

编译这个文件，没有任何问题：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
   7452     172     104    7728    1e30 cortexm0.elf
```

第一次编译会耗费一些时间。编译的二进制结果占用了 7624 字节（文本和数据）的 Flash 空间，对于一个什么都没有做的程序来说，占用的空间有点大。还剩下 8760 字节的空间去做一些有用的事情。

对于传统的 *Hello, World!* 代码如何：

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
```

很不幸，出错了：

```bash
$ egc
/usr/local/arm/bin/arm-none-eabi-ld: /home/michal/P/go/src/github.com/ziutek/emgo/egpath/src/stm32/examples/f030-demo-board/blog/cortexm0.elf section `.text' will not fit in region `Flash'
/usr/local/arm/bin/arm-none-eabi-ld: region `Flash' overflowed by 10880 bytes
exit status 1
```

*Hello, World!* 需要 STM32F030x6 至少 32KB 的 Flash 空间。

*fmt* 包强制包含整个 *strconv* 和 *reflect* 包。甚至在精简版本的 Emgo 中，这三个在一起都非常大。我们不能实现这个例子了。其实许多的应用程序不需要花哨的格式化文本输出。通常情况下，一个或多个 LED 或是 7 段数码管显示就足够了。但是，在第二部分中，我将会尝试使用 *strconv* 包去格式化并在 UART 上打印一些数字或文本。

## 闪烁

我们的开发板有一个 LED 连接于 PA4 引脚和 VCC。这次我们编写多一点代码：

```go
package main

import (
	"delay"

	"stm32/hal/gpio"
	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
)

var led gpio.Pin

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	gpio.A.EnableClock(false)
	led = gpio.A.Pin(4)

	cfg := &gpio.Config{Mode: gpio.Out, Driver: gpio.OpenDrain}
	led.Setup(cfg)
}

func main() {
	for {
		led.Clear()
		delay.Millisec(100)
		led.Set()
		delay.Millisec(900)
	}
}
```

按照惯例，*init* 函数负责初始化和配置外设。

`system.SetupPLL(8, 1, 48/8)` 配置 RCC 去使用外部 8 MHz 振荡器的 PLL 作为系统时钟源。PLL 分频器设置为 1，倍频数为 48/8 = 6，这样就提供 48 MHz 的系统频率。

`systick.Setup(2e6)` 设置 Cortex-M SYSTICK 时钟作为系统时钟，每隔 2e6 纳秒运行一次（每秒 500 次）。

`gpio.A.EnableClock(false)` 为 GPIO A 口使能时钟。*False* 意思是时钟在低功耗模式下会被禁用，但是在 STM32F0 中没有实现低功耗模式。

`led.Setup(cfg)` 设置 PA4 引脚为开漏输出。

`led.Clear()` 设置 PA4 引脚为低电平，在开漏配置下，打开 LED。

`led.Set()` 设置 PA4 为高电平状态，关掉 LED。

编译这个代码：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
   9772     172     168   10112    2780 cortexm0.elf
```

正如你看到的，闪烁程序比最小程序多占用 2320 字节的空间。这里仍然还有 6440 字节的剩余空间。

让我们看看代码是否工作：

```bash
$ openocd -d0 -f interface/stlink.cfg -f target/stm32f0x.cfg -c 'init; program cortexm0.elf; reset run; exit'
Open On-Chip Debugger 0.10.0+dev-00319-g8f1f912a (2018-03-07-19:20)
Licensed under GNU GPL v2
For bug reports, read
        http://openocd.org/doc/doxygen/bugs.html
debug_level: 0
adapter speed: 1000 kHz
adapter_nsrst_delay: 100
none separate
adapter speed: 950 kHz
target halted due to debug-request, current mode: Thread
xPSR: 0xc1000000 pc: 0x0800119c msp: 0x20000da0
adapter speed: 4000 kHz
** Programming Started **
auto erase enabled
target halted due to breakpoint, current mode: Thread
xPSR: 0x61000000 pc: 0x2000003a msp: 0x20000da0
wrote 10240 bytes from file cortexm0.elf in 0.817425s (12.234 KiB/s)
** Programming Finished **
adapter speed: 950 kHz
```

在这篇文章中，这是我人生第一次把短视频转换为 [动画 PNG](https://en.wikipedia.org/wiki/APNG)。对此我印象深刻，告别了 YouTube 同时对 IE 用户说声抱歉。了解更多，请访问 [apngasm](http://apngasm.sourceforge.net/)。我应该学习 HTML5 基础的，但是现在 APNG 是我喜欢的展现循环短视频的方式了。

![STM32F030F4P6](https://ziutek.github.io/images/mcu/f030-demo-board/blinky.png)

## 更多 Go 编程

如果你不是一个 Go 的程序员，但是你已经听过 Go 语言的一些事情，你可能会说：“这种语法很好，但是相较于 C  没有明显的提升。给我展示 *Go 语言* 的 *channels* 和 *goroutines*！”

下面是代码：

```go
import (
	"delay"

	"stm32/hal/gpio"
	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
)

var led1, led2 gpio.Pin

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	gpio.A.EnableClock(false)
	led1 = gpio.A.Pin(4)
	led2 = gpio.A.Pin(5)

	cfg := &gpio.Config{Mode: gpio.Out, Driver: gpio.OpenDrain}
	led1.Setup(cfg)
	led2.Setup(cfg)
}

func blinky(led gpio.Pin, period int) {
	for {
		led.Clear()
		delay.Millisec(100)
		led.Set()
		delay.Millisec(period - 100)
	}
}

func main() {
	go blinky(led1, 500)
	blinky(led2, 1000)
}
```

代码改动很小：第二个 LED 被添加，前面的 *main* 函数被重命名为 *blinky*，函数需要两个参数。*Main* 在一个新的 goroutine 中启动第一个 *blinky* 函数，这样两个 LED 同时 *并行* 运行。有必要提一下，*gpio.Pin* 类型支持并发访问在同一 GPIO 口的不同引脚。

Emgo 仍然还有许多缺点。其中一个就是你必须提前对 goroutines（tasks）指定一个最大数值。是时候编辑一下 *script.Id* 了：

```go
ISRStack = 1024;
MainStack = 1024;
TaskStack = 1024;
MaxTasks = 2;

INCLUDE stm32/f030x4
INCLUDE stm32/loadflash
INCLUDE noos-cortexm
```

栈是用猜的方式确定的大小，现在我们还不会关心这些事情。

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  10020     172     172   10364    287c cortexm0.elf
```

另外一个 LED 和 goroutine 花费了 248 字节的 Flash 空间。

![STM32F030F4P6](https://ziutek.github.io/images/mcu/f030-demo-board/goroutines.png)

## Channels

Channels 是 Go 中 goroutines 之间通信 [最好的方式](https://blog.golang.org/share-memory-by-communicating)。Emgo 做的更多，它允许通过 *中断处理* 去使用 *缓冲* channels。下面的例子实际展示了这种情况。

```go
package main

import (
	"delay"
	"rtos"

	"stm32/hal/gpio"
	"stm32/hal/irq"
	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
	"stm32/hal/tim"
)

var (
	leds  [3]gpio.Pin
	timer *tim.Periph
	ch    = make(chan int, 1)
)

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	gpio.A.EnableClock(false)
	leds[0] = gpio.A.Pin(4)
	leds[1] = gpio.A.Pin(5)
	leds[2] = gpio.A.Pin(9)

	cfg := &gpio.Config{Mode: gpio.Out, Driver: gpio.OpenDrain}
	for _, led := range leds {
		led.Set()
		led.Setup(cfg)
	}

	timer = tim.TIM3
	pclk := timer.Bus().Clock()
	if pclk < system.AHB.Clock() {
		pclk *= 2
	}
	freq := uint(1e3) // Hz
	timer.EnableClock(true)
	timer.PSC.Store(tim.PSC(pclk/freq - 1))
	timer.ARR.Store(700) // ms
	timer.DIER.Store(tim.UIE)
	timer.CR1.Store(tim.CEN)

	rtos.IRQ(irq.TIM3).Enable()
}

func blinky(led gpio.Pin, period int) {
	for range ch {
		led.Clear()
		delay.Millisec(100)
		led.Set()
		delay.Millisec(period - 100)
	}
}

func main() {
	go blinky(leds[1], 500)
	blinky(leds[2], 500)
}

func timerISR() {
	timer.SR.Store(0)
	leds[0].Set()
	select {
	case ch <- 0:
		// Success
	default:
		leds[0].Clear()
	}
}

//c:__attribute__((section(".ISRs")))
var ISRs = [...]func(){
	irq.TIM3: timerISR,
}
```

与之前例子的不同之处对比：

1. 第三个 LED 被添加，连接到 PA9 引脚（UART 头部的 TXD 引脚）。
2. 定时器（TIM3）被引入作为中断源。
3. 新的 *timerISR* 方法处理 *irp.TIM3* 中断。
4. 新增的容量为 1 的缓冲 channel 用于 *timerISR* 和 *blinky* 协程之间进行通信。
5. *ISRs* 数组作为中断向量表，是更大的异常向量表的一部分。
6. *blinky 的 for 语句* 被替换为 *range 语句*。

为了方便，所有的 LED 或者其引脚都被集中放入 *leds* 数组中。除此之外，所有的引脚都已经在它们被配置为输出之前设置为已知的初始状态（高电平）。

在这个例子中，我们想计时器以 1kHz 跳动。为了配置 TIM3 预分频器，我们需要知道它的输入时钟频率。根据参考手册，当 APBCLK = AHBCLK 时，输入时钟频率等于 APBCLK，否则为 2 倍 APBCLK。

如果 CNT 寄存器增加 1 kHz，那么 ARR 寄存器的值对应于以毫秒表示的更新事件（重载事件）的计数周期。为了让更新事件产生中断，在 DIER 寄存器中的 UIE 比特位必须被置位。CEN 比特位使能计时器。

外部定时器在低功耗模式下应该保持可用，这是为了在 CPU 睡眠时保持跳动：`timer.EnableClock(true)`。在 STM32F0 中这个没有关系，但是它对于代码的可移植性很重要。

*timerISR* 方法处理 *irq.TIM3* 中断请求。`timer.SR.Store(0)` 清除 SR 寄存器中的所有事件标志让 IRQ 到 [NVIC](http://infocenter.arm.com/help/topic/com.arm.doc.ddi0432c/Cihbecee.html) 无效。根据经验规则一般是在处理程序开始时，立即清除中断标志，因为 IRQ 无效会有延时。这就阻止了不明所以的再次调用处理器的情况。为了完全放心，清除读序列应该被运行，但是在我们的例子中，清理一下就足够了。

以下代码：

```go
select {
case ch <- 0:
	// Success
default:
	leds[0].Clear()
}
```

是使用 Go 的方式在一个 channel 上非阻塞地发送消息。没有一个中断处理程序能够在等待 channel 中的空闲空间。如果 channel 满了，执行 default，那么开发板上 LED 被点亮，直到下一次中断。

*ISRs* 数组包含中断向量。`//c:__attribute__((section(".ISRs")))` 会造成连接器将会把它插入到 .ISRs section 中。

新的 *blinky 的 for* 循环：

```go
for range ch {
	led.Clear()
	delay.Millisec(100)
	led.Set()
	delay.Millisec(period - 100)
}
```

等价于：

```go
for {
	_, ok := <-ch
	if !ok {
		break // Channel closed.
	}
	led.Clear()
	delay.Millisec(100)
	led.Set()
	delay.Millisec(period - 100)
}
```

注意，在这个例子中，我们对从 channel 中接收到的值不感兴趣。我们只在意这里能够接收到东西就行。我们可以通过声明 channel 的元素类型给予它空的结构体表达式 `struct{}` 而不是 *int*，同时发送 `struct{}{}` 值而不是 0，但是它会让才看到这个的人略感陌生。

让我们来编译这个代码：

```go
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  11096     228     188   11512    2cf8 cortexm0.elf
```

这个新的例子占用了 11324 字节的 Flash 空间，比之前的多了 1132 字节。

使用当前的时序，两个 *blinky* goroutines 从 channel 消费的速度比 *timerISR* 发送给它的速度快得多。因此，它们同时等待新数据到来，你可以观察到 [Go规范](https://golang.org/ref/spec#Select_statements) 所要求的 *select* 的随机性。

![STM32F030F4P6](https://ziutek.github.io/images/mcu/f030-demo-board/channels1.png)

开发板上的 LED 总是关闭的，因此 channel 没有出现溢出。

让我们来加快发送的速度，改变 `timer.ARR.Store(700)` 为 `timer.ARR.Store(200)`。现在 *timerISR* 每秒发送 5 条数据，但是两个接收者每秒同时只能接收 4 条消息。

![STM32F030F4P6](https://ziutek.github.io/images/mcu/f030-demo-board/channels2.png)

正如你所看到的，*timerISR* 点亮了黄色 LED，意味着在 channel 中没有空间了。

到这里，我完成了本文的第一部分。你应该清除这一部分没有为你展示 Go 语言中最重要的东西，*接口*。

Goroutines 和 channels 是很棒很便捷的语法。你可以用你自己的代码替换它们 - 这不容易但是可行。接口是 Go 的本质，这就是我将在本文的第二部分开始的内容。

我们仍然有空闲的 Flash 空间。

---

via: https://ziutek.github.io/2018/03/30/go_on_very_small_hardware.html

作者：[Michał Derkacz ](https://ziutek.github.io)
译者：[PotoYang](https://github.com/PotoYang)
校对：[DingdingZhou](https://blog.zhoudingding.com)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
首发于：https://studygolang.com/articles/24877

# Go 最小硬件编程（第三部分）

[![STM32F030F4P6](https://ziutek.github.io/images/mcu/f030-demo-board/board.jpg)](https://ziutek.github.io/2018/05/03/go_on_very_small_hardware3.html)

本系列的第一部分和第二部分中讨论的大多数示例都是以一种或另一种方式闪烁 LED。起初它可能很有趣，但过了一段时间它变得有点无聊。让我们做一些更有趣的事情......

......让我们点亮更多 LED！

## WS281x LEDs

[WS281x](http://www.world-semi.com/solution/list-4-1.html) RGB LED（和它们的克隆版）非常的流行。你可以将它们作为单个元素购买、链接成长条或组装成矩阵、环或其他形状。

![WS2812B](https://ziutek.github.io/images/led/ws2812b.jpg)

它们可以串联连接，由于这个原因，你可以通过 MCU 的单个引脚控制长 LED 条。不幸的是，它们的内部控制器使用的物理协议并不适合你可以在 MCU 中找到的任何外设。你必须使用 bit-banging 或以不寻常的方式使用可用的外设。

哪种可用解决方案最有效，取决于同时控制的 LED 灯条的数量。如果你必须驱动 4 到 16 个条带，最有效的方法是 [使用定时器和 DMA](http://www.martinhubacek.cz/arm/improved-stm32-ws2812b-library)（不要忽视 Martin 的文章末尾的链接）。

如果你只需要控制一个或两个条带，请使用可用的 SPI 或 UART 外设。对于 SPI，你只能在发送的一个字节中编码两个 WS281x 位。通过巧妙地使用起始位和停止位，UART 允许更密集的编码：每个字节发送 3 位。

我在 [这个网站](http://mikrokontrolery.blogspot.com/2011/03/Diody-WS2812B-sterowanie-XMega-cz-2.html) 上找到的 UART 协议如何适合 WS281x 协议的最佳解释。如果你不懂波兰语，这里是 [英文翻译版](https://translate.google.pl/translate?sl=pl&tl=en&u=http://mikrokontrolery.blogspot.com/2011/03/Diody-WS2812B-sterowanie-XMega-cz-2.html)。

基于 WS281x 的 LED 仍然是最受欢迎的，但市场上也有 SPI 控制的 LED：APA102](http://neon-world.com/en/product.php)，[SK9822](http://www.normandled.com/index.php/Product/view/id/800.html)。关于它们的三篇有趣的文章：[1](https://cpldcpu.wordpress.com/2014/08/27/apa102/)，[2](https://cpldcpu.wordpress.com/2014/11/30/understanding-the-apa102-superled/)，[3](https://cpldcpu.wordpress.com/2016/12/13/sk9822-a-clone-of-the-apa102/)。

## LED 环

市场上有许多基于 WS2812 的环状 LED。我弄了这个：

![WS2812B](https://ziutek.github.io/images/led/rgbring.jpg)

它有 24 个可单独寻址的 RGB LED（WS2812B），并有四个端子：GND、5V、DI 和 DO。你可以通过将 DI（数据输入）终端连接到前一个终端的 DO（数据输出）终端来链接更多环或其他基于 WS2812 的东西。

让我们将这个环连接到我们的 STM32F030 开发板上。我们将使用基于 UART 的驱动器，因此 DI 应连接到 UART 接头上的 TXD 引脚。WS2812B LED 需要至少 3.5V 电源。24 个 LED 可以消耗相当多的电流，因此在编程/调试过程中，最好将环上的 GND 和 5V 端子直接连接到 ST-LINK 编程器上的 GND 和 5V 引脚：

![WS2812B](https://ziutek.github.io/images/led/ring-stlink-f030.jpg)

我们的 STM32F030F4P6 MCU 和整个 STM32 F0、F3、F7、L4 系列有一个重要的东西，而 F1、F4、L1 MCU 没有：它允许反转 UART 信号，因此我们可以将环直接连接到 UART TXD 引脚。如果你不知道我们需要这样的反转，你可能没有阅读我上面提到的 [文章](https://translate.google.pl/translate?sl=pl&tl=en&u=http://mikrokontrolery.blogspot.com/2011/03/Diody-WS2812B-sterowanie-XMega-cz-2.html)。

所以你不能用这种方式使用 [Blue Pill](https://jeelabs.org/article/1649a/) 或 [STM32F4-DISCOVERY](http://www.st.com/en/evaluation-tools/stm32f4discovery.html) 。使用 SPI 外设或外部逆变器。请参阅 [Christmas Tree Lights](https://github.com/ziutek/emgo/tree/master/egpath/src/stm32/examples/minidev/treelights) 项目作为 UART + 逆变器的示例或使用 SPI 的 NUCLEO-F411RE 的 [WS2812 示例](https://github.com/ziutek/emgo/tree/master/egpath/src/stm32/examples/nucleo-f411re/ws2812)。

顺便说一句，可能大多数 DISCOVERY 开发板都有一个问题：它们工作在 VDD = 3V 而不是 3.3V。对于 DI 高电平，WS281x 至少需要 0.7 倍 *供给电压*。对于 5V 电源就是 3.5V，如果是 4.7V，则可以在 DISCOVERY 的 5V 引脚上找到 3.3V。如你所见，即使在我们的情况下，第一个 LED 工作电压低于额定电压 0.2V。在 DISCOVERY 的情况下，如果供电 4.7V，则工作电压低于额定电压 0.3V，如果供电 5V，则工作电压低于额定电压 0.5V。

让我们结束这个冗长的介绍并转到代码：

```go
package main

import (
	"delay"
	"math/rand"
	"rtos"

	"led"
	"led/ws281x/wsuart"

	"stm32/hal/dma"
	"stm32/hal/gpio"
	"stm32/hal/irq"
	"stm32/hal/system"
	"stm32/hal/system/timer/systick"
	"stm32/hal/usart"
)

var tts *usart.Driver

func init() {
	system.SetupPLL(8, 1, 48/8)
	systick.Setup(2e6)

	gpio.A.EnableClock(true)
	tx := gpio.A.Pin(9)

	tx.Setup(&gpio.Config{Mode: gpio.Alt})
	tx.SetAltFunc(gpio.USART1_AF1)

	d := dma.DMA1
	d.EnableClock(true)

	tts = usart.NewDriver(usart.USART1, d.Channel(2, 0), nil, nil)
	tts.Periph().EnableClock(true)
	tts.Periph().SetBaudRate(3000000000 / 1390)
	tts.Periph().SetConf2(usart.TxInv)
	tts.Periph().Enable()
	tts.EnableTx()

	rtos.IRQ(irq.USART1).Enable()
	rtos.IRQ(irq.DMA1_Channel2_3).Enable()
}

func main() {
	var rnd rand.XorShift64
	rnd.Seed(1)
	rgb := wsuart.GRB
	strip := wsuart.Make(24)
	black := rgb.Pixel(0)
	for {
		c := led.Color(rnd.Uint32()).Scale(127)
		pixel := rgb.Pixel(c)
		for i := range strip {
			strip[i] = pixel
			tts.Write(strip.Bytes())
			delay.Millisec(40)
		}
		for i := range strip {
			strip[i] = black
			tts.Write(strip.Bytes())
			delay.Millisec(20)
		}
	}
}

func ttsISR() {
	tts.ISR()
}

func ttsDMAISR() {
	tts.TxDMAISR()
}

//c:__attribute__((section(".ISRs")))
var ISRs = [...]func(){
	irq.USART1:          ttsISR,
	irq.DMA1_Channel2_3: ttsDMAISR,
}
```

### *import* 区域

对比之前的例子来说，*import* 区域新增的是 *rand/math* 包和 *led* 包及 *led/ws281x* 子包。*led* 包本身包含 *Color* 类型的定义。*led/ws281x/wsuart* 定义了 *ColorOrder*、*Pixel* 和 *Strip* 类型。

我想知道从 *image/color* 使用 *Color* 或是 *RGBA* 类型以及如何定义 *Strip*，它将实现 *image.Image* 接口，但是由于使用了 [伽马校正](https://en.wikipedia.org/wiki/Gamma_correction) 和 *image/draw* 大包，所以我简单的实现了：

```go
type Color uint32
type Strip []Pixel
```

同时加入一些有用的方法。然而，这个在未来是可以改变的。

### *init* 函数

*init* 函数没有太多新奇之处。UART 波特率从 115200 变为 3000000000/1390 ≈ 2158273，相当于每个 WS2812 位耗费 1390 纳秒。CR2 寄存器中的 *TxInv* 位设置为反转 TXD 信号。

### *main* 函数

*XorShift64* 伪随机数生成器用于生成随机颜色。[XORSHIFT](https://en.wikipedia.org/wiki/Xorshift) 是目前 *math/rand* 包实现的唯一算法。你必须使用带有非零参数的 *Seed* 方法显式地初始化它。

*rgb* 变量的类型为 *wsuart.ColorOrder*，并设置为 WS2812 使用的 GRB 颜色顺序（WS2811 使用 RGB 顺序）。然后它用于将颜色转换为像素。

`wsuart.Make(24)` 创建了 24 像素的初始化条带。它相当于：

```go
strip := make(wsuart.Strip, 24)
strip.Clear()
```

其余代码使用随机颜色绘制类似于 “Please Wait ...” 微调器的内容。

*strip* 切片充当帧缓冲区。`tts.Write(strip.Bytes())` 将帧缓冲区的内容发送到环。

### 中断

该程序使用处理中断的代码，与前一个 [UART 示例](https://ziutek.github.io/2018/04/14/go_on_very_small_hardware2.html#uart) 相同。

让我们编译并运行：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text	   data	    bss	    dec	    hex	filename
  14088	    240	    204	  14532	   38c4	cortexm0.elf
$ openocd -d0 -f interface/stlink.cfg -f target/stm32f0x.cfg -c 'init; program cortexm0.elf; reset run; exit'
```

我已经跳过了 openod 的输出。下面这个视频展示了这个程序是如何运行的：

<video width="576" height="324" controls="" preload="auto">
	<source src="https://ziutek.github.io/videos/rgbspinner.mp4" type="video/mp4">
	Sorry, your browser doesn't support embedded videos.
</video>

## 让我们来做一点有用的事......

在第一部分的开头，我问过：“我们使用 Go 最低能到多低，仍然做一些有用的事情？”。我们的 MCU 实际上是一个低端设备（8 位可能我的还是不赞同的），但到目前为止我们还没有做任何有用的事情。

那么......让我们做一些有用的事情......*让我们制作一个时钟吧*！

互联网上有许多由 RGB LED 构成的时钟示例。让我们自己使用我们的小开发板和 RGB 环。我们更改以前的代码，如下所述。

### *import* 区域

去掉 *math/rand* 包并添加 *stm32/hal/exti*。

### 全局变量

添加两个新的全局变量：*btn* 和 *btnev*：

```go
var (
	tts   *usart.Driver
	btn   gpio.Pin
	btnev rtos.EventFlag
)
```

它们将会被用于处理 button，用于设置我们的时钟。我们的开发板没有重置按钮，但是一定层度上我们可以不用它也能管理。

### *init* 函数

把这个代码添加到 *init* 函数中：

```go
btn = gpio.A.Pin(4)

btn.Setup(&gpio.Config{Mode: gpio.In, Pull: gpio.PullUp})
ei := exti.Lines(btn.Mask())
ei.Connect(btn.Port())
ei.EnableFallTrig()
ei.EnableRiseTrig()
ei.EnableIRQ()

rtos.IRQ(irq.EXTI4_15).Enable()
```

PA4 引脚被配置为输入，并使能内部上拉电阻。它连接到板载 LED，但不会妨碍任何事情。更重要的是它位于 GND 引脚旁边，因此我们可以使用任何金属物体来模拟按钮并设置时钟。作为奖励，我们可以从板载 LED 获得额外的反馈。

我们使用 EXTI 外设来跟踪 PA4 状态。它被配置为在任何更改时生成中断。

### *btnWait* 函数

定义一个新的辅助函数：

```go
func btnWait(state int, deadline int64) bool {
	for btn.Load() != state {
		if !btnev.Wait(1, deadline) {
			return false // timeout
		}
		btnev.Reset(0)
	}
	delay.Millisec(50) // debouncing
	return true
}
```

它等待 button 引脚上的指定状态，但仅在 *deadline* 出现之前。这是略微改进的轮询代码：

```go
for btn.Load() != state {
	if rtos.Nanosec() >= deadline {
		// timeout
	}
}
```

我们的 *btnWait* 函数，不是忙于等待 *state* 或 *deadline*，而是使用类型为 *rtos.EventFlag* 的 *btnev* 变量来睡眠，直到发生某些事情。你当然可以使用 channel 而不是 *rtos.EventFlag*，但后者要便宜得多。

### *main* 函数

我们需要完全新的 *main* 函数：

```go
func main() {
	rgb := wsuart.GRB
	strip := wsuart.Make(24)
	ds := 4 * 60 / len(strip) // Interval between LEDs (quarter-seconds).
	adjust := 0
	adjspeed := ds
	for {
		qs := int(rtos.Nanosec() / 25e7) // Quarter-seconds since reset.
		qa := qs + adjust

		qa %= 12 * 3600 * 4 // Quarter-seconds since 0:00 or 12:00.
		hi := len(strip) * qa / (12 * 3600 * 4)

		qa %= 3600 * 4 // Quarter-seconds in the current hour.
		mi := len(strip) * qa / (3600 * 4)

		qa %= 60 * 4 // Quarter-seconds in the current minute.
		si := len(strip) * qa / (60 * 4)

		hc := led.Color(0x550000)
		mc := led.Color(0x005500)
		sc := led.Color(0x000055)

		// Blend the colors if the hands of the clock overlap.
		if hi == mi {
			hc |= mc
			mc = hc
		}
		if mi == si {
			mc |= sc
			sc = mc
		}
		if si == hi {
			sc |= hc
			hc = sc
		}

		// Draw the clock and write to the ring.
		strip.Clear()
		strip[hi] = rgb.Pixel(hc)
		strip[mi] = rgb.Pixel(mc)
		strip[si] = rgb.Pixel(sc)
		tts.Write(strip.Bytes())

		// Sleep until the button pressed or the second hand should be moved.
		if btnWait(0, int64(qs+ds)*25e7) {
			adjust += adjspeed
			// Sleep until the button is released or timeout.
			if !btnWait(1, rtos.Nanosec()+100e6) {
				if adjspeed < 5*60*4 {
					adjspeed += 2 * ds
				}
				continue
			}
			adjspeed = ds
		}
	}
}
```

我们使用 *rtos.Nanosec* 函数而不是 *time.Now* 来获取当前时间。这节省了大量的 Flash 空间，但也让我们的时钟减弱为原始设备，不知道几天、几个月和几年，最糟糕的是它不处理夏令时的变化。

我们的环有 24 个 LED，因此秒针的精度可达 2.5 秒。为了不牺牲这种精度并获得平稳运行，我们使用四分之一秒作为基本间隔。半秒就足够了，但是四分之一秒更准确，并且适用于 16 和 48 个 LED。

红色、绿色和蓝色分别用于时针、分针和秒针。这允许我们使用简单的逻辑或操作进行颜色混合。我们有 *Color.Blend* 方法可以混合任意颜色，但我们的 Flash 空间很小，所以我们更喜欢最简单的解决方案。

我们只在秒针移动时才重绘时钟：

```go
btnWait（0，int64（qs + ds）* 25e7）
```

正在等待那个时刻或按下按钮。

每次按下按钮都会向前调整时钟。按住按钮一段时间后会有加速。

### 中断

定义新的中断处理程序：

```go
func exti4_15ISR() {
	pending := exti.Pending() & 0xFFF0
	pending.ClearPending()
	if pending&exti.Lines(btn.Mask()) != 0 {
		btnev.Signal(1)
	}
}
```

同时添加 `irq.EXTI4_15: exti4_15ISR`，进入 *ISR* 数组。

此处理程序（或中断服务程序）处理 EXTI4_15 IRQ。Cortex-M0 CPU 支持的 IRQ 明显少于其兄弟，因此你经常可以看到一个 IRQ 由多个中断源共享。在我们的例子中，12 个 EXTI 线共享一个 IRQ。

*exti4_15ISR* 读取所有 *pending* 位并选择其中 12 个更重要的位。接下来，它清除 EXTI 中的选择的位并开始处理它们。在我们的例子中，只检查第 4 位。`btnev.Signal(1)` 导致 `btnev.Wait(1, deadline)` 唤醒并返回 *true*。

你可以在 [Github](https://github.com/ziutek/emgo/tree/master/egpath/src/stm32/examples/f030-demo-board/ws2812-clock) 上找到完整的代码。我们来编译它：

```bash
$ egc
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  15960     240     216   16416    4020 cortexm0.elf
```

任何 iprovements 只有 184 个字节。让我们再次重建所有内容，但这次没有 typeinfo 中的任何类型和字段名称：

```bash
$ cd $HOME/emgo
$ ./clean.sh
$ cd $HOME/firstemgo
$ egc -nf -nt
$ arm-none-eabi-size cortexm0.elf
   text    data     bss     dec     hex filename
  15120     240     216   15576    3cd8 cortexm0.elf
```

现在，利用一千字节的空闲空间，你可以做一些事情了。让我们来看看程序如何运行的：

<video width="576" height="324" controls="" preload="auto">
	<source src="https://ziutek.github.io/videos/rgbclock.mp4" type="video/mp4">
	Sorry, your browser doesn't support embedded videos.
</video>
我不知道我怎么才能准确的显示 3:00！？

这就是全部了，大兄弟（大妹子）！ 在第 4 部分（结束本系列）中，我们将尝试在 LCD 上显示某些内容。

---

via: https://ziutek.github.io/2018/05/03/go_on_very_small_hardware3.html

作者：[Michał Derkacz](https://ziutek.github.io)
译者：[PotoYang](https://github.com/PotoYang)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

首发于：https://studygolang.com/articles/21762

# 使用 Go 实现简单的事件总线

![pic_1](https://raw.githubusercontent.com/studygolang/gctt-images/master/let%E2%80%99s-write-a-simple-event-bus-in-go/pic_1.png)

事件驱动架构是计算机科学中一种高度可扩展的范例。它允许我们可以多方系统异步处理事件。

事件总线是[发布/订阅模式](https://en.wikipedia.org/wiki/Publish%E2%80%93subscribe_pattern)的实现，其中发布者发布数据，并且感兴趣的订阅者可以监听这些数据并基于这些数据作出处理。这使发布者与订阅者松耦合。发布者将数据事件发布到事件总线，总线负责将它们发送给订阅者。

![pic_2](https://raw.githubusercontent.com/studygolang/gctt-images/master/let%E2%80%99s-write-a-simple-event-bus-in-go/pic_2.png)

传统的实现事件总线的方法会涉及到使用回调。订阅者通常实现接口，然后事件总线通过接口传播数据。

使用 Go 的并发模型，我们知道在大多数地方可以使用 `channel` 来替代回调。在本文中，我们将重点介绍如何使用 `channel` 来实现事件总线。

> 我们专注于**基于主题（topic）的事件**。发布者发布到主题，订阅者可以收听它们。

## 定义数据结构

为了实现事件总线，我们需要定义要传递的数据结构。我们可以使用 `struct` 简单地创建一个新的数据类型。我们定义一个 `DataEvent` 的结构体如下：

```go
type DataEvent struct {
   Data interface{}
   Topic string
}
```

在这里，我们已经将基础数据定义为接口，这意味着它可以是任何值。我们还将主题定义为结构的成员。订阅者可能会收听**多个主题**，因此，我们通过主题来让订阅者可以区分不同的事件的做法是不错的。

## 介绍 channels

现在我们已经为事件总线定义了我们主要的数据结构，我们还需要一种方法来传递它。为此，我们可以定义一个可以传播 `DataEvent` 的 `DataChannel` 类型。

```go
// DataChannel 是一个能接收 DataEvent 的 channel
type DataChannel chan DataEvent

// DataChannelSlice 是一个包含 DataChannels 数据的切片
type DataChannelSlice [] DataChannel
```

`DataChannelSlice` 的创建是为了保留 `DataChannel` 的切片并轻松引用它们。

## 事件总线

```go
// EventBus 存储有关订阅者感兴趣的特定主题的信息
type EventBus struct {
   subscribers map[string]DataChannelSlice
   rm sync.RWMutex
}
```

`EventBus` 有 `subscribers`，这是一个包含 `DataChannelSlices` 的 map。我们使用互斥锁来保护并发访问的读写。

通过使用 `map` 和定义 `topics` ，它允许我们轻松地组织事件。主题被视为 `map` 的键。当有人发布它时，我们可以通过键轻松找到主题，然后将事件传播到 `channel` 中以进行进一步处理。

## 订阅主题

对于订阅主题，使用 `channel`。它就像传统方法中的回调一样。当发布者向主题发布数据时，`channel`将接收数据。

```go
func (eb *EventBus)Subscribe(topic string, ch DataChannel)  {
   eb.rm.Lock()
   if prev, found := eb.subscribers[topic]; found {
      eb.subscribers[topic] = append(prev, ch)
   } else {
      eb.subscribers[topic] = append([]DataChannel{}, ch)
   }
   eb.rm.Unlock()
}
```

简单地说，我们将订阅者添加到 `channel` 切片中然后给该结构加锁，最后在操作后将其解锁。

## 发布主题

要发布事件，发布者需要提供广播给订阅者所需要的主题和数据。

```go
func (eb *EventBus) Publish(topic string, data interface{}) {
   eb.rm.RLock()
   if chans, found := eb.subscribers[topic]; found {
      // 这样做是因为切片引用相同的数组，即使它们是按值传递的
      // 因此我们正在使用我们的元素创建一个新切片，从而能正确地保持锁定
      channels := append(DataChannelSlice{}, chans...)
      go func(data DataEvent, dataChannelSlices DataChannelSlice) {
         for _, ch := range dataChannelSlices {
            ch <- data
         }
      }(DataEvent{Data: data, Topic: topic}, channels)
   }
   eb.rm.RUnlock()
}
```

在此方法中，首先我们检查主题是否存在任何订阅者。然后我们只是简单地遍历与主题相关的 `channel` 切片并把事件发布给它们。

> 请注意，我们在发布方法中使用了 Goroutine 来避免阻塞发布者

## 开始

首先，我们需要创建一个事件总线的实例。在实际场景中，你可以从包中导出单个 `EventBus`，**使其像单例一样运行**。

```go
var eb = &EventBus{
   subscribers: map[string]DataChannelSlice{},
}
```

为了测试新创建的事件总线，我们将创建一个以随机间隔时间发布到指定主题的方法。

```go
func publisTo(topic string, data string)  {
   for {
      eb.Publish(topic, data)
      time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
   }
}
```

接下来，我们需要一个可以收听主题的 main 函数。它使用辅助方法打印出事件的数据。

```go
func printDataEvent(ch string, data DataEvent)  {
   fmt.Printf("Channel: %s; Topic: %s; DataEvent: %v\n", ch, data.Topic, data.Data)
}
func main()  {
   ch1 := make(chan DataEvent)
   ch2 := make(chan DataEvent)
   ch3 := make(chan DataEvent)
   eb.Subscribe("topic1", ch1)
   eb.Subscribe("topic2", ch2)
   eb.Subscribe("topic2", ch3)
   go publisTo("topic1", "Hi topic 1")
   go publisTo("topic2", "Welcome to topic 2")
   for {
      select {
      case d := <-ch1:
         go printDataEvent("ch1", d)
      case d := <-ch2:
         go printDataEvent("ch2", d)
      case d := <-ch3:
         go printDataEvent("ch3", d)
      }
   }
}
```

我们创建了三个可以订阅主题的 `channels` 订阅者（ch1，ch2，ch3）。其中 ch2 和 ch3 这两个监听同一事件。

我们使用 select 语句从最快返回的 `channel` 中获取数据。然后它使用另一个 Goroutine 打印输出数据。用 Goroutine 也不是必需的。但在某些情况下，你必须对事件进行一些繁重的操作处理。为了防止阻塞 select，我们使用了 Goroutine。

示例输出将如下所示

```
Channel: ch1; Topic: topic1; DataEvent: Hi topic 1
Channel: ch2; Topic: topic2; DataEvent: Welcome to topic 2
Channel: ch3; Topic: topic2; DataEvent: Welcome to topic 2
Channel: ch3; Topic: topic2; DataEvent: Welcome to topic 2
Channel: ch2; Topic: topic2; DataEvent: Welcome to topic 2
Channel: ch1; Topic: topic1; DataEvent: Hi topic 1
Channel: ch3; Topic: topic2; DataEvent: Welcome to topic 2
...
```

你可以看到事件总线通过 `channel` 分发事件。

基于简单 `channel` 的事件总线的源代码。

## 完整的代码

```go
package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type DataEvent struct {
	Data  interface{}
	Topic string
}

// DataChannel 是一个能接收 DataEvent 的 channel
type DataChannel chan DataEvent

// DataChannelSlice 是一个包含 DataChannels 数据的切片
type DataChannelSlice []DataChannel

// EventBus 存储有关订阅者感兴趣的特定主题的信息
type EventBus struct {
	subscribers map[string]DataChannelSlice
	rm          sync.RWMutex
}

func (eb *EventBus) Publish(topic string, data interface{}) {
	eb.rm.RLock()
	if chans, found := eb.subscribers[topic]; found {
		// 这样做是因为切片引用相同的数组，即使它们是按值传递的
		// 因此我们正在使用我们的元素创建一个新切片，从而正确地保持锁定
		channels := append(DataChannelSlice{}, chans...)
		go func(data DataEvent, dataChannelSlices DataChannelSlice) {
			for _, ch := range dataChannelSlices {
				ch <- data
			}
		}(DataEvent{Data: data, Topic: topic}, channels)
	}
	eb.rm.RUnlock()
}

func (eb *EventBus) Subscribe(topic string, ch DataChannel) {
	eb.rm.Lock()
	if prev, found := eb.subscribers[topic]; found {
		eb.subscribers[topic] = append(prev, ch)
	} else {
		eb.subscribers[topic] = append([]DataChannel{}, ch)
	}
	eb.rm.Unlock()
}

var eb = &EventBus{
	subscribers: map[string]DataChannelSlice{},
}

func printDataEvent(ch string, data DataEvent) {
	fmt.Printf("Channel: %s; Topic: %s; DataEvent: %v\n", ch, data.Topic, data.Data)
}

func publisTo(topic string, data string) {
	for {
		eb.Publish(topic, data)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	}
}

func main() {
	ch1 := make(chan DataEvent)
	ch2 := make(chan DataEvent)
	ch3 := make(chan DataEvent)

	eb.Subscribe("topic1", ch1)
	eb.Subscribe("topic2", ch2)
	eb.Subscribe("topic2", ch3)

	go publisTo("topic1", "Hi topic 1")
	go publisTo("topic2", "Welcome to topic 2")

	for {
		select {
		case d := <-ch1:
			go printDataEvent("ch1", d)
		case d := <-ch2:
			go printDataEvent("ch2", d)
		case d := <-ch3:
			go printDataEvent("ch3", d)
		}
	}
}
```

## 使用 channel 取代回调的理由

传统的回调方式要求实现某种接口。

例如，

```go
type Subscriber interface {
   onData(event Event)
}
```

使用回调的话，如果你想订阅一个事件，你需要实现该接口，以便事件总线可以传播它。

```go
type MySubscriber struct {
}
func (m MySubscriber) onData(event Event)  {
   // 处理事件
}
```

而 `channel` 允许你在没有接口的情况下在一个简单的函数中注册订阅者。

```go
func main() {
   ch1 := make(chan DataEvent)
   eb.Subscribe("topic1", ch1)
   fmt.Println((<-ch1).Data)
   ...
}
```

## 结论

本文的目的是指出编写事件总线的不同实现方法。

> 这可能不是理想的解决方案。

例如，`channel` 被阻塞直到有人消费它们。这有一定的局限性。

> 我已经使用切片来存储主题的所有订阅者。这用于简化文章。这需要用 **SET** 替换，以至于列表中不存在重复的订阅者。

传统的回调方法可以使用提供的相同的原理去简单地实现。你可以轻松地在 Goroutine 中进行异步装饰发布事件。

我很想听听你对这篇文章的看法。 :)

---

via: https://levelup.gitconnected.com/lets-write-a-simple-event-bus-in-go-79b9480d8997

作者：[Kasun Vithanage](https://levelup.gitconnected.com/@kasvith)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网] (https://studygolang.com/) 荣誉推出

**Go语言是Google的语言，不是我们的。**

在Twitter上，我看到了下面的问题：

“在Go话题下有很多人在讨论泛型，为什么我们不能拥有像OpenGo这样的东西，通过社区就可以实现Go的泛型，而非要等待官方的实现。就像OpenJDK一样”

对于这个问题有很多答案，但是没有人大大方方的说出来：Go是Google的语言，不是社区的。

是的，是有一个社区在为Go做贡献，其中不乏一些重要和有价值的东西;你只需要去看看[贡献者](https://github.com/golang/go/blob/master/CONTRIBUTORS)排行榜或者是[commit](https://github.com/golang/go/blob/master/CONTRIBUTORS)提交榜单。但是google是这些社区贡献的守护者，只有它决定了Go可以接受什么。某种程度来说，如果说需要一套社区流程来决定接受什么，那么房间里可能会有一个800磅的大猩猩。如果google反对，那么没有什么可以添加到Go中去，如果google决定某些东西需要在Go中，它就会发生。

(最明显和显而意见的例子就是Go的Models，Google公司的Go核心团队的一名成员放弃了整个系统，但外部的go社区一直支持着一个[相对完全不同的Model](https://research.swtch.com/vgo)。可以[参见](https://peter.bourgon.org/blog/2018/07/27/a-response-about-dep-and-vgo.html)这个故事的另一个版本)

简而言之，Go有社区贡献者但却不是一个社区项目。他是google的项目。这是无可争辩的事情，无论你认为他是好是坏，他都需要我们接受。举例来说，如果你有一些重要的想法想让go接受，那么说服Go核心团队远比努力在社区中建立共识要重要的多。

(因此，将大量的时间和精力投入到一个没有Go核心团队大力支持的社区工作中可能是在浪费时间，最多你的工作可以帮助Go团队更好的理解问题。同样，请参阅Go模块中的实际操作。)

总的来说，很明显社区的声音对Go的发展并不重要，我们这些工作在Google之外的人只能接受这个结果。如果我们非常幸运，我们的优先级和Google相匹配，并且我们足够幸运的话，Go核心团队和google会确定他们有足够关心我们的优先级并处理他们。好消息是目前为止，Google和Go核心团队对于Go是否在外部取得成功非常关心，而不仅仅在google内部，所以他们愿意为了解决痛点而工作。

(无论是好是坏，一种普通的感觉就是Go做的很好，因为它有一个小巧的核心团队，拥有良好的品味和一致的语言愿景，这个团队不受外界的影响并且行动迟缓，倾向于尽量不做改变。)

PS：我喜欢Go语言并且到现在有一段时间了，我对Go目前的演变和Go核心团队的管理感到满意。我当然认为慢慢的采用泛型是个好主意。但同时,Go的去模块化发展在我看来是一个不好的尝试,现在我无法想象我自己成为一个贡献者,即使是微小的变化(一个方法,我没有兴趣知道我总是会一个二等公民)。我将继续提交bug报告，但仅此而已。整个情形给我留下了模棱两可的感觉，所以我通常会完全忽略它。

Go团队声称他们非常关心社区，希望大家参与其中，这听起来很可笑。我承认他们很关心社区，但只停留在某些点上。我认为Go核心团队应该更真诚的面对，而不是假装和含蓄的领导人们。

题外话：Google和Go核心团队
自从Go的发展方向被这个核心团队控制之后，你可以问问Go是Google的语言还是Go核心团队的语言。然而目前我认为绝大部分活跃的Go核心团队都在Google工作，实际中不可能察觉到区别(至少从Google外部看来)。事实上我们只有等到Go核心团队成员开始离开Google并继续保持活跃来确定Go的方向的时候才会知道Go到底属于谁。如果这发生了，尤其是大多数成员都离开了google，我们才会说Go大概是他们的语言，不再属于Google。就像python一直是Guido van Rossum的语言一样。

在实践过程中，不可否认的是Google目前提供了很多基础设施和资源来支持Go，比如[golang.org](https://golang.org)，并且因此拥有了域名等等。Google还拥有作为编程语言的“Go”商标，[商标列表](https://www.google.com/permissions/trademark/trademark-list/)参考.

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoIsGooglesLanguage

作者：Chris Siebenmann

译者：carsickcars(https://github.com/carsickcars)

本文由 GCTT 原创编译，Go 中文网 荣誉推出
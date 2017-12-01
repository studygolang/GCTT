
【翻译中 by arthurlee】

# [注释你的代码](https://npf.io/2017/11/comments/)
2017年11月17日 作者：[Nate Finch](https://npf.io)

每隔一段时间，网上总会出现一些令人不安的帖子，说注释是不好的，它存在的唯一原因是因为你们的代码不够干净。我完全不能认同。

## 不良代码
他们这样说也不完全是错的，我们的代码肯定不是足够好的。代码总会慢慢被腐蚀，你知道什么时候代码最烂吗？当你六个月不碰这些代码，回过头再度回顾的时候，你会好奇：“作者到底是怎么想的？”（于是使用 git blame 来查看历史记录，没想到代码竟然是自己写的。）

反对注释者的论点是只有不够“干净”的代码才需要注释。如果重构、命名、书写都非常好，那么也就不需要注释了。

今天，当整个项目和问题空间都在你的脑袋里的时候，你自然会觉得代码是干净、清晰和优雅的。但是，当六个月后，这些代码可能对你已经有些模糊；又或是，CTO 正好在生产系统上突然发现一个关键性的 bug，在主管紧盯的情况下，一些可怜的家伙不得不去调试你的代码。

一段你理解的代码，去想象其他人不能理解的原因，这个是非常难以掌握的技能。不过，这个具有无可估量的价值，几乎和第一次就能写出好的代码的能力一样重要。在软件行业中，几乎没有人是独行侠。即使真的一个人写代码，你也会忘记为什么写这样的代码或者昨天深夜“工程代码”的核心部分的确切目的。一旦你离职，接替你的人不得不去理解每一个仅在你的脑袋里的小偏好和诀窍。

所以，写一个即使在现在看来过于浅显的注释不是一个坏事情。有时候，甚至会带来巨大的帮助。

## 不写注释经常导致代码更难以理解
一些人声称如果移除注释，将会使代码更好，因为你会用更清晰的代码来补偿。我对此亦不以为然，因为我不认为有人会实际写上一些次佳的代码，并且放注释来解释（除了 // TODO: 这是一个临时的解决方法，我会稍后修正 之外）。我们都会写出在各种外部条件下（通常是时间）认为最好的代码。

为去除注释而重构代码的问题在于，这个经常会导致更坏的代码，而不是更好的代码。权威的例子是重构一行复杂的代码，将之提取到一个名字望文生义的函数中。这个听起来不错，除了现在你为正在阅读代码的人引入了一个上下午切换。替代真实代码的是一个函数调用，他们不得不滚动到函数定义的地方，记住和对照函数声明和调用的参数，并且将函数返回代入到调用的地方。

另外，清晰的函数名仅仅适合非常短小的注释。任何超过一小段短语的注释不能（或者不应该）形成一个函数名。因此，你最终会得到一个头有注释的函数。

的确，一个非常短小的函数就可能导致困惑和更复杂的代码。如果看到这样的函数，我就会去搜索这个函数被调用的所有地方。如果只有只有一个地方使用，我就会去考虑这是否确实是一个封装了全局逻辑的通用代码块（譬如 `NameToUserID`），或者这个函数是严重依赖调用端的特定状态和实现的定制代码，并且不能在其他地方正确工作。随着把这些放到一个函数里面，你本质上在其余的代码中暴露了实现细节，不应该这么草率的做出这样的决定。即使你知道这不是一个其他人应该调用的函数，当即便这样做不合适，其他人还会在某个时刻调用之。

小函数的问题在 Cindy Sridharan在 [medium 网站上的帖子](https://medium.com/@copyconstruct/small-functions-considered-harmful-91035d316c29)中有更加详细的阐述。

我们甚至可以深入到厂变量名和短的的比较，但是就此打住吧，一般你不可能接受更长的变量名了。除非你的变量名就是你想写的完整的注释，否则你还是会丢失信息而不得不添加到注释中。我认为我们可以达成一致，`usernameStrippedOfSpacesWithDotCSVExtension` 是一个可怕的变量名称。

我不是说我们不应该尽量去让我们的代码清晰和优雅，当然需要，这是一个出色的开发人员的特点。但是，代码清晰性和有注释是正交的，好的注释也是出色的开发人员的特点。

## 没有不良的注释
The examples of bad comments often given in these discussions are trivially bad, and almost never encountered in code written outside of a programming 101 class.

// instantiate an error
var err error
Yes, clearly, this is not a useful comment. But at the same time, it’s not really harmful. It’s some noise that is easily ignored when browsing the code. I would rather see a hundred of the above comments if it means the dev leaves in one useful comment that saves me hours of head banging on keyboard.

I’m pretty sure I’ve never read any code and said “man, this code would be so much easier to understand if it weren’t for all these comments.” It’s nearly 100% the opposite.

In fact, I’ll even call out some code that I think is egregious in its lack of comments - the Go standard library. While the code may be very correct and well structured.. in many cases, if you don’t have a deep understanding of what the code is doing before you look at the it, it can be a challenge to understand why it’s doing what it’s doing. A sprinkling of comments about what the logic is doing and why would make a lot of the go standard library a lot easier to read. In this I am specifically talking about comments inside the implementation, not doc comments on exported functions in general (those are generally pretty good).

## Any comment is better than no comment
Another chestnut the anti-commenters like to bring out is the wisdom can be illustrated with a pithy image:


Ah, hilarious, someone updated the contents and didn’t update the comment.

But, that was a problem 20 years ago, when code reviews were not (generally) a thing. But they are a thing now. And if checking that comments match the implementation isn’t part of your code review process, then you should probably review your code review process.

Which is not to say that mistakes can’t be made… in fact I filed a “comment doesn’t match implementation” bug just yesterday. The saying goes something like “no comment is better than an incorrect comment” which sounds obviously true, except when you realize that if there is no comment, then devs will just guess what the code does, and probably be wrong more often than a comment would be wrong.

Even if this does happen, and the code has changed, you still have valuable information about what the code used to do. Chances are, the code still does basically the same thing, just slightly differently. In this world of versioning and backwards compatbility, how often does the same function get drastically changed in functionality while maintaining the same name and signature? Probably not often.

Take the bug I filed yesterday… the place where we were using the function was calling client.SetKeepAlive(60). The comment on SetKeepAlive was “SetKeepAlive will set the amount of time (in seconds) that the client should wait before sending a PING request”. Cool, right? Except I noticed that SetKeepAlive takes a time.Duration. Without any other units specified for the value of 60, Go’s duration type defaults to…. nanoseconds. Oops. Someone had updated the function to take a Duration rather than an Int. Interestingly, it did still round the duration down to the nearest second, so the comment was not incorrect per se, it was just misleading.

## Why?
The most important comments are the why comments. Why is the code doing what it’s doing? Why must the ID be less than 24 characters? Why are we hiding this option on Linux? etc. The reason these are important is that you can’t figure out the why by looking at the code. They document lessons learned by the devs, outside constraints imposed by the business, other systems, etc. These comments are invaluable, and almost impossible to capture in other ways (e.g. function names should document what the function does, not why).

Comments that document what the code is doing are less useful, because you can generally always figure out what the code is doing, given enough time and effort. The code tells you what it is doing, by definition. Which is not to say that you should never write what comments. Definitely strive to write the clearest code you can, but comments are free, so if you think someone might misunderstand some code or otherwise have difficulty knowing what’s going on, throw in a comment. At least, it may save them a half hour of puzzling through your code, at best it may save them from changing it or using it in incorrect ways that cause bugs.

## Tests
Some people think that tests serve as documentation for functions. And, in a way, this is true. But they’re generally very low on my list of effective documentation. Why? Well, because they have to be incredibly precise, and thus they are verbose, and cover a narrow strip of functionality. Every test tests exactly one specific input and one specific output. For anything other than the most simple function, you probably need a bunch of code to set up the inputs and construct the outputs.

For much of programming, it’s easier to describe briefly what a function does than to write code to test what it does. Often times my tests will be multiple times as many lines of code as the function itself… whereas the doc comment on it may only be a few sentences.

In addition, tests only explain the what of a function. What is it supposed to do? They don’t explain why, and why is often more important, as stated above.

You should definitely test your code, and tests can be useful in figuring out the expected behavior of code in some edge cases… but if I have to read tests to understand your code in general, then that’s red flag that you really need to write more/better comments.

## Conclusion
I feel like the line between what’s a useful comment and what’s not is difficult to find (outside of trivial examples), so I’d rather people err on the side of writing too many comments. You never know who may be reading your code next, so do them the favor you wish was done for you… write a bunch of comments. Keep writing comments until it feels like too many, then write a few more. That’s probably about the right amount.

----------------

via: https://npf.io/2017/11/comments/

作者：[Nate Finch](https://npf.io/about/)
译者：[译者ID](https://github.com/译者ID)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推

# 当在 Go 中使用微服务还不够时：介绍 DDD Lite
当我开始用 Go 工作时，社区并不看好类似 DDD（Domain-Driven Design 领域驱动设计）和清晰架构这样的技术。我很多次听到这样的声音：*“不要在 Golang 中用 Java！”，“我已经在 Java 中见过了，请别这样做！”*。

这些时候，我已经有了近10年的 PHP 和 Python 经验。我已经见过太多糟糕的事情了。我记得所有那些“八千行”（有着8千行以上代码的方法 😉）和没有人愿意维护的应用。我查看了这些丑陋怪物以前的 git 历史，他们最初看起来是无害的。但是随着时间的推移，微小，无辜的问题开始变得越发明显且越发严重。**我也同样见过 DDD 和清晰架构如何解决了这些问题。**

也许 Golang 是不一样的？也许用 Golang 写微服务可以修复这些问题？

**原本应当是美好的**

现在，与许多人交流了经验，且能够看到很多代码库之后，我的观点较三年前更清晰了一些。很不幸，我现在并不认为仅靠着 Golang 和微服务就可以解决我之前面对过的那些问题。我开始回顾过去困难的那些时候。

由于是相对初期的代码库，这些问题不太明显。由于 Golang 设计，这些问题还是不太明显。但是我确定随着时间推移，我们会有越来越多的，没人愿意维护的老的 Golang 应用。

幸亏，3年前我并没有因别人的冷嘲热讽而放弃。我觉定在 Go 中尝试应用我原来在工作中用到的 DDD 以及相关技术。我与 Milosz 一起带领团队3年，并且都成功使用了 DDD，清晰架构，和所有相关的，但在 Golang 中还不受欢迎的技术。**这些技术使我们能够以恒定的速度开发我们的应用和产品，而不管代码的年头有多久。**

（DDD 以及相关技术）从一开始就效果显著，其他技术的1：1 移动模式将不会起作用。必不可少的，我们不会抛弃惯用的 Go 代码和微服务架构——它们完美契合！

今天我会先分享最简单直接的技术—— DDD lite。

**Golang 中 DDD 的状态**

坐下来写这篇文章之前，我在 Google 上查看了关于 Go 语言中 DDD 的几个文章。不客气地说，这些文章都没抓住应用 DDD 最核心的东西。**如果我读这些文章的时候根本不了解 DDD 的话，我想我不会在自己的团队应用这些技术。这种肤浅的方法也许也是 DDD 没有在 Go 社区推广的一个原因。**

在这个系列中，我们会展示所有必要的技术，同时以使用的方式来实现。在描述任何一个模式之前，我们以一个问题开始：它能给我们带来什么？这是一个挑战我们当前思想的一个不错的方式。

我确定我们可以通过这系列文章改变 Go 社区对这些技术的接受度。我们相信这些技术是实现复杂业务项目的最好方法。**我相信我们会在确立 Go 的地位上能有所贡献，使其成为构建基础设施以及业务软件的出色的语言。**

**你需要慢下来，才能走得更快**

以最简单的方式实现项目或许是诱人的，甚至当你感受到来自“上层”的压力时会更加诱人。我们用了微服务吗？如果需要的话，我们会仅仅重写该服务吗？我听到这种事情好多次了，绝大部分最后都不如人意。😉**老实说最简单方式实现会短期节省一些时间。但仅仅是短期。**

考虑下任何形式的测试用例的例子。你可以在项目开始时省略写测试用例。你明显会节省一些时间，并且管理者也会满意。**计算方式似乎很简单——项目更快交付了。**

但是长远来看，这种走捷径的方式是不值得的。随着工程的成长，团队会开始惧怕对其做任何的修改。最终，（开发）消耗的总时间会高于在一开始实现了测试用例所需要的时间。**开始为了快速的性能提升而牺牲质量的做法在长远上会拖慢脚步。** 另一方面——如果项目不是很关键，且需要快速创建，可以省略测试用例。这应当是一个务实的决定，而不仅仅是所谓 *“我们了解的更多，我们不会有 bug”。*

DDD 的情形也类似。当你想要使用 DDD，你需要在开始的时候需要一点更多的时间，但是长线来说会节省更多。然而，并不是所有的项目都足够复杂到需要使用类似 DDD 这样的高级技术。

**没有质量与开发速度的权衡，如果想要长期快速迭代，就需要保持高质量**

![高质量的软件值得吗？](https://github.com/studygolang/gctt-images2/blob/master/20200701-When-microservices-in-Go-are-not-enough-introduction-to-DDD-Lite/quality-fowler.png?raw=true)

**听起来很棒，但是是否有证据证明它可行？**

如果你在两年前问我这个问题，我会说： *“好吧，我觉得这样效果更好！”*。但是仅仅相信我的话似乎还不够。😉有太多的教程展示了愚蠢的想法和主张，而它们可没有什么证据——不要盲目相信他们。

要记住：如果某人有成千上万的 Twitter 关注这，[仅凭这个可不是相信他们的理由](https://en.wikipedia.org/wiki/Authority_bias)！

幸运的是，2年前，[《Accelerate: The Science of Lean Software and DevOps: Building and Scaling High Performing Technology Organizations》](https://www.amazon.com/Accelerate-Software-Performing-Technology-Organizations/dp/1942788339)发布了。简单来说，这本书描述了影响开发团队表现的因素。但是这本书可不是靠着一堆未经证实的想法而出名的—— **而是基于科学探究。**

**我最感兴趣的是展示到底是什么塑造了高水平团队的那部分内容**。这本书展示了几个明显的因素，比如引进了 DevOps，CI/CD，以及松耦合架构，这些都是高水平团队必不可少的因素。

> 如果像 DevOps 和 CI/CD 对你来说并不明显，可以先看看这些书：[《The Phoenix Project》](https://www.amazon.com/Phoenix-Project-DevOps-Helping-Business/dp/0988262592) 和 [《The DevOps Handbook》](https://www.amazon.com/DevOps-Handbook-World-Class-Reliability-Organizations/dp/1942788002)。

那么 《Accelerate》所告诉我们的塑造高水平团队的因素有什么？

> 我们发现，只要系统与构建以及维护它们的团队是松耦合的，所有类型的系统是有可能表现出高性能的。
>
> 这一关键架构属性使得团队即使在组织以及运作的系统数量不断增长也可以很容易地测试和部署单个组件或服务。它使组织可以在扩展规模时提高生产力。

所以，我们使用微服务的话，就可以了吗？如果使用微服务就足够的话，我就不用写这篇文章了。 😉

> - 对系统设计进行大规模修改，而不依赖其他团队修改它们的系统或者不造成其他团队的大量工作。
> - 无需与外部团队沟通协调就可以完整自己的工作。
> - 按需求部署或发布服务，不需要关系依赖的其他服务。
> - 不需要集成测试环境即可按需求进行大多数测试，并且可以在正常时间部署，停机时间可以忽略不计。
>
> 不幸的是，在现实生活中，许多所谓的面向服务架构不允许服务间独立测试和部署。因此无法使团队获得更高的表现。
>
> [...] 如果忽视了这些特点，即使采用了最新的微服务架构，在容器上进行部署，也无法获得更高的表现。[...] 为了获得这些特点，设计系统要松耦合——这可以让服务间独立更改和验证。

**仅仅使用微服务架构将服务拆分为更小单位是不够的。如果以错误的方式实现，这么做会增加额外的复杂度并行拖慢团队的节奏**。DDD 可以帮助我们。

我提到 DDD 这一术语好几次了。DDD 实际上是什么？

## 什么是 DDD （领域驱动设计）

先看下 Wikipedia 的定义：
> 领域驱动设计（DDD）是一个概念，代码的结构和语言（类名称，类方法，类变量）应当与业务领域相匹配。比如，如果是处理贷款申请的软件，该软件可能会有诸如 LoanApplication 和 Customer 的类，以及像 AcceptOffer 和 Withdraw 之类的方法。

![](https://github.com/studygolang/gctt-images2/blob/master/20200701-When-microservices-in-Go-are-not-enough-introduction-to-DDD-Lite/no-god-please-no.jpg?raw=true)

好吧，这不是一个完美的解释。 😅 它仍然缺少了一些重要的要点。

值得一提的是，DDD 在2003年提出。算是很早了。它一些留下来的精华也许能对 DDD 在2020年的今天以及 Go 的环境下的应用有所帮助。

> 如果你对 DDD 诞生的历史背景有兴趣，可以看看[解决软件核心中的复杂性](https://youtu.be/dnUFEg68ESM?t=1109)，演讲者是 DDD 的创造者—— Eric Evans

![Eric Evans DDD 的创造者，请把这个图片答应下来挂在床头，来获得 +10 的 DDD 保佑](https://github.com/studygolang/gctt-images2/blob/master/20200701-When-microservices-in-Go-are-not-enough-introduction-to-DDD-Lite/eric-evans.jpg?raw=true)

我对 DDD 的简单定义是：保证以**最佳方式**解决**有效问题**。之后**以你的业务逻辑会被理解的方式实施解决方案，而不需要技术语言的额外翻译。**

如何实施呢？

**编码就好比打仗，取胜需要策略！**

我喜欢说 *“5天的编码能节约15分钟的计划时间”。*

在开始编码之前，需要确定要解决的是一个有效的问题。听起来似乎是废话，但是以我的经验看，这并不像听起来那么容易。通常的情况是由工程师所创建的解决方案并没有实际解决业务需要解决的问题。在这一领域对我们有帮助的一系列模式被命名为**策略 DDD 模式（Strategic DDD patterns）。**

根据我的经验，DDD 策略模式（DDD Strategic Patterns）经常被忽略。原因很简单：我们都是开发，我们更喜欢写代码而不是与 *“业务人员（business people）”* 交谈。😉 不幸的是，自我封闭，不与业务人员交流的方法有很多缺点。对业务缺少信任，对系统如何运作缺少认知（业务侧和研发侧都有这个问题），解决错误的问题——这些仅仅是最常见的问题中的一部分。

好消息是大多数情况是由于缺少类似事件风暴（Event Storming）这样合适的技术导致的。这些技术可以为双方都带来好处。令人惊讶的是交流业务逻辑可能是工作中最有意思的一部分。

除此之外，我们会从适用于代码的模式开始。这些模式会带给我们**一些** DDD 的好处。它们也会更快的对你产生用处。**没有策略模式，我得说这样你仅仅会拥有 DDD 可以带来的优势中的 30%，在下一个文章中，我们会回到策略模式上。**

## Go中的 DDD Lite
在相当长的介绍之后，终于到了接触一些代码的时候了！在这篇文章中，我们会涵盖 **Go中的战术领域驱动设计模式（Tactical Domain-Driven Design patterns in Go）** 的一些基本知识。请记住这仅仅是开始。会有更多文章来涵盖整个主题。

战术 DDD 中最关键的部分之一是试图直接在代码中反映领域逻辑。

但是这依然是一些非特定的定义——并且现在并不需要。我也不想从描述什么是 *值对象，实体，集合* 开始。从实际例子开始会更好。

### Wild workouts
> **这不是另一篇带着随机代码片段的文章。**
>
> 这篇博客是一个更大的系列文章的一部分，而在这些文章中，我们会展示如何构建**长期上易于开发，维护，且能够愉快工作的 Go 应用程序。** 我们通过分享已被证明的技术来做到这一点，这些技术基于我们与团队所做的许多实验，以及基于[科学研究](https://threedots.tech/post/ddd-lite-in-go-introduction/?utm_source=about-wild-workouts#thats-great-but-do-you-have-any-evidence-if-that-is-working)。
>
> 你可以通过与我们一起构建[功能完整](https://threedots.tech/post/serverless-cloud-run-firebase-modern-go-application/?utm_source=about-wild-workouts#what-wild-workouts-can-do)的实例 Go web 应用——**Wild Workouts** 来学习这些模式。
>
> 我们用完全不同的方式做着同一件事—— **我们在最初的 Wild Workouts 实现中引入了一些微小的问题。** 我们是失去了理智了吗？不是的。😉 这些问题是很多 Go 项目的共性问题。长远看，这些小小的问题会变得棘手，进而让新功能的添加变得困难无比。
>
> 对高级或首席开发人员来说，最重要的技能一直就是就是，需要始终关注长期影响。
>
> 我们会重构“太现代的” Wild Workouts 来修复这些问题。通过这种方式，你会轻松理解我们分享的技术。
>
> 你了解那种读了一些技术文章之后，试着实现的时候却因为几个指南中略过的问题而被卡住的感觉吗？省略这些细节无疑会让文章更简短，并且提升浏览量，但这不是我们的目标。我们的目标是创建提供了足够的技术诀窍以应用所介绍技术的文章。如果你还没有读过[这个系列之前的文章](https://threedots.tech/tags/building-business-applications/)，我们强烈建议读一下。
>
> 我们相信在某些方面没有捷径可走。如果想用快速且高效的方式构建复杂的应用，你仅仅需要花些时间学习这些技术。如果问题很简单的话，就不会有这么让人头疼的遗留代码了。
>
> 这里是目前为止发布的 [8 篇文章的完整列表](https://threedots.tech/tags/building-business-applications/?utm_source=about-wild-workouts)。
>
> Wild Workouts 的**全部源码**可以从 [GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example?utm_source=about-wild-workouts) 上获得。

我还没有提到，我们特别为了这些文章，创建了一个叫做 Wild Workouts 的整个应用。有趣的是，我们在这个应用中引入了一些微妙的问题，以进行重构。如果 Wild Workouts 看起来像是你原来接触过的应用——最好多在我们这里驻足一会儿😉。

### 重构 `trainer` 服务
我们开始重构的第一个（微）服务是 `trainer`。我们现在先不动其他服务——以后会回过头处理他们。

这个服务的职责是维护教练的时间表，并保证我们每一个小时之内只安排一个培训。该服务统一维护有效时间（教练的时间表）的信息。

最初的实现不是最好的。即使没有大量的逻辑，代码的一些部分也开始变得混乱了。基于我的经验，我感觉这些代码随着时间的推移会变得更糟。 😉

```go
func (g GrpcServer) UpdateHour(ctx context.Context, req *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
   trainingTime, err := grpcTimestampToTime(req.Time)
   if err != nil {
      return nil, status.Error(codes.InvalidArgument, "unable to parse time")
   }

   date, err := g.db.DateModel(ctx, trainingTime)
   if err != nil {
      return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get data model: %s", err))
   }

   hour, found := date.FindHourInDate(trainingTime)
   if !found {
      return nil, status.Error(codes.NotFound, fmt.Sprintf("%s hour not found in schedule", trainingTime))
   }

   if req.HasTrainingScheduled && !hour.Available {
      return nil, status.Error(codes.FailedPrecondition, "hour is not available for training")
   }

   if req.Available && req.HasTrainingScheduled {
      return nil, status.Error(codes.FailedPrecondition, "cannot set hour as available when it have training scheduled")
   }
   if !req.Available && !req.HasTrainingScheduled {
      return nil, status.Error(codes.FailedPrecondition, "cannot set hour as unavailable when it have no training scheduled")
   }
   hour.Available = req.Available

   if hour.HasTrainingScheduled && hour.HasTrainingScheduled == req.HasTrainingScheduled {
      return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("hour HasTrainingScheduled is already %t", hour.HasTrainingScheduled))
   }

   hour.HasTrainingScheduled = req.HasTrainingScheduled
```

尽管这还不是最糟糕的代码，它也让我想起了我检查代码的 git 历史记录时所看到的。我可以想象到，一段时间后，经过几次新功能迭代后，这些代码的情况会变得更糟。

这些代码同样难以 mock 依赖，所以同样没有单元测试。

#### 第一条规则——直白地去反映你的业务逻辑
实现 domain 的时候，请不要总去想着类似拘泥数据结构这样的结构体，或者是带着一大堆 setter 和 getter 的“类 ORM” 实体。相反，应该将他们看作**带着行为的类型**。

当与业务相关的人聊天时，他们会说 *“我在 13：00 安排了训练”*，而不是 *“我将 13：00 的属性状态设置为了‘安排训练’”*。

他们也不会说： *“你无法将属性状态设置为‘安排训练’”*。而是：*“如果时间不合适的话，就无法安排训练”*。那么如何直接把这些反映在代码里面呢？

```go
func (h *Hour) ScheduleTraining() error {
   if !h.IsAvailable() {
      return ErrHourNotAvailable
   }

   h.availability = TrainingScheduled
   return nil
}
```

一个可以帮助我们更好实现代码的问题是：*“业务人员可以不需要技术术语的翻译就能够读懂我的代码吗？”*。你可以看下上面的片段，**即使是非技术人员，也能够明白什么时候你可以安排培训**。

这个方法的代价不高，并且有助于应对复杂情况，是规则更加易于理解。即使带来的变化不大，我们也摆脱了一大串 `if`，而这一大串 if 在未来会是代码变得更加复杂。

我们也同样能够容易地添加单元测试。这很好——我们不需要 mock 什么了。这些测试同样是有助于我们理解 `Hour` 行为的文档。

```go
func TestHour_ScheduleTraining(t *testing.T) {
   h, err := hour.NewAvailableHour(validTrainingHour())
   require.NoError(t, err)

   require.NoError(t, h.ScheduleTraining())

   assert.True(t, h.HasTrainingScheduled())
   assert.False(t, h.IsAvailable())
}

func TestHour_ScheduleTraining_with_not_available(t *testing.T) {
   h := newNotAvailableHour(t)
   assert.Equal(t, hour.ErrHourNotAvailable, h.ScheduleTraining())
}
```

现在，如果有人问“什么时候我可以安排训练”，你可以很快回答。在一个更大的系统中，这一类问题的答案相对不怎么明显——好几次我花费数小时去寻找一些对象被意外应用的地方。下一条规则会进一步帮助我们。

#### 第二条规则——在内存中始终保持一个有效的状态

> 我意识到我的代码会以我无法预料的方式使用，也会以并非其设计的方式被使用，并且这种错误使用比按照预期使用的时间更长。——[坚固宣言（The Rugged Manifesto）](https://ruggedsoftware.org/)

如果所有人将这段引言记住那这个世界会更好。我在这里也不是没有错。😉

据我观察，当你确定你使用的对象始终是有效的，那么这会避免许多的 `if` 语句以及避免许多的 bug。无法对当前的代码做任何愚蠢的事情会让你感到更加的自信。

我多次提到我害怕去做一些修改，因为我不清楚会带来什么副作用。**在没有把握是否正确使用代码的情况下，开发新功能要慢得多！**

我们的目标是仅在一个地方去做校验（良好的 DRY）并且确保没有人可以修改 `Hour` 的内在状态。该对象的唯一公共 API 应当是描述行为的方法。而不是愚蠢的 getter 和 setter ！我们还需要将类型分开包装，并且所有属性设置为私有。

```go
type Hour struct {
   hour time.Time

   availability Availability
}

// ...

func NewAvailableHour(hour time.Time) (*Hour, error) {
   if err := validateTime(hour); err != nil {
      return nil, err
   }

   return &Hour{
      hour:         hour,
      availability: Available,
   }, nil
}
```

我们也应当保证没有破坏我们类型内在的任何规则。

不好的例子：

```go
h := hour.NewAvailableHour("13:00")

if h.HasTrainingScheduled() {
    h.SetState(hour.Available)
} else {
    return errors.New("unable to cancel training")
}
```

好的例子：

```go
func (h *Hour) CancelTraining() error {
   if !h.HasTrainingScheduled() {
      return ErrNoTrainingScheduled
   }

   h.availability = Available
   return nil
}

// ...

h := hour.NewAvailableHour("13:00")
if err := h.CancelTraining(); err != nil {
    return err
}
```

#### 第三条规则——领域需要与数据库无关

这里有很多流派——其中一些会告诉你，领域受数据库客户端影响是可以的。以我们的经验来说，严格保证领域不受任何数据库影响会更好。

主要的原因是：

- 领域类型不受所使用的数据库方案影响——他们应当仅受业务规则的影响。
- 这样我们能够以更好的方式将数据保存在数据库中。
- 由于 Go 的设计以及缺少类似注解这样的“魔法”，ORM 或者任何数据库解决方案都会以更显著的方式产生影响。

> 领域优先方法
>
> 如果项目足够复杂，我们甚至可以花上2-4周时间在领域层上，仅适用纯内存的数据库实现。这样的话，我们可以更深层地探索想法，并且延后选择数据库层的决定。所有我们的实现仅仅基于单元测试。
>
> 我们尝试了几次这个方法，效果都还不错。在这里使用一些时间框也是不错的主意，这些不会消耗很长时间。
>
> 请记住，这一方法需要与业务人员有着良好的关系以及充分的信任！**如果与业务人员的关系远远说不上良好，策略 DDD 模式会改善这一情况。去过也做过！**

为了不使这个文章太长，这里仅介绍下 Repository 接口，并且假定可以正常工作。😉 在接下来的文章中我会更深入地涵盖这个主题。

```go
type Repository interface {
   GetOrCreateHour(ctx context.Context, time time.Time) (*Hour, error)
   UpdateHour(
      ctx context.Context,
      hourTime time.Time,
      updateFn func(h *Hour) (*Hour, error),
   ) error
}
```

> 你也许会问 为什么 `UpdateHour` 会有 `updateFn func(h *Hour) (*Hour, error)` —— 我们会用它以一种巧妙的方式处理事务。更多信息请看关于 repositories 的文章！😉

**使用领域对象**

我对我们的 gRPC 端点进行了小小的重构，提供更“行为导向”而不是 [CRUD](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) 导向的 API。它更好地反映了领域的新特征。以我的经验来说，维护多个，小的方法，相比维护一个，“全能”的，可以让我们更新所有东西的方法要容易的多。

```diff
--- a/api/protobuf/trainer.proto
+++ b/api/protobuf/trainer.proto
@@ -6,7 +6,9 @@ import "google/protobuf/timestamp.proto";

 service TrainerService {
   rpc IsHourAvailable(IsHourAvailableRequest) returns (IsHourAvailableResponse) {}
-  rpc UpdateHour(UpdateHourRequest) returns (EmptyResponse) {}
+  rpc ScheduleTraining(UpdateHourRequest) returns (EmptyResponse) {}
+  rpc CancelTraining(UpdateHourRequest) returns (EmptyResponse) {}
+  rpc MakeHourAvailable(UpdateHourRequest) returns (EmptyResponse) {}
 }

 message IsHourAvailableRequest {
@@ -19,9 +21,6 @@ message IsHourAvailableResponse {

 message UpdateHourRequest {
   google.protobuf.Timestamp time = 1;
-
-  bool has_training_scheduled = 2;
-  bool available = 3;
 }

 message EmptyResponse {}
```

现在的实现比原来简单多了，并且容易理解。我们这里也没有逻辑——只是一些编排。我们的 gRPC hander 现在有18行，并且没有领域逻辑！

```go
func (g GrpcServer) MakeHourAvailable(ctx context.Context, request *trainer.UpdateHourRequest) (*trainer.EmptyResponse, error) {
   trainingTime, err := protoTimestampToTime(request.Time)
   if err != nil {
      return nil, status.Error(codes.InvalidArgument, "unable to parse time")
   }

   if err := g.hourRepository.UpdateHour(ctx, trainingTime, func(h *hour.Hour) (*hour.Hour, error) {
      if err := h.MakeAvailable(); err != nil {
         return nil, err
      }

      return h, nil
   }); err != nil {
      return nil, status.Error(codes.Internal, err.Error())
   }

   return &trainer.EmptyResponse{}, nil
}
```

> **不要再有八千（Eight-thousanders）**
> 依据我过去的记忆，许多的所谓八千行代码实际上是在 HTTP controller 中有着大量领域逻辑的 controller。
>
> 在我们的领域类型内部隐蔽复杂性，并且坚持我提到的那些规则，我们可以阻止在这个地方代码不可阻止的增长。

### 今天就到这里

我不希望让这片文章太冗长——咱们一步一步来！

如果你等不及，重构的整个 diff 可以在 [GitHub](https://github.com/ThreeDotsLabs/wild-workouts-go-ddd-example/commit/0249977c58a310d343ca2237c201b9ba016b148e) 上面找到。在下一篇文章中，我会覆盖这些 diff 中没有讲到的一部分：repositories。

即使还在刚开始的阶段，在我们代码中一些简化还是明显的。

这个模型目前的实现还是不完美的——这很好！在一开始你不会实现完美的模型。**最好是去准备好轻松地修改这个模型，而不是浪费时间取让它变得完美。** 给这个模型添加了测试代码，以及将他与应用中的其他部分分离开来后，我不再惧怕对它进行修改。

**我可以在我的简历上说我了解 DDD 了吗？**

还不行。

从听说 DDD 之后到融会贯通我花费了3年时间（比我听说 Go 语言还要早 😉 ）。此外，我知道为什么我们下一篇文章会讲的所有技术十分重要。在将这些技术融会贯通之前，需要一些耐心而且要相信这些技术会起作用。这是值得的！你不需要想我一样需要3年时间，但是我们现在计划了大约10篇关于策略和战术模式的文章。😉 在 Wild Workouts 项目中还有很多新的特性以及要重构的部分！

我知道，现如今有许多人保证经过10分钟的文章或视频之后，你会成为某个领域的专家。如果这可能的话世界会是美好的，然而，在现实生活中并不会如此简单。

幸运的是，我们所分析的大部分知识是通用且可以应用在多个技术上面，而不仅仅是 Go。长远上，你可以将这些学校视为对你职业生涯以及心理健康的投资 😉。没有比解决正确的问题，同时没有不可维护的代码更好的事情了。

**你在 Go 语言中应用 DDD 有什么经验吗？是好是坏？与我们的做法有什么不同吗？你觉得 DDD 在你的项目中是否有用？请在评论中告诉我们！**

**有没有身边的同事你觉得可能会对这个主题感兴趣？请把这篇文章分享给他们！即使他们不使用 Go。** 😉

---
via: https://threedots.tech/post/ddd-lite-in-go-introduction/

作者：[Robert Laszczak](https://twitter.com/roblaszczak)
译者：[dust347](https://github.com/dust347)
校对：[unknwon](https://github.com/unknwon)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

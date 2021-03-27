首发于：https://studygolang.com/articles/33990

# Go 官方 2020 年开发者调查报告

2021 年 3 月 9 日，在 Go 官方博客发布了 Go 开发者 2020 年调查报告。一起来看看该报告的内容吧。

> 2020 年，一共有 9648 人参与投票，大约相当于 2019 年的投票人数。

说明：你可能会注意到有些问题的样本量比其他问题小 (“n =”)。这是因为有些问题是向所有人展示的，而另一些只是向随机的一部分受访者展示。

## 01 报告摘要

- Go 的使用场景和企业都在扩大，76% 的受访者工作中使用 Go；66% 的人说 Go 对他们公司的成功至关重要；
- 92% 的受访者对 Go 感到满意；
- 在使用不到 3 个月的时间里，大多数受访者感觉使用 Go 非常高效（生产力很高），占比达 81%；
- 大家倾向升级到 Go 最新版本，在前 5 个月中达到了 76％；
- 使用 pkg.go.dev 用户更容易找到想要的包（91% vs 82%）；
- Go 模块的采用率几乎达到了普遍水平，满意度为 77％，但受访者还强调需要改进文档；
- Go 继续大量用于 API，CLI，Web，DevOps 和数据处理；
- 代表性不足的群体在社区中往往会受到较少的欢迎；

## 02 受访者群体

人口统计学问题有助于我们区分哪些年度差异可能源于调查对象的变化，哪些是情绪或行为的变化。因为我们的人口统计数据与去年相似，我们有理由相信，其他的年度变化主要不是由于人口统计学的变化。

例如，从 2019 年到 2020 年，组织规模、开发人员经验和行业的分布基本保持不变。

![Bar chart of organization size for 2019 to 2020 where the majority have fewer than 1000 employees](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/orgsize.svg)

![Bar chart of years of professional experience for 2019 to 2020 with the majority having 3 to 10 years of experience](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/devex_yoy.svg)

![Bar chart of organization industries for 2019 to 2020 with the majority in Technology](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/industry_yoy.svg)

近一半（48%）的受访者使用 Go 不到两年。在 2020 年，我们收到的回应是少于一年。（GCTT 注：可见 Go 还是很年轻，这两年增长也迅速，很多新人进入了）

![Bar chart of years of experience using Go](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/goex_yoy.svg)

大多数人说他们在工作中(76%)和工作之外(62%)使用 Go。在工作中使用 Go 的受访者的百分比逐年呈上升趋势。

![Bar chart where Go is being used at work or outside of work](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/where_yoy.svg)

今年我们提出了一个关于主要工作职责的新问题。我们发现，70% 的受访者的主要职责是开发软件和应用程序，但有相当一部分(10%)是设计 IT 系统和架构。

![Primary job responsibilities](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/job_responsibility.svg)

与往年一样，我们发现大多数受访者并不是 Go 开源项目的经常贡献者，75% 的受访者表示他们“不经常”或“从来没有”贡献过。（GCTT 注：看来开源不是那么容易的，还是使用者居多）

![How often respondents contribute to open source projects written in Go from 2017 to 2020 where results remain about the same each year and only 7% contribute daily](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/foss_yoy.svg)

## 03 开发人员工具和实践

与前几年一样，绝大多数的调查对象报告说他们在使用 Linux (63%)和 macOS (55%)系统。随着时间的推移，主要在 Linux 上开发的受访者的比例似乎略有下降。（GCTT 注：Windows 对开发还是不够友好，不过近一年 Windows 做了很多改变，拭目以待！）

![Primary operating system from 2017 to 2020](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/os_yoy.svg)

第一次，首选的编辑器似乎已经稳定: VS Code 仍然是最受欢迎的编辑器(41%) ，GoLand 是强有力的次优(35%)。它们俩加起来占比超过 76% ，其他编辑器并没有像前几年那样继续减少。（GCTT 注：VS Code 确实很棒，插件也已经归属 Go 官方，而 GoLand 同样很棒，但毕竟收费的。其他的没减少，大概率是某个编辑器的忠实粉丝吧！）

![Editor preferences from 2017 to 2020](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/editor_pref_yoy.svg)

今年我们询问受访者，假设他们有 100 个 “gophercoin”（一个虚构的货币） ，他们会花多少钱来优先改进他们编辑器的什么功能。代码完成（Code completion）得到的 “gophercoin” 最多。一半的受访者给出了前四项特性（代码完成、导航代码、编辑器性能和重构），它们 10 个或更多的 gophercoin。

![Bar char of average number of GopherCoins spent per respondent](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/editor_improvements_means.svg)

大多数受访者(63%)花费 10-30% 的时间进行重构，这表明这是一项常见的任务，我们希望研究改进它的方法。这也解释了为什么重构支持是最受资助的编辑器改进之一。

![Bar chart of time spent refactoring](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/refactor_time.svg)

去年我们询问了特定的开发人员，发现几乎 90% 的受访者使用文本日志进行调试，因此今年我们增加了一个后续问题来找出原因。结果显示，43% 的人使用它是因为它允许他们在不同的语言中使用相同的调试策略，42% 的人更喜欢使用文本日志而不是其他调试技术。然而，27% 的人不知道如何开始使用 Go 的调试工具，24% 的人从来没有尝试过使用 Go 的调试工具，因此有机会改进调试工具的可发现性、可用性和文档性。此外，由于四分之一的受访者从未尝试过使用调试工具，痛处可能被低估了。（GCTT 注：PHPer 喜欢文本日志调试，哈哈哈哈，你懂的）

![img](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/why_printf.svg)

## 04 对 Go 的看法

今年，我们第一次询问了总体满意度。92% 的受访者表示，在过去一年中，他们对 Go 的使用非常满意或略感满意。

![Bar chart of overall satisfaction on a 5 points scale from very dissatisfied to very satisfied](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/csat.svg)

这是我们第三年提出“你推荐... ...”[网络推广分数](https://en.wikipedia.org/wiki/Net_Promoter)(NPS)的问题。今年我们的 NPS 结果是 61(68% 的“推动者”减去 6% 的“诋毁者”) ，与 2019 年和 2018 年统计数据相同。(GCTT 注：也就是说 68% 的人会推荐 Go 语言，而 6% 的人说 Go 不好之类的)

![Stacked bar chart of promoters, passives, and detractors](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/nps.svg)

与前几年一样，91% 的受访者表示他们更愿意在下一个新项目中使用 Go。89% 的人认为 Go 在他们的团队中表现良好。今年，我们看到越来越多的受访者认为 Go 对他们公司的成功至关重要，从 2019 年的 59% 上升到 2020 年的 66% 。在 5000 人以上的组织工作的受访者不太可能同意(63%) ，而在较小的组织工作的受访者更可能同意(73%)。（GCTT 注：看来小公司更认为 Go 对他们的成功很关键，比如国内的七牛？！）

![Bar chart of agreement with statements I would prefer to use Go for my next project, Go is working well for me team, 89%, and Go is critical to my company's success](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/attitudes_yoy.svg)

与去年一样，我们要求受访者根据满意度和重要性对 Go 开发的具体方向进行评级。使用云服务、调试和使用模块(去年被强调为改进的领域)的满意度有所提高，而大多数重要性得分保持不变。我们还介绍了一些新的主题: API 和 Web 框架。我们发现 Web 框架的满意度低于其他领域(64%)。对于大多数当前用户来说，它并不是那么重要(只有 28% 的受访者认为它非常重要) ，但是对于潜在的 Go 开发者来说，它可能是一个缺失的关键特性。

![Bar chart of satisfaction with aspects of Go from 2019 to 2020, showing highest satisfaction with build speed, reliability and using concurrency and lowest with Web frameworks](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/feature_sat_yoy.svg)

81% 的受访者说他们觉得使用 Go 非常有效率。大型组织的受访者比小型组织的受访者更有可能感到极其富有成效。（GCTT 注：看来确实是一门面向大型工程的语言）

![Stacked bar chart of perceived productivity on 5 point scale from not all to extremely productive ](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/prod.svg)

我们听说 Go 很容易变得高效。我们询问了那些觉得自己至少有一点生产力的受访者，他们花了多长时间才变得有生产力。93% 的人说他们花了不到一年的时间，大多数人在 3 个月内就感觉到有效率。（GCTT 注：这一定程度上还是说明 Go 简单，容易快速进行开发，提高效率）

![Bar chart of length of time before feeling productive](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/prod_time.svg)

虽然与去年大致相同，但随着时间的推移，同意”我感到在 Go 社区受欢迎”这一说法的受访者百分比似乎有所下降，或者至少不像其他领域那样保持同样的上升趋势。

我们还发现，认为 Go Team 理解自己需求的受访者比例(63%)逐年显著上升。（GCTT 注：这是说 Go Team 开发的新特性，大部分都是社区需要的）

![Bar chart showing agreement with statements I feel welcome in the Go community, I am confident in the Go leadership, I feel welcome to contribute, The Go project leadership understands my needs, and The process of contributing to the Go project is clear to me](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/attitudes_community_yoy.svg)

我们就如何使 Go 社区更受欢迎提出了一个公开问题，最常见的建议(21%)涉及学习资源和文档的不同形式或改进/增加。

![Bar chart of recommendations for improving the welcomeness of the Go community](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/more_welcoming.svg)

## 05 用 Go 干什么

构建 API/RPC 服务(74%)和 cli (65%)仍然是 Go 最常见的用途。与去年相比，我们没有看到任何重大变化，当时我们在选项排序中引入了随机化。(在 2019 年之前，名单开头的选项被不成比例地选中。)我们还根据组织规模对这一问题进行了分析，发现受访者在大型企业或者小型组织中使用 Go 的情况类似，尽管大型组织使用返回 HTML 的 Go for Web 服务的可能性有所降低。（GCTT 注：竟然有 8% 的人用 Go 写桌面 GUI 应用？厉害了）

![Bar chart of Go use cases from 2019 to 2020 including API or RPC services, CLIs, frameworks, Web services, automation, agents and daemons, data processing, GUIs, games and mobile apps](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/app_yoy.svg)

今年，我们更好地区分了调查者在业余时间使用 Go 和在工作中使用 Go 开发的不同软件。虽然返回 HTML 的 Web 服务是第四个最常见的用例，但这是由于与工作无关的使用。与返回 HTML 的 Web 服务相比，更多的受访者使用 Go 进行自动化/脚本、代理和守护进程以及工作数据处理。很大一部分最不常用的应用(桌面/GUI 应用、游戏和移动应用)是在工作之外编写的。（GCTT 注：看来 GUI 之类的，还是个人爱好的尝试）

![Stacked bar charts of proportion of use case is at work, outside of work, or both ](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/app_context.svg)

另一个新问题是询问受访者对每个用例的满意程度。CLI 满意度最高，85% 的受访者说他们非常、中等或稍微满意使用 Go for cli。一般对 Go 的使用往往有较高的满意度分数，但满意度和受欢迎程度并不完全一致。例如，代理和守护进程的满意度排名第二，但在使用率上排名第六。

![Bar chart of satisfaction with each use case](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/app_sat_bin.svg)

其他的后续问题探讨了不同的用例，例如，开发 CLI 的用户一般使用哪个平台，Linux (93%)和 macOS (59%) 具有很高的代表性并不奇怪，因为 Linux 和 macOS 的开发人员使用频率很高，而且 Linux 云的使用频率也很高) ，但是即使是 Windows 也有近三分之一 CLI 开发人员使用。

![Bar chart of platforms being targeted for CLIs](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/cli_platforms.svg)

对 Go 在数据处理的深入研究表明， Kafka 是唯一被广泛采用的引擎，但大多数受访者表示他们使用的是一个定制的数据处理引擎。

![Bar chart of data processing engines used by those who use Go for data processing](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/dpe.svg)

我们还询问了受访者使用 Go 的其他更大领域。最常见的领域是网络开发(68%) ，其他常见领域包括数据库(46%) ，DevOps (42%)、网络编程(41%)和系统编程(40%)。

![Bar chart of the kind of work where Go is being used](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/domain_yoy.svg)

与去年类似，我们发现 76% 的受访者表示将当前的 Go 版本用于生产用途，但今年我们改进了我们的时间表，发现 60%  的人在新版本发布前或两个月内开始试用新版本。这突出了平台即服务提供商（PaaS）快速支持新的稳定版 Go 的重要性。

![Bar chart of how soon respondents begin evaluating a new Go release](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/update_time.svg)

## 06 Module（模块）

今年我们发现几乎所有人都采用了 Go 模块，并且只使用模块进行包管理的受访者的比例显著增加。96% 的受访者表示他们正在使用模块管理包，高于去年的 89% 。87% 的受访者表示，他们只使用模块管理包，而去年这一比例为 71% 。同时，其他软件包管理工具的使用也在减少。（GCTT 注：这调查感觉意义不大，这是必然的，官方大力推广，可不用嘛）

![Bar chart of methods used for Go package management](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/modules_adoption_yoy.svg)

与去年相比，用户对模块的满意度也有所提高。77% 的受访者表示，他们对模块非常、中等或稍微满意，而 2019 年这一比例为 68% 。（GCTT 注：看来不满意的人也不少）

![Stacked bar chart of satisfaction with using modules on a 7 point scale from very dissatisfied to very satisfied](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/modules_sat_yoy.svg)

## 07 官方文档

大多数受访者表示，他们对官方文档感到头疼。62% 的受访者难以找到足够的信息来实现他们应用程序的一个特性，超过三分之一的人难以开始做他们以前从未做过的事情。（GCTT 注：看来问题不小）

![Bar chart of struggles using official Go documentation](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/doc_struggles.svg)

官方文档中问题最多的领域是使用模块和 CLI 开发，20% 的受访者认为模块文档稍微有点帮助或者根本没有帮助，16% 的受访者认为有关 CLI 开发的文档有帮助。（GCTT 注：所以现在官网上增加了一个模块相关的教程）

![Stacked bar charts on helpfulness of specific areas of documentation including using modules, CLI tool development, error handling, Web service development, data access, concurrency and file input/output, rated on a 5 point scale from not at all to very helpful](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/doc_helpfulness.svg)

## 08 云上 Go

在设计时 Go 就考虑到了现代的分布式计算服务，我们希望继续提高开发者使用 Go 构建云服务的体验。

- 全球最大的三家云服务提供商(亚马逊网络服务、谷歌云平台和微软 Azure)的受访者使用率持续上升，而大多数其他云服务提供商的受访者比例每年都在下降。特别是 Azure，从 7% 上升到了 12%。（GCTT 注：阿里云等国内云用的少，多半是国人参与这个调查的不多吧）
- 作为最常见的部署目标，对自有或公司拥有的服务器的 On-prem 部署继续减少；

![Bar chart of cloud providers used to deploy Go programs where AWS is the most common at 44%](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/cloud_yoy.svg)

部署到 AWS 和 Azure 的受访者发现，部署到管理的 Kubernetes 平台的受访者增加了，目前分别为 40% 和 54% 。发现将 Go 程序部署到 VMs 的用户比例显著下降，容器使用率从 18% 增长到 25% 。与此同时，GCP (已经有很高比例的受访者报告使用管理的 Kubernetes)部署到 serverless 云的比例从 10% 增长到 17% 。

![Bar charts of proportion of services being used with each provider](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/cloud_services_yoy.svg)

总体而言，大多数受访者对三大主要云供应商都使用 Go 感到满意，而且这些数据与去年相比没有统计上的变化。受访者对 AWS 和 GCP 的 Go 开发的满意程度相当(82%)。Azure 的满意度得分较低（58% 的满意度） ，大家在备注中表示需要对 Azure 的 Go SDK 和 Go 支持 Azure 功能进行改进。

![Stacked bar chart of satisfaction with using Go with AWS, GCP and Azure](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/cloud_csat.svg)

## 09 痛苦的地方

受访者表示无法使用 Go 的主要原因是，他们仍然在用另一种语言进行项目(54%) ，在一个更喜欢使用另一种语言的团队工作(34%) ，Go 本身缺乏某些关键特性(26%)。（GCTT 注：看来项目领导很重要）

今年，我们引入了一个新的选项：“I already use Go everywhere I would like to”，这样受访者就可以不受限制，只要我喜欢 Go，总有可以让我使用的场景。这显著降低了其他选项的选择率，但没有改变它们的相对顺序。我们还引入了“ Go 缺少关键框架”的选项。

如果我们只看那些选择不使用 Go 的原因的受访者，我们可以更好地了解每年的趋势。随着时间的推移，「用另一种语言从事现有项目」、「项目/团队/领导对另一种语言的偏好」正在减少。

![Bar charts of reasons preventing respondents from using Go more](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/goblockers_yoy_sans_na.svg)

26% 的受访者认为 Go 缺乏他们需要的语言特性，其中 88% 的人选择了泛型作为一个关键的缺失功能。其他关键的缺失特性是更好的错误处理(58%)、nil 安全(44%)、函数式编程特性(42%)和更强/扩展类型系统(41%)。（GCTT 注：可见泛型呼声是最高的）

需要明确的是，这些数字来自于那些表示如果不缺少一个或多个关键特性他们将能够更多使用 Go 的受访者，而不是整个调查受访者群体。换个角度来看，18% 的受访者因为缺乏泛型而不能使用 Go。

![Bar chart of missing critical features](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/missing_features.svg)

受访者在使用 Go 时报告的最大挑战仍然是 Go 缺乏泛型(18%) ，而「模块/包管理」和「学习曲线/最佳实践/文档」方面的问题各占 13% 。

![Bar chart of biggest challenges respondents face when using Go](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/biggest_challenge.svg)

## 10 Go 社区

今年我们询问了受访者查询 Go 相关问题的 5 大资源。去年我们只要求排名前三，所以结果不能直接比较，然而，StackOverflow 仍然是最受欢迎的资源，占 65% 。阅读源代码(57%)仍然是另一个受欢迎的资源，而对 godoc. org (39%)的依赖已经显著减少。包发现网站 pkg.go.dev 是今年榜单中的新成员，是 32% 的受访者的首选资源。使用 pkg.go.dev 的受访者更有可能同意他们能够快速找到他们需要的 Go 软件包/库：pkg.go.dev 用户占 91% ，而其他用户占 82% 。（GCTT 注：开篇已经总结，也就是说，通过 pkg.go.dev 更容易找到想要的包）

![Bar chart of top 5 resources respondents use to answer Go-related questions](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/resources.svg)

多年来，不参加 Go 相关活动的受访者比例呈上升趋势。由于 2019 冠状病毒疾病的缘故，今年我们修改了围绕 Go 活动的问题，发现超过四分之一的受访者比往年花更多的时间在 Go 频道上，14% 的人参加了虚拟 Go 会议，是去年的两倍。64% 参加虚拟活动的人说这是他们第一次参加虚拟活动。（GCTT 注：这里说的 virtual Go Meetup 应该指线上吧）

![Bar chart of respondents participation in online channels and events](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/events.svg)

我们发现 12% 的受访者认同传统上代表性不足的群体(例如，种族，性别认同，等等) ，与 2019 年相同，2% 的受访者认同女性，少于 2019 年(3%)。认同代表性不足群体的受访者比不认同代表性不足群体的受访者对“我在 Go 社区感到受欢迎”这句话的不同意率更高(10% 对 4%)。这些问题使我们能够衡量社区的多样性，并突出外联和增长的机会。

![Bar chart of underrepresented groups](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/underrep.svg)

![Bar chart of those who identify as women](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/underrep_groups_women.svg)

![Bar chart of welcomeness of underrepresented groups](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/welcome_underrep.svg)

今年，我们增加了一个关于辅助技术使用（assistive technology usage）的问题，发现 8% 的受访者正在使用某种形式的辅助技术。最常用的辅助技术是对比度或颜色设置(2%)。这是一个很好的提醒，我们有需要辅助功能的用户，帮助我们在由 Go 团队管理的网站上做出一些设计决策。

![Bar chart of assistive technology usage](https://raw.githubusercontent.com/studygolang/gctt-images/master/2020-go-survey/at.svg)

团队重视多样性和包容性，不仅仅是因为这是正确的事情，而是因为不同的声音可以照亮我们的盲点，最终使所有用户受益。我们询问敏感信息的方式，包括性别和传统上未被充分代表的群体，已经根据数据隐私条例发生了变化，我们希望这些问题，特别是围绕性别多样性，在未来更具包容性。

## 11 总结

感谢您 Review 我们的调查结果：我们 2020 开发者调查报告！理解开发人员的经验和挑战有助于我们衡量我们的进展和指导 Go 的未来。再次感谢所有参与这项调查的人，没有你们，我们不可能完成这项工作。我们希望明年见到你！

> 原文链接：<https://blog.golang.org/survey2020-results>
>
> 国内访问链接：<https://docs.studygolang.com/blog/survey2020-results>
>
> 编译：GCTT（polarisxu），并非完全直译

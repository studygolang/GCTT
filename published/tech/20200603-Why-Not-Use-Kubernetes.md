首发于：https://studygolang.com/articles/30256

# 为什么不使用 Kubernetes

![When to choose Kubernetes?](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200603-Why-Not-Use-Kubernetes/00.png)

很多团队都很兴奋地开始使用 Kubernetes。其中一些团队希望能充分利用它的弹性、灵活性、可移植性、可靠性以及其他的一些 Kubernetes 能原生地提供的优势。也有些团队只是热衷于技术，仅仅想使用下这个平台，来更好地了解它。还有一些开发者想获得一些使用它的经验，这样他们的简历上就可以添加一项很多公司急需的技能。总之，现在大部分开发者出于不同的目的都想要使用 Kubernetes。

使用 Kubernetes 有好处也有坏处。

## 设计 Kubernetes 的初衷是用来解决分布式架构的问题的

来看[官网文档网站](https://kubernetes.io/docs/concepts/overview/what-is-kubernetes/)的定义：

> “Kubernetes 为你提供了能灵活运行分布式系统的框架。它的能力体现在对你的应用进行扩缩容和故障转移，提供部署模式，等等”。

它不是仅用于分布式系统的，而是用于容器化应用。即便如此，它提供了很多可以让管理分布式系统和扩缩容变得更容易地资源，就像微服务解决方案一样。它也被认为是一个编排系统。

> [自动化](https://www.redhat.com/en/topics/automation)和编排不一样，但是也有关联。自动化通过减少和替换掉人与 IT 系统的交互使用软件来执行任务，这样能减少资源消耗，降低复杂度和减少错误，进而使得系统更加高效。
>
> 总之，自动化表示使某单一的任务自动执行。它与编排不同，编排是指在多个不同的系统间的多个步骤中如何自动化执行某一个处理或工作流。当你开始把自动化编进你的处理流程中时，你可以编排它们，让他们自动执行。
>
> — [编排是什么？RedHat 官方网站](https://www.redhat.com/en/topics/automation/what-is-orchestration)

换句话说，Kubernetes 使管理复杂的解决方案变得更加容易，而如果没有适当的编排系统，这些解决方案将很难维护。虽然您可以自己实施 DevOps 工程实践，但如果要从数十种服务扩展到数百种服务，则无法扩展。

## Kubernetes 很复杂

为了充分利用它的各个功能，开发者和 IT 操作者们必须掌握容器、网络、安全、移植性、弹性和 Kubernetes 本身相关的知识。为了合理地使用它的负载，你应该先了解每个组件是如何工作的。为了管理一个集群，你应该了解它的架构、存储、API 和后台管理系统，而这可能与传统的虚拟化环境不一样。为了实施某个方案，你应该先了解如何集成工具来部署、监控以及追踪服务，诸如 [Helm](https://helm.sh/) 和 [Istio](https://istio.io/)。这里涉及大量的新概念，因此你的团队要做好充足的准备来迎接挑战。

## Kubernetes 对于小的解决方案花费很大

为了理解原因，我们先来加深下对 Kubernetes 的一个很重要的概念的认识 — 弹性。为了体现弹性，你需要更多的节点 — 比运行应用程序所需的最少节点数要多一点。当某个节点挂掉时，请求的 pod 会迁移到可用的节点。在生产中的工作负载，为了集群有弹性，推荐至少部署三个节点。

如果你只需要维护一个单一的应用，那么显然不需要像上面那样做。但是即使你有十几个应用，你仍然要考虑维护集群的收益和集群的开销是否平衡。

**维护环境的开销还包括运维支持。**平台越复杂，对运维人员的专业性要求就越高。你可能需要雇佣第三方的专业团队来提供支持，或者需要一个诸如 Openshift 的包含支持服务的解决方案。

## 什么时候该选择 Kubernetes

基于你使用的架构，应用的数量和各应用间的依赖程度，你的团队能提供的运维能力，你可以在众多可用的技术中判断 Kubernetes 是否是最佳选择。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200603-Why-Not-Use-Kubernetes/01.jpeg)

按照[基于容器的 Web 应用](https://azure.microsoft.com/en-us/services/app-service/containers/)部署完后，你有了一套可以用于生产的环境。在遵照流程做了完整的计划，有 SSL 特性，并安装了 [Application Insights](https://docs.microsoft.com/en-us/azure/azure-monitor/app/cloudservices) 后，你的环境会变得安全、可伸缩，几乎不需要运维工作。

如果你的应用都是独立的，或者只连接了少量的应用，也许在同一个虚拟网络中组合使用 Azure Web Apps 和 [容器实例](https://azure.microsoft.com/en-us/services/container-instances/)就足够了。

然而，如果你的容器化的应用数量会增长，那么使用 Kubernetes 管理他们会很有趣。你会在单一的、集中式的环境中管理像 Web 应用，API 和 循环的任务等不同类型的应用。你的团队也能把精力放在 Kubernetes 而不是浪费在选择不同的云原生解决方案上。

如果你要处理分布式的场景，像微服务，那么请选择 Kubernetes。分布式架构很复杂，而 Kubernetes 就是为它设计的。能完全契合分布式的应用，以及可以根据应用的需求进行扩展，除了 Kubernetes，我想不到更好的任何其他平台。

## 总结

当你只需要处理少量的容器化的应用、各应用相互独立的或者应用间几乎不相互依赖时，选择其他的管理方式如[基于容器的 Web 应用](https://azure.microsoft.com/en-us/services/app-service/containers/)或[Azure 容器实例](https://azure.microsoft.com/en-us/services/container-instances/) — 或者两者结合 — 可能更简单，花费更少。

如果你的团队对 Kubernetes 的能力很满意，并且你的容器化的应用数量会越来越多，那么可能值得你在一个 Kubernetes 平台（如 [Azure Kubernetes 服务](https://azure.microsoft.com/en-us/services/kubernetes-service/)）中进行集中式管理。

*Kubernetes 是一个用来提高性能和减少分布式系统中运维工作的平台。*它主要用来降低复杂场景（如 微服务）中运维的复杂度。

如果你不需要处理大量的应用，没有使用分布式架构，或者团队中没有技术专家，那么你就不能享受到 Kubernetes 带来的便利 — 因为它不是为你设计的。使用 Kubernetes 只会给你的解决方案增加意外和不符合预期的复杂度。

如果你想更好地了解容器化应用的管理方式和如何选择最适合的方式，那么请查阅下面的文章：

[为你的应用选择 Azure 计算服务](https://docs.microsoft.com/en-us/azure/architecture/guide/technology-choices/compute-decision-tree)

---
via: https://medium.com/better-programming/why-not-use-kubernetes-52a89ada5e22

作者：[Grazi Bonizi](https://medium.com/@grazibonizi)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

已发布：https://studygolang.com/articles/12799

# Microservices in Golang - Part 7 - Terraform a Cloud
 
在之前的文章中，我们简要介绍了用户界面和Web客户端以及如何使用微工具包rpc代理与我们新创建的rpc服务进行交互。

本文将讨论如何创建云环境来托管我们的服务。 我们将使用 Terraform 在 Google Cloud 平台上构建我们的云群集。这应该是一篇相当短的文章，但它也很重要。

## 为什么选择 Terraform？

我已经使用了几种不同的云供应解决方案，但对我而言，Hashicorps Terraform 感觉最容易使用并且得到最好的支持。近年来出现了一个术语：'基础设施作为代码'。为什么你想要你的基础设施作为代码？那么，基础设施很复杂，它描述了很多移动部件。跟踪基础架构的变更和版本控制变更也很重要。

Terraform 完美地做到了这一点。他们实际上已经创建了自己的DSL（域特定语言），它允许您描述您的基础设施应该是什么样子。

Terraform 还允许您进行原子更改。所以在出现失败的情况下，它会将所有东西都退回来，并将其恢复到原来的状态。 Terraform 甚至允许您通过执行转出计划来预览更改。这将准确描述您的更改将对您的基础架构做什么。这给了你很多的信心，曾经是一个可怕的前景。

所以让我们开始吧！

## 创建您的云计划

转到 Google Cloud 并创建一个新项目。 如果您之前没有使用过它，您可能会发现您有300英镑的免费试用版。太好了！ 无论如何，你应该看到一个空白的新项目。 现在在你的左边，你应该看到一个IAM＆Admin tabb，然后在那里创建一个新的服务密钥。 确保选择“提供新密钥”，然后确保您选择了json类型。 安全保管，我们稍后需要。 这允许程序代表您执行对 Google Cloud API 的更改。 现在我们应该掌握一切我们需要开始的事情。
![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-micro/Screen-Shot-2018-02-10-at-10.58.07.png)
所以创建一个新的回购。 我称之为我的 shippy-infrastructure。

## 描述我们的基础设施

创建一个名为 variables.tf 的新文件。 这将包含关于我们项目的基本信息。 在这里，我们有我们的项目 ID，我们的地区，我们的项目名称和我们的平台名称。 该地区是我们希望集群运行的数据中心。 该区域是该区域内的可用区域。 项目名称是我们Google项目的项目ID，最后，平台名称是我们的群集的名称。

```c
variable "gcloud-region"    { default = "europe-west2" }
variable "gcloud-zone"      { default = "europe-west2-a" }
variable "gcloud-project"   { default = "shippy-freight" }
variable "platform-name"    { default = "shippy-platform" }
```

创建一个名为 providers.tf 的新文件，这是 Google 特定的部分：

```c
provider "google" {
	credentials = "${file("google-cred.json")}"
	project     = "${var.gcloud-project}"
	region      = "${var.gcloud-region}"
}
```

现在让我们创建一个名为 global.tf 的文件。 这里是我们设置的一部分：

```c
# Creates a network layer
resource "google_compute_network" "shippy-network" {
	name = "${var.platform-name}"
}

# Creates a firewall with some sane defaults, allowing ports 22, 80 and 443 to be open
# This is ssh, http and https.
resource "google_compute_firewall" "ssh" {
	name    = "${var.platform-name}-ssh"
	network = "${google_compute_network.shippy-network.name}"

	allow {
		protocol = "icmp"
	}

	allow {
		protocol = "tcp"
		ports    = ["22", "80", "443"]
	}

	 source_ranges = ["0.0.0.0/0"]
}

# Creates a new DNS zone
resource "google_dns_managed_zone" "shippy-freight" {
	name        = "shippyfreight-com"
	dns_name    = "shippyfreight.com."
	description = "shippyfreight.com DNS zone"
}

# Creates a new subnet for our platform within our selected region
resource "google_compute_subnetwork" "shippy-freight" {
	name          = "dev-${var.platform-name}-${var.gcloud-region}"
	ip_cidr_range = "10.1.2.0/24"
	network       = "${google_compute_network.shippy-network.self_link}"
	region        = "${var.gcloud-region}"
}

# Creates a container cluster called 'shippy-freight-cluster'
# Attaches new cluster to our network and our subnet,
# Ensures at least one instance is running
resource "google_container_cluster" "shippy-freight-cluster" {
	name = "shippy-freight-cluster"
	network = "${google_compute_network.shippy-network.name}"
	subnetwork = "${google_compute_subnetwork.shippy-freight.name}"
	zone = "${var.gcloud-zone}"

	initial_node_count = 1

	master_auth {
		username = <redacted>
		password = <redacted>
	}

	node_config {

		# Defines the type/size instance to use
		# Standard is a sensible starting point
		machine_type = "n1-standard-1"

		# Grants OAuth access to the following API's within the cluster
		oauth_scopes = [
			"https://www.googleapis.com/auth/projecthosting",
			"https://www.googleapis.com/auth/devstorage.full_control",
			"https://www.googleapis.com/auth/monitoring",
			"https://www.googleapis.com/auth/logging.write",
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/cloud-platform"
		]
	}
}

# Creates a new DNS range for cluster
resource "google_dns_record_set" "dev-k8s-endpoint-shippy-freight" {
	name  = "k8s.dev.${google_dns_managed_zone.shippy-freight.dns_name}"
	type  = "A"
	ttl   = 300

	managed_zone = "${google_dns_managed_zone.shippy-freight.name}"

	rrdatas = ["${google_container_cluster.shippy-freight-cluster.endpoint}"]
}
```

现在将您之前创建的密钥移至此项目的根目录并将其命名为 google-cred.json。

你现在应该拥有一切你需要创建一个新的集群！ 但是，我们不要发疯，你应该先进行测试，并检查一切是否正常。

运行 `$ terraform init` - 这将下载任何缺少的提供 providers/plugins。 这会发现我们正在使用 Google Cloud 模块并自动获取这些依赖关系。

现在，如果您运行 `$ terraform` 计划，它会向您显示它将做出的更改。 这几乎就像在你的整个基础设施上做 `$ git diff`。 现在很酷！

浏览部署计划后，我认为很好。

> $ terraform apply

注意：您可能会被要求启用一些 API 以完成此操作，没关系，单击链接，确保启用它们并重新运行 `$ terraform apply`。 如果您想节省时间，请转到 Google 云端控制台中的 API 部分，并启用 DNS，Kubernetes 和计算引擎 API。

这可能需要一段时间，但那是因为发生了很多事情。 但是一旦完成，如果您转到Google云端控制台右侧菜单中的 Kubernetes 引擎或计算引擎细分，则应该能够看到您的新群集。

注意：如果你没有使用免费试用期，这将立即开始花费你的钱。 请务必查看实例定价列表。 哦，而且我在开展这些职位的工作之间已经开动了我的力量。 这不会产生费用，因为这些费用是根据资源使用情况收取的。

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-micro/Screen-Shot-2018-02-10-at-12.25.11-1.png)

而已！ 我们有一个完全正常运行的cluster / vm，准备好让我们开始部署我们的容器。 本系列的下一部分将介绍Kubernetes，以及如何设置和部署容器到Kubernetes。

如果你发现这个系列有用，并且你使用了一个广告拦截器（谁可以责怪你）。 请考虑为我的时间和努力捐赠几块钱。 干杯! https://monzo.me/ewanvalentine


--- 

via: https://ewanvalentine.io/microservices-in-golang-part-7/

作者：[Ewan Valentine](https://ewanvalentine.io/author/ewan/)
译者：[zhangyang9](https://github.com/zhangyang9)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
首发于：https://studygolang.com/articles/23979

# 以编程方式查找 DNS 记录

DNS 记录是与 DNS 服务器关联的映射文件，无论每个域名与哪个 IP 地址关联，它们都能处理发送到每个域名的请求。net 包包含各种方法来查找 DNS 记录的细节。让我们运行一些示例，收集有关 DNS 服务器的信息以及目标域名的相应记录：

## Go 程序查找域名的 A 记录

net.LookupIP() 函数接受一个字符串（domain-name）并返回一个包含主机的 IPv4 和 IPv6 地址的 net.IP 对象切片。

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	iprecords, _ := net.LookupIP("facebook.com")
	for _, ip := range iprecords {
		fmt.Println(ip)
	}
}
```

上述程序的输出列出了以 IPv4 和 IPv6 格式返回的 facebook.com 的 A 记录。

```bash
C:\golang\dns> Go run example1.go
2a03:2880:f12f:83:face:b00c:0:25de
31.13.79.35
```

## Go 程序查找域名的 CNAME 记录

CNAME 是规范名称的缩写。CNAME 本质上是绑定路径的域名和子域名的文本别名。net.LookupCNAME()  函数接受主机域名（m.facebook.com）作为字符串，并返回给定主机的单个规范域名

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	cname, _ := net.LookupCNAME("m.facebook.com")
	fmt.Println(cname)
}
```

m.facebook.com 域名返回的 CNAME 记录如下所示：

```bash
C:\golang\dns> Go run example2.go
star-mini.c10r.facebook.com。
```

## Go 程序查找域名的 PTR 指针记录

这些记录提供从地址到名称的反向绑定。PTR 记录应与正向记录完全匹配。net.LookupAddr() 函数对地址执行反向查找，并返回映射到给定地址的名称列表。

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	ptr, _ := net.LookupAddr("6.8.8.8")
	for _, ptrvalue := range ptr {
		fmt.Println(ptrvalue)
	}
}
```

对于给定的地址，上述程序返回单个反向记录，如下所示：

```bash
C:\golang\dns>go run example3.go
tms_server.yuma.army.mil.
```

## Go 程序查找域名的名称服务器（NS）记录

NS 记录描述了区域的授权名称服务器。NS 还将子域名委托给区域文件上的其他组织。net.LookupNS() 函数将域名（facebook.com）作为字符串，并返回 DNS-NS 记录作为 NS 结构的切片。

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	nameserver, _ := net.LookupNS("facebook.com")
	for _, ns := range nameserver {
		fmt.Println(ns)
	}
}
```

支持该域名的 NS 记录如下所示：

```bash
C:\golang\dns>go run example4.go
&{a.ns.facebook.com.}
&{b.ns.facebook.com.}
```

## Go 程序查找域的 MX 记录

这些记录用来记录可以交换电子邮件的服务器。net.LookupMX() 函数将域名作为字符串，并返回按首选项排序的 MX 结构切片。MX 结构由类型为字符串的 HOST 和 类型为 uint16 的 Pref 组成。

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	mxrecords, _ := net.LookupMX("facebook.com")
	for _, mx := range mxrecords {
		fmt.Println(mx.Host, mx.Pref)
	}
}
```

域名（facebook.com）的输出列表 MX 记录。

```bash
C:\golang\dns>go run example5.go
msgin.vvv.facebook.com. 10
```

## Go 程序查找域名的 SRV 服务记录

LookupSRV 函数尝试解析给定服务，协议和域名的 SRV 查询。第二个参数是 “tcp” 或 “udp”。返回的记录按优先级排序，并按照权重随机化。

```go
package main

import (
	"fmt"
	"net"
)

func main() {
	cname, srvs, err := net.LookupSRV("xmpp-server", "tcp", "golang.org")
	if err != nil {
		panic(err)
	}

	fmt.Printf("\ncname: %s \n\n", cname)

	for _, srv := range srvs {
		fmt.Printf("%v:%v:%d:%d\n", srv.Target, srv.Port, srv.Priority, srv.Weight)
	}
}
```

下面的输出演示了 CNAME 返回，后跟由冒号分隔的 SRV 记录的目标，端口，优先级和权重。

```bash
C:\golang\dns>go run example6.go
cname: _xmpp-server._tcp.golang.org.
```

## Go 程序查找域名的 TXT 记录

TXT 记录存储有关 SPF 的信息，该信息可以识别授权服务器以代表您的组织发送电子邮件。net.LookupTXT() 函数将域名（facebook.com）作为字符串，并返回 DNS TXT 记录的字符串切片。

```bash
package main

import (
	"fmt"
	"net"
)

func main() {
	txtrecords, _ := net.LookupTXT("facebook.com")

	for _, txt := range txtrecords {
		fmt.Println(txt)
	}
}
```

gmail.com 的单个 TXT 记录如下所示。

```bash
C:\golang\dns>go run example7.go
v=spf1 redirect=_spf.facebook.com
```

---

via: http://www.golangprograms.com/find-dns-records-programmatically.html

作者：[golangprograms](http://www.golangprograms.com)
译者：[lovechuck](https://github.com/lovechuck)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

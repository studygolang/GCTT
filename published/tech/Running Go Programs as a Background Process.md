已发布：https://studygolang.com/articles/12729

# 以后台进程方式运行 Go 程序

从 1999 年那时开始我就为 windows 写过服务，一开始用 C/C++，后来用 C#。现在我在 Linux 上用 Go 编写服务端软件。然而我没辙了。更令人沮丧的是，我一开始编写软件所用的操作系统并不是我即将部署所用的操作系统。当然，那是后话了。

我想要在我的 Mac 上以后台进程（守护进程）的方式运行代码。而我的问题也是，我无从下手。

我很幸运在 Bitbucket 上找到了由 Daniel Theophanes 编写的名为 [service](https://bitbucket.org/kardianos/service/src) 的开源项目。通过它的代码我学会了如何在 Mac OS 上创建，安装，启动和停止守护进程。当然，这个项目也支持 Linux 和 Windows。

## Mac OS 上的后台进程

Mac OS 上有两种类型的后台进程。守护进程（Daemons）和代理进程（Agents）。解释如下：

**守护进程** 作为整个系统的一部分运行在后台（也就是说，它并不和某个特定用户关联）。守护进程没有图形化界面，甚至不允许连接窗口服务器（window server，并非指 Microsoft 的 Windows）。Web 服务器就是一个典型的例子。

**代理进程** 则不同，它在后台以特定用户的身份运行。它能做很多守护进程做不到的事，比如可靠地访问用户的主目录或者连接窗口服务器。

更多信息，访问：
[http://developer.apple.com/library/mac/#documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/Introduction.html](http://developer.apple.com/library/mac/#documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/Introduction.html)

我们来看看如何在 Mac OS 上配置一个守护进程吧。

![mac-os-daemons](https://raw.githubusercontent.com/studygolang/gctt-images/master/background-process/Screen-Shot-2013-07-28-at-9.51.35-AM.png)

打开 Finder 你可以找到如下目录。Library 下的 LaunchDaemons 目录是用来给我们存放 launchd .plist 文件的。/System 下面也有一个 Library/LaunchDaemons 目录，它则用来为操作系统的守护进程服务。

在 Mac OS 上，launchd 程序是用来管理（包括启动和停止）守护进程，应用程序，进程和脚本的一个服务管理框架。一旦内核启动了 launchd 程序，它就会开始扫描系统上的一系列目录，包括 /etc 下的脚本，/Library 和 /System/Library 下的 LaunchAgents，LaunchDaemons 目录。LaunchDaemons 下找到的程序会以 root 的身份运行。

这是一份 launchd .plist 文件的样例，包含了基本的配置信息：

```xml
<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"http://www.apple.com/DTDs/PropertyList-1.0.dtd\" >
<plist version='1.0'>
<dict>
 <key>Label</key><string>My Service</string>
 <key>ProgramArguments</key>
 <array>
  <string>/Users/bill/MyService/MyService</string>
 </array>
 <key>WorkingDirectory</key><string>/Users/bill/MyService</string>
 <key>StandardOutPath</key><string>/Users/bill/MyService/My.log</string>
 <key>KeepAlive</key><true/>
 <key>Disabled</key><false/>
</dict>
</plist>
```

这里你能找到 .plist 文件中各种不同的配置项：
[https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/launchd.plist.5.html](https://developer.apple.com/library/mac/documentation/Darwin/Reference/ManPages/man5/launchd.plist.5.html)

ProgramArguments 字段非常重要：

```xml
<key>ProgramArguments</key>
<array>
	<string>/Users/bill/MyService/MyService</string>
</array>
```

通过这个字段，你能够指定运行的程序和传递给该程序 main 方法所需的其他参数。

WorkingDirectory 和 StandardOutPath 两个字段也同样很有用：

```xml
<key>WorkingDirectory</key><string>/Users/bill/MyService</string>
<key>StandardOutPath</key><string>/Users/bill/MyService/My.log</string>
```

有了 launchd .plist 文件之后，我们就可以通过一个叫 launchctl 的程序来让我们的程序以守护进程的方式启动。

```
launchctl load /Library/LaunchDaemons/MyService.plist
```

launchctl 程序提供了服务控制和报告的功能。load 命令用以基于 launchd .plist 启动一个守护进程。要验证一下程序是否已经成功启动可以使用 list 命令：

```
launchctl list

PID  Status  Label
948  -       0x7ff4a9503410.anonymous.launchctl
946  -       My Service
910  -       0x7ff4a942ce00.anonymous.bash
```

可以看到，My Service 正在运行，PID 为 946。现在要停止程序可以用 unload 命令：

```
launchctl unload /Library/LaunchDaemons/MyService.plist
launchctl list

PID  Status  Label
948  -       0x7ff4a9503410.anonymous.launchctl
910  -       0x7ff4a942ce00.anonymous.bash
```

程序被终止了。我们还需要处理一下程序在启动和终止的时候操作系统发送的启动和停止请求。

## 操作系统相关的 Go 文件

你可以编写只针对特定平台的 Go 文件。

![go-specific-platform](https://raw.githubusercontent.com/studygolang/gctt-images/master/background-process/Screen-Shot-2013-07-28-at-9.52.55-AM.png)

在我的 LiteIDE 项目中你可以看到 5 个 Go 源文件。其中有 3 个文件名称中包含了其针对的目标平台的名字（darwin (Mac)，linux 和 windows）。

因为我现在是在 Mac OS 下进行构建，所以 `service_linux.go` 和 `service_windows.go` 两个文件会被编译器忽略。

编译器默认就能识别这种命名规范。

这是个很酷的特性，因为在不同的平台上总是要处理一些不同的事，使用一些不同的包。比如在 `service_windows.go` 文件中，就引用了下面这些：

```
"code.google.com/p/winsvc/eventlog"
"code.google.com/p/winsvc/mgr"
"code.google.com/p/winsvc/svc"
```

目前我并没有安装这些包，因为我并不打算在 windows 上运行它。但这并不影响构建因为 `service_windows.go` 被忽略了。

这还有另一个好处。因为这些文件中只有一个会被编译，所以我可以复用这些文件中的类型和方法名。也就是说，任何使用这个包的代码在更改平台之后也不需要做任何修改。实在是酷。

## 服务接口

每个服务为了提供命令和控制功能都必须实现三个接口。

```go
type Service interface {
	Installer
	Controller
	Runner
}

type Installer interface {
	Install(config *Config) error
	Remove() error
}

type Controller interface {
	Start() error
	Stop() error
} 

type Runner interface {
	Run(config *Config) error
}
```

Installer 接口提供了在特定的操作系统上安装和卸载后台进程的逻辑。Controller 接口提供了从命令行启动和停止程序的逻辑。最后 Runner 接口是用来实现当被请求的时候作为服务执行的所有应用逻辑。

## Darwin 下的实现

既然这篇文章是针对 Mac OS 的，我就专注于说明 service_darwin.go 源文件的实现。

Installer 接口需要实现两个方法，Install 和 Remove。按照上文所说，我们需要为服务创建一个 launchd .plist 文件。要完成这步最好的方式是使用 text/template 包。

_InstallScript 函数使用了一个多行的字符串（字符串面值）来创建 launchd .plist 文件的模版。

```go
func _InstallScript() (script string) {
	return `<?xml version='1.0' encoding='UTF-8'?>
<!DOCTYPE plist PUBLIC \"-//Apple Computer//DTD PLIST 1.0//EN\" \"[ http://www.apple.com/DTDs/PropertyList-1.0.dtd ](../../broken-link.html)\" >
<plist version='1.0'>
<dict>
  <key>Label</key><string><b>{{.DisplayName}}</b></string>
  <key>ProgramArguments</key>
  <array>
	<string><b>{{.WorkingDirectory}}</b>/<b>{{.ExecutableName}}</b></string>
  </array>
  <key>WorkingDirectory</key><string><b>{{.WorkingDirectory}}</b></string>
  <key>StandardOutPath</key><string><b>{{.LogLocation}}</b>/<b>{{.Name}}</b>.log</string>
  <key>KeepAlive</key><true/>
  <key>Disabled</key><false/>
</dict>
</plist>`
}
```

多行字符串很酷的一点是其中的回车，换行和空格也不会被忽略。因为这只是个模版，所以我们需要让其中的变量能够被真实的数据替换。{{.variable_name}} 表达式用来定义这些变量。

下面是 Install 方法的实现：

```go
func (service *_DarwinLaunchdService) Install(config *Config) error {
	confPath := service._GetServiceFilePath()

	_, err := os.Stat(confPath)
	if err == nil {
		return fmt.Errorf("Init already exists: %s", confPath)
	}

	file, err := os.Create(confPath)
	if err != nil {
		return err
	}
	defer file.Close()

	parameters := struct {
		ExecutableName string
		WorkingDirectory string
		Name string
		DisplayName string
		LongDescription string
		LogLocation string
	}{
		service._Config.ExecutableName,
		service._Config.WorkingDirectory,
		service._Config.Name,
		service._Config.DisplayName,
		service._Config.LongDescription,
		service._Config.LogLocation,
	}

	template := template.Must(template.New("launchdConfig").Parse(_InstallScript()))
	return template.Execute(file, &parameters)
}
```

`_GetServiceFilePath()` 函数用来在不同的平台下都能获取到配置文件的路径。Darwin 下是这样的：

```go
func (service *_DarwinLaunchdService) _GetServiceFilePath() string {
	return fmt.Sprintf("/Library/LaunchDaemons/%s.plist", service._Config.Name)
}
```

然后代码会检查文件是否已经存在，不存在则会创建一个空文件。紧接着我们创建一个结构体并填充好 template.Execute 函数调用需要的各项参数。注意字段的名称要和模版中 `{{.variable_name}}` 变量的名称匹配。

Execute 函数会处理好模版并将其写入磁盘文件。

Controller 接口需要两个方法，Start 和 Stop。在 Darwin 中代码实现很简单：

```go
func (service *_DarwinLaunchdService) Start() error {
	confPath := service._GetServiceFilePath()

	cmd := exec.Command("launchctl", "load", confPath)
	return cmd.Run()
}

func (service *_DarwinLaunchdService) Stop() error {
	confPath := service._GetServiceFilePath()

	cmd := exec.Command("launchctl", "unload", confPath)
	return cmd.Run()
}
```

如同我们前面所说的一样，每个方法都会调用 launchctl 程序。这是个启动和停止守护进程的简便方法。

最后要实现的 Runner 接口只有一个叫 Run 的方法。

```go
func (service *_DarwinLaunchdService) Run(config *Config) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("******> SERVICE PANIC:", r)
		}
	}()

	fmt.Print("******> Initing Service\n")

	if config.Init != nil {
		if err := config.Init(); err != nil {
			return err
		}
	}

	fmt.Print("******> Starting Service\n")

	if config.Start != nil {
		if err := config.Start(); err != nil {
			return err
		}
	}

	fmt.Print("******> Service Started\n")

	// Create a channel to talk with the OS
	var sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	// Wait for an event
	whatSig := <-sigChan

	fmt.Print("******> Service Shutting Down\n")

	if config.Stop != nil {
		if err := config.Stop(); err != nil {
			return err
		}
	}

	fmt.Print("******> Service Down\n")
	return nil
}
```

当程序开始以守护进程的方式运行时 Run 方法就会被调用。它首先会调用用户的 onInit 和 onStart 方法。用户会做一些初始化工作，执行他们的程序，然后返回。

接着创建一个 channel 与操作系统进行交互。signal.Notify 绑定了 channel 用来接收操作系统的各类事件。代码接着会执行一个无限循环直到操作系统有事件通知程序。代码会寻找关闭事件。一旦接受到了一个关闭事件，用户的 onStop 方法就会被调用，然后 Run 方法返回并终止程序。

## 服务管理器

服务管理器提供了所有的样板代码，因此任何程序都能很轻松地实现服务。下面的代码展示了 Config 的 Run 方法：

```go
func (config *Config) Run() {
	var err error
	config.Service, err = NewService(config)
	if err != nil {
		fmt.Printf("%s unable to start: %s", config.DisplayName, err)
		return
	}

	// Perform a command and then return
	if len(os.Args) > 1 {
		verb := os.Args[1]

		switch verb {
			case "install":
				if err := service.Install(config); err != nil {
					fmt.Println("Failed to install:", err)
					return
				}

				fmt.Printf("Service \"%s\" installed.\n", config.DisplayName)
				return

			case "remove":
				if err := service.Remove(); err != nil {
					fmt.Println("Failed to remove:", err)
					return
				}

				fmt.Printf("Service \"%s\" removed.\n", config.DisplayName)
				return

			case "debug":
				config.Start(config)

				fmt.Println("Starting Up In Debug Mode")

				reader := bufio.NewReader(os.Stdin)
				reader.ReadString('\n')

				fmt.Println("Shutting Down")

				config.Stop(config)
				return

			case "start":
				if err := service.Start(); err != nil {
					fmt.Println("Failed to start:", err)
					return
				}

				fmt.Printf("Service \"%s\" started.\n", config.DisplayName)
				return

			case "stop":
				if err := service.Stop(); err != nil {
					fmt.Println("Failed to stop:", err)
					return
			 	}

				fmt.Printf("Service \"%s\" stopped.\n", config.DisplayName)
				return

			default:
				fmt.Printf("Options for \"%s\": (install | remove | debug | start | stop)\n", os.Args[0])
				return
		}
	}

	// Run the service
	service.Run(config)
}
```

Run 方法一开始通过提供的配置创建了 service 对象。接着查询传入的命令行参数。如果参数是一个命令，则相应的命令便会执行，接着程序终止。如果命令是 debug，程序就会以类似服务的方式启动除非它没有被操作系统钩入。点击 `<enter>` 键可以结束程序。

如果没有传入任何命令行参数，代码就会通过调用 service.Run 方法尝试以守护进程的方式启动。

## 服务实现

下面的代码是一个使用 service 框架的例子：

```go
package main

import (
	"fmt"
	"path/filepath"

	"github.com/goinggo/service/v1"
)

func main() {
	// Capture the working directory
	workingDirectory, _ := filepath.Abs("")

	// Create a config object to start the service
	config := service.Config{
		ExecutableName: "MyService",
		WorkingDirectory: workingDirectory,
		Name: "MyService",
		DisplayName: "My Service",
		LongDescription: "My Service provides support for…",
		LogLocation: _Straps.Strap("baseFilePath"),

		Init: InitService,
		Start: StartService,
		Stop: StopService,
	}

	// Run any command line options or start the service
	config.Run()
}

func InitService() {
	fmt.Println("Service Inited")
}

func StartService() {
	fmt.Println("Service Started")
}

func StopService() {
	fmt.Println("Service Stopped")
}
```

Init，Start 和 Stop 方法执行完必须立即返回 config.Run 方法。

我的代码已经在 Mac OS 上测试过了。Linux 下的代码除了创建和安装的脚本以外都是一样的。当然 Start 和 Stop 方法的实现使用了不同的程序。不久以后我会测试一部分 Linux 下的代码。至于 Windows，则需要重构一下代码，我就不再构建了。如果你计划使用 Windows，可以从 Daniel 的代码开始自行编写。

一旦代码构建完成，就可以打开终端并执行不同的命令。

```
./MyService debug

./MyService install

./MyService start

./MyService stop
```

和往常一样，我希望这份代码能帮助你创建和运行自己的服务。


----------------

via: https://www.ardanlabs.com/blog/2013/06/running-go-programs-as-background.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[alfred-zhong](https://github.com/alfred-zhong)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
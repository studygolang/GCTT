首发于：https://studygolang.com/articles/22242

# 使用 CGO 和 GoReleaser 进行跨平台编译

我参与了一个开源项目 [Mailchain](https://github.com/mailchain/mailchain) ，该项目使用 Go 语言。我想使用 `CI\CD` 轻松的创建版本。 Golang 是一种允许以简单的方式编译代码，并在不同的操作系统上执行的语言。我发现了一个很棒的工具 [https://goreleaser.com/](https://goreleaser.com/) ，它可以用来构建，打包和发布二进制文件。

这是在 Mac 上构建的命令。

```bash
goreleaser --rm-dist --snapshot
```

- `rm-dist` 该发布需要一个干净的目录，该标志确保 `/dist` 目录会被删除。
- `snapshot` 默认情况下，发布设置为 `release` 。此标志将关闭此操作。

而当我在 Mac 上运行它时，它在构建 Linux 的二进制文件时失败了。

```bash
• releasing using Goreleaser 0.106.0...
• loading config file       file=.goreleaser.yml
... [TRUNCATED]
• BUILDING BINARIES
   • building                  binary=dist/darwin_amd64/mailchain
   • building                  binary=dist/linux_amd64/mailchain
   ⨯ release failed after 9.40s error=failed to build for Linux_amd64: # os/user
/usr/local/Cellar/go/1.12.4/libexec/src/os/user/getgrouplist_unix.go:16:35: warning: passing 'gid_t *' (aka 'unsigned int *') to parameter of type 'int *' converts between pointers to integer types with different sign [-Wpointer-sign]
... [TRUNCATED]
ld: warning: option -s is obsolete and being ignored
ld: warning: ignoring file
... [TRUNCATED]
  "__cgoexp_07a0021afc18_secp256k1GoPanicError", referenced from:
	  _secp256k1GoPanicError in 000023.o
  "__cgoexp_07a0021afc18_secp256k1GoPanicIllegal", referenced from:
	  _secp256k1GoPanicIllegal in 000023.o
  "_crosscall2", referenced from:
	  _secp256k1GoPanicIllegal in 000023.o
	  _secp256k1GoPanicError in 000023.o
  "_main", referenced from:
	 implicit entry/start for main executable
ld: symbol(s) not found for architecture x86_64
clang: error: linker command failed with exit code 1 (use -v to see invocation)
```

这是怎么回事？

问题是一个库依赖于 C 并需要 CGO 。我很喜欢在我的工具包中使用 [GoReleaser](https://goreleaser.com/) ！所以我深入研究了一下，发现这是一个针对不同操作系统进行编译需要安装正确的依赖包的问题。 Docker 似乎非常适合解决这个问题，甚至还有一个 CGO 的 docker 容器。

```bash
docker run --rm --privileged \
-v ($pwd):/go/src/github.com/mailchain/mailchain \
-v /var/run/docker.sock:/var/run/docker.sock \
-w /go/src/github.com/mailchain/mailchain \
goreleaser/goreleaser:latest-cgo release --snapshot --rm-dist
```

- `rm` 每次都从一个新的 docker 容器开始。
- `-v ($pwd):/go/src/github.com/:org/:repo` 在 docker 容器上创建一个卷，其中包含当前目录的内容。
- `-w /go/src/github.com/:org/:repo` 将工作目录设置为代码所在的目录 `go/src` 。

注：记得用你代码库的信息替换 `:org` 和 `:repo` 。

```bash
releasing using Goreleaser 0.106.0...
loading config file       file=.goreleaser.
... [TRUNCATED]
building                  binary=dist/darwin_amd64/mailchain
⨯ release failed after 163.82s error=failed to build for darwin_amd64: Go build GitHub.com/ethereum/go-ethereum/crypto/secp256k1: build constraints exclude all Go files in /go/pkg/mod/github.com/ethereum/go-ethereum@v1.8.26/crypto/secp256k1
# GitHub.com/ethereum/go-ethereum/rpc
/go/pkg/mod/github.com/ethereum/go-ethereum@v1.8.26/rpc/endpoints.go:96:19: undefined: ipcListen
/go/pkg/mod/github.com/ethereum/go-ethereum@v1.8.26/rpc/ipc.go:50:10: undefined: newIPCConnection
```

我们正在取得进展 ......

但是这次我得到了一个看起来与 CGO 无关的错误。然而，这个错误还是和 CGO 相关的，因为 `github.com/ethereum/go-ethereum@v.1.8.26/crypto/secp256k1` 是一个 C 依赖。看上去需要的是一个具有 CGO 库的 docker 容器，并且能够编译到 Linux ， OSX 和 Windows 中。可以使用 [goreleaser-xcgo](https://github.com/mailchain/goreleaser-xcgo) ，它是一个包含所有必需的依赖库和最新版 GoReleaser 的 docker 容器。用法和之前类似。

```bash
docker run --rm --privileged \
-v ($pwd):/go/src/github.com/mailchain/mailchain \
-v /var/run/docker.sock:/var/run/docker.sock \
-w /go/src/github.com/mailchain/mailchain \
mailchain/goreleaser-xcgo Goreleaser --snapshot --rm-dist
```

成功！ :)

这次它创建了所有二进制文件。

```bash
• releasing using Goreleaser 0.106.0...
... [TRUNCATED]
• BUILDING BINARIES
  • building                  binary=dist/darwin_amd64/mailchain
  • building                  binary=dist/linux_amd64/mailchain
  • building                  binary=dist/windows_amd64/mailchain.exe
... [TRUNCATED]
• release succeeded after 330.99s
```

它的工作配置文件 `.goreleaser.yml` [在这里]((https://github.com/mailchain/mailchain/blob/3ffc95a23a82e37f1831dd9e397b2e6f104f18e3/.goreleaser.yml))，至于如何与 travis 集成，请看 [`.travis.yml`](https://github.com/mailchain/mailchain/blob/3ffc95a23a82e37f1831dd9e397b2e6f104f18e3/.travis.yml) 配置文件。

如果你想更多的了解我们关于 Mailchain 的工作，请访问 [https://github.com/mailchain/mailchain](https://github.com/mailchain/mailchain)

谢谢

---

via: <https://medium.com/@robdefeo/cross-compile-with-cgo-and-goreleaser-6af884731222>

作者：[Rob De Feo](https://medium.com/@robdefeo)
译者：[lovechuck](https://github.com/lovechuck)
校对：[magichan](https://github.com/magichan)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

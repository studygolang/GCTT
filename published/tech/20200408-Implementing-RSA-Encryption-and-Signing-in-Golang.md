首发于：https://studygolang.com/articles/28458

# 用 Golang 实现 RSA 加密和签名（有示例）

本文介绍 RSA 干了什么，以及我们怎样用 Go 实现它。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200408-Implementing-RSA-Encryption-and-Signing-in-Golang/00.jpg)

RSA（*Rivest–Shamir–Adleman*）加密是使用最广的安全数据加密算法之一。

它是一种非对称加密算法，也叫”单向加密“。用这种方式，任何人都可以很容易地对数据进行加密，而只有用正确的”秘钥“才能解密。

> 如果你想跳过解释直接看源码，点击[这里](https://gist.github.com/sohamkamani/08377222d5e3e6bc130827f83b0c073e)。

## RSA 加密，一言以蔽之

RSA 是通过生成一个公钥和一个私钥进行加/解密的。公钥和私钥是一起生成的，组成一对秘钥对。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200408-Implementing-RSA-Encryption-and-Signing-in-Golang/01.svg)

公钥可以用来加密任意的数据，但不能用来解密。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200408-Implementing-RSA-Encryption-and-Signing-in-Golang/02.svg)

私钥可以用来解密由它对应的公钥加密的数据。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200408-Implementing-RSA-Encryption-and-Signing-in-Golang/03.svg)

这意味着我们可以把我们的公钥给任何想给的人。之后他们可以把想发送给我们的信息进行加密，唯一能访问这些信息的方式就是用我们的私钥进行解密。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200408-Implementing-RSA-Encryption-and-Signing-in-Golang/04.svg)

> 秘钥的生成过程，以及信息的加密解密过程不在本文讨论范围内，但是如果你想研究详细信息，这里有一个关于此主题的[强大视频](https://www.youtube.com/watch?v=wXB-V_Keiu8)。

## 秘钥的生成

我们要做的第一件事就是生成公钥私钥对。这些秘钥是随机生成的，在后面所有的处理中都会用到。

我们用标准库 [crypto/rsa](https://pkg.go.dev/crypto/rsa?tab=doc) 来生成秘钥，用 [crypto/rand](https://pkg.go.dev/crypto/rand?tab=doc) 库来生成随机数。

```go
// The GenerateKey method takes in a reader that returns random bits, and
// the number of bits
privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
if err != nil {
	panic(err)
}

// The public key is a part of the *rsa.PrivateKey struct
publicKey := privateKey.PublicKey

// use the public and private keys
// ...
```

`publicKey` 和 `privateKey` 变量分别用于加密和解密。

## 加密

我们用 [EncryptOEAP](https://pkg.go.dev/crypto/rsa?tab=doc#EncryptOAEP) 函数来加密一串随机的信息。我们需要为这个函数提供一些输入：

1. 一个哈希函数，用了它之后要能保证即使输入做了微小的改变，输出哈希也会变化很大。SHA256 适合于此。
2. 一个用来生成随机位的 random reader，这样相同的内容重复输入时就不会有相同的输出
3. 之前生成的公钥
4. 我们想加密的信息
5. 可选的标签参数（本文中我们忽略）

```go
encryptedBytes, err := rsa.EncryptOAEP(
	sha256.New(),
	rand.Reader,
	&publicKey,
	[]byte("super secret message"),
	nil)
if err != nil {
	panic(err)
}

fmt.Println("encrypted bytes: ", encryptedBytes)
```

这段代码会打印加密后的字节，看起来有点像无用的信息。

## 解密

如果想访问加密字节承载的信息，就需要对它们进行解密。

解密它们的唯一方法就是使用与加密时的公钥对应的私钥。

`*rsa.PrivateKey` 结构体有一个方法 [Decrypt](https://pkg.go.dev/crypto/rsa?tab=doc#PrivateKey.Decrypt)，我们使用这个方法从加密数据中解出原始的信息。

解密时我们需要输入的参数有：1. 被加密的数据（称为*密文*）2. 加密数据用的哈希

```go
// The first argument is an optional random data generator (the rand.Reader we used before)
// we can set this value as nil
// The OEAPOptions in the end signify that we encrypted the data using OEAP, and that we used
// SHA256 to hash the input.
decryptedBytes, err := privateKey.Decrypt(nil, encryptedBytes, &rsa.OAEPOptions{Hash: crypto.SHA256})
if err != nil {
	panic(err)
}

// We get back the original information in the form of bytes, which we
// the cast to a string and print
fmt.Println("decrypted message: ", string(decryptedBytes))
```

## 签名和校验

RSA 秘钥也用于签名和校验。签名不同于加密，签名可以让你宣示真实性，而不是机密性。

也就是说，由原始信息生成一段数据，称为“签名”，而不是伪装原始信息的内容（像[加密](https://www.sohamkamani.com/golang/rsa-encryption/#encryption)中做的那样）。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200408-Implementing-RSA-Encryption-and-Signing-in-Golang/05.svg)

有签名、信息和公钥的任何人，可以用 RSA 校验来确保信息就是来自拥有公钥的人。如果数据和签名不匹配，校验不通过。

![](https://raw.githubusercontent.com/studygolang/gctt-images2/master/20200408-Implementing-RSA-Encryption-and-Signing-in-Golang/06.svg)

请注意，只有拥有私钥的人才能对信息进行签名，但是有公钥的人可以验证它。

```go
msg := []byte("verifiable message")

// Before signing, we need to hash our message
// The hash is what we actually sign
msgHash := sha256.New()
_, err = msgHash.Write(msg)
if err != nil {
	panic(err)
}
msgHashSum := msgHash.Sum(nil)

// In order to generate the signature, we provide a random number generator,
// our private key, the hashing algorithm that we used, and the hash sum
// of our message
signature, err := rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, msgHashSum, nil)
if err != nil {
	panic(err)
}

// To verify the signature, we provide the public key, the hashing algorithm
// the hash sum of our message and the signature we generated previously
// there is an optional "options" parameter which can omit for now
err = rsa.VerifyPSS(&publicKey, crypto.SHA256, msgHashSum, signature, nil)
if err != nil {
	fmt.Println("could not verify signature: ", err)
	return
}
// If we don't get any error from the `VerifyPSS` method, that means our
// signature is valid
fmt.Println("signature verified")
```

## 总结

本文中我们看到了如何生成 RSA 公钥和私钥，以及怎样使用它们进行加密、解密、签名和验证任意数据。

在将它们用于你的数据之前，你需要了解一些使用限制。首先，你要加密的数据必须比你的秘钥短。例如，[EncryptOAEP 文档](https://pkg.go.dev/crypto/rsa?tab=doc#EncryptOAEP) 中说“（要加密的）信息不能比公布的模数减去哈希长度的两倍后再减去 2 长”。

使用的哈希算法要适合你的需求。SHA256（在本例中用的就是 SHA256）可以用于大部分案例，但是如果是对数据要求更高的应用，你可能需要用 SHA512。

你可以在[这里](https://gist.github.com/sohamkamani/08377222d5e3e6bc130827f83b0c073e)找到所有示例的源码。

---

via: https://www.sohamkamani.com/golang/rsa-encryption/

作者：[Soham Kamani](https://twitter.com/sohamkamani)
译者：[lxbwolf](https://github.com/lxbwolf)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

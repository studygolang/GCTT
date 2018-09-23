已发布：https://studygolang.com/articles/12701

# 如何在 JAVA 中进行消息签名并在 GO 中进行验证

在我的公司中，我们使用 Java 和 Go 作为开发平台，当然有时候这些项目彼此之间会进行交互。在这篇文章中，我想要介绍我们的关于在 Java 端进行消息签名并在 Go 服务程序中进行验证的解决方案。

首先，我们聊一聊下面这个架构，我们的 Java 应用程序运行在云上新建虚拟机实例中，并且这个基础镜像实例包含了一个小的 Go 服务程序。这个服务程序是我们的配置管理系统的主入口，我们不希望有来自不可信的客户端可以修改节点。在请求中包含签名的双向 SSL 看起来足以信任客户端。但由于这两个组件都是开源的，所以我们在二进制文件中没有任何“秘密”，因此我们选择了RSA非对称秘钥对来生成和验证签名。Java 端拥有私钥，Go 端拥有公钥。

Java 是一个古老的平台（个人有多年的Java经验）因此，Java 有很多的库，但是我开始使用Go。我没有第六感，但我认为 Go 应该是支持协议的列表中最弱的。好消息是， Go 有一个内置的 crypto/rsa 软件包，坏消息是，它只支持 PKCS#1。在研究期间，我发现了一个支持 PKCS#8 的第三方库，我们不得不在这个计划点上停下来并重点考察：

1. 使用在较老的标准上建立的，经过良好测试的库
2. 使用在新的标准上的未知的库

PKCS#1 不是最新的标准版本，但是另一方面第三方库看起来风险太大。所以我们还是选择了第一个选项。

在业务中，我们有这样一个库，它只有一个功能，即通过 VerifyPSS 函数去验证 PSS（随机签名方案）签名。

```go
func CheckSignature(rawSign string, pubPem []byte, data []byte) error {
	var err error
	var sign []byte
	var pub interface{}
	sign, err = base64.StdEncoding.DecodeString(rawSign)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(pubPem)
	if block == nil {
		return errors.New("Failed to decode public PEM")
	}
	pub, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return err
	}
	newHash := crypto.SHA256.New()
	newHash.Write(data)
	opts := rsa.PSSOptions{SaltLength: 20} // Java default salt length
	err = rsa.VerifyPSS(pub.(*rsa.PublicKey), crypto.SHA256, newHash.Sum(nil), sign, &opts)
	return err
}
```

在 Java 端，客户端将额外的头部放到请求中，该请求中包含了用私钥生成的请求主体签名。下一步就是找到签名并调用之前实现的功能。

```go
func Wrap(handler func(w http.ResponseWriter, req *http.Request), signatureKey []byte) http.Handler {
 return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	body := new(bytes.Buffer)
	defer r.Body.Close()
	ioutil.ReadAll(io.TeeReader(r.Body, body))
	r.Body = ioutil.NopCloser(body) // 我们读取主体两次, 我们必须包装原始的 ReadCloser
	signature := strings.TrimSpace(r.Header.Get("signature"))
	if err := CheckSignature(signature, signatureKey, body.Bytes()); err != nil {
		// Error handling
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 Not Acceptable"))
		return
	}
	http.HandlerFunc(handler).ServeHTTP(w, r)
 })
}
```

最后实现 HTTP 处理程序并将校验程序一起包装起来。

```go
func PostItHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("ok"))
}

func RegisterHandler() {
	signature, _ := ioutil.ReadFile("/path/of/public/key")

	r := mux.NewRouter()
	r.Handle("/postit", Wrap(PostItHandler, signature)).Methods("POST")
	http.Handle("/", r)
	http.ListenAndServe("8080", nil)
}
```

我写了单元测试以确保校验是按照设计进行的。

```go
type TestWriter struct {
	header  http.Header
	status  int
	message string
}

func (w *TestWriter) Header() http.Header {
	return w.header
}

func (w *TestWriter) Write(b []byte) (int, error) {
	w.message = string(b)
	return len(b), nil
}

func (w *TestWriter) WriteHeader(s int) {
	w.status = s
}

func TestWrapAllValid(t *testing.T) {
	pk, _ := rsa.GenerateKey(rand.Reader, 1024)
	pubDer, _ := x509.MarshalPKIXPublicKey(&pk.PublicKey)
	pubPem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Headers: nil, Bytes: pubDer})

	content := "body"
	newHash := crypto.SHA256.New()
	newHash.Write([]byte(content))
	opts := rsa.PSSOptions{SaltLength: 20}
	sign, _ := rsa.SignPSS(rand.Reader, pk, crypto.SHA256, newHash.Sum(nil), &opts)

	body := bytes.NewBufferString(content)
	req, _ := http.NewRequest("GET", "http://valami", body)
	req.Header.Add("signature", base64.StdEncoding.EncodeToString(sign))
	writer := new(TestWriter)
	writer.header = req.Header
	handler := Wrap(func(w http.ResponseWriter, req *http.Request) {}, pubPem)
	handler.ServeHTTP(writer, req)
	if writer.status != 0 {
		t.Errorf("writer.status 0 == %d", writer.status)
	}
}
```

看起来我们已经完成了服务端的实现，现在让我们编写一些 Java 代码吧。我研究了如何在 Java 中生成 PSS 签名，并且我还发现了我们的一个依赖中已经包含了我们所需的功能。 [Bouncy Castle Crypto API](http://bouncycastle.org/)， 在 Java 世界中一个非常出名的库，应用它非常的简单。

```java
// privateKeyPem -  PEM 格式的私钥
// data - 签名数据
public static String generateSignature(String privateKeyPem, byte[] data) {
	try (PEMParser pEMParser = new PEMParser(new StringReader(clarifyPemKey(privateKeyPem)))) {
		PEMKeyPair pemKeyPair = (PEMKeyPair) pEMParser.readObject();

		KeyFactory factory = KeyFactory.getInstance("RSA");
		X509EncodedKeySpec publicKeySpec = new X509EncodedKeySpec(pemKeyPair.getPublicKeyInfo().getEncoded());
		PublicKey publicKey = factory.generatePublic(publicKeySpec);
		PKCS8EncodedKeySpec privateKeySpec = new PKCS8EncodedKeySpec(pemKeyPair.getPrivateKeyInfo().getEncoded());
		PrivateKey privateKey = factory.generatePrivate(privateKeySpec);
		KeyPair kp = new KeyPair(publicKey, privateKey);
		RSAPrivateKeySpec privKeySpec = factory.getKeySpec(kp.getPrivate(), RSAPrivateKeySpec.class);

		PSSSigner signer = new PSSSigner(new RSAEngine(), new SHA256Digest(), 20); //确保我们使用默认的 salt lenght
		signer.init(true, new RSAKeyParameters(true, privKeySpec.getModulus(), privKeySpec.getPrivateExponent()));
		signer.update(data, 0, data.length);
		byte[] signature = signer.generateSignature();

		return BaseEncoding.base64().encode(signature);
	} catch (NoSuchAlgorithmException | IOException | InvalidKeySpecException | CryptoException e) {
		throw new RuntimeException(e);
	}
}

private static String clarifyPemKey(String rawPem) {
	return "-----BEGIN RSA PRIVATE KEY-----\n" + rawPem.replaceAll("-----(.*)-----|\n", "") + "\n-----END RSA PRIVATE KEY-----"; // PEMParser nem kedveli a sortöréseket
}
```

就是这样。。。

ps： 我不为你介绍如何使用 Java 生成秘钥对，因为你可以在网上找到很多关于它的知识。

---

via：https://mhmxs.blogspot.hk/2018/03/how-to-sign-messages-in-java-and-verify.html

作者：[Richárd Kovács](https://mhmxs.blogspot.hk/)
译者：[fredvence](https://github.com/fredvence)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go中文网](https://studygolang.com/) 荣誉推出



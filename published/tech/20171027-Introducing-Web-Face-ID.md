已发布：https://studygolang.com/articles/12299

# 介绍 Web Face ID，如何使用 HTML5，Go 和 Facebox 进行人脸识别

使用你的脸授权解锁你的手机正在成为一种趋势，但是在 Web 上，这种情况看到的并不多，也许是因为这种功能对 Web 来说不太重要，不足以被实现。那么，仅通过 HTML5，Javascript 和一些 Go 语句，能够实现人脸识别的功能么？答案是肯定的，我使用 [Facebox](https://machinebox.io/docs/facebox/teaching-facebox) 在一个小时完成了这个工作，并将 [代码作为一个名为 Web Face ID 的开源项目发布](https://github.com/machinebox/webFaceID)。

## 你怎么从使用人脸识别的网站上获益？

- 你可以使用你的脸作为双重因素认证，比如在完成付款时。
- 你可以证明你不是一个 “机器人”。
- 你可以使用你的脸与一些诸如银行或保险中需要的身份验证的服务进行交互。
- 你可以在一些情况下，加快您的登录系统的速度。
- 你可以减少你系统中的欺诈与模仿登录。

![verify](https://raw.githubusercontent.com/studygolang/gctt-images/master/go-facebox/1.png)
通过 Machinebox 使用 Web Face ID 来验证自己。

## 如何使用标准来实现它

作为一个通用的方法，我们将会使用 HTML5 通过网络摄像头来获取用户头像，用 Javascript 发送一张照片到服务器端。一旦照片到了服务器上，我们将使用 Go 解码照片并使用 Facebox 进行检查，以便能够发出响应。

我们把整个过程分成以下几个步骤。

## 先决条件

你需要启动 [Facebox](https://machinebox.io/docs/facebox) 并运行，为此您 [只需注册一个账户](http://machinebox.io/account) 即可将 Facebox 作为 Docker 容器运行。还需要使用你想要识别的人的照片给 [Facebox 进行训练](https://machinebox.io/docs/facebox/teaching-facebox)，一张照片就够，但多张照片会使结果更准确。

## 使用 HTML5 和 Javascript 捕捉摄像头

对于网站，我们可以利用 HTML5 的 video 标签和 canvas 标签。

```html
<div class='options'>
   <button id="snap">
    <i class='camera icon'></i>
    </button>
    <video id="video" width="100%" autoplay></video>
    <canvas id="canvas" width="400" height="225" style="display:none;"></canvas>
</div>
```

我们将使用 video 标签去捕捉摄像头，并使用 canvas 标签拍摄照片，然后发送到服务器端，下面是 Javascript 代码

```javascript
var video = document.getElementById('video');
if (navigator.mediaDevices &&
          navigator.mediaDevices.getUserMedia){
  navigator.mediaDevices.getUserMedia({video: true}).then(
      function (stream) {
         video.src = window.URL.createObjectURL(stream);
         video.play();
  });
}
var canvas = document.getElementById('canvas');
var context = canvas.getContext('2d');
var video = document.getElementById('video');
var button = $('#snap')
button.click(function(){
   button.addClass('loading')
   context.drawImage(video, 0, 0, 400, 225);
   var dataURL = canvas.toDataURL();
   $.ajax({
       type: "POST",
       url: "/webFaceID",
       data: {
         imgBase64: dataURL
       },
       success: { // ommited }
   })
})
```
上面的代码基本上实现了这样的过程，当你单击该按钮时，将摄像头的照片信息捕获到 canvas 中，并将照片发送到服务器端的断点（endpoint) /webFaceID。这张照片将是一个以 base64 编码的 PNG。

## 使用 Go 处理服务器端的人脸验证

现在我们在服务器端有了你的脸部图像，我们只需要解码图像，将解码后的数据发送给 [Facebox with the SDK](https://github.com/machinebox/sdk-go) 完成后面复杂的工作，然后将处理后的结果返回给前端.
这里我们可以写一个 Go http handler 来做这个工作。

```go
func (s *Server) handlewebFaceID(w http.ResponseWriter, r *http.Request) {
    img := r.FormValue("imgBase64")
    // remove the Data URI before decode
    b64data := img[strings.IndexByte(img, ',')+1:]
    imgDec, err := base64.StdEncoding.DecodeString(b64data)
    if err != nil {
        // omitted error handling
    }
    // validate the face on facebox
    faces, err := s.facebox.Check(bytes.NewReader(imgDec))
    if err != nil {
        // omitted error handling
    }
    var response struct {
        FaceLen int    `json:"faces_len"`
        Matched bool   `json:"matched"`
        Name    string `json:"name"`
    }
    response.FaceLen = len(faces)
    if len(faces) == 1 {
        response.Matched = faces[0].Matched
        response.Name = faces[0].Name
    }
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}
```
通过这几行代码，我们可以在任何网站上使用人脸验证。

## 安全性怎么样？

那么，任何类型的生物识别（比如使用你的脸，虹膜或是指纹）只能作为 “用户名”，而不能作为 “密码”。所以如果你的网站要实现这个功能，那么它可以作为一个理想的第二身份认证，或是某种降低风险的工作，但是它不能取代密码。

另外请记住，恶意攻击者可以拍下你的照片，并用你的照片来仿照你的身份。

> 重点提示：Facebox 经过优化，可以在任何类型的场景下识别任何照片中的人物，但是端点 `/check` 具有可以调整的可选参数 `tolerance`。如果您的脸部验证的条件不会改变（例如相同的位置，相同的环境照明），您可以减小容差，使得验证时系统更加严格。

如果你想看看整个代码，[请访问 Github 上的 Web Face ID](https://github.com/machinebox/webFaceID)。它是开源的。

## 免费试一下吧

你可以很容易地使用我们的盒子来实现这样的功能。[立即注册并免费开始使用此功能](https://machinebox.io/)。

---

via: https://blog.machinebox.io/introducing-web-face-id-how-to-use-html5-go-and-facebox-to-verify-your-face-b75cf2aee5e8

作者：[David Hernandez](https://blog.machinebox.io/@dahernan)
译者：[Titanssword](https://github.com/Titanssword)
校对：[rxcai](https://github.com/rxcai)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

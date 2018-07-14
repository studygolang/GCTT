首发于：https://studygolang.com/articles/13609

# 在 Go 中使用 Time, Timezones 和 Location

今天我遇到个问题。我在编写代码处理 NOAA 的潮汐站 XML 文档时，很快意识到我遇到了麻烦。这是一小段 XML 文档：

```xml
<timezone>LST/LDT</timezone>
<item>
<date>2013/01/01</date>
<day>Tue</day>
<time>02:06 AM</time>
<predictions_in_ft>19.7</predictions_in_ft>
<predictions_in_cm>600</predictions_in_cm>
<highlow>H</highlow>
</item>
```

如果您注意到 timezone 标签，它代表当地标准时间/当地日时。这是一个真实存在的问题因为您需要用 UTC 格式存储这些数据。如果没有正确的时区我就会迷失。我的生意伙伴抓了我的头后，给我看了俩个使用经纬度位置并返回时区信息的 API。很幸运，每个潮汐站我都有经纬度位置信息。

如果您打开这个网页您就能读到这个 Google's Timezone API 文档：

https://developers.google.com/maps/documentation/timezone/

这个 API 相当简单。它需要一个位置，时间戳和一个标志来识别请求的应用是否正在使用传感器（如 GPS 设备）来确定位置。

这是一个简单的 Google API 调用和响应：

```
https://maps.googleapis.com/maps/api/timezone/json?location=38.85682,-92.991714&sensor=false&timestamp=1331766000

{
    "dstOffset" : 3600.0,
    "rawOffset" : -21600.0,
    "status" : "OK",
    "timeZoneId" : "America/Chicago",
    "timeZoneName" : "Central Daylight Time"
}
```

它限制一天只能访问2500次。对于我的潮汐站初始加载，我知道我将达到这个限制，而且我不想等几天再加载所有数据。所有我的商业伙伴从 GeoNames 发现了这个 timezone API。

如果您打开这个网页您就能读到这个 GeoNames's API 文档：

http://www.geonames.org/export/web-services.html#timezone

这个 API 需要一个免费帐号，它相当快就可以设置好。一旦您激活您的帐号，为了使用这个 API您需要找到帐号页去激活您的用户名。

这是一个简单的 GeoNames API 调用和响应：

```
http://api.geonames.org/timezoneJSON?lat=47.01&lng=10.2&username=demo

{
    "time":"2013-08-09 00:54",
    "countryName":"Austria",
    "sunset":"2013-08-09 20:40",
    "rawOffset":1,
    "dstOffset":2,
    "countryCode":"AT",
    "gmtOffset":1,
    "lng":10.2,
    "sunrise":"2013-08-09 06:07",
    "timezoneId":"Europe/Vienna",
    "lat":47.01
}
```

这个 API 返回的信息多一些。而且没有访问限制但是响应时间不能保证。目前我访问了几千次没有遇到问题。

至此我们有两个不同的 web 请求能帮我们获得 timezone 信息。让我们看看怎么使用 Go 去使用 Google web 请求并获得一个返回对象用在我们的程序中。

首先，我们需要定义一个新的类型来包含从 API 返回的信息。

```go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)

const googleURI = "https://maps.googleapis.com/maps/api/timezone/json?location=%f,%f&timestamp=%d&sensor=false "

type GoogleTimezone struct {
    DstOffset    float64 bson:&quot;dstOffset&quot;
    RawOffset    float64 bson:&quot;rawOffset&quot;
    Status       string  bson:&quot;status&quot;
    TimezoneID   string  bson:&quot;timeZoneId&quot;
    TimezoneName string  bson:&quot;timeZoneName&quot;
}
```

Go 对 JSON 和 XML 有非常好的支持。如果您看 GoogleTimezone 结构，您会看到每个字段都包含一个"标签"。标签是额外的数据附加在每个字段，它能通过使用反射获取到。要了解标签的更多信息可以读这个文档。

http://golang.org/pkg/reflect/#StructTag

encoding/json 包定义了一组标签，它可以帮助封装和拆封 JSON 数据。要了解更多关于 Go 对 JSON 的支持可以读这些文档。

http://golang.org/doc/articles/json_and_go.html

http://golang.org/pkg/encoding/json/

如果在结构中您定义的字段名与 JSON 文档中的字段名相同，您就不需要使用标签了。我没那样做是因为标签能告诉 Unmarshal 函数如何映射数据。

让我们来看下这个函数，它能访问 Google API 并将 JSON 文档 Unmarshal 到我们的新类型上：

```go
func RetrieveGoogleTimezone(latitude float64, longitude float64) (googleTimezone *GoogleTimezone, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("%v", r)
        }
    }()

    uri := fmt.Sprintf(googleURI, latitude, longitude, time.Now().UTC().Unix())

    resp, err := http.Get(uri)
    if err != nil {
        return googleTimezone, err
    }

    defer resp.Body.Close()

    // Convert the response to a byte array
    rawDocument, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        return googleTimezone, err
    }

    // Unmarshal the response to a GoogleTimezone object
    googleTimezone = new(GoogleTimezone)
    if err = json.Unmarshal(rawDocument, googleTimezone); err != nil {
        return googleTimezone, err
    }

    if googleTimezone.Status != "OK" {
        err = fmt.Errorf("Error : Google Status : %s", googleTimezone.Status)
        return googleTimezone, err
    }

    if len(googleTimezone.TimezoneId) == 0 {
        err = fmt.Errorf("Error : No Timezone Id Provided")
        return googleTimezone, err
    }

    return googleTimezone, err
}
```

这个 web 请求和错误处理是相当的模式化，所以让我们只简单的谈论下 Unmarshal 调用。

```go
rawDocument, err = ioutil.ReadAll(resp.Body)

err = json.Unmarshal(rawDocument, googleTimezone)
```

当这个 web 调用返回时，我们获取到响应数据并把它存储在一个字节数组中。然后我们调用这个 json Unmarshal 函数，传递字节数组和一个引用到我们返回的指针类型变量。这个 Unmarshal 调用能创建一个 GoogleTimezone类型对象，从返回的 JSON 文档提取并拷贝数据，然后设置这个值到我们的指针变量。它相当聪明，如果任务字段不能映射就被忽略。如果有异常发 Unmarshal 调用会返回一个错误。

所以这很好，我们能得到 timezone 数据并把它解封为一个只有三行代码的对象。现在唯一的问题是我们如何使用 timezoneid 来设置我们的位置？

这又有个问题。我们必须从 feed 文档提取本地时间，使用 timezone 信息转换所有 UTC。

让我们再看一下 feed 文档：

```xml
<timezone>LST/LDT</timezone>
<item>
<date>2013/01/01</date>
<day>Tue</day>
<time>02:06 AM</time>
<predictions_in_ft>19.7</predictions_in_ft>
<predictions_in_cm>600</predictions_in_cm>
<highlow>H</highlow>
</item>
```

假设我们已经从文档提取了数据，我们怎样使用 timezoneid 是我们摆脱困境？看一下我在 main 函数里写的代码。它使用 time.LoadLocation 函数和我们从 API 调用获得的时区 ID 来解决这个问题。

```go
func main() {
    // Call to get the timezone for this lat and lng position
    googleTimezone, err := RetrieveGoogleTimezone(38.85682, -92.991714)
    if err != nil {
        fmt.Printf("ERROR : %s", err)
        return
    }

    // Pretend this is the date and time we extracted
    year := 2013
    month := 1
    day := 1
    hour := 2
    minute := 6

    // Capture the location based on the timezone id from Google
    location, err := time.LoadLocation(googleTimezone.TimezoneId)
    if err != nil {
        fmt.Printf("ERROR : %s", err)
        return
    }

    // Capture the local and UTC time based on timezone
    localTime := time.Date(year, time.Month(month), day, hour, minute, 0, 0, location)
    utcTime := localTime.UTC()

    // Display the results
    fmt.Printf("Timezone:\t%s\n", googleTimezone.TimezoneId)
    fmt.Printf("Local Time: %v\n", localTime)
    fmt.Printf("UTC Time: %v\n", utcTime)
}
```

这是输出：

```
Timezone:   America/Chicago
Local Time: 2013-01-01 02:06:00 -0600 CST
Time:       2013-01-01 08:06:00 +0000 UTC
```

一切运行像冠军一样。我们的 localTime 变量设置为 CST 或 中央标准时间，这是芝加哥所在的位置。Google API 为经纬度提供了正确的时区，因为该位置属于密苏里州。

https://maps.google.com/maps?q=39.232253,-92.991714&z=6

我们要问的最后一个问题是 LoadLocation 函数如果获取时区 ID 字符串并使其工作。时区 ID 包含一个国家和城市（美国/芝加哥）。一定有数以千计这样的时区 ID。

如果我们看一下 LoadLocation 的 time 包文档，我们就能找到答案：

http://golang.org/pkg/time/#LoadLocation

这是 LoadLocation 文档：

```
LoadLocation returns the Location with the given name.

If the name is "" or "UTC", LoadLocation returns UTC. If the name is "Local", LoadLocation returns Local.

Otherwise, the name is taken to be a location name corresponding to a file in the IANA Time Zone database, such as "America/New_York".

The time zone database needed by LoadLocation may not be present on all systems, especially non-Unix systems. LoadLocation looks in the directory or uncompressed zip file named by the ZONEINFO environment variable, if any, then looks in known installation locations on Unix systems, and finally looks in $GOROOT/lib/time/zoneinfo.zip.
```

如果您读最后一段，您将看到 LoadLocation 函数正读取数据库文件获取信息。我没有下载任何数据库，也没设置名为 ZONEINFO 的环境变量。唯一的答案是在 GOROOT 下的 zoneinfo.zip文件。让我们看下：

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/using-time-timezones-and-location-in-go/Screen+Shot+2013-08-08+at+8.06.04+PM.png)

果然有个 zoneinfo.zip 文件在 Go 的安装位置下的 lib/time 目录下。非常酷！！！

您有它了。现在您知道如何使用 time.LoadLocation函数来帮助确保您的时间值始终在正确的时区。如果您有经纬度，则可以使用任一 API 获取该时区 ID。

如果您想要这两个API 都被调的代码可重用副本的话，我已经在 Github 的 GoingGo 库中添加了一个名为 timezone 的新包。以下是整个工作示例程序：

```go
package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"
)

const (
    googleURI = "https://maps.googleapis.com/maps/api/timezone/json?location=%f,%f&timestamp=%d&sensor=false"
)

type GoogleTimezone struct {
    DstOffset    float64 bson:&quot;dstOffset&quot;
    RawOffset    float64 bson:&quot;rawOffset&quot;
    Status       string  bson:&quot;status&quot;
    TimezoneID   string  bson:&quot;timeZoneId&quot;
    TimezoneName string  bson:&quot;timeZoneName&quot;
}

func main() {
    // Call to get the timezone for this lat and lng position
    googleTimezone, err := RetrieveGoogleTimezone(38.85682, -92.991714)
    if err != nil {
        fmt.Printf("ERROR : %s", err)
        return
    }

    // Pretend this is the date and time we extracted
    year := 2013
    month := 1
    day := 1
    hour := 2
    minute := 6

    // Capture the location based on the timezone id from Google
    location, err := time.LoadLocation(googleTimezone.TimezoneID)
    if err != nil {
        fmt.Printf("ERROR : %s", err)
        return
    }

    // Capture the local and UTC time based on timezone
    localTime := time.Date(year, time.Month(month), day, hour, minute, 0, 0, location)
    utcTime := localTime.UTC()

    // Display the results
    fmt.Printf("Timezone:\t%s\n", googleTimezone.TimezoneID)
    fmt.Printf("Local Time: %v\n", localTime)
    fmt.Printf("UTC Time: %v\n", utcTime)
}

func RetrieveGoogleTimezone(latitude float64, longitude float64) (googleTimezone *GoogleTimezone, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("%v", r)
        }
    }()

    uri := fmt.Sprintf(googleURI, latitude, longitude, time.Now().UTC().Unix())

    resp, err := http.Get(uri)
    if err != nil {
        return googleTimezone, err
    }

    defer resp.Body.Close()

    // Convert the response to a byte array
    rawDocument, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return googleTimezone, err
    }

    // Unmarshal the response to a GoogleTimezone object
    googleTimezone = new(GoogleTimezone)
    if err = json.Unmarshal(rawDocument, &googleTimezone); err != nil {
        return googleTimezone, err
    }

    if googleTimezone.Status != "OK" {
        err = fmt.Errorf("Error : Google Status : %s", googleTimezone.Status)
        return googleTimezone, err
    }

    if len(googleTimezone.TimezoneID) == 0 {
        err = fmt.Errorf("Error : No Timezone Id Provided")
        return googleTimezone, err
    }

    return googleTimezone, err
}
```

---

via: https://www.ardanlabs.com/blog/2013/08/using-time-timezones-and-location-in-go.html

作者：[William Kennedy](https://github.com/ardanlabs/gotraining)
译者：[themoonbear](https://github.com/themoonbear)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

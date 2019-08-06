首发于：https://studygolang.com/articles/22433

# 使用 Go 和 ReactJS 构建聊天系统（五）：优化前端

本节完整代码：[GitHub](https://github.com/watermelo/realtime-chat-go-react/tree/part-5)

> 本文是关于使用 ReactJS  和 Go 构建聊天应用程序的系列文章的第 5 部分。你可以在这里找到第 4 部分 - [处理多个客户端](https://studygolang.com/articles/22430)

欢迎来到本系列的第 5 部分！如果你已经学到这儿了，那么我希望你享受学习 Go 的乐趣并运用 Go 和 React 建立自己的聊天系统！

在本节中，我们将再次关注前端，并对其进行优化，以便可以输入自定义的聊天消息，并且以更好的方式显示新的聊天消息。

## 聊天输入组件

我们需要创建一个新的组件。该组件基本上只渲染 `<input />` 的内容，然后监听 `onKeyDown` 事件（译者注：`onkeydown` 事件会在用户按下键盘按键时触发）。当用户在 `<input />` 元素内按键时，它将触发 `onKeyDown` 事件的函数。

```js
import React, { Component } from "react";
import "./ChatInput.scss";

class ChatInput extends Component {
	render() {
		return (
			<div className="ChatInput">
				<input onKeyDown={this.props.send} />
			</div>
		);
	}
}

export default ChatInput;
```

然后，我们将为新的输入组件定义一些样式：

```css
.ChatInput {
	width: 95%;
	display: block;
	margin: auto;

	input {
		padding: 10px;
		margin: 0;
		font-size: 16px;
		border: none;
		border-radius: 5px;
		border: 1px solid rgba(0, 0, 0, 0.1);
		width: 98%;
		box-shadow: 0 5px 15px -5px rgba(0, 0, 0, 0.1);
	}
}
```

定义了组件和样式，现在只需要导出它。

```js
import ChatInput from "./ChatInput.jsx";

export default ChatInput;
```

### 更新 App.js

我们创建了 `ChatInput` 组件，现在需要更新 `App.js`，以便它使用新组件并将已经定义的 `send()` 函数传递给该组件。

```js
render() {
	return (
		<div className="App">
			<Header />
			<ChatHistory chatHistory={this.state.chatHistory} />
			<ChatInput send={this.send} />
		</div>
	);
}
```

我们已经传入了定义的 `send()` 函数，该函数现在只是向 WebSocket 端点发送一个简单的 “Hello” 字符串。我们需要修改它，以便接收触发它的事件的上下文。

通过传递这个事件，我们将能够查询按下的键是否是 `Enter` 键，如果是，我们将 `<input />` 字段的值发送到 WebSocket 端点，然后清除 `<input />`：

```js
send(event) {
	if(event.keyCode === 13) {
		sendMsg(event.target.value);
		event.target.value = "";
	}
}
```

### 测试

现在已经创建了 `ChatInput` 组件，我们来运行 Go WebSocket 服务和前端，尝试发送一些自定义消息，看看是否都按预期工作。

## 优化聊天记录组件

现在，我们有一个相当丑陋但功能正常的聊天记录界面，它显示从 WebSocket 服务向连接的客户端广播的每一条消息。

这条消息只是以 JSON 格式显示，没有额外的样式，所以现在让我们看一下通过创建另一个 `Message` 组件来优化它。

### Message 组件

我们先定义 `Message.jsx` 文件。该组件将通过  `prop` 展示接收的消息。然后它将解析成名为 `message` 的 `prop`，并将其存储在组件状态中，然后我们可以在 `render` 函数中使用它。

```js
// src/components/Message/Message.jsx
import React, { Component } from "react";
import "./Message.scss";

class Message extends Component {
	constructor(props) {
		super(props);
		let temp = JSON.parse(this.props.message);
		this.state = {
			message: temp
		};
	}

	render() {
		return <div className="Message">{this.state.message.body}</div>;
	}
}

export default Message;
```

跟之前一样，我们还需要定义一个 `index.js` 文件，以使其在项目的其余部分中可导出：

```js
// src/components/Message/index.js
import Message from "./Message.jsx";

export default Message;
```

到此为止，我们的组件样式还是比较基本的，只是在一个框中显示消息，我们再设置一些 `box-shadow`，使聊天界面有点视觉深度。

```css
.Message {
	display: block;
	background-color: white;
	margin: 10px auto;
	box-shadow: 0 5px 15px -5px rgba(0, 0, 0, 0.2);
	padding: 10px 20px;
	border-radius: 5px;
	clear: both;

	&.me {
		color: white;
		float: right;
		background-color: #328ec4;
	}
}
```

## 更新历史聊天记录组件

创建好了 `Message` 组件，我们现在可以在 `ChatHistory` 组件中使用它。我们需要更新 `render()` 函数，如下所示：

```js
render() {
	console.log(this.props.chatHistory);
	const messages = this.props.chatHistory.map(msg => <Message message={msg.data} />);

	return (
		<div className='ChatHistory'>
			<h2>Chat History</h2>
			{messages}
		</div>
	);
};
```

在第 3 行，可以看到已更新的 `.map` 函数返回 `<Message />`组件，并将消息 `prop` 设置为 `msg.data`。随后会将 JSON 字符串传递给每个消息组件，然后它将能够按照自定义的格式解析和展示它。

现在我们可以看到，每当我们从 WebSocket 端点收到新消息时，它就会在 `ChatHistory` 组件中很好地展示出来！

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/chat-system-in-go-and-react-course-series/image_1.png)

> 下一节：Part 6 - [Docker 部署](https://studygolang.com/articles/22434)

---

via: https://tutorialedge.net/projects/chat-system-in-go-and-react/part-5-improved-frontend/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

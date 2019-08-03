首发于：https://studygolang.com/articles/22429

# 使用 Go 和 ReactJS 构建聊天系统（三）：前端实现

本节完整代码：[GitHub](https://github.com/watermelo/realtime-chat-go-react/tree/part-3)

> 本文是关于使用 ReactJS 和 Go 构建聊天应用程序的系列文章的第 3 部分。你可以在这里找到第 2 部分 - [后端实现](https://studygolang.com/articles/22426)

## Header 组件

我们先来创建一个非常简单的 Header 组件。我们需要在 `frontend/src/` 目录下 创建一个叫 `components/` 的新目录，并在其中添加一个 `Header/` 目录，它将容纳 Header 组件的所有文件。

```plain
- src/
- - components/
- - - Header/
- - - - Header.jsx
- - - - index.js
- - - - Header.scss
```

> 注意 - 每当我们创建一个新组件时，我们将在 `components/` 目录中创建一个新目录，我们会在该目录中创建这三个文件（*.jsx，*.js，*.scss）。

### Header.jsx

我们需要在 `Header.jsx` 文件中实现函数组件。这将用于呈现网站的标题：

```js
import React from "react";
import "./Header.scss";

const Header = () => (
	<div className="header">
		<h2>Realtime Chat App</h2>
	</div>
);

export default Header;
```

### Header.scss

接下来，我们需要加上一些样式。 由于 ReactJS 项目没有处理 `scss` 文件的能力，因此我们首先需要在 `frontend/` 目录中运行以下命令来安装 `node-sass`：

```shell
$ yarn add node-sass
```

安装完成后，我们就可以添加样式了：

```css
.header {
	background-color: #15223b;
	width: 100%;
	margin: 0;
	padding: 10px;
	color: white;

	h2 {
		margin: 0;
		padding: 0;
	}
}
```

### index.js
最后，我们要导出 `Header` 组件以至于其他组件可以导入它并在它们自己的 `render()` 函数中展示它：

```js
import Header from "./header.jsx";

export default Header;
```

### 更新 App.js

现在已经创建好了 `Header` 组件，我们需要将它导入 `App.js`，然后通过将它添加到我们的 `render()` 函数中来展示它，如下所示：

```js
// App.js
// 从相对路径导入组件
import Header from './components/Header/Header';
// ...
render() {
	return (
		<div className="App">
			<Header />
			<button onClick={this.send}>Hit</button>
		</div>
	);
}
```

保存这个文件后，我们的前端应用程序需要重新编译，然后可以看到 `Header` 组件成功展示在浏览器页面的顶部。

> 恭喜 - 你已经成功创建了第一个 React 组件！

## 历史聊天记录组件

我们已经构建并渲染了一个非常简单的组件，所以我们再来构建一个更复杂一点的组件。

在这个小节中，我们将创建一个历史聊天记录组件，它用来显示我们从 WebSocket 服务收到的所有消息。

我们将在 `components/` 目录中创建一个新文件夹叫 `ChatHistory/`。同样，我们需要为这个组件创建三个文件。

### ChatHistory.jsx

我们从 `ChatHistory.jsx` 文件开始吧。它比之前的要稍微复杂一些，因为我们将构建一个 `Class` 组件，而不是我们上面 Header 组件的 `Function` 组件。

> 注意 - 我们可以使用 `ES6 calss` 定义类组件。如果你想了解更多有关信息，建议查看官方文档：[功能和类组件](https://reactjs.org/docs/components-and-props.html#function-and-class-components)

在这个组件中，你会注意到有一个 `render()` 函数。  `render()` 函数返回一个用于展示此特定组件的 `jsx`。

该组件将通过 `props` 从 App.js 函数中接收一组聊天消息，然后将它们按列表由上往下展示。

```js
import React, { Component } from "react";
import "./ChatHistory.scss";

class ChatHistory extends Component {
	render() {
		const messages = this.props.chatHistory.map((msg, index) => (
			<p key={index}>{msg.data}</p>
		));

		return (
			<div className="ChatHistory">
				<h2>Chat History</h2>
				{messages}
			</div>
		);
	}
}

export default ChatHistory;
```

### ChatHistory.scss

我们在 `ChatHistory.scss` 中来为 `ChatHistory` 组件添加一个小样式，只是简单的修改一下背景颜色和填充及边距：

```css
.ChatHistory {
	background-color: #f7f7f7;
	margin: 0;
	padding: 20px;
	h2 {
		margin: 0;
		padding: 0;
	}
}
```

### Index.js

最后，我们需要导出新组件，就像使用 `Header` 组件一样，这样它就可以在 `App.js` 中被导入并展示：

```js
import ChatHistory from "./ChatHistory.jsx";

export default ChatHistory;
```

## 更新 App.js 和 api/index.js

现在我们又添加了 `ChatHistory` 组件，我们需要实际提供一些消息。

在本系列的前一部分中，我们建立了双向通信，回显发送给它的任何内容，因此每当我们点击应用程序中的发送消息按钮时，都会收到一个新消息。

来更新一下 `api/index.js` 文件和 `connect()` 函数，以便它从 WebSocket 连接收到新消息时用于回调：

```js
let connect = cb => {
	console.log("connecting");

	socket.onopen = () => {
		console.log("Successfully Connected");
	};

	socket.onmessage = msg => {
		console.log(msg);
		cb(msg);
	};

	socket.onclose = event => {
		console.log("Socket Closed Connection: ", event);
	};

	socket.onerror = error => {
		console.log("Socket Error: ", error);
	};
};
```

因此，我们在函数中添加了一个 `cb` 参数。每当我们收到消息时，都会在第 10 行调用此 `cb` 会调函数。

当我们完成这些修改，就可以通过 `App.js` 来添加此回调函数，并在获取新消息时使用 `setState` 来更新状态。

我们将把 `constructor` 函数 `connect()` 移动到 `componentDidMount()` 函数中调用，该函数将作为组件生命周期的一部分自动调用（译者注：在 render() 方法之后调用）。

```js
// App.js
componentDidMount() {
	connect((msg) => {
		console.log("New Message")
		this.setState(prevState => ({
			chatHistory: [...this.state.chatHistory, msg]
		}))
		console.log(this.state);
	});
}
```

然后更新 `App.js` 的 `render()` 函数并展示 `ChatHistory` 组件：

```js
render() {
	return (
		<div className="App">
			<Header />
			<ChatHistory chatHistory={this.state.chatHistory} />
			<button onClick={this.send}>Hit</button>
		</div>
	);
}
```

当我们编译并运行前端和后端项目时，可以看到每当点击前端的发送消息按钮时，它会继续通过 `WebSocket` 连接向后端发送消息，然后后端将其回传给前端，最终在 `ChatHistory` 组件中成功展示！

![](https://raw.githubusercontent.com/studygolang/gctt-images/master/chat-system-in-go-and-react-course-series/image_3.png)

## 总结

我们成功地改进了前端应用程序，并将其视为聊天应用程序。在本系列的下一部分中，将重点关注以下内容：

- 改进前端：添加新的发送消息组件以允许我们发送自定义消息
- 改进后端：处理多个客户端以及跨客户端的通信。

> 下一节：Part 4 - [处理多客户端](https://studygolang.com/articles/22430)

---

via: https://tutorialedge.net/projects/chat-system-in-go-and-react/part-3-designing-our-frontend/

作者：[Elliot Forbes](https://twitter.com/elliot_f)
译者：[咔叽咔叽](https://github.com/watermelo)
校对：[polaris1119](https://github.com/polaris1119)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出

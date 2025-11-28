# 🚀 Easy-IM — 简易 WebSocket 实时聊天室

基于 Go 语言和 `gorilla/websocket` 实现的轻量级、多用户在线聊天服务器。支持：
- 公共广播聊天
- 查看在线用户（`who`）
- 自定义用户名（`rename|新名字`）
- 私聊功能（`to|用户名|消息内容`）
- 自动超时断开（5 分钟无活动）

非常适合学习 WebSocket 编程或作为小型 IM 系统的基础。

---

## ✨ 功能演示

```text
> rename|Alice
更新用户名成功:Alice

> who
[127.0.0.1:54321]Alice:在线...

> to|Bob|你好啊！
（仅 Bob 会收到）Alice私聊：你好啊！

> 大家好！
[Alice]: 大家好！
```

## 🛠 快速启动

### 1. 安装依赖  
确保已安装 Go 1.19+

```bash
go mod tidy

```
### 2. 启动服务器
```bash
go run .

```
默认监听 0.0.0.0:8081，WebSocket 地址：ws://localhost:8081/ws

### 3. 连接测试
可使用以下任一方式连接：

- **浏览器控制台**（需支持 WebSocket）：
```js
  const ws = new WebSocket("ws://localhost:8081/ws");
  ws.onmessage = (e) => console.log(e.data);
  ws.send("hello");
```
- **wscat**（推荐）：
```bash
npm install -g wscat
wscat -c ws://localhost:8081/ws

```
## 📝 指令说明
| 指令 | 说明 |
|------|------|
| `who` | 列出所有在线用户 |
| `rename`|新名字 | 修改自己的用户名（不能重复） |
| `to`|用户名|消息 | 向指定用户发送私聊消息 |
| `任意其他文本` | 广播给所有在线用户 |

## ⚙️ 代码结构
```
├── main.go          # 服务器入口
├── client.go        # Client 结构与方法（上线/下线/消息处理）
└── server.go        # Server 结构、广播逻辑与 WebSocket 路由
```

## 🔒 注意事项
- **当前允许所有跨域请求（CheckOrigin: true），生产环境请限制来源。**
- **用户名默认为 IP:Port，可通过 rename|xxx 修改。**
- **消息通道带缓冲（64 条），防止瞬时高并发阻塞。**
- **5 分钟无消息将自动断开连接。**











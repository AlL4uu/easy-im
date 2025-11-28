package main

import (
	"github.com/gorilla/websocket"
	"strings"
)

type Client struct {
	Name   string
	Addr   string
	C      chan string
	conn   *websocket.Conn
	server *Server
}

func NewClient(conn *websocket.Conn, server *Server) *Client {
	userAddr := conn.RemoteAddr().String()
	user := &Client{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string, 64), // 建议带缓冲，防止阻塞
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()
	return user
}

func (cli *Client) Online() {
	cli.server.mapLock.Lock()
	cli.server.OnlineMap[cli.Name] = cli
	cli.server.mapLock.Unlock()

	cli.server.BroadCast(cli, "已上线")
}

func (cli *Client) Offline() {
	cli.server.mapLock.Lock()
	delete(cli.server.OnlineMap, cli.Name)
	cli.server.mapLock.Unlock()

	cli.server.BroadCast(cli, "已下线")
}

// SendMsg 发送文本消息（自动加换行）
func (cli *Client) SendMsg(msg string) {
	cli.conn.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
}

func (cli *Client) DoMessage(msg string) {
	if msg == "who" {
		cli.server.mapLock.Lock()
		for _, user := range cli.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线..."
			cli.SendMsg(onlineMsg)
		}
		cli.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		if _, ok := cli.server.OnlineMap[newName]; ok {
			cli.SendMsg("当前用户名已存在")
		} else {
			cli.server.mapLock.Lock()
			delete(cli.server.OnlineMap, cli.Name)
			cli.server.OnlineMap[newName] = cli
			cli.server.mapLock.Unlock()

			cli.Name = newName
			cli.SendMsg("更新用户名成功:" + cli.Name)
		}
	} else if len(msg) > 3 && strings.HasPrefix(msg, "to|") {
		parts := strings.SplitN(msg, "|", 3)
		if len(parts) != 3 {
			cli.SendMsg("消息格式错误，请使用 to|张三|消息内容")
			return
		}
		remoteName := parts[1]
		content := parts[2]

		if remoteName == "" || content == "" {
			cli.SendMsg("消息格式错误，请使用 to|张三|消息内容")
			return
		}

		remoteUser, ok := cli.server.OnlineMap[remoteName]
		if !ok {
			cli.SendMsg("该用户不存在")
			return
		}

		remoteUser.SendMsg(cli.Name + "私聊：" + content)
	} else {
		cli.server.BroadCast(cli, msg)
	}
}

func (cli *Client) ListenMessage() {
	for msg := range cli.C {
		cli.conn.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
	}
}

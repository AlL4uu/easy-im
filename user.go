package main

import (
	"github.com/gorilla/websocket"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   *websocket.Conn
	server *Server
}

func NewUser(conn *websocket.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string, 64), // 建议带缓冲，防止阻塞
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()
	return user
}

func (u *User) Online() {
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "已上线")
}

func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	u.server.BroadCast(u, "已下线")
}

// SendMsg 发送文本消息（自动加换行）
func (u *User) SendMsg(msg string) {
	u.conn.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
}

func (u *User) DoMessage(msg string) {
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线..."
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		if _, ok := u.server.OnlineMap[newName]; ok {
			u.SendMsg("当前用户名已存在")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			u.Name = newName
			u.SendMsg("更新用户名成功:" + u.Name)
		}
	} else if len(msg) > 3 && strings.HasPrefix(msg, "to|") {
		parts := strings.SplitN(msg, "|", 3)
		if len(parts) != 3 {
			u.SendMsg("消息格式错误，请使用 to|张三|消息内容")
			return
		}
		remoteName := parts[1]
		content := parts[2]

		if remoteName == "" || content == "" {
			u.SendMsg("消息格式错误，请使用 to|张三|消息内容")
			return
		}

		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("该用户不存在")
			return
		}

		remoteUser.SendMsg(u.Name + "私聊：" + content)
	} else {
		u.server.BroadCast(u, msg)
	}
}

func (u *User) ListenMessage() {
	for msg := range u.C {
		u.conn.WriteMessage(websocket.TextMessage, []byte(msg+"\n"))
	}
}

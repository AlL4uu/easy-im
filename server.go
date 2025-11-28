package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有跨域（生产环境限制）
	},
}

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	return &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string, 64),
	}
}

func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			select {
			case cli.C <- msg:
			default:
				// 防止 channel 阻塞，丢弃旧消息或处理背压
			}
		}
		s.mapLock.Unlock()
	}
}

func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

// WebSocket 处理函数
func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	user := NewUser(conn, s)
	user.Online()
	defer user.Offline()

	// 活跃检测 channel
	isLive := make(chan struct{}, 1) // 带缓冲避免阻塞

	// 读取消息
	go func() {
		defer close(isLive)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				return
			}
			msg := string(message)
			user.DoMessage(msg)
			select {
			case isLive <- struct{}{}:
			default:
			}
		}
	}()

	// 超时检测
	ticker := time.NewTicker(time.Second * 300)
	defer ticker.Stop()

	for {
		select {
		case <-isLive:
			// 活跃，重置 ticker（通过重新创建）
			ticker.Stop()
			ticker = time.NewTicker(time.Second * 300)
		case <-ticker.C:
			user.SendMsg("超时未活动，请重新连接")
			return
		}
	}
}

func (s *Server) Start() {
	http.HandleFunc("/ws", s.wsHandler)

	addr := fmt.Sprintf("%s:%d", s.Ip, s.Port)
	fmt.Printf("WebSocket 服务器启动在: ws://%s/ws\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

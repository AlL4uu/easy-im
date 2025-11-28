package main

func main() {
	server := NewServer("0.0.0.0", 8081)

	// 启动广播协程
	go server.ListenMessage()

	server.Start()
}

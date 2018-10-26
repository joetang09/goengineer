package webserver

import (
	"fmt"
	"sync"
	"testing"
)

var (
	dealer *ChatCallback
)

func chatDemo() {
	callback := new(ChatCallback)
	var err error
	callback.controller, err = RegisterWebSocket("/chat-demo", "chat", callback)
	if err != nil {
		panic(err)
	}
	dealer = callback
}

type ChatCallback struct {
	controller *WebSocketController
	members    sync.Map
}

func (c *ChatCallback) OnReciveTextMessage(id string, msg []byte) {
	c.members.Range(func(k, v interface{}) bool {
		member, _ := k.(string)
		err := c.controller.SendTextMessage(member, append([]byte("revice "+id+" : "), msg...))
		fmt.Println(err)
		return true
	})
}

func (c *ChatCallback) OnReciveBinaryMessage(id string, msg []byte) {}

func (c *ChatCallback) OnConnection(id string, request *WSRequest) bool {
	fmt.Println("new connection : " + id)
	c.members.Store(id, "")
	return true
}
func (c *ChatCallback) OnCloseConnection(id string) {
	fmt.Println("close connection : " + id)
	c.members.Delete(id)
}

func (c *ChatCallback) OnError(err error) {}

func init() {
	WebServer{}.Init(Config{
		Addr:       ":8080",
		Mode:       DEBUG,
		Pprof:      false,
		Host:       "",
		WebSockets: map[string]WebSocketConfig{},
	})

	chatDemo()

	go WebServer{}.Serve()
}

func TestWebsocket(t *testing.T) {
	// websocket.NewClient()

	select {}
}

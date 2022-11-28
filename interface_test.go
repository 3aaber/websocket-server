package ws

import (
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func handler(id string) bool {
	return true
}

func TestInitWebSocketServer(t *testing.T) {
	InitWebSocket(handler)

	time.Sleep(1 * time.Second)

	u := url.URL{
		Scheme: "ws",
		Host:   "localhost:8080",
		Path:   "/ws/sessions/1234",
	}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	if n := NoOfWebSocketClients(); n < 1 {
		t.Errorf("wrong number of clients, expect one, recieved : %d", n)
	}

	ok, _ := GetWebSocketSession("1234")
	if !ok {
		t.Error("error in get web socket session ")
	}
}

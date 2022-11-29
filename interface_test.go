package ws

import (
	"log"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func handler(id string) bool {
	return true
}

func TestInitWebSocketServer(t *testing.T) {
	Host := "localhost:8080"

	u := url.URL{
		Scheme: "ws",
		Host:   Host,
		Path:   "/ws/sessions/",
	}
	InitWebSocket(handler, Host)

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

func TestCallWServerInLoop(t *testing.T) {
	Host := "localhost:8080"

	baseURL := "/ws/sessions/"
	u := url.URL{
		Scheme: "ws",
		Host:   Host,
		Path:   baseURL,
	}
	InitWebSocket(handler, Host)

	for i := 0; i < 1000; i++ {
		id := uuid.New()
		u.Path = baseURL + id.String()

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer c.Close()

		if n := NoOfWebSocketClients(); n != i+1 {
			t.Errorf("wrong number of clients, expect one, recieved : %d", n)
		}

		ok, _ := GetWebSocketSession(id.String())
		if !ok {
			t.Error("error in get web socket session ")
		}
	}

}

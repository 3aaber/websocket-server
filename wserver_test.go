package ws

import (
	"log"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocket(t *testing.T) {

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

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < 10; i++ {
		t := <-ticker.C
		err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}

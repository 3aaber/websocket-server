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
	Host := "localhost:8081"

	u := url.URL{
		Scheme: "ws",
		Host:   Host,
		Path:   "/ws/sessions/",
	}
	InitWebSocket(handler, Host)

	log.Printf("connecting to %s", u.String())

	id := uuid.New()
	u.Path = u.Path + id.String()

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	if n := NoOfWebSocketClients(); n < 1 {
		t.Errorf("wrong number of clients, expect one, recieved : %d", n)
	}

	ok, _ := GetWebSocketSession(id.String())
	if !ok {
		t.Error("error in get web socket session ")
	}
}
func TestSendRecieve(t *testing.T) {

	sampleText := "This is Test"
	Host := "localhost:8080"

	baseURL := "/ws/sessions/"
	u := url.URL{
		Scheme: "ws",
		Host:   Host,
		Path:   baseURL,
	}
	InitWebSocket(handler, Host)

	id := uuid.New()
	u.Path = baseURL + id.String()

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Error(err)
	}

	ok, wsession := GetWebSocketSession(id.String())
	if !ok {
		t.Error("problem in get web socket sessions")
		return
	}

	err = wsession.WriteMessage(1, []byte(sampleText))
	if err != nil {
		t.Error(err)
		return
	}

	_, message, err := c.ReadMessage()
	if err != nil {
		t.Error(err)
		return
	}

	if string(message) != sampleText {
		t.Errorf("recieved message differ from sent one : recieved :%s, sent: %s", string(message), sampleText)
		return
	}

	err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}
func BenchmarkCallWServerInLoop(b *testing.B) {
	Host := "localhost:8080"

	baseURL := "/ws/sessions/"
	u := url.URL{
		Scheme: "ws",
		Host:   Host,
		Path:   baseURL,
	}
	InitWebSocket(handler, Host)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		id := uuid.New()
		u.Path = baseURL + id.String()

		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		defer c.Close()

		ok, _ := GetWebSocketSession(id.String())
		if !ok {
			b.Error("error in get web socket session ")
		}
	}

}

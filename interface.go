package ws

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/umpc/go-sortedmap"
	"github.com/umpc/go-sortedmap/asc"
)

// InitWebSocket initialize the web socket server
//
// Input:
//
//	handler: the handler to verify the session id
//	addr : the server address
func InitWebSocket(handler func(string) bool, addr string) {
	if wsserverInternal == nil {
		wsserverInternal = &wsserver{
			RWMutex:       sync.RWMutex{},
			internalWSMap: map[string]*websocket.Conn{},
		}
	}

	wsserverInternal.webSocketMapTTL = sortedmap.New(1, asc.Time)

	wsserverInternal.upgrader = websocket.Upgrader{
		HandshakeTimeout: time.Second * handshakeTimeout,
		ReadBufferSize:   readBufferSize,
		WriteBufferSize:  writeBufferSize,
		WriteBufferPool:  nil,
		Subprotocols:     []string{},
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		},

		CheckOrigin: func(r *http.Request) bool {
			return true
		},

		EnableCompression: EnableCompression,
	}
	gin.SetMode(gin.ReleaseMode)

	wsserverInternal.webServer = gin.New()

	wsserverInternal.webServer.GET(defaultPath, wsserverInternal.getWebSocketHandler(handler))
	wsserverInternal.webServer.DELETE(defaultPath, wsserverInternal.deleteWebSocketHandler(handler))

	// Wait for gin server to initialize and run in background
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(wg *sync.WaitGroup, addr string) {
		wg.Done()
		wsserverInternal.webServer.Run(addr)

	}(wg, addr)
	wg.Wait()

	wsserverInternal.checkTTLofRecords()

}

// GetWebSocketSession get the web socket session for  session id
//
// Input:
// sessionID : the client web socket session id
//
// Output:
// ok : the session related to id exist
// ws : the web socket session
func GetWebSocketSession(sessionID string) (ok bool, ws *websocket.Conn) {
	return wsserverInternal.getWebSocketSession(sessionID)
}

// NoOfWebSocketClients return the number of clients
// Output:
// Number of clients
func NoOfWebSocketClients() int {
	return wsserverInternal.len()
}

// SendMessage send message to websocket session related to sessionID
//
// Input:
// sessionID : session id
// message: message to write to session
//
// Output:
// error
func SendMessage(sessionID string, message []byte) error {
	ok, ws := wsserverInternal.getWebSocketSession(sessionID)
	if !ok {
		return fmt.Errorf("error in session key : %s, no assosiated web socekt session exist", sessionID)
	}
	err := ws.WriteMessage(1, message)
	if err != nil {
		return fmt.Errorf("error in weite to session : %s, error %s", sessionID, err.Error())
	}
	return nil
}

// CloseSession close the session related to ID
//
// Input:
// sessionID : web socket session id
//
// Output:
// error
func CloseSession(sessionID string) error {
	ok, ws := wsserverInternal.getWebSocketSession(sessionID)
	if !ok {
		return fmt.Errorf("error in session key : %s, no assosiated web socekt session exist", sessionID)
	}
	err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return fmt.Errorf("error in close session : %s, error %s", sessionID, err.Error())
	}
	return nil
}

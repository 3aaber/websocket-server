package ws

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/umpc/go-sortedmap"
	"github.com/umpc/go-sortedmap/asc"
)

func GetWebSocketSession(sessionID string) (ok bool, ws *websocket.Conn) {
	return wsserverInternal.getWebSocketSession(sessionID)
}

func InitWebSocket(webSocketChannel chan *websocket.Conn, wg *sync.WaitGroup) {
	wsserverInternal = wsserver{
		RWMutex:       sync.RWMutex{},
		internalWSMap: map[string]*websocket.Conn{},
	}

	wsserverInternal.wsmapTTL = sortedmap.New(4, asc.Time)

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
	wsserverInternal.ginEngine = gin.Default()

	wsserverInternal.ginEngine.GET(path, wsserverInternal.getwshandler(webSocketChannel))
	wsserverInternal.ginEngine.DELETE(path, wsserverInternal.delwshandler(webSocketChannel))

	wsserverInternal.ginEngine.Run()

	go wsserverInternal.checkTTLofRecords()

	wg.Done()
}

package ws

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const path = "/ws/sessions/:id"

type wsserver struct {
	sync.RWMutex
	internalWSMap map[string]*websocket.Conn
	ginEngine     *gin.Engine
	upgrader      websocket.Upgrader
}

var wsserverInternal wsserver

func InitWebSocket(webSocketChannel chan *websocket.Conn, wg *sync.WaitGroup) {
	wsserverInternal = wsserver{
		RWMutex:       sync.RWMutex{},
		internalWSMap: map[string]*websocket.Conn{},
	}

	wsserverInternal.upgrader = websocket.Upgrader{
		HandshakeTimeout: time.Second * 3,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		WriteBufferPool:  nil,
		Subprotocols:     []string{},
		Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		},
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}
	wsserverInternal.ginEngine = gin.Default()

	wsserverInternal.ginEngine.GET(path, wsserverInternal.getwshandler(webSocketChannel))
	wsserverInternal.ginEngine.DELETE(path, wsserverInternal.delwshandler(webSocketChannel))

	wsserverInternal.ginEngine.Run()

	wg.Done()
}

func (w *wsserver) delwshandler(webSocketChannel chan *websocket.Conn) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		sessionID := c.Param("id")

		res := len(sessionID) > 0

		if res {
			res = w.isClientExist(sessionID)
			if res {
				w.onDelClient(sessionID)
			}
		}
	}

	return gin.HandlerFunc(fn)
}

func (w *wsserver) getwshandler(webSocketChannel chan *websocket.Conn) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		sessionID := c.Param("id")

		res := len(sessionID) > 0

		// if the sessionID dont exist, we return 401 status code
		if !res {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		//upgrade get request to websocket protocol
		ws, err := w.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		w.onAddClient(sessionID, ws)

		webSocketChannel <- ws

	}

	return gin.HandlerFunc(fn)

}

func (w *wsserver) onAddClient(id string, ws *websocket.Conn) {
	w.Lock()
	defer w.Unlock()
	w.internalWSMap[id] = ws
}

func (w *wsserver) onDelClient(id string) {
	w.Lock()
	defer w.Unlock()
	delete(w.internalWSMap, id)
}

func (w *wsserver) isClientExist(id string) bool {
	w.Lock()
	defer w.Unlock()
	_, ok := w.internalWSMap[id]
	return ok
}

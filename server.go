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

var internalWSMap map[string]*websocket.Conn

var (
	upgrader = websocket.Upgrader{
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

	ginEngine *gin.Engine
)

func InitWebSocket(webSocketChannel chan *websocket.Conn, wg *sync.WaitGroup) {
	ginEngine = gin.Default()
	ginEngine.GET(path, getwshandler(webSocketChannel))
	ginEngine.DELETE(path, delwshandler(webSocketChannel))
	ginEngine.Run()
	wg.Done()
}

func delwshandler(webSocketChannel chan *websocket.Conn) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		sessionID := c.Param("id")

		res := len(sessionID) > 0

		if res {
			onDelClient(sessionID)
		}
	}

	return gin.HandlerFunc(fn)
}

func getwshandler(webSocketChannel chan *websocket.Conn) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		sessionID := c.Param("id")

		res := len(sessionID) > 0

		// if the sessionID dont exist, we return 401 status code
		if !res {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		//upgrade get request to websocket protocol
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		onAddClient(sessionID, ws)

		webSocketChannel <- ws

	}

	return gin.HandlerFunc(fn)

}

func onAddClient(id string, ws *websocket.Conn) {
	internalWSMap[id] = ws
}

func onDelClient(id string) {
	delete(internalWSMap, id)
}

func isClientExist(id string) bool {
	_, ok := internalWSMap[id]
	if ok {
		return true
	}
	return false
}

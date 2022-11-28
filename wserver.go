package ws

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/umpc/go-sortedmap"
)

const (
	defaultPath       = "/ws/sessions/:id"
	handshakeTimeout  = 3
	readBufferSize    = 1024
	writeBufferSize   = 1024
	EnableCompression = false
	ttlTime           = time.Hour * 1
)

type wsserver struct {
	sync.RWMutex                             // Lock for internal map
	internalWSMap map[string]*websocket.Conn // internal map : session id -> web socket instance
	wsmapTTL      *sortedmap.SortedMap       // sorted Map to save TTL data
	ginEngine     *gin.Engine                // gin engine
	upgrader      websocket.Upgrader         // websocket upgrader
}

var wsserverInternal wsserver

func (w *wsserver) delwshandler(handler func(string) bool) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		sessionID := c.Param("id")

		res := false

		if len(sessionID) > 0 && handler != nil {
			res = handler(sessionID)
		}
		if res {
			res = w.isClientExist(sessionID)
			if res {
				w.deleteClient(sessionID)
			}
		}
	}

	return gin.HandlerFunc(fn)
}

func (w *wsserver) getwshandler(handler func(string) bool) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		sessionID := c.Param("id")

		res := false

		if len(sessionID) > 0 && handler != nil {
			res = handler(sessionID)
		}

		// check if the sessionID dont authorized , we return 401 status code
		if !res {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// check if sessionID already exist in map
		res = w.isClientExist(sessionID)
		if res {
			c.AbortWithStatus(http.StatusBadRequest)
		}

		//upgrade get request to websocket protocol
		ws, err := w.upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println(err)
			return
		}

		w.addClient(sessionID, ws)

	}

	return gin.HandlerFunc(fn)

}

func (w *wsserver) len() int {
	w.Lock()
	defer w.Unlock()
	return len(w.internalWSMap)
}

func (w *wsserver) addClient(id string, ws *websocket.Conn) {
	w.Lock()
	defer w.Unlock()
	w.wsmapTTL.Insert(id, time.Now().Add(ttlTime))
	w.internalWSMap[id] = ws
}

func (w *wsserver) deleteClient(id string) {
	w.Lock()
	defer w.Unlock()
	w.wsmapTTL.Delete(id)
	delete(w.internalWSMap, id)
}

func (w *wsserver) isClientExist(id string) bool {
	w.Lock()
	defer w.Unlock()
	_, ok := w.internalWSMap[id]
	return ok
}

func (w *wsserver) getWebSocketSession(sessionID string) (ok bool, ws *websocket.Conn) {
	w.RLock()
	defer w.RUnlock()
	ws, ok = w.internalWSMap[sessionID]
	return ok, ws
}

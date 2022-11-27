package ws

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/umpc/go-sortedmap"
	"github.com/umpc/go-sortedmap/asc"
)

const (
	path              = "/ws/sessions/:id"
	handshakeTimeout  = 3
	readBufferSize    = 1024
	writeBufferSize   = 1024
	EnableCompression = false
	TTLTime           = time.Hour * 1
)

type wsserver struct {
	sync.RWMutex                             // Lock for internal map
	internalWSMap map[string]*websocket.Conn // internal map : session id -> web socket instance
	wsmapTTL      *sortedmap.SortedMap
	ginEngine     *gin.Engine        // gin engine
	upgrader      websocket.Upgrader // websocket upgrader
}

var wsserverInternal wsserver

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
	w.wsmapTTL.Insert(id, time.Now().Add(TTLTime))
	w.internalWSMap[id] = ws
}

func (w *wsserver) onDelClient(id string) {
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

func (w *wsserver) checkTTLofRecords() {
	reversed := true
	lowerBound := time.Date(1994, 1, 1, 0, 0, 0, 0, time.UTC)
	upperBound := time.Now()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {

		iterCh, err := w.wsmapTTL.BoundedIterCh(reversed, lowerBound, upperBound)
		if err != nil {
			continue
		}
		// defer iterCh.Close()

		for rec := range iterCh.Records() {
			w.onDelClient(rec.Key.(string))
		}
	}

}

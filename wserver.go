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

const (
	defaultPath       = "/ws/sessions/:id"
	handshakeTimeout  = 3
	readBufferSize    = 1024
	writeBufferSize   = 1024
	EnableCompression = false
	ttlTime           = time.Hour * 1
)

type wsserver struct {
	sync.RWMutex                         // Lock for internal map
	internalWSMap   sync.Map             //map[string]*websocket.Conn // internal map : session id -> web socket instance
	webSocketMapTTL *sortedmap.SortedMap // sorted Map to save TTL data
	webServer       *gin.Engine          // gin engine
	upgrader        websocket.Upgrader   // websocket upgrader
}

var (
	wsserverInternal *wsserver
	initialOnce      sync.Once
)

func initializeWebSocketServer(handler func(string) bool, addr string) {

	initialOnce.Do(func() {
		wsserverInternal = &wsserver{
			RWMutex:       sync.RWMutex{},
			internalWSMap: sync.Map{}, //map[string]*websocket.Conn{},
		}
	})

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
func (w *wsserver) deleteWebSocketHandler(handler func(string) bool) gin.HandlerFunc {
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

func (w *wsserver) getWebSocketHandler(handler func(string) bool) gin.HandlerFunc {
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
			c.AbortWithStatus(http.StatusInternalServerError)
			fmt.Println(err)
			return
		}
		w.addClient(sessionID, ws)
	}
	return gin.HandlerFunc(fn)
}

func (w *wsserver) len() int {
	var i int
	w.internalWSMap.Range(func(k, v interface{}) bool {
		i++
		return true
	})
	return i

}

func (w *wsserver) addClient(id string, ws *websocket.Conn) {
	w.Lock()
	defer w.Unlock()
	w.webSocketMapTTL.Insert(id, time.Now().Add(ttlTime))
	w.internalWSMap.Store(id, ws)
}

func (w *wsserver) deleteClient(id string) {
	w.Lock()
	defer w.Unlock()
	w.webSocketMapTTL.Delete(id)
	w.internalWSMap.Delete(id)
}

func (w *wsserver) isClientExist(id string) bool {
	w.Lock()
	defer w.Unlock()
	_, ok := w.internalWSMap.Load(id)
	return ok
}

func (w *wsserver) getWebSocketSession(sessionID string) (ok bool, ws *websocket.Conn) {
	w.RLock()
	defer w.RUnlock()
	returnVal, ok := w.internalWSMap.Load(sessionID)
	return ok, returnVal.(*websocket.Conn)
}

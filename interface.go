package ws

import (
	"fmt"

	"github.com/gorilla/websocket"
)

// InitWebSocket initialize the web socket server
//
// Input:
//
//	handler: the handler to verify the session id
//	addr : the server address
func InitWebSocket(handler func(string) bool, addr string) {
	initializeWebSocketServer(handler, addr)
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

// ShutDownServer shutdown web server
func ShutDownServer() {
	wsserverInternal.shutdown()
}

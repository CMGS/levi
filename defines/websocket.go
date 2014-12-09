package defines

import (
	"net/http"

	"../logs"
	"github.com/gorilla/websocket"
)

type WebSocketWrapper struct {
	*websocket.Conn
	WriteJSON func(interface{}) error
}

func NewWebSocket(endpoint string, readBufferSize, writeBufferSize int) *WebSocketWrapper {
	var dialer = &websocket.Dialer{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
	}

	conn, _, err := dialer.Dial(endpoint, http.Header{})
	if err != nil {
		logs.Assert(err, "Master")
	}

	ws := &WebSocketWrapper{Conn: conn}
	var makeWebSocketWrapper func(*WebSocketWrapper, *websocket.Conn) *WebSocketWrapper
	MakeWrapper(&makeWebSocketWrapper)
	return makeWebSocketWrapper(ws, conn)
}

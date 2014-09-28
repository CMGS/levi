package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketWrapper struct {
	*websocket.Conn
	WriteJSON func(interface{}) error
}

var Ws *WebSocketWrapper

func NewWebSocket(endpoint string) *WebSocketWrapper {
	var dialer = &websocket.Dialer{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
	}

	conn, _, err := dialer.Dial(endpoint, http.Header{})
	if err != nil {
		logger.Assert(err, "Master")
	}

	ws := &WebSocketWrapper{Conn: conn}
	var makeWebSocketWrapper func(*WebSocketWrapper, *websocket.Conn) *WebSocketWrapper
	MakeWrapper(&makeWebSocketWrapper)
	return makeWebSocketWrapper(ws, conn)
}

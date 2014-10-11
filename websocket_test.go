package main

import (
	"testing"

	. "./defines"
)

func Test_MockWebSocket(t *testing.T) {
	load("levi.yaml")
	Ws = NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	MockWebSocket(Ws)
	defer Ws.Close()
	if err := Ws.WriteJSON("aaa"); err != nil {
		t.Error(err)
	}
}

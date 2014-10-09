package main

import (
	"testing"
)

func Test_MockWebSocket(t *testing.T) {
	load("levi.yaml")
	Ws = NewWebSocket(config.Master)
	MockWebSocket(Ws)
	defer Ws.Close()
	if err := Ws.WriteJSON("aaa"); err != nil {
		t.Error(err)
	}
}

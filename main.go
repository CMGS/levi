package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	LoadConfig()
	Etcd = NewEtcdClient(config.Etcd.Machines)

	defer os.Remove(config.PidFile)
	WritePid(config.PidFile)

	var dialer = websocket.Dialer{
		ReadBufferSize:  config.ReadBufferSize,
		WriteBufferSize: config.WriteBufferSize,
	}
	ws, _, err := dialer.Dial(config.Master, http.Header{})
	if err != nil {
		logger.Assert(err, "Master")
	}
	defer ws.Close()

	levi := NewLevi(ws, config.Docker.Endpoint)
	status := NewStatus()
	go status.Listen()
	go status.Load()
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		signal.Notify(c, syscall.SIGHUP)
		signal.Notify(c, syscall.SIGKILL)
		logger.Info("Catch", <-c)
		levi.Exit()
	}()
	go levi.Read()
	levi.Loop()
}

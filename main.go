package main

import (
	"github.com/CMGS/websocket"
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

	var levi = Levi{}
	var dialer = websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, _, err := dialer.Dial(config.Master, http.Header{})
	if err != nil {
		logger.Assert(err, "Master")
	}
	defer ws.Close()

	levi.Connect(config.Docker.Endpoint)
	levi.Load()
	go levi.Report(ws, config.CheckInterval)
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		signal.Notify(c, syscall.SIGHUP)
		logger.Info("Catch", <-c)
		levi.Close()
	}()
	levi.Loop(ws, config.TaskNum, config.TaskInterval)
}

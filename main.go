package main

import (
	"os"
	"os/signal"
	"syscall"

	"./defines"
	"./utils"
)

var Ws *defines.WebSocketWrapper
var Etcd *defines.EtcdWrapper
var Docker *defines.DockerWrapper

var Status *StatusMoniter
var Lenz *LenzForwarder
var Metrics *MetricsRecorder

func main() {
	LoadConfig()
	Etcd = defines.NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	Docker = defines.NewDocker(config.Docker.Endpoint)
	Lenz = NewLenz()
	Status = NewStatus()
	Metrics = NewMetricsRecorder()

	defer os.Remove(config.PidFile)
	utils.WritePid(config.PidFile)

	Ws = defines.NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	defer Ws.Close()

	levi := NewLevi()
	go Status.Listen()
	go Status.Report(STATUS_IDENT)
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		signal.Notify(c, syscall.SIGHUP)
		signal.Notify(c, syscall.SIGKILL)
		signal.Notify(c, syscall.SIGQUIT)
		utils.Logger.Info("Catch", <-c)
		Metrics.StopAll()
		levi.Exit()
	}()
	go levi.Read()
	levi.Loop()
}

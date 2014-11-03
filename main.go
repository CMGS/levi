package main

import (
	"os"
	"os/signal"
	"syscall"

	"./common"
	"./defines"
	"./lenz"
	"./logs"
	"./metrics"
	"./utils"
)

var Ws *defines.WebSocketWrapper
var Etcd *defines.EtcdWrapper
var Docker *defines.DockerWrapper

var Lenz *lenz.LenzForwarder
var Metrics *metrics.MetricsRecorder

var Status *StatusMoniter

func main() {
	LoadConfig()

	Etcd = defines.NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	Docker = defines.NewDocker(config.Docker.Endpoint)

	Lenz = lenz.NewLenz(Docker, config.Lenz)
	Metrics = metrics.NewMetricsRecorder(config.HostName, config.Metrics, Docker)

	utils.WritePid(config.PidFile)
	defer os.Remove(config.PidFile)

	Ws = defines.NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	defer Ws.Close()

	Status = NewStatus()
	levi := NewLevi()
	go Status.Listen()
	go Status.Report(common.STATUS_IDENT)
	go Metrics.Report()
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		signal.Notify(c, syscall.SIGHUP)
		signal.Notify(c, syscall.SIGKILL)
		signal.Notify(c, syscall.SIGQUIT)
		logs.Info("Catch", <-c)
		Metrics.Stop()
		levi.Exit()
	}()
	go levi.Read()
	levi.Loop()
}

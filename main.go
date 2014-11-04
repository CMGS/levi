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

var Lenz *lenz.LenzForwarder
var Metrics *metrics.MetricsRecorder

var Status *StatusMoniter

func main() {
	LoadConfig()

	common.Etcd = defines.NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	common.Docker = defines.NewDocker(config.Docker.Endpoint)

	Lenz = lenz.NewLenz(config.Lenz)
	Metrics = metrics.NewMetricsRecorder(config.HostName, config.Metrics)

	utils.WritePid(config.PidFile)
	defer os.Remove(config.PidFile)

	common.Ws = defines.NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	defer common.Ws.Close()

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

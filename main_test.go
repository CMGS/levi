package main

import (
	"./defines"
	"./lenz"
	"./metrics"
)

func InitTest() {
	load("levi.yaml")
	Docker = defines.NewDocker(config.Docker.Endpoint)
	defines.MockDocker(Docker)
	Ws = defines.NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	defines.MockWebSocket(Ws)
	Etcd = defines.NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	defines.MockEtcd(Etcd)
	if Status == nil {
		Status = NewStatus()
	}
	if Lenz == nil {
		Lenz = lenz.NewLenz(Docker, config.Lenz)
	}
	if Metrics == nil {
		Metrics = metrics.NewMetricsRecorder(config.HostName, config.Metrics, Docker)
	}
}

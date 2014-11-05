package main

import (
	"./common"
	"./defines"
	"./lenz"
	"./metrics"
)

func InitTest() {
	load("levi.yaml")
	common.Docker = defines.NewDocker(config.Docker.Endpoint)
	defines.MockDocker(common.Docker)
	common.Ws = defines.NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	defines.MockWebSocket(common.Ws)
	common.Etcd = defines.NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	defines.MockEtcd(common.Etcd)
	if Status == nil {
		Status = NewStatus()
	}
	if Lenz == nil {
		Lenz = lenz.NewLenz(config.Lenz)
	}
	if Metrics == nil {
		Metrics = metrics.NewMetricsRecorder(config.HostName, config.Metrics)
	}
}

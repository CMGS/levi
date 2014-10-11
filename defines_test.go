package main

import (
	"testing"

	. "./defines"
	"github.com/fsouza/go-dockerclient"
)

func init() {
	load("levi.yaml")
}

func Test_MockWebSocket(t *testing.T) {
	Ws = NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	MockWebSocket(Ws)
	defer Ws.Close()
	if err := Ws.WriteJSON("aaa"); err != nil {
		t.Error(err)
	}
}

func Test_MockEtcd(t *testing.T) {
	load("levi.yaml")
	Etcd = NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	MockEtcd(Etcd)
	resp, err := Etcd.Get("/test", false, false)
	if err != nil || resp != nil {
		t.Error(err)
	}
}

func Test_MockDocker(t *testing.T) {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	err := Docker.PushImage(docker.PushImageOptions{}, docker.AuthConfiguration{})
	if err != nil {
		t.Error(err)
	}
}

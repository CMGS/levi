package main

import (
	"path"
	"testing"
)

var status *Status

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Etcd = NewEtcdClient(config.Etcd.Machines)
	status = NewStatus()
}

func Test_GetInfo(t *testing.T) {
	var containerName string
	containerName = "nbetest_1234"
	appname, p := status.getInfo(containerName)
	if appname != "nbetest" {
		t.Fatal("Get appname failed")
	}
	if p != path.Join("/NBE/_Apps/nbetest/apps", config.Name) {
		t.Fatal("Get path failed")
	}
}

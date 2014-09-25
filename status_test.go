package main

import (
	"fmt"
	"path"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
)

var status *Status

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
	status = NewStatus()
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		return &docker.Container{Name: "/test_1234"}, nil
	}
}

func Test_GetInfo(t *testing.T) {
	var containerName string
	containerName = "test_1234"
	appname, p := status.getInfo(containerName)
	if appname != "test" {
		t.Fatal("Get appname failed")
	}
	if p != path.Join("/NBE/_Apps/test/apps", config.Name) {
		t.Fatal("Get path failed")
	}
}

func Test_StatusAdd(t *testing.T) {
	id := "xxx"
	Etcd.Create = func(p string, o string, _ uint64) (*etcd.Response, error) {
		if p != fmt.Sprintf("/NBE/_Apps/test/apps/%s/%s", config.Name, id) {
			t.Fatal("Write to wrong path")
		}
		return &etcd.Response{}, nil
	}
	status.add(id)
}

func Test_StatusClean(t *testing.T) {
	id := "xxx"
	Etcd.Delete = func(p string, _ bool) (*etcd.Response, error) {
		if p != fmt.Sprintf("/NBE/_Apps/test/apps/%s/%s", config.Name, id) {
			t.Fatal("Delete to wrong path")
		}
		node := &etcd.Node{Nodes: etcd.Nodes{&etcd.Node{}, &etcd.Node{}}}
		return &etcd.Response{Node: node}, nil
	}
	Etcd.Get = func(_ string, _ bool, _ bool) (*etcd.Response, error) {
		node := &etcd.Node{Nodes: etcd.Nodes{&etcd.Node{}, &etcd.Node{}}}
		return &etcd.Response{Node: node}, nil
	}
	status.clean(id)
}

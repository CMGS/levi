package main

import (
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
}

func Test_Stop(t *testing.T) {
	container := Container{"abcdefg", "test"}
	if err := container.Stop(); err != nil {
		t.Fatal(err)
	}
}

func Test_RemoveContainer(t *testing.T) {
	cpath := "/tmp/t1"
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		m := map[string]string{}
		m["/test/config.yaml"] = cpath
		return &docker.Container{Volumes: m}, nil
	}
	f, err := os.Create(cpath)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("test")
	f.Sync()
	f.Close()
	if err := RemoveContainer("abcdefg", false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cpath); err == nil {
		t.Error("Not clean")
	}
}

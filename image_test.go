package main

import (
	"fmt"
	"strings"
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

func Test_Pull(t *testing.T) {
	image := Image{"test", "v1", 12345}
	if err := image.Pull(); err != nil {
		t.Error(err)
	}
}

func Test_Run(t *testing.T) {
	image := Image{"test", "v1", 12345}
	job := &AddTask{
		ident:     "12345",
		CpuShares: 512,
		CpuSet:    "0,1",
		Memory:    12345,
		Cmd:       []string{"python", "xxx.py"},
		Port:      9999,
		Bind:      5000,
	}
	Docker.CreateContainer = func(opts docker.CreateContainerOptions) (*docker.Container, error) {
		if opts.Name != "test_12345" {
			t.Fatal("Name invaild")
		}
		if opts.Config.Env[0] != fmt.Sprintf("NBE_RUNENV=%s", PRODUCTION) {
			t.Fatal("Env invaild")
		}
		return &docker.Container{ID: "abcdefg"}, nil
	}
	Docker.StartContainer = func(id string, opts *docker.HostConfig) error {
		if strings.LastIndex(opts.Binds[0], "/test/config.yaml:ro") == -1 {
			t.Fatal("Bind invaild")
		}
		return nil
	}
	c, err := image.Run(job, 4001)
	if err != nil {
		t.Fatal(err)
	}
	if c.ID != "abcdefg" {
		t.Error("Cid invaild")
	}
}

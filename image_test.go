package main

import (
	"fmt"
	"strings"
	"testing"

	"./common"
	"github.com/fsouza/go-dockerclient"
)

func init() {
	InitTest()
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
	common.Docker.CreateContainer = func(opts docker.CreateContainerOptions) (*docker.Container, error) {
		if opts.Name != "test_12345" {
			t.Error("Name invaild")
		}
		if opts.Config.Env[0] != fmt.Sprintf("NBE_RUNENV=%s", common.PRODUCTION) {
			t.Error("Env invaild")
		}
		return &docker.Container{ID: "abcdefg"}, nil
	}
	common.Docker.StartContainer = func(id string, opts *docker.HostConfig) error {
		if strings.LastIndex(opts.Binds[0], "/test/config.yaml:ro") == -1 {
			t.Error("Bind invaild")
		}
		return nil
	}
	c, err := image.Run(job, 4001)
	if err != nil {
		t.Error(err)
	}
	if c.ID != "abcdefg" {
		t.Error("Cid invaild")
	}
}

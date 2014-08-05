package main

import (
	"fmt"
	"github.com/CMGS/go-dockerclient"
	"os"
	"strconv"
)

type Image struct {
	client      *docker.Client
	appname     string
	version     string
	config_path string
	port        int64
	registry    string
}

func (self *Image) Pull() error {
	url := fmt.Sprintf("%s/%s", self.registry, self.appname)
	if err := self.client.PullImage(
		docker.PullImageOptions{url, self.registry, self.version, os.Stdout},
		docker.AuthConfiguration{}); err != nil {
		return err
	}
	return nil
}

func (self *Image) Run(job *Task, uid int) (*docker.Container, error) {
	image := fmt.Sprintf("%s/%s:%s", self.registry, self.appname, self.version)

	exposedPorts := make(map[docker.Port]struct{})
	port := docker.Port(fmt.Sprintf("%d/tcp", job.Port))
	exposedPorts[port] = struct{}{}

	config := docker.Config{
		CpuShares:    job.Cpus,
		Memory:       job.Memory,
		User:         strconv.Itoa(uid),
		Image:        image,
		Cmd:          job.Cmd,
		Env:          []string{"RUNENV=PROD"},
		WorkingDir:   fmt.Sprintf("/%s", self.appname),
		ExposedPorts: exposedPorts,
	}
	opts := docker.CreateContainerOptions{
		fmt.Sprintf("%s_%d", self.appname, job.Bind),
		&config,
	}
	container, err := self.client.CreateContainer(opts)
	if err != nil {
		return nil, err
	}
	portBindings := make(map[docker.Port][]docker.PortBinding)
	portBindings[port] = []docker.PortBinding{{
		HostIp:   "0.0.0.0",
		HostPort: strconv.FormatInt(job.Bind, 10),
	}}
	hostConfig := docker.HostConfig{
		Binds:        []string{fmt.Sprintf("%s:%s", self.config_path, fmt.Sprintf("/%s/config.yaml", self.appname))},
		PortBindings: portBindings,
		NetworkMode:  BRIDGE_MODE,
	}
	if err := self.client.StartContainer(container.ID, &hostConfig); err != nil {
		return nil, err
	}
	return container, nil
}

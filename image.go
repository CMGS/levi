package main

import (
	"github.com/fsouza/go-dockerclient"
	"os"
	"strconv"
	"strings"
)

type Image struct {
	client      *docker.Client
	appname     string
	version     string
	config_path string
	port        int
}

func (self *Image) Pull(registry *string) error {
	url := strings.Join([]string{*registry, self.appname}, "/")
	if err := self.client.PullImage(
		docker.PullImageOptions{url, *registry, self.version, os.Stdout},
		docker.AuthConfiguration{}); err != nil {
		return err
	}
	return nil
}

func (self *Image) Run(job *Task, registry *string, user string) (*docker.Container, error) {
	image := strings.Join([]string{strings.Join([]string{*registry, self.appname}, "/"), self.version}, ":")
	config := docker.Config{
		CpuShares:  job.Cpus,
		Memory:     job.Memory,
		User:       user,
		Image:      image,
		Entrypoint: []string{job.Entrypoint},
		Env:        []string{"RUNENV=PROD"},
	}
	opts := docker.CreateContainerOptions{
		strings.Join([]string{self.appname, strconv.Itoa(job.Bind)}, "_"),
		&config,
	}
	container, err := self.client.CreateContainer(opts)
	if err != nil {
		return nil, err
	}
	portBindings := make(map[docker.Port][]docker.PortBinding)
	port := docker.Port(strings.Join([]string{strconv.Itoa(job.Port), "tcp"}, "/"))
	portBindings[port] = []docker.PortBinding{{
		HostIp:   "0.0.0.0",
		HostPort: strconv.Itoa(job.Bind),
	}}
	hostConfig := docker.HostConfig{
		Binds:        []string{strings.Join([]string{self.config_path, "/config.yaml"}, ":")},
		PortBindings: portBindings,
	}
	if err := self.client.StartContainer(container.ID, &hostConfig); err != nil {
		return nil, err
	}
	return container, nil
}

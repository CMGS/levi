package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"strconv"
)

type Image struct {
	appname string
	version string
	port    int64
}

func (self *Image) Pull() error {
	url := UrlJoin(config.Docker.Registry, self.appname)
	if err := Docker.PullImage(
		docker.PullImageOptions{url, config.Docker.Registry, self.version, GetBuffer(), false},
		docker.AuthConfiguration{}); err != nil {
		return err
	}
	return nil
}

func (self *Image) Run(job *Task, uid int, runenv string) (*docker.Container, error) {
	image := fmt.Sprintf("%s/%s:%s", config.Docker.Registry, self.appname, self.version)
	configPath := GenerateConfigPath(self.appname, job.ident)
	permdir := GeneratePermdirPath(self.appname, job.ident, runenv == TESTING)

	containerConfig := docker.Config{
		CpuShares: job.CpuShares,
		CpuSet:    job.CpuSet,
		Memory:    job.Memory,
		User:      strconv.Itoa(uid),
		Image:     image,
		Cmd:       job.Cmd,
		Env: []string{
			fmt.Sprintf("NBE_RUNENV=%s", runenv),
			fmt.Sprintf("NBE_PERMDIR=%s", permdir),
		},
		WorkingDir: fmt.Sprintf("/%s", self.appname),
	}

	hostConfig := docker.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:%s:ro", configPath, fmt.Sprintf("/%s/config.yaml", self.appname)),
			fmt.Sprintf("%s:%s", permdir, fmt.Sprintf("/%s/permdir", self.appname)),
			"/var/run:/var/run",
		},
		NetworkMode: config.Docker.Network,
	}

	if job.ShouldExpose() {
		port := docker.Port(fmt.Sprintf("%d/tcp", job.Port))
		exposedPorts := make(map[docker.Port]struct{})
		exposedPorts[port] = struct{}{}
		containerConfig.ExposedPorts = exposedPorts

		portBindings := make(map[docker.Port][]docker.PortBinding)
		portBindings[port] = []docker.PortBinding{{
			HostIp:   "0.0.0.0",
			HostPort: strconv.FormatInt(job.Bind, 10),
		}}
		hostConfig.PortBindings = portBindings
	}

	opts := docker.CreateContainerOptions{
		fmt.Sprintf("%s_%s", self.appname, job.ident),
		&containerConfig,
	}

	container, err := Docker.CreateContainer(opts)
	if err != nil {
		return nil, err
	}

	if err := Docker.StartContainer(container.ID, &hostConfig); err != nil {
		return nil, err
	}
	return container, nil
}

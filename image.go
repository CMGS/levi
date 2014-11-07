package main

import (
	"fmt"
	"strconv"

	"./common"
	"./defines"
	"./lenz"
	"./logs"
	"./utils"
	"github.com/fsouza/go-dockerclient"
)

type Image struct {
	appname string
	version string
	port    int64
}

func (self *Image) Pull() error {
	url := utils.UrlJoin(config.Docker.Registry, self.appname)
	outputStream := lenz.GetDevBuffer(config.Lenz.Stdout)
	if err := common.Docker.PullImage(
		docker.PullImageOptions{url, config.Docker.Registry, self.version, outputStream, false},
		docker.AuthConfiguration{}); err != nil {
		return err
	}
	return nil
}

func (self *Image) Run(job *defines.AddTask, uid int) (*docker.Container, error) {
	image := fmt.Sprintf("%s/%s:%s", config.Docker.Registry, self.appname, self.version)
	configPath := GenerateConfigPath(self.appname, job.Ident)
	mPermdir := fmt.Sprintf("/%s/permdir", self.appname)
	runenv := common.PRODUCTION
	if job.IsTest() {
		runenv = common.TESTING
	}
	permdir := GeneratePermdirPath(self.appname, job.Ident, job.IsTest())

	containerConfig := docker.Config{
		CpuShares: job.CpuShares,
		CpuSet:    job.CpuSet,
		Memory:    job.Memory,
		User:      strconv.Itoa(uid),
		Image:     image,
		Cmd:       job.Cmd,
		Env: []string{
			fmt.Sprintf("NBE_RUNENV=%s", runenv),
			fmt.Sprintf("NBE_PERMDIR=%s", mPermdir),
		},
		WorkingDir: fmt.Sprintf("/%s", self.appname),
	}

	hostConfig := docker.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:%s:ro", configPath, fmt.Sprintf("/%s/config.yaml", self.appname)),
			fmt.Sprintf("%s:%s", permdir, mPermdir),
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
		fmt.Sprintf("%s_%s", self.appname, job.Ident),
		&containerConfig,
	}

	container, err := common.Docker.CreateContainer(opts)
	if err != nil {
		return nil, err
	}

	if err := common.Docker.StartContainer(container.ID, &hostConfig); err != nil {
		// Have to remove resource when start failed
		logs.Debug("Rollback add files")
		RemoveContainer(container.ID, job.IsTest(), false)
		return nil, err
	}
	return container, nil
}

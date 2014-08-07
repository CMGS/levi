package main

import (
	"bytes"
	"fmt"
	"github.com/CMGS/go-dockerclient"
	"path"
	"strconv"
)

type Image struct {
	client     *docker.Client
	appname    string
	version    string
	configPath string
	port       int64
}

func (self *Image) Pull() error {
	url := fmt.Sprintf("%s/%s", RegEndpoint, self.appname)
	buf := bytes.Buffer{}
	if err := self.client.PullImage(
		docker.PullImageOptions{url, RegEndpoint, self.version, &buf},
		docker.AuthConfiguration{}); err != nil {
		logger.Debug(buf.String())
		return err
	}
	logger.Debug(buf.String())
	return nil
}

func (self *Image) Run(job *Task, uid int) (*docker.Container, error) {
	image := fmt.Sprintf("%s/%s:%s", RegEndpoint, self.appname, self.version)

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
	permdir := path.Join(Permdirs, self.appname)
	hostConfig := docker.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:%s", self.configPath, fmt.Sprintf("/%s/config.yaml", self.appname)),
			fmt.Sprintf("%s:%s", permdir, fmt.Sprintf("/%s/permdir", self.appname)),
		},
		PortBindings: portBindings,
		NetworkMode:  NetworkMode,
	}
	if err := self.client.StartContainer(container.ID, &hostConfig); err != nil {
		return nil, err
	}
	return container, nil
}

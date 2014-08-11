package main

import (
	"github.com/CMGS/go-dockerclient"
	"os"
	"path"
)

type Container struct {
	client  *docker.Client
	id      string
	appname string
}

func (self *Container) Stop() error {
	if err := self.client.StopContainer(self.id, CONTAINER_STOP_TIMEOUT); err != nil {
		logger.Info(err)
		if err := self.client.KillContainer(docker.KillContainerOptions{ID: self.id}); err != nil {
			return err
		}
	}
	return nil
}

func (self *Container) Remove() error {
	container, err := self.client.InspectContainer(self.id)
	if err != nil {
		return err
	}
	configPath := container.Volumes[path.Join("/", self.appname, "config.yaml")]
	if err := os.Remove(configPath); err != nil {
		return err
	}
	if err := self.client.RemoveContainer(docker.RemoveContainerOptions{ID: self.id}); err != nil {
		return err
	}
	return nil
}

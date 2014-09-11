package main

import (
	"github.com/CMGS/go-dockerclient"
	"os"
	"path"
)

type Container struct {
	id      string
	appname string
}

func (self *Container) Stop() error {
	if err := Docker.StopContainer(self.id, CONTAINER_STOP_TIMEOUT); err != nil {
		logger.Info(err)
		if err := Docker.KillContainer(docker.KillContainerOptions{ID: self.id}); err != nil {
			return err
		}
	}
	return nil
}

func (self *Container) Remove() error {
	container, err := Docker.InspectContainer(self.id)
	if err != nil {
		return err
	}
	if err := Die(container.ID[:12], container.Name); err != nil {
		return err
	}
	configPath := container.Volumes[path.Join("/", self.appname, "config.yaml")]
	if err := os.Remove(configPath); err != nil {
		return err
	}
	if err := Docker.RemoveContainer(docker.RemoveContainerOptions{ID: self.id}); err != nil {
		return err
	}
	return nil
}

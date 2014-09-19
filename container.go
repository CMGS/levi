package main

import (
	"github.com/fsouza/go-dockerclient"
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
	return Remove(self.id, self.appname, false)
}

func Remove(id, appname string, test bool) error {
	container, err := Docker.InspectContainer(id)
	if err != nil {
		return err
	}
	configPath := container.Volumes[path.Join("/", appname, "config.yaml")]
	if err := os.Remove(configPath); err != nil {
		return err
	}
	if test {
		permdirPath := container.Volumes[path.Join("/", appname, "permdir")]
		if err := os.RemoveAll(permdirPath); err != nil {
			return err
		}
	}
	if err := Docker.RemoveContainer(docker.RemoveContainerOptions{ID: id}); err != nil {
		return err
	}
	return nil
}

package main

import "github.com/fsouza/go-dockerclient"

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

func RemoveContainer(id string, test bool, rmi bool) error {
	container, err := Docker.InspectContainer(id)
	if err != nil {
		return err
	}
	if err := Docker.RemoveContainer(docker.RemoveContainerOptions{ID: id, RemoveVolumes: true}); err != nil {
		return err
	}
	if rmi {
		if err := Docker.RemoveImage(container.Image); err != nil {
			return err
		}
	}
	return nil
}

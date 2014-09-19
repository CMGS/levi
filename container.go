package main

import (
	"github.com/fsouza/go-dockerclient"
	"os"
	"strings"
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

func RemoveContainer(id string, test bool) error {
	container, err := Docker.InspectContainer(id)
	if err != nil {
		return err
	}
	for p, rp := range container.Volumes {
		switch {
		case strings.HasSuffix(p, "/config.yaml"):
			if err := os.Remove(rp); err != nil {
				return err
			}
		case test && strings.HasSuffix(p, "/permdir"):
			if err := os.Remove(rp); err != nil {
				return err
			}
		}
	}
	if err := Docker.RemoveContainer(docker.RemoveContainerOptions{ID: id}); err != nil {
		return err
	}
	return nil
}

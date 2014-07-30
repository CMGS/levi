package main

import (
	"github.com/fsouza/go-dockerclient"
	"os"
	"path"
)

type Container struct {
	client      *docker.Client
	id          string
	appname     string
	config_path string
}

func (self *Container) Stop() error {
	info, err := self.client.InspectContainer(self.id)
	if err != nil {
		return err
	}
	ports_mapping := info.NetworkSettings.PortMappingAPI()
	public_port := ports_mapping[0].PublicPort
	file_name := GenerateConfigPath(self.appname, public_port)
	self.config_path = path.Join(DEFAULT_HOME_PATH, self.appname, file_name)
	if err := self.client.StopContainer(self.id, CONTAINER_STOP_TIMEOUT); err != nil {
		if err := self.client.KillContainer(docker.KillContainerOptions{ID: self.id}); err != nil {
			return err
		}
	}
	return nil
}

func (self *Container) Remove() error {
	if err := os.Remove(self.config_path); err != nil {
		return err
	}
	if err := self.client.RemoveContainer(docker.RemoveContainerOptions{ID: self.id}); err != nil {
		return err
	}
	return nil
}

package main

import (
	"github.com/fsouza/go-dockerclient"
	"os"
	"path"
	"strconv"
	"strings"
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
	file_name := strings.Join([]string{self.appname, strconv.FormatInt(public_port, 10)}, "_")
	file_name = strings.Join([]string{file_name, "yaml"}, ".")
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

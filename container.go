package main

import (
	"github.com/CMGS/go-dockerclient"
	"os"
)

type Container struct {
	client     *docker.Client
	id         string
	appname    string
	configPath string
}

func (self *Container) Stop() error {
	info, err := self.client.InspectContainer(self.id)
	if err != nil {
		return err
	}
	var portsMapping = info.NetworkSettings.PortMappingAPI()
	var publicPort = portsMapping[0].PublicPort
	self.configPath = GenerateConfigPath(self.appname, publicPort)
	if err := self.client.StopContainer(self.id, CONTAINER_STOP_TIMEOUT); err != nil {
		if err := self.client.KillContainer(docker.KillContainerOptions{ID: self.id}); err != nil {
			return err
		}
	}
	return nil
}

func (self *Container) Remove() error {
	if err := os.Remove(self.configPath); err != nil {
		return err
	}
	if err := self.client.RemoveContainer(docker.RemoveContainerOptions{ID: self.id}); err != nil {
		return err
	}
	return nil
}

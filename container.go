package main

import (
	"./common"
	"./logs"
	"./utils"
	"github.com/fsouza/go-dockerclient"
)

type Container struct {
	id      string
	appname string
}

func (self *Container) GetIdent() (ident string, err error) {
	container, err := common.Docker.InspectContainer(self.id)
	if err != nil {
		return
	}
	_, ident, _ = utils.GetAppInfo(container.Name)
	return
}

func (self *Container) Stop() error {
	if err := common.Docker.StopContainer(self.id, common.CONTAINER_STOP_TIMEOUT); err != nil {
		logs.Info("Stop Container", err)
		if err := common.Docker.KillContainer(docker.KillContainerOptions{ID: self.id}); err != nil {
			return err
		}
	}
	return nil
}

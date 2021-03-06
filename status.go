package main

import (
	"fmt"
	"strings"

	"./common"
	"./defines"
	"./logs"
	"./utils"
	"github.com/fsouza/go-dockerclient"
)

type StatusMoniter struct {
	events    chan *docker.APIEvents
	Removable map[string]struct{}
}

func NewStatus() *StatusMoniter {
	status := &StatusMoniter{}
	status.events = make(chan *docker.APIEvents)
	status.Removable = map[string]struct{}{}
	logs.Assert(common.Docker.AddEventListener(status.events), "Attacher")
	return status
}

func (self *StatusMoniter) Listen() {
	logs.Info("Status Monitor Start")
	for event := range self.events {
		logs.Debug("Status:", event.Status, event.ID, event.From)
		if event.Status == common.STATUS_DIE {
			Metrics.Remove(event.ID[:12])
			if _, ok := self.Removable[event.ID]; ok {
				self.die(event.ID)
			}
		}
	}
}

func (self *StatusMoniter) getStatus(s string) string {
	switch {
	case strings.HasPrefix(s, "Up"):
		return common.STATUS_START
	default:
		return common.STATUS_DIE
	}
}

func (self *StatusMoniter) writeBack(result *defines.Result) {
	if err := common.Ws.WriteJSON(result); err != nil {
		logs.Info(err, result)
	}
}

func (self *StatusMoniter) Report(id string) {
	containers, err := common.Docker.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		logs.Info(err, "Load")
	}

	logs.Info("Load container")
	for index, container := range containers {
		if !strings.HasPrefix(container.Image, config.Docker.Registry) {
			continue
		}
		status := self.getStatus(container.Status)
		name, aid, at := utils.GetAppInfo(container.Names[0])
		shortID := container.ID[:12]
		logs.Debug("Container", name, shortID, status)
		if status != common.STATUS_DIE {
			Metrics.Add(name, shortID, at)
			if at != common.TEST_TYPE {
				Lenz.Attacher.Attach(shortID, name, aid, at)
			}
		}
		self.Removable[container.ID] = struct{}{}
		result := &defines.Result{
			Id:    id,
			Done:  true,
			Index: index,
			Type:  common.INFO_TASK,
			Data:  fmt.Sprintf("%s|%s|%s", status, name, container.ID),
		}
		self.writeBack(result)
	}
}

func (self *StatusMoniter) die(id string) {
	container, err := common.Docker.InspectContainer(id)
	if err != nil {
		logs.Info("Status inspect docker failed", err)
		return
	}
	appname, _, _ := utils.GetAppInfo(container.Name)
	result := &defines.Result{
		Id:    common.STATUS_IDENT,
		Done:  true,
		Index: 0,
		Type:  common.INFO_TASK,
		Data:  fmt.Sprintf("%s|%s|%s", common.STATUS_DIE, appname, id),
	}
	self.writeBack(result)
}

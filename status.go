package main

import (
	"strings"

	"github.com/fsouza/go-dockerclient"
)

var Status *StatusMoniter

type StatusMoniter struct {
	events    chan *docker.APIEvents
	Removable map[string]struct{}
}

func NewStatus() *StatusMoniter {
	status := &StatusMoniter{}
	status.events = make(chan *docker.APIEvents)
	status.Removable = map[string]struct{}{}
	logger.Assert(Docker.AddEventListener(status.events), "Attacher")
	return status
}

func (self *StatusMoniter) Listen() {
	logger.Debug("Status Monitor Start")
	for event := range self.events {
		logger.Debug("Status:", event.Status, event.ID, event.From)
		if _, ok := self.Removable[event.ID]; !ok {
			continue
		}
		if event.Status == "die" {
			delete(self.Removable, event.ID)
			self.die(event.ID)
		}
	}
}

func (self *StatusMoniter) getStatus(s string) string {
	switch {
	case strings.HasPrefix(s, "Up"):
		return STATUS_START
	default:
		return STATUS_DIE
	}
}

func (self *StatusMoniter) Report(id string) {
	containers, err := Docker.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		logger.Info(err, "Load")
	}

	result := TaskResult{Id: id}
	result.Status = []*StatusInfo{}

	logger.Info("Load container")
	for _, container := range containers {
		logger.Debug("Container", container)
		if !strings.HasPrefix(container.Image, config.Docker.Registry) {
			continue
		}
		status := self.getStatus(container.Status)
		if status == STATUS_START {
			self.Removable[container.ID] = struct{}{}
		}
		s := &StatusInfo{status, self.getName(container.Names[0]), container.ID}
		result.Status = append(result.Status, s)
	}
	if err := Ws.WriteJSON(result); err != nil {
		logger.Info(err, result)
	}
}

func (self *StatusMoniter) die(id string) {
	result := TaskResult{Id: STATUS_IDENT}
	result.Status = make([]*StatusInfo, 1)

	container, err := Docker.InspectContainer(id)
	if err != nil {
		logger.Info("Status inspect docker failed", err)
		return
	}
	appname := self.getName(container.Name)
	logger.Info("Status Remove:", appname, id)
	result.Status[0] = &StatusInfo{STATUS_DIE, appname, id}
	if err := Ws.WriteJSON(result); err != nil {
		logger.Info(err, result)
	}
}

func (self *StatusMoniter) getName(containerName string) string {
	containerName = strings.TrimLeft(containerName, "/")
	if pos := strings.LastIndex(containerName, "_daemon_"); pos > -1 {
		appname := containerName[:pos]
		return appname
	}
	if pos := strings.LastIndex(containerName, "_test_"); pos > -1 {
		appname := containerName[:pos]
		return appname
	}
	appinfo := strings.Split(containerName, "_")
	appname := strings.Join(appinfo[:len(appinfo)-1], "_")
	return appname
}

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

type Info struct {
	Type    string
	Appname string
	Id      string
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

func (self *StatusMoniter) Report() {
	containers, err := Docker.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		logger.Info(err, "Load")
	}

	info := map[string][]*Info{}
	info[STATUS_IDENT] = []*Info{}

	logger.Info("Load container")
	for _, container := range containers {
		logger.Debug("Container", container)
		if _, ok := self.Removable[container.ID]; !ok || !strings.HasPrefix(container.Image, config.Docker.Registry) {
			continue
		}
		status := self.getStatus(container.Status)
		if status == STATUS_START {
			self.Removable[container.ID] = struct{}{}
		}
		i := &Info{status, self.getName(container.Names[0]), container.ID}
		info[STATUS_IDENT] = append(info[STATUS_IDENT], i)
	}
	if err := Ws.WriteJSON(info); err != nil {
		logger.Info(err, info)
	}
}

func (self *StatusMoniter) die(id string) {
	info := map[string][]*Info{}
	info[STATUS_IDENT] = make([]*Info, 1)

	container, err := Docker.InspectContainer(id)
	if err != nil {
		logger.Info("Status inspect docker failed", err)
		return
	}
	appname := self.getName(container.Name)
	logger.Info("Status Remove:", appname, id)
	info[STATUS_IDENT][0] = &Info{STATUS_DIE, appname, id}
	if err := Ws.WriteJSON(info); err != nil {
		logger.Info(err, info)
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

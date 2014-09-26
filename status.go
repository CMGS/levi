package main

import (
	"strings"

	"github.com/fsouza/go-dockerclient"
)

type Status struct {
	events chan *docker.APIEvents
}

type Info struct {
	Type    string
	Appname string
	Id      string
}

func NewStatus() *Status {
	status := &Status{}
	status.events = make(chan *docker.APIEvents)
	logger.Assert(Docker.AddEventListener(status.events), "Attacher")
	return status
}

func (self *Status) Load() {
	containers, err := Docker.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		logger.Assert(err, "Load")
	}
	logger.Info("Load container")
	for _, container := range containers {
		event := &docker.APIEvents{
			Status: "start",
			ID:     container.ID,
			From:   container.Image,
			Time:   container.Created,
		}
		self.events <- event
	}
}

func (self *Status) Listen() {
	logger.Debug("Status Listener Start")
	for event := range self.events {
		logger.Debug("Status:", event.Status, event.ID, event.From)
		id := event.ID[:12]
		if !strings.HasPrefix(event.From, config.Docker.Registry) {
			continue
		}
		switch event.Status {
		case "start":
			self.add(id)
		case "die":
			self.clean(id)
		}
	}
}

func (self *Status) add(id string) {
	info := map[string][]*Info{}
	info[STATUS_IDENT] = make([]*Info, 1)

	container, err := Docker.InspectContainer(id)
	if err != nil {
		logger.Info("Status inspect docker failed", err)
		return
	}
	appname := self.getName(container.Name)
	logger.Info("Status Add:", appname, id)
	info[STATUS_IDENT][0] = &Info{STATUS_ADD, appname, id}
	if err := Ws.WriteJSON(info); err != nil {
		logger.Info(err, info)
	}
}

func (self *Status) clean(id string) {
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
	RemoveContainer(id, strings.LastIndex(container.Name, "_test_") > -1)
}

func (self *Status) getName(containerName string) string {
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

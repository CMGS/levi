package main

import (
	"strings"

	. "./utils"
	"github.com/fsouza/go-dockerclient"
)

type StatusMoniter struct {
	events    chan *docker.APIEvents
	Removable map[string]struct{}
	lenz      *Lenz
}

func NewStatus() *StatusMoniter {
	status := &StatusMoniter{}
	status.events = make(chan *docker.APIEvents)
	status.lenz = NewLenz()
	status.Removable = map[string]struct{}{}
	Logger.Assert(Docker.AddEventListener(status.events), "Attacher")
	return status
}

func (self *StatusMoniter) Listen() {
	Logger.Info("Status Monitor Start")
	for event := range self.events {
		Logger.Debug("Status:", event.Status, event.ID, event.From)
		if _, ok := self.Removable[event.ID]; ok && event.Status == "die" {
			self.die(event.ID)
		}
		if strings.HasPrefix(event.From, config.Docker.Registry) && event.Status == "start" {
			container, err := Docker.InspectContainer(event.ID)
			if err != nil {
				Logger.Info("Status:", "Inspect Container", event.ID, "Failed")
				continue
			}
			shortID := event.ID[:12]
			if !self.lenz.Attacher.Attached(shortID) {
				self.lenz.Attacher.Attach(shortID, self.getName(container.Name))
			}
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
		Logger.Info(err, "Load")
	}

	result := &TaskResult{Id: id}
	result.Status = []*StatusInfo{}

	Logger.Info("Load container")
	for _, container := range containers {
		if !strings.HasPrefix(container.Image, config.Docker.Registry) {
			continue
		}
		name := self.getName(container.Names[0])
		shortID := container.ID[:12]
		status := self.getStatus(container.Status)
		Logger.Debug("Container", name, shortID, status)
		if !self.lenz.Attacher.Attached(shortID) && status != STATUS_DIE {
			self.lenz.Attacher.Attach(shortID, name)
		}
		self.Removable[container.ID] = struct{}{}
		s := &StatusInfo{status, name, container.ID}
		result.Status = append(result.Status, s)
	}
	if err := Ws.WriteJSON(result); err != nil {
		Logger.Info(err, result)
	}
}

func (self *StatusMoniter) die(id string) {
	result := &TaskResult{Id: STATUS_IDENT}
	result.Status = make([]*StatusInfo, 1)

	container, err := Docker.InspectContainer(id)
	if err != nil {
		Logger.Info("Status inspect docker failed", err)
		return
	}
	appname := self.getName(container.Name)
	result.Status[0] = &StatusInfo{STATUS_DIE, appname, id}
	if err := Ws.WriteJSON(result); err != nil {
		Logger.Info(err, result)
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

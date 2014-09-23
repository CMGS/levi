package main

import (
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v1"
	"path"
	"strings"
)

type Status struct {
	events chan *docker.APIEvents
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
		logger.Debug(event.ID, event.From)
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
	container, err := Docker.InspectContainer(id)
	if err != nil {
		logger.Info("Status inspect docker failed", err)
		return
	}
	appname, p := self.getInfo(container.Name)
	logger.Info("Status Add:", appname, id)
	Etcd.CreateDir(p, 0)
	out, err := yaml.Marshal(container)
	if err != nil {
		logger.Info("Status marshal info failed", err)
		return
	}
	Etcd.Create(path.Join(p, id), string(out), 0)
}

func (self *Status) clean(id string) {
	container, err := Docker.InspectContainer(id)
	if err != nil {
		logger.Info("Status inspect docker failed", err)
		return
	}
	appname, p := self.getInfo(container.Name)
	logger.Info("Status Remove:", appname, id)
	resp, err := Etcd.Get(p, false, false)
	if err != nil {
		logger.Info("Status get levi dir failed", err)
		return
	}
	if _, err := Etcd.Delete(path.Join(p, id), true); err != nil {
		logger.Info("Status delete info file failed", err)
		return
	}
	if len(resp.Node.Nodes)-1 <= 0 {
		Etcd.DeleteDir(p)
	}
	RemoveContainer(id, strings.LastIndex(container.Name, "_test_") > -1)
}

func (self *Status) getInfo(containerName string) (string, string) {
	containerName = strings.TrimLeft(containerName, "/")
	if pos := strings.LastIndex(containerName, "_daemon_"); pos > -1 {
		appname := containerName[:pos]
		return appname, path.Join("/NBE/_Apps", appname, "daemons", config.Name)
	}
	if pos := strings.LastIndex(containerName, "_test_"); pos > -1 {
		appname := containerName[:pos]
		return appname, path.Join("/NBE/_Apps", appname, "tests", config.Name)
	}
	appinfo := strings.Split(containerName, "_")
	appname := strings.Join(appinfo[:len(appinfo)-1], "_")
	return appname, path.Join("/NBE/_Apps", appname, "apps", config.Name)
}

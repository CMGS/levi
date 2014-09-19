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

func (self *Status) Listen() {
	logger.Debug("Status Listener Start")
	for msg := range self.events {
		id := msg.ID[:12]
		if !strings.HasPrefix(msg.From, config.Docker.Registry) {
			continue
		}
		switch msg.Status {
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
	p := self.getPath(container.Name)
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
	p := self.getPath(container.Name)
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

func (self *Status) getPath(containerName string) string {
	if pos := strings.LastIndex(containerName, "_daemon_"); pos > -1 {
		return path.Join("/NBE/_Apps", containerName[:pos], "daemons", config.Name)
	}
	if pos := strings.LastIndex(containerName, "_test_"); pos > -1 {
		return path.Join("/NBE/_Apps", containerName[:pos], "tests", config.Name)
	}
	appinfo := strings.Split(containerName, "_")
	return path.Join("/NBE/_Apps", strings.Join(appinfo[:len(appinfo)-1], "_"), "apps", config.Name)
}

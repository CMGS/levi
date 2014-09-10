package main

import (
	"github.com/CMGS/go-dockerclient"
	"github.com/CMGS/websocket"
	"net"
	"strings"
	"sync"
	"time"
)

var Docker *docker.Client

type Levi struct {
	finish bool
	events chan *docker.APIEvents
}

func (self *Levi) Connect(endpoint string) {
	var err error
	Docker, err = docker.NewClient(endpoint)
	if err != nil {
		logger.Assert(err, "Docker")
	}
	self.events = make(chan *docker.APIEvents)
	logger.Assert(Docker.AddEventListener(self.events), "Attacher")
}

func (self *Levi) Load() []docker.APIContainers {
	containers, err := Docker.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		logger.Info(err)
		self.Close()
	}
	return containers
}

func (self *Levi) Read(ws *websocket.Conn, apptask *AppTask) bool {
	switch err := ws.ReadJSON(apptask); {
	case err != nil:
		if e, ok := err.(net.Error); !ok || !e.Timeout() {
			self.Close()
			logger.Info(err)
		}
	case err == nil:
		logger.Debug(apptask)
		if apptask.Id != "" {
			return true
		}
	}
	return false
}

func (self *Levi) Close() {
	self.finish = true
	Docker.RemoveEventListener(self.events)
}

func (self *Levi) Status() {
	logger.Debug("Status Listener Start")
	for msg := range self.events {
		id := msg.ID[:12]
		logger.Debug("event:", id, msg.Status)
		switch msg.Status {
		case "start":
			if err := Start(id); err != nil {
				logger.Info(err)
			}
		case "stop":
			if err := Die(id); err != nil {
				logger.Info(err)
			}
		}
	}
}

func (self *Levi) NewNginx() *Nginx {
	nginx := &Nginx{
		make(map[string]*Upstream),
		make(map[string]struct{}),
	}
	for _, container := range self.Load() {
		var appinfo = strings.SplitN(strings.TrimLeft(container.Names[0], "/"), "_", 2)
		if strings.Contains(appinfo[1], "daemon_") {
			continue
		}
		appname, apport := appinfo[0], appinfo[1]
		nginx.New(appname, container.ID, apport)
	}
	return nginx
}

func (self *Levi) NewDeploy() *Deploy {
	return &Deploy{
		make([]*AppTask, 0, config.TaskNum),
		&sync.WaitGroup{},
		self.NewNginx(),
	}
}

func (self *Levi) Loop(ws *websocket.Conn) {
	var newtask bool
	var deploy *Deploy
	deploy = self.NewDeploy()
	for !self.finish {
		apptask := AppTask{wg: &sync.WaitGroup{}}
		ws.SetReadDeadline(time.Now().Add(time.Duration(config.TaskInterval) * time.Second))
		logger.Debug("Timeout check")
		if newtask = self.Read(ws, &apptask); newtask {
			deploy.tasks = append(deploy.tasks, &apptask)
		}
		if (len(deploy.tasks) != 0 && !newtask) || len(deploy.tasks) == cap(deploy.tasks) {
			deploy.Deploy(ws)
			deploy = self.NewDeploy()
		}
	}
}

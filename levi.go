package main

import (
	"container/list"
	"github.com/CMGS/go-dockerclient"
	"github.com/CMGS/websocket"
	"net"
	"sync"
	"time"
)

var Docker *docker.Client

type Levi struct {
	containers []docker.APIContainers
	finish     bool
}

func (self *Levi) Connect(endpoint string) {
	var err error
	Docker, err = docker.NewClient(endpoint)
	if err != nil {
		logger.Assert(err, "Docker")
	}
}

func (self *Levi) Load() {
	var err error
	self.containers, err = Docker.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		logger.Assert(err, "Docker")
	}
}

func (self *Levi) Process(ws *websocket.Conn, deploy *Deploy) {
	deploy.Deploy()
	result := deploy.Result()
	if err := ws.WriteJSON(&result); err != nil {
		self.Close()
		logger.Info(err)
	}
	deploy.Reset()
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
}

func (self *Levi) Report(ws *websocket.Conn, sleep int) {
	for !self.finish {
		if err := ws.WriteJSON(&self.containers); err != nil {
			logger.Assert(err, "JSON")
		}
		time.Sleep(time.Duration(sleep) * time.Second)
	}
}

func (self *Levi) Loop(ws *websocket.Conn, num, wait int) {
	var newtask bool
	deploy := &Deploy{
		make(map[string][]interface{}),
		list.New(),
		&sync.WaitGroup{},
		&self.containers,
		&Nginx{
			make(map[string]*Upstream),
			make(map[string]struct{}),
		},
	}
	for !self.finish {
		apptask := AppTask{}
		ws.SetReadDeadline(time.Now().Add(time.Duration(wait) * time.Second))
		logger.Debug(time.Now())
		if newtask = self.Read(ws, &apptask); newtask {
			deploy.tasks.PushBack(apptask)
		}
		if (deploy.tasks.Len() != 0 && !newtask) || deploy.tasks.Len() >= num {
			self.Process(ws, deploy)
			self.Load()
		}
	}
}

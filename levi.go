package main

import (
	"sync"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
)

var Docker interface {
	ListImages(bool) ([]docker.APIImages, error)
	BuildImage(docker.BuildImageOptions) error
	PullImage(docker.PullImageOptions, docker.AuthConfiguration) error
	PushImage(docker.PushImageOptions, docker.AuthConfiguration) error
	RemoveImage(string) error

	AddEventListener(chan<- *docker.APIEvents) error

	ListContainers(docker.ListContainersOptions) ([]docker.APIContainers, error)
	CreateContainer(docker.CreateContainerOptions) (*docker.Container, error)
	InspectContainer(string) (*docker.Container, error)
	KillContainer(docker.KillContainerOptions) error
	RemoveContainer(docker.RemoveContainerOptions) error
	StopContainer(string, uint) error
	StartContainer(string, *docker.HostConfig) error
	WaitContainer(string) (int, error)
}

type Levi struct {
	deploy *Deploy
	ws     *websocket.Conn
	finish bool
	task   chan *AppTask
	err    chan error
}

func NewLevi(ws *websocket.Conn, endpoint string) *Levi {
	var levi *Levi = &Levi{ws: ws}
	var err error

	Docker, err = docker.NewClient(endpoint)
	if err != nil {
		logger.Assert(err, "Docker")
	}

	levi.err = make(chan error)
	levi.task = make(chan *AppTask)
	levi.finish = false
	levi.deploy = &Deploy{
		ws: ws,
		wg: &sync.WaitGroup{},
	}
	levi.deploy.Init()

	return levi
}

func (self *Levi) Exit() {
	self.finish = true
}

func (self *Levi) Clean() {
}

func (self *Levi) Read() {
	for {
		apptask := &AppTask{wg: &sync.WaitGroup{}}
		if err := self.ws.ReadJSON(apptask); err != nil {
			self.err <- err
			continue
		}
		self.task <- apptask
	}
}

func (self *Levi) Loop() {
	for !self.finish {
		select {
		case err := <-self.err:
			logger.Info(err)
			if len(self.deploy.tasks) != 0 {
				self.deploy.Deploy()
			}
			self.Exit()
		case task := <-self.task:
			self.deploy.tasks = append(self.deploy.tasks, task)
			if len(self.deploy.tasks) == cap(self.deploy.tasks) {
				self.deploy.Deploy()
			}
		case <-time.After(time.Second * time.Duration(config.TaskInterval)):
			logger.Debug("Time Check")
			if len(self.deploy.tasks) != 0 {
				self.deploy.Deploy()
			}
		}
	}
	self.Clean()
}

package main

import (
	"container/list"
	"fmt"
	"github.com/CMGS/websocket"
	"github.com/fsouza/go-dockerclient"
	"log"
	"net"
	"sync"
	"time"
)

type Levi struct {
	client     *docker.Client
	containers []docker.APIContainers
	finish     bool
}

func (self *Levi) Connect(url *string) {
	var err error
	self.client, err = docker.NewClient(*url)
	if err != nil {
		log.Fatal("Connect docker failed")
	}
}

func (self *Levi) Load() {
	var err error
	self.containers, err = self.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatal("Query docker failed")
	}
}

func (self *Levi) Process(ws *websocket.Conn, deploy *Deploy) {
	deploy.Deploy()
	result := deploy.Result()
	if err := ws.WriteJSON(&result); err != nil {
		log.Fatal("Write Fail:", err)
	}
	deploy.Reset()
}

func (self *Levi) Read(ws *websocket.Conn, apptask *AppTask) bool {
	switch err := ws.ReadJSON(apptask); {
	case err != nil:
		if e, ok := err.(net.Error); !ok || !e.Timeout() {
			log.Fatal("Read Fail:", err)
		}
	case err == nil:
		if apptask.Id != "" {
			return true
		}
	}
	return false
}

func (self *Levi) Close() {
	self.finish = true
}

func (self *Levi) Report(ws *websocket.Conn, sleep *int) {
	for !self.finish {
		if err := ws.WriteJSON(&self.containers); err != nil {
			log.Fatal("Write Fail:", err)
		}
		time.Sleep(time.Duration(*sleep) * time.Second)
	}
}

func (self *Levi) Loop(ws *websocket.Conn, wait, num *int, dst, ngx, registry *string) {
	var got_task bool
	deploy := &Deploy{
		make(map[string][]interface{}),
		list.New(),
		&sync.WaitGroup{},
		&self.containers,
		&Nginx{
			*ngx, *dst,
			make(map[string]*Upstream),
		},
		self.client,
		registry,
	}
	for !self.finish {
		apptask := AppTask{}
		ws.SetReadDeadline(time.Now().Add(time.Duration(*wait) * time.Second))
		fmt.Println(time.Now())
		if got_task = self.Read(ws, &apptask); got_task {
			deploy.tasks.PushBack(apptask)
		}
		if (deploy.tasks.Len() != 0 && !got_task) || deploy.tasks.Len() >= *num {
			self.Process(ws, deploy)
			self.Load()
		}
	}
}

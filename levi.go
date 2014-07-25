package main

import (
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
	tasks      []AppTask
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

func (self *Levi) Clear() {
	self.tasks = []AppTask{}
}

func (self *Levi) Append(apptask *AppTask) {
	self.tasks = append(self.tasks, *apptask)
}

func (self *Levi) Process(ws *websocket.Conn, dst, ngx *string) {
	deploy := Deploy{
		make(map[string][]interface{}),
		&self.tasks,
		&sync.WaitGroup{},
		&self.containers,
		&Nginx{
			*ngx, *dst,
			make(map[string]*Upstream),
		},
		self.client,
	}
	deploy.Deploy()
	result := deploy.Result()
	if err := ws.WriteJSON(&result); err != nil {
		log.Fatal("Write Fail:", err)
	}
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

func (self *Levi) Loop(ws *websocket.Conn, wait, num *int, dst, ngx *string) {
	var got_task bool
	self.Clear()
	for !self.finish {
		apptask := AppTask{}
		ws.SetReadDeadline(time.Now().Add(time.Duration(*wait) * time.Second))
		fmt.Println(time.Now())
		if got_task = self.Read(ws, &apptask); got_task {
			self.Append(&apptask)
		}
		if (len(self.tasks) != 0 && !got_task) || len(self.tasks) >= *num {
			self.Process(ws, dst, ngx)
			self.Load()
			self.Clear()
		}
	}
}

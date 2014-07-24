package main

import (
	"fmt"
	"github.com/CMGS/websocket"
	"github.com/fsouza/go-dockerclient"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type container_info struct {
	id   string
	port string
}

type Levi struct {
	client     *docker.Client
	containers []docker.APIContainers
	info       map[string][]container_info
	tasks      []AppTask
	finish     bool
}

func (self *Levi) Connect(docker_url *string) {
	var err error
	self.client, err = docker.NewClient(*docker_url)
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
	self.info = make(map[string][]container_info)
	for _, container := range self.containers {
		split_names := strings.SplitN(strings.TrimLeft(container.Names[0], "/"), "_", 2)
		info := container_info{container.ID, split_names[1]}
		self.info[split_names[0]] = append(self.info[split_names[0]], info)
	}
}

func (self *Levi) clear() {
	self.tasks = []AppTask{}
}

func (self *Levi) appendTask(apptask *AppTask) {
	self.tasks = append(self.tasks, *apptask)
}

func (self *Levi) process() map[string][]int {
	deploy := Deploy{
		make(map[string][]int),
		&self.tasks,
		&sync.WaitGroup{},
	}
	deploy.Deploy()
	deploy.Wait()
	return deploy.Result()
}

func (self *Levi) read(ws *websocket.Conn, apptask *AppTask) bool {
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

func (self *Levi) Loop(ws *websocket.Conn, wait *int, num *int, dst_dir *string) {
	self.clear()
	for !self.finish {
		apptask := AppTask{}
		ws.SetReadDeadline(time.Now().Add(time.Duration(*wait) * time.Second))
		fmt.Println(time.Now())
		switch got_task := self.read(ws, &apptask); {
		case !got_task && len(self.tasks) != 0:
			result := self.process()
			if err := ws.WriteJSON(&result); err != nil {
				log.Fatal("Write Fail:", err)
			}
			self.clear()
		case got_task:
			self.appendTask(&apptask)
			if len(self.tasks) >= *num {
				result := self.process()
				if err := ws.WriteJSON(&result); err != nil {
					log.Fatal("Write Fail:", err)
				}
				self.clear()
			}
		}
	}
}

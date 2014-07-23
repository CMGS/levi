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
}

func (self *Levi) Connect(docker_url *string) {
	var err error
	self.client, err = docker.NewClient(*docker_url)
	if err != nil {
		log.Fatal("Connect docker failed")
	}
	self.containers, err = self.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatal("Query docker failed")
	}
}

func (self *Levi) Load() {
	self.info = make(map[string][]container_info)
	for _, container := range self.containers {
		split_names := strings.SplitN(strings.TrimLeft(container.Names[0], "/"), "_", 2)
		info := container_info{container.ID, split_names[1]}
		self.info[split_names[0]] = append(self.info[split_names[0]], info)
	}
	fmt.Println(self.info)
}

func (self *Levi) clear() {
	self.tasks = []AppTask{}
}

func (self *Levi) appendTask(apptask *AppTask) {
	self.tasks = append(self.tasks, *apptask)
}

func (self *Levi) process() {
	deploy := Deploy{
		make(map[string][]int),
		&self.tasks,
		&sync.WaitGroup{},
	}
	deploy.Deploy()
	deploy.Wait()
	fmt.Println(deploy.Result())
}

func (self *Levi) Loop(ws *websocket.Conn, sleep *int, num *int, dst_dir *string) {
	ws.SetPingHandler(nil)
	self.clear()
	for {
		got_task := false
		apptask := AppTask{}
		ws.SetReadDeadline(time.Now().Add(time.Duration(*sleep) * time.Second))
		fmt.Println(time.Now())
		if err := ws.ReadJSON(&apptask); err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				log.Fatal("Read Fail:", err)
			}
		} else {
			got_task = true
		}
		switch {
		case !got_task && len(self.tasks) != 0:
			self.process()
			self.clear()
		case got_task:
			self.appendTask(&apptask)
			if len(self.tasks) >= *num {
				self.process()
				self.clear()
			}
		}
	}
}

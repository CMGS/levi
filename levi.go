package main

import (
	"fmt"
	"github.com/CMGS/websocket"
	"github.com/fsouza/go-dockerclient"
	"log"
	"net"
	"strings"
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

func (self *Levi) Connect() {
	var err error
	self.client, err = docker.NewClient(LOCAL_DOCKER)
	if err != nil {
		log.Fatal("Connect docker failed")
	}
	self.containers, err = self.client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatal("Query docker failed")
	}
}

func (self *Levi) Parse() {
	self.info = make(map[string][]container_info)
	for _, container := range self.containers {
		split_names := strings.SplitN(strings.TrimLeft(container.Names[0], "/"), "_", 2)
		info := container_info{container.ID, split_names[1]}
		self.info[split_names[0]] = append(self.info[split_names[0]], info)
	}
	fmt.Println(self.info)
}

func (self *Levi) Clear() {
	self.tasks = []AppTask{}
}

func (self *Levi) AppendTask(apptask *AppTask) {
	self.tasks = append(self.tasks, *apptask)
}

func (self *Levi) Process() {
	h := Taskhub{
		tasks: make(map[string]bool),
		done:  make(chan string),
	}

	for _, task := range self.tasks {
		fmt.Println("Process", task.Name)
		h.Process(task)
	}

	for {
		if h.CheckDone() {
			break
		}
	}
}

func (self *Levi) Loop(ws *websocket.Conn, sleep *int, num *int) {
	ws.SetPingHandler(nil)
	self.Clear()
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
			self.Process()
			self.Clear()
		case got_task:
			self.AppendTask(&apptask)
			if len(self.tasks) >= *num {
				self.Process()
				self.Clear()
			}
		}
	}
}

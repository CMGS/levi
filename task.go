package main

import (
	"fmt"
)

type Task struct {
	Type       int
	Image      string
	Version    string
	Bind       int
	Port       int
	Container  string
	Entrypoint string
	Memory     int
	Cpus       float64
}

type AppTask struct {
	Id    string
	Name  string
	Tasks []Task
}

type Taskhub struct {
	tasks map[string]bool
	done  chan string
}

func (self *Taskhub) Process(task AppTask) {
	self.tasks[task.Id] = false
	go func() {
		//TODO add/remove/update container
		for _, job := range task.Tasks {
			fmt.Println("process", Methods[job.Type], task.Name)
		}
		self.done <- task.Id
	}()
}

func (self *Taskhub) CheckDone() bool {
	select {
	case tid := <-self.done:
		self.tasks[tid] = true
		fmt.Println(tid, "done")
	}
	for _, done := range self.tasks {
		if !done {
			return false
		}
	}
	return true
}

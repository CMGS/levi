package main

import (
	"fmt"
)

type Task struct {
	Version   string
	Bind      int64
	Port      int64
	Container string
	Cmd       []string
	Memory    float64
	Cpus      int64
	Config    interface{}
	Daemon    string
	ident     string
}

func (self *Task) IsDaemon() bool {
	return self.Daemon != ""
}

func (self *Task) CheckDaemon() bool {
	daemon_ident := fmt.Sprintf("daemon_%s", self.Daemon)
	return daemon_ident == self.ident
}

func (self *Task) SetAsDaemon() {
	self.ident = fmt.Sprintf("daemon_%s", self.Daemon)
}

func (self *Task) SetAsService() {
	self.ident = fmt.Sprintf("%d", self.Bind)
}

type AppTask struct {
	Id    string
	Uid   int
	Name  string
	Type  int
	Tasks []Task
}

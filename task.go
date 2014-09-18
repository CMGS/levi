package main

import (
	"fmt"
	"sync"
)

type BuildInfo struct {
	Group   string
	Name    string
	Version string
	Base    string
	Build   string
	Static  string
	Schema  string
}

type Task struct {
	Build     BuildInfo
	Version   string
	Bind      int64
	Port      int64
	Container string
	Cmd       []string
	Memory    int64
	CpuShares int64
	CpuSet    string
	Daemon    string
	Test      string
	ident     string
}

func (self *Task) IsDaemon() bool {
	return self.Daemon != ""
}

func (self *Task) IsTest() bool {
	return self.Test != ""
}

func (self *Task) ShouldExpose() bool {
	return self.Daemon == "" && self.Test == ""
}

func (self *Task) CheckTest() bool {
	test_ident := fmt.Sprintf("test_%s", self.Test)
	return test_ident == self.ident
}

func (self *Task) CheckDaemon() bool {
	daemon_ident := fmt.Sprintf("daemon_%s", self.Daemon)
	return daemon_ident == self.ident
}

func (self *Task) SetAsTest() {
	self.ident = fmt.Sprintf("test_%s", self.Test)
}

func (self *Task) SetAsDaemon() {
	self.ident = fmt.Sprintf("daemon_%s", self.Daemon)
}

func (self *Task) SetAsService() {
	self.ident = fmt.Sprintf("%d", self.Bind)
}

type AppTask struct {
	Id     string
	Uid    int
	Name   string
	Type   int
	Tasks  []Task
	wg     *sync.WaitGroup
	result map[string][]interface{}
}

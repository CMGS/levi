package main

import (
	"fmt"
	"sync"
)

type BuildTask struct {
	Group   string
	Name    string
	Version string
	Base    string
	Build   string
	Static  string
	Schema  string
}

type RemoveTask struct {
	Container string
	RmImage   bool
	RunEnv    string
}

func (self *RemoveTask) IsTest() bool {
	return self.RunEnv == TESTING
}

func (self *RemoveTask) IsRemoveImage() bool {
	return self.RmImage
}

type AddTask struct {
	Version   string
	Bind      int64
	Port      int64
	Cmd       []string
	Memory    int64
	CpuShares int64
	CpuSet    string
	Daemon    string
	Test      string
	ident     string
}

func (self *AddTask) IsDaemon() bool {
	return self.Daemon != ""
}

func (self *AddTask) IsTest() bool {
	return self.Test != ""
}

func (self *AddTask) ShouldExpose() bool {
	return self.Daemon == "" && self.Test == ""
}

func (self *AddTask) CheckTest() bool {
	test_ident := fmt.Sprintf("test_%s", self.Test)
	return test_ident == self.ident
}

func (self *AddTask) CheckDaemon() bool {
	daemon_ident := fmt.Sprintf("daemon_%s", self.Daemon)
	return daemon_ident == self.ident
}

func (self *AddTask) SetAsTest() {
	self.ident = fmt.Sprintf("test_%s", self.Test)
}

func (self *AddTask) SetAsDaemon() {
	self.ident = fmt.Sprintf("daemon_%s", self.Daemon)
}

func (self *AddTask) SetAsService() {
	self.ident = fmt.Sprintf("%d", self.Bind)
}

type Task struct {
	Build  *BuildTask
	Add    *AddTask
	Remove *RemoveTask
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

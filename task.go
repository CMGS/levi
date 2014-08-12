package main

import (
	"fmt"
	"strings"
)

type Ident string

func (self Ident) IsDaemon(appname string) bool {
	prefix := fmt.Sprintf("%s_daemon_", appname)
	return strings.HasPrefix(self.String(), prefix)
}

func (self Ident) String() string {
	return string(self)
}

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
	ident     Ident
}

type AppTask struct {
	Id    string
	Uid   int
	Name  string
	Type  int
	Tasks []Task
}

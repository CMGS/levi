package main

import (
	"fmt"
	"sync"
	"testing"
)

var apptask *AppTask

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Ws = NewWebSocket(config.Master)
	MockWebSocket(Ws)
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
	apptask = &AppTask{
		Id:   "abc",
		Uid:  4001,
		Name: "nbetest",
	}
	apptask.wg = &sync.WaitGroup{}
	apptask.Tasks = &Tasks{}
	apptask.result = &TaskResult{Id: apptask.Id}
}

func Test_SetAddTaskType(t *testing.T) {
	task := AddTask{
		Bind:   9999,
		Daemon: "abc",
		Test:   "def",
	}
	task.SetAsTest()
	if !task.IsTest() || task.ident != "test_def" {
		t.Error("Test ident invaild")
	}
	task.SetAsDaemon()
	if !task.IsDaemon() || task.ident != "daemon_abc" {
		t.Error("Daemon ident invaild")
	}
	task.Daemon = ""
	task.Test = ""
	task.SetAsService()
	if !task.ShouldExpose() || task.ident != "9999" {
		t.Error("Service ident invaild")
	}
}

func Test_TaskBuildImage(t *testing.T) {
	ver := "082d405"
	name := "nbetest"
	job := &BuildTask{
		Group:   "platform",
		Name:    name,
		Version: ver,
		Base:    fmt.Sprintf("%s/nbeimage/ubuntu:python-2014.9.30", config.Docker.Registry),
		Build:   "pip install -r requirements.txt",
		Static:  "static",
	}
	apptask.Tasks.Build = []*BuildTask{job}
	apptask.result.Build = make([]string, len(apptask.Tasks.Build))
	apptask.wg.Add(len(apptask.Tasks.Build))
	apptask.BuildImage(0)
	if len(apptask.result.Build) == 0 {
		t.Error("Wrong Result")
	}
	if apptask.result.Build[0] != fmt.Sprintf("%s/%s:%s", config.Docker.Registry, name, ver) {
		t.Error("Wrong Data")
	}
}

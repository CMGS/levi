package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
)

var apptask *AppTask

func init() {
	InitTest()
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
	apptask.result.Build = make([]string, 1)
	apptask.wg.Add(1)
	apptask.BuildImage(0)
	if len(apptask.result.Build) == 0 {
		t.Error("Wrong Result")
	}
	if apptask.result.Build[0] != fmt.Sprintf("%s/%s:%s", config.Docker.Registry, name, ver) {
		t.Error("Wrong Data")
	}
}

func Test_TaskRemoveContainer(t *testing.T) {
	id := "abcdefg"
	job := &RemoveTask{id, true}
	nginx := NewNginx()
	apptask.Tasks.Remove = []*RemoveTask{job}
	apptask.result.Remove = make([]bool, 1)
	apptask.wg.Add(1)
	apptask.RemoveContainer(0, nginx)
	if len(apptask.result.Remove) == 0 {
		t.Error("Wrong Result")
	}
	if apptask.result.Remove[0] {
		t.Error("Wrong Data")
	}
	Status.Removable[id] = struct{}{}
	apptask.wg.Add(1)
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		return &docker.Container{Volumes: map[string]string{}, ID: id}, nil
	}
	apptask.RemoveContainer(0, nginx)
	if len(apptask.result.Remove) == 0 {
		t.Error("Wrong Result")
	}
	if !apptask.result.Remove[0] {
		t.Error("Wrong Data")
	}
	if _, ok := Status.Removable[id]; ok {
		t.Error("Wrong Status")
	}
	Status.Removable[id] = struct{}{}
	apptask.result.Remove = make([]bool, 1)
	apptask.wg.Add(1)
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		return nil, errors.New("made by test")
	}
	apptask.RemoveContainer(0, nginx)
	if apptask.result.Remove[0] {
		t.Error("Wrong Data")
	}
	if _, ok := Status.Removable[id]; !ok {
		t.Error("Wrong Status")
	}
}

func Test_TaskAddContainer(t *testing.T) {
	cid := "12345"
	appname := "nbetest"
	ver := "v1"
	tid := "abcdefg"
	job := &AddTask{
		Version:   ver,
		Cmd:       []string{"python", "xxx.py"},
		Memory:    99999,
		CpuShares: 512,
		CpuSet:    "0",
		Test:      tid,
	}
	job.SetAsTest()
	cpath := GenerateConfigPath(appname, job.ident)
	dpath := GeneratePermdirPath(appname, job.ident, true)
	nginx := NewNginx()
	env := &Env{appname, 4011}
	apptask.Tasks.Add = []*AddTask{job}
	apptask.result.Add = make([]string, 1)
	apptask.wg.Add(1)
	Etcd.Get = func(string, bool, bool) (*etcd.Response, error) {
		ret := &etcd.Response{Node: &etcd.Node{Value: ""}}
		return ret, nil
	}
	Docker.CreateContainer = func(docker.CreateContainerOptions) (*docker.Container, error) {
		return &docker.Container{ID: cid}, nil
	}
	defer os.RemoveAll(cpath)
	defer os.RemoveAll(dpath)
	apptask.AddContainer(0, env, nginx)
	if len(apptask.result.Add) == 0 {
		t.Error("Wrong Result")
	}
	if apptask.result.Add[0] != cid {
		t.Error("Wrong Data")
	}
}

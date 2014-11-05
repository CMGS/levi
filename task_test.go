package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"./common"
	"./defines"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
)

var apptask *AppTask

func init() {
	InitTest()
	apptask = &AppTask{
		Id:    "abc",
		Uid:   4001,
		Name:  "nbetest",
		wg:    &sync.WaitGroup{},
		Tasks: &Tasks{},
	}
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
	common.Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*defines.Result)
		if !ok {
			t.Error("Wrong Data")
		}
		if x.Id != apptask.Id {
			t.Error("Wrong Id")
		}
		if x.Type != common.BUILD_TASK {
			t.Error("Wrong Type")
		}
		if x.Index != 0 {
			t.Error("Wrong Index")
		}
		return nil
	}
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
	apptask.wg.Add(1)
	apptask.BuildImage(0)
}

func Test_TaskRemoveContainer(t *testing.T) {
	common.Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*defines.Result)
		if !ok {
			t.Error("Wrong Data")
		}
		if x.Id != apptask.Id {
			t.Error("Wrong Id")
		}
		if x.Type != common.REMOVE_TASK {
			t.Error("Wrong Type")
		}
		if x.Index != 0 {
			t.Error("Wrong Index")
		}
		return nil
	}
	id := "abcdefg"
	job := &RemoveTask{id, true}
	nginx := NewNginx()
	apptask.Tasks.Remove = []*RemoveTask{job}
	apptask.wg.Add(1)
	apptask.RemoveContainer(0, nginx)
	Status.Removable[id] = struct{}{}
	apptask.wg.Add(1)
	common.Docker.InspectContainer = func(string) (*docker.Container, error) {
		return &docker.Container{Volumes: map[string]string{}, ID: id}, nil
	}
	apptask.RemoveContainer(0, nginx)
	if _, ok := Status.Removable[id]; ok {
		t.Error("Wrong Status")
	}
	Status.Removable[id] = struct{}{}
	apptask.wg.Add(1)
	common.Docker.InspectContainer = func(string) (*docker.Container, error) {
		return nil, errors.New("made by test")
	}
	apptask.RemoveContainer(0, nginx)
	if _, ok := Status.Removable[id]; !ok {
		t.Error("Wrong Status")
	}
}

func Test_TaskAddContainer(t *testing.T) {
	common.Docker.InspectContainer = func(string) (*docker.Container, error) {
		m := map[string]string{}
		return &docker.Container{Volumes: m}, nil
	}
	common.Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*defines.Result)
		if !ok {
			t.Error("Wrong Data")
		}
		if x.Id != apptask.Id {
			t.Error("Wrong Id")
		}
		if x.Type == common.ADD_TASK && x.Done == true {
			t.Error("Wrong Type")
		}
		if x.Type == common.TEST_TASK && x.Done == false {
			t.Error("Wrong Type")
		}
		if x.Index != 0 {
			t.Error("Wrong Index")
		}
		return nil
	}
	cid := "1234567890abcdefg"
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
	apptask.wg.Add(1)
	common.Etcd.Get = func(string, bool, bool) (*etcd.Response, error) {
		ret := &etcd.Response{Node: &etcd.Node{Value: ""}}
		return ret, nil
	}
	common.Docker.CreateContainer = func(docker.CreateContainerOptions) (*docker.Container, error) {
		return &docker.Container{ID: cid}, nil
	}
	defer os.RemoveAll(cpath)
	defer os.RemoveAll(dpath)
	apptask.AddContainer(0, env, nginx)
	if _, ok := Status.Removable[cid]; ok {
		t.Error("Wrong Container Flag")
	}
}

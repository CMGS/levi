package main

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func init() {
	InitTest()
}

func Test_GetName(t *testing.T) {
	var containerName string
	containerName = "test_1234"
	appname := Status.getName(containerName)
	if appname != "test" {
		t.Error("Get appname failed")
	}
}

func Test_GetStatus(t *testing.T) {
	a := "Exited (0) 9 days ago"
	if Status.getStatus(a) != STATUS_DIE {
		t.Error("Wrong Status")
	}
	a = "Up 8 days"
	if Status.getStatus(a) != STATUS_START {
		t.Error("Wrong Status")
	}
}

func Test_StatusReport(t *testing.T) {
	id := "xxxxxxxxxxxx"
	Docker.ListContainers = func(opt docker.ListContainersOptions) ([]docker.APIContainers, error) {
		c1 := docker.APIContainers{
			Names:  []string{"/test_1234"},
			ID:     id,
			Image:  config.Docker.Registry,
			Status: "Exited (0) 9 days ago",
		}
		c := []docker.APIContainers{c1}
		return c, nil
	}
	tid := "zzzzzzzzzzzz"
	Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*TaskResult)
		if !ok {
			t.Error("Wrong Data")
		}
		if x.Id != tid {
			t.Error("Wrong Task ID")
		}
		if len(x.Status) == 0 {
			t.Error("Wrong Status")
		}
		i := x.Status[0]
		if i.Appname != "test" {
			t.Error("Wrong Appname")
		}
		if i.Id != id {
			t.Error("Wrong Id")
		}
		if i.Type != STATUS_DIE {
			t.Error("Wrong Status")
		}
		return nil
	}
	Status.Report(tid)
	if _, ok := Status.Removable[id]; !ok {
		t.Error("Wrong Data")
	}
}

func Test_StatusDie(t *testing.T) {
	id := "xxx"
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		return &docker.Container{Name: "/test_1234"}, nil
	}
	Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*TaskResult)
		if !ok {
			t.Error("Wrong Data")
		}
		if len(x.Status) == 0 {
			t.Error("Wrong Status")
		}
		i := x.Status[0]
		if i.Appname != "test" {
			t.Error("Wrong Appname")
		}
		if i.Id != id {
			t.Error("Wrong Id")
		}
		if i.Type != STATUS_DIE {
			t.Error("Wrong Status")
		}
		return nil
	}
	Status.die(id)
}

func Test_StatusListen(t *testing.T) {
	go Status.Listen()
	id := "abc"
	event := &docker.APIEvents{"die", id, "zzz", 12345}
	Status.events <- event
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		t.Error("Wrong event")
		return nil, nil
	}
	Status.Removable[id] = struct{}{}
	Docker.InspectContainer = func(i string) (*docker.Container, error) {
		if i != id {
			t.Error("Wrong event")
		}
		return &docker.Container{ID: id, Name: "/test_1234"}, nil
	}
	Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*TaskResult)
		if !ok {
			t.Error("Wrong Data")
		}
		if len(x.Status) == 0 {
			t.Error("Wrong Status")
		}
		i := x.Status[0]
		if i.Appname != "test" {
			t.Error("Wrong Appname")
		}
		if i.Id != id {
			t.Error("Wrong Id")
		}
		if i.Type != STATUS_DIE {
			t.Error("Wrong Status")
		}
		return nil
	}
	Status.events <- event
}

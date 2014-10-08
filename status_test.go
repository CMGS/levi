package main

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Ws = NewWebSocket(config.Master)
	MockWebSocket(Ws)
	Status = NewStatus()
}

func Test_GetName(t *testing.T) {
	var containerName string
	containerName = "test_1234"
	appname := Status.getName(containerName)
	if appname != "test" {
		t.Fatal("Get appname failed")
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
	id := "xxx"
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
	tid := "zz"
	Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*TaskResult)
		if !ok {
			t.Fatal("Wrong Data")
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
}

func Test_StatusDie(t *testing.T) {
	id := "xxx"
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		return &docker.Container{Name: "/test_1234"}, nil
	}
	Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*TaskResult)
		if !ok {
			t.Fatal("Wrong Data")
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

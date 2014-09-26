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

func Test_StatusDie(t *testing.T) {
	id := "xxx"
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		return &docker.Container{Name: "/test_1234"}, nil
	}
	Ws.WriteJSON = func(d interface{}) error {
		if x, ok := d.(map[string][]*Info); !ok {
			i := x[STATUS_IDENT][0]
			t.Fatal("Wrong Data")
			if i.Type != STATUS_DIE {
				t.Error("Wrong Status")
			}
			if i.Appname != "test" {
				t.Error("Wrong Appname")
			}
			if i.Id != id {
				t.Error("Wrong Id")
			}
		}
		return nil
	}
	Status.die(id)
}

package main

import (
	"errors"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Ws = NewWebSocket(config.Master)
	MockWebSocket(Ws)
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
}

func Test_WaitForTester(t *testing.T) {
	tester := Tester{"xxxxxx", map[string]string{}}
	tester.cids["abc"] = "def"
	Docker.WaitContainer = func(id string) (int, error) {
		return 0, errors.New(id)
	}
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		m := map[string]string{}
		return &docker.Container{Volumes: m}, nil
	}
	Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*TaskResult)
		if !ok {
			t.Fatal("Wrong Data")
		}
		if len(x.Test) == 0 {
			t.Fatal("Wrong Data")
		}
		r := x.Test["abc"]
		if r.ExitCode != 0 {
			t.Fatal("Wrong Exit Code")
		}
		if r.Err != "def" {
			t.Fatal("Wrong ErrStr")
		}
		return nil
	}
	tester.WaitForTester()
}

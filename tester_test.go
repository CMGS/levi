package main

import (
	"errors"
	"strconv"
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
	tester := Tester{"test", "abc", map[string]string{}}
	tester.cids["aa"] = "1"
	tester.cids["bb"] = "2"
	tester.cids["cc"] = "3"
	Docker.WaitContainer = func(id string) (int, error) {
		return 0, errors.New(id)
	}
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		m := map[string]string{}
		return &docker.Container{Volumes: m}, nil
	}
	Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(map[string][]*Result)
		if !ok {
			t.Fatal("Wrong Data")
		}
		y, ok := x["abc"]
		if !ok {
			t.Fatal("Wrong Map")
		}
		for i, z := range y {
			if z.ExitCode != 0 {
				t.Fatal("Wrong RetCode")
			}
			if n, err := strconv.Atoi(z.Err); err != nil || n != i+1 {
				t.Fatal("Parser Invaild")
			}
		}
		return nil
	}
	tester.WaitForTester()
}

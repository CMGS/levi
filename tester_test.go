package main

import (
	"errors"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func init() {
	InitTest()
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
			t.Error("Wrong Data")
		}
		if len(x.Test) == 0 {
			t.Error("Wrong Data")
		}
		r := x.Test["abc"]
		if r.ExitCode != 0 {
			t.Error("Wrong Exit Code")
		}
		if r.Err != "def" {
			t.Error("Wrong ErrStr")
		}
		return nil
	}
	tester.WaitForTester()
}

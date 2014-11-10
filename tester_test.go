package main

import (
	"errors"
	"testing"

	"./common"
	"./defines"
	"github.com/fsouza/go-dockerclient"
)

var tester *Tester

func init() {
	InitTest()
	tester = &Tester{
		"xxxxxx",
		"zzzzzzzzzzzz",
		"test",
		"abcdefg",
		0,
		0,
	}
}

func Test_TesterGetLogs(t *testing.T) {
	// Not Implement Yet
}

func Test_TesterWait(t *testing.T) {
	errData := "abc"
	common.Docker.WaitContainer = func(id string) (int, error) {
		return 0, errors.New(errData)
	}
	common.Docker.InspectContainer = func(string) (*docker.Container, error) {
		m := map[string]string{}
		return &docker.Container{Volumes: m}, nil
	}
	common.Ws.WriteJSON = func(d interface{}) error {
		x, ok := d.(*defines.Result)
		if !ok {
			t.Error("Wrong Data")
		}
		if x.Id != "xxxxxx" {
			t.Error("Wrong Id")
		}
		if x.Done != true {
			t.Error("Wrong Done")
		}
		if x.Index != 0 {
			t.Error("Wrong Index")
		}
		if x.Type != common.TEST_TASK {
			t.Error("Wrong Type")
		}
		if x.Data != errData {
			t.Error("Wrong Data")
		}
		return nil
	}
	tester.Wait()
}

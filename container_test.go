package main

import (
	"os"
	"path"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

func init() {
	InitTest()
}

func Test_Stop(t *testing.T) {
	container := Container{"abcdefg", "test"}
	if err := container.Stop(); err != nil {
		t.Fatal(err)
	}
}

func Test_RemoveContainer(t *testing.T) {
	ppath := path.Join(config.App.Home, "d1")
	cpath := path.Join(config.App.Home, "t1")
	image := "testimage"
	Docker.InspectContainer = func(string) (*docker.Container, error) {
		m := map[string]string{}
		m["/test/config.yaml"] = cpath
		m["/test/permdir"] = ppath
		return &docker.Container{Volumes: m, Image: image}, nil
	}
	f, err := os.Create(cpath)
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(ppath, 0755)
	f.WriteString("test")
	f.Sync()
	f.Close()
	Docker.RemoveImage = func(p string) error {
		if p != image {
			t.Error("Wrong Image")
		}
		return nil
	}
	if err := RemoveContainer("abcdefg", true, true); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(cpath); err == nil {
		t.Error("Not clean")
	}
	if _, err := os.Stat(ppath); err == nil {
		t.Error("Not clean")
	}
}

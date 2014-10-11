package main

import (
	"testing"

	. "./defines"
	"github.com/fsouza/go-dockerclient"
)

func Test_MockDocker(t *testing.T) {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	err := Docker.PushImage(docker.PushImageOptions{}, docker.AuthConfiguration{})
	if err != nil {
		t.Error(err)
	}
}

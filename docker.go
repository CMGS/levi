package main

import (
	"github.com/fsouza/go-dockerclient"
	"reflect"
)

type DockerWrapper struct {
	*docker.Client
	PushImage       func(docker.PushImageOptions, docker.AuthConfiguration) error
	PullImage       func(docker.PullImageOptions, docker.AuthConfiguration) error
	CreateContainer func(docker.CreateContainerOptions) (*docker.Container, error)
	StartContainer  func(string, *docker.HostConfig) error
	BuildImage      func(docker.BuildImageOptions) error
}

func NewDocker(endpoint string) *DockerWrapper {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		logger.Assert(err, "Docker")
	}
	d := &DockerWrapper{Client: client}
	v := reflect.ValueOf(d).Elem()
	vt := v.Type()
	for i := 1; i < reflect.TypeOf(*d).NumField(); i++ {
		field := v.Field(i)
		f := reflect.ValueOf(d.Client).MethodByName(vt.Field(i).Name)
		field.Set(f)
	}
	return d
}

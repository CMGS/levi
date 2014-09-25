package main

import (
	"github.com/fsouza/go-dockerclient"
	"reflect"
)

type DockerWrapper struct {
	*docker.Client
	PushImage  func(docker.PushImageOptions, docker.AuthConfiguration) error
	BuildImage func(docker.BuildImageOptions) error
}

func NewDocker(endpoint string) *DockerWrapper {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		logger.Assert(err, "Docker")
	}
	d := &DockerWrapper{Client: client}
	v := reflect.ValueOf(d).Elem()
	for i := 1; i < reflect.TypeOf(*d).NumField(); i++ {
		field := v.Field(i)
		f := reflect.ValueOf(d.Client).MethodByName(v.Type().Field(i).Name)
		field.Set(f)
	}
	return d
}

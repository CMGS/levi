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

func NewDocker(endpoint string, mock bool) *DockerWrapper {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		logger.Assert(err, "Docker")
	}
	d := &DockerWrapper{Client: client}
	if !mock {
		return originDocker(d)
	}
	return mockDocker(d)
}

func originDocker(d *DockerWrapper) *DockerWrapper {
	v := reflect.ValueOf(d).Elem()
	for i := 1; i < reflect.TypeOf(*d).NumField(); i++ {
		field := v.Field(i)
		f := reflect.ValueOf(d.Client).MethodByName(v.Type().Field(i).Name)
		field.Set(f)
	}
	return d
}

func mockDocker(d *DockerWrapper) *DockerWrapper {
	errorType := reflect.TypeOf(make([]error, 1)).Elem()
	v := reflect.ValueOf(d).Elem()
	for i := 1; i < reflect.TypeOf(*d).NumField(); i++ {
		field := v.Field(i)
		f := reflect.MakeFunc(field.Type(), func(in []reflect.Value) []reflect.Value {
			return []reflect.Value{reflect.Zero(errorType)}
		})
		field.Set(f)
	}
	return d
}

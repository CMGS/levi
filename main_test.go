package main

import (
	"github.com/fsouza/go-dockerclient"
	"reflect"
)

func MockDocker(d *DockerWrapper) {
	var makeMockedDockerWrapper func(*DockerWrapper, *docker.Client) *DockerWrapper
	MakeMockedWrapper(&makeMockedDockerWrapper)
	makeMockedDockerWrapper(d, d.Client)
}

func MakeMockedWrapper(fptr interface{}) {
	var maker = func(in []reflect.Value) []reflect.Value {
		wrapper := in[0].Elem()
		client := in[1]
		wrapperType := wrapper.Type()
		for i := 1; i < wrapperType.NumField(); i++ {
			field := wrapper.Field(i)
			fd, ok := client.Type().MethodByName(wrapperType.Field(i).Name)
			if !ok {
				logger.Info("Reflect Failed")
				continue
			}
			fdt := fd.Type
			ret := make([]reflect.Value, 0, fdt.NumOut())
			f := reflect.MakeFunc(field.Type(), func(in []reflect.Value) []reflect.Value {
				for i := 0; i < fdt.NumOut(); i++ {
					ret = append(ret, reflect.Zero(fdt.Out(i)))
				}
				return ret
			})
			field.Set(f)
		}
		return []reflect.Value{in[0]}
	}
	fn := reflect.ValueOf(fptr).Elem()
	v := reflect.MakeFunc(fn.Type(), maker)
	fn.Set(v)
}

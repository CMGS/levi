package main

import (
	"reflect"
)

func MockDocker(d *DockerWrapper) {
	v := reflect.ValueOf(d).Elem()
	vt := v.Type()
	for i := 1; i < reflect.TypeOf(*d).NumField(); i++ {
		field := v.Field(i)
		fd, ok := reflect.TypeOf(d.Client).MethodByName(vt.Field(i).Name)
		if !ok {
			logger.Info("Reflect Failed")
		}
		f := reflect.MakeFunc(field.Type(), func(in []reflect.Value) []reflect.Value {
			ret := []reflect.Value{}
			for i := 0; i < fd.Type.NumOut(); i++ {
				ret = append(ret, reflect.Zero(fd.Type.Out(i)))
			}
			return ret
		})
		field.Set(f)
	}
}

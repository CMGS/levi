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
}

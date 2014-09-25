package main

import (
	"reflect"
)

func MockDocker(d *DockerWrapper) {
	errorType := reflect.TypeOf(make([]error, 1)).Elem()
	v := reflect.ValueOf(d).Elem()
	for i := 1; i < reflect.TypeOf(*d).NumField(); i++ {
		field := v.Field(i)
		f := reflect.MakeFunc(field.Type(), func(in []reflect.Value) []reflect.Value {
			return []reflect.Value{reflect.Zero(errorType)}
		})
		field.Set(f)
	}
}

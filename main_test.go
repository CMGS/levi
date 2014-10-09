package main

import (
	"reflect"

	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
)

func InitTest() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Ws = NewWebSocket(config.Master)
	MockWebSocket(Ws)
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
	Status = NewStatus()
}

func MockDocker(d *DockerWrapper) {
	var makeMockedDockerWrapper func(*DockerWrapper, *docker.Client) *DockerWrapper
	MakeMockedWrapper(&makeMockedDockerWrapper)
	makeMockedDockerWrapper(d, d.Client)
}

func MockEtcd(e *EtcdWrapper) {
	var makeMockedEtcdWrapper func(*EtcdWrapper, *etcd.Client) *EtcdWrapper
	MakeMockedWrapper(&makeMockedEtcdWrapper)
	makeMockedEtcdWrapper(e, e.Client)
}

func MockWebSocket(w *WebSocketWrapper) {
	var makeMockedWebSocketWrapper func(*WebSocketWrapper, *websocket.Conn) *WebSocketWrapper
	MakeMockedWrapper(&makeMockedWebSocketWrapper)
	makeMockedWebSocketWrapper(w, w.Conn)
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
			f := reflect.MakeFunc(field.Type(), func(in []reflect.Value) []reflect.Value {
				ret := make([]reflect.Value, 0, fdt.NumOut())
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

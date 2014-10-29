package main

import (
	"reflect"

	"./defines"
	"./lenz"
	"./logs"
	"./metrics"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
)

func InitTest() {
	load("levi.yaml")
	Docker = defines.NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Ws = defines.NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	MockWebSocket(Ws)
	Etcd = defines.NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	MockEtcd(Etcd)
	if Status == nil {
		Status = NewStatus()
	}
	if Lenz == nil {
		Lenz = lenz.NewLenz(Docker, config.Lenz)
	}
	if Metrics == nil {
		Metrics = metrics.NewMetricsRecorder(config.HostName, config.Metrics)
	}
}

func MockDocker(d *defines.DockerWrapper) {
	var makeMockedDockerWrapper func(*defines.DockerWrapper, *docker.Client) *defines.DockerWrapper
	MakeMockedWrapper(&makeMockedDockerWrapper)
	makeMockedDockerWrapper(d, d.Client)
}

func MockEtcd(e *defines.EtcdWrapper) {
	var makeMockedEtcdWrapper func(*defines.EtcdWrapper, *etcd.Client) *defines.EtcdWrapper
	MakeMockedWrapper(&makeMockedEtcdWrapper)
	makeMockedEtcdWrapper(e, e.Client)
}

func MockWebSocket(w *defines.WebSocketWrapper) {
	var makeMockedWebSocketWrapper func(*defines.WebSocketWrapper, *websocket.Conn) *defines.WebSocketWrapper
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
				logs.Info("Reflect Failed")
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

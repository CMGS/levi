package lenz

import (
	"fmt"
	"io"
	"os"
	"strings"

	"../common"
	"../defines"
	"../logs"
)

type Writer interface {
	io.Writer
	Close()
}

type ForwardOutput struct {
	result   *defines.Result
	name     string
	version  string
	typ      string
	routes   []*defines.Route
	channels []chan *defines.Log
}

func NewForwardOutput(result *defines.Result, name, version, typ string, stdout bool, routes []*defines.Route) *ForwardOutput {
	o := &ForwardOutput{result: result, name: name, version: version, typ: typ}
	o.routes = routes
	o.channels = make([]chan *defines.Log, len(routes))
	for i, route := range routes {
		o.channels[i] = make(chan *defines.Log)
		go Streamer(route, o.channels[i], stdout)
	}
	return o
}

func (self ForwardOutput) Write(p []byte) (n int, err error) {
	data := fmt.Sprintf("%s", p)
	data = strings.TrimSuffix(data, "\n")
	data = strings.TrimSuffix(data, "\r")
	self.send(data)
	self.report(data)
	return len(p), nil
}

func (self ForwardOutput) report(data string) {
	self.result.Data = data
	if err := common.Ws.WriteJSON(self.result); err != nil {
		logs.Info(err, self.result)
	}
}

func (self ForwardOutput) send(data string) {
	for _, chann := range self.channels {
		o := &defines.Log{
			Data:    data,
			ID:      self.version,
			Name:    self.name,
			AppID:   "",
			AppType: self.typ,
			Type:    "stdout",
		}
		chann <- o
	}
}

func (self ForwardOutput) Close() {
	for _, chann := range self.channels {
		close(chann)
	}
}

func GetBuffer(Lenz *LenzForwarder, result *defines.Result, name, version, typ string, stdout bool) Writer {
	routes, err := Lenz.Router.GetAll()
	if err != nil {
		return Stdout{}
	}
	return NewForwardOutput(result, name, version, typ, stdout, routes)
}

type Stdout struct{}

func (self Stdout) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (self Stdout) Close() {
}

func GetDevBuffer(stdout bool) io.Writer {
	if stdout {
		return os.Stdout
	} else {
		return DevBuffer{}
	}
}

type DevBuffer struct {
}

func (self DevBuffer) Write(p []byte) (n int, err error) {
	return len(p), nil
}
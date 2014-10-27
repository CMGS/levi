package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"./defines"
	"./lenz"
)

type Writer interface {
	io.Writer
	Close()
}

type ForwardOutput struct {
	name     string
	version  string
	typ      string
	routes   []*defines.Route
	channels []chan *defines.Log
}

func NewForwardOutput(name, version, typ string, routes []*defines.Route) *ForwardOutput {
	o := &ForwardOutput{name: name, version: version, typ: typ}
	o.routes = routes
	o.channels = make([]chan *defines.Log, len(routes))
	for i, route := range routes {
		o.channels[i] = make(chan *defines.Log)
		go lenz.Streamer(route, o.channels[i], config.Lenz.Stdout)
	}
	return o
}

func (self ForwardOutput) Write(p []byte) (n int, err error) {
	data := fmt.Sprintf("%s", p)
	data = strings.TrimSuffix(data, "\n")
	data = strings.TrimSuffix(data, "\r")
	self.send(data)
	return len(p), nil
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

func GetBuffer(name, version, typ string) Writer {
	routes, err := Lenz.Router.GetAll()
	if err != nil {
		return Stdout{}
	}
	return NewForwardOutput(name, version, typ, routes)
}

type Stdout struct{}

func (self Stdout) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

func (self Stdout) Close() {
}

func GetDevBuffer() io.Writer {
	if config.Lenz.Stdout {
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

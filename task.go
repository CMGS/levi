package main

type Task struct {
	Version    string
	Bind       int
	Port       int
	Container  string
	Entrypoint string
	Memory     int
	Cpus       float64
	Config     interface{}
}

type AppTask struct {
	Id    string
	Uid   int
	Name  string
	Type  int
	Tasks []Task
}

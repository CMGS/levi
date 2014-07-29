package main

type Task struct {
	Version    string
	Bind       int
	Port       int
	Container  string
	Entrypoint string
	Memory     uint64
	Cpus       int64
	Config     interface{}
}

type AppTask struct {
	Id    string
	User  string
	Uid   int
	Name  string
	Type  int
	Tasks []Task
}

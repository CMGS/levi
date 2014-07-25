package main

type Task struct {
	Image      string
	Version    string
	Bind       int
	Port       int
	Container  string
	Entrypoint string
	Memory     int
	Cpus       float64
}

type AppTask struct {
	Id    string
	Uid   int
	Name  string
	Type  int
	Tasks []Task
}

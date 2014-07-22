package main

type Task struct {
	Type       int
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
	Name  string
	Tasks []Task
}

type Taskhub struct {
	task map[string]bool
	done chan string
}

var Type map[int]string = map[int]string{
	1: "ADD",
	2: "REMOVE",
	3: "UPDATE",
}

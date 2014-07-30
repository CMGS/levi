package main

type Task struct {
	Version   string
	Bind      int64
	Port      int64
	Container string
	Cmd       []string
	Memory    float64
	Cpus      int64
	Config    interface{}
}

type AppTask struct {
	Id    string
	Uid   int
	Name  string
	Type  int
	Tasks []Task
}

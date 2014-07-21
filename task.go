package main

type Info struct {
	Name      string
	Image     string
	Version   string
	Bind      int
	Port      int
	Container string
}

type Config struct {
	Entrypoint string
	Memory     int
	Cpus       float64
}

type Message struct {
	Id               string
	Type             string
	App_info         Info
	Container_config Config
}

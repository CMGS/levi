package main

import (
	"flag"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
)

type GitConfig struct {
	Endpoint  string
	WorkDir   string
	ExtendDir string
}

type NginxConfig struct {
	Configs    string
	Template   string
	DyUpstream string
}

type DockerConfig struct {
	Endpoint string
	Registry string
	Network  string
}

type AppConfig struct {
	Home     string
	Permdirs string
}

type EtcdConfig struct {
	Sync     bool
	Machines []string
}

type LeviConfig struct {
	Name            string
	Master          string
	PidFile         string
	TaskNum         int
	TaskInterval    int
	ReadBufferSize  int
	WriteBufferSize int

	Git    GitConfig
	Nginx  NginxConfig
	Docker DockerConfig
	App    AppConfig
	Etcd   EtcdConfig
}

var config = LeviConfig{}

func LoadConfig() {
	var configPath string
	flag.BoolVar(&logger.Mode, "DEBUG", false, "enable debug")
	flag.StringVar(&configPath, "c", "levi.yaml", "config file")
	flag.Parse()

	if _, err := os.Stat(configPath); err != nil {
		logger.Assert(err, "config file invaild")
	}

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		logger.Assert(err, "Read config file failed")
	}

	if err := yaml.Unmarshal(b, &config); err != nil {
		logger.Assert(err, "Load config file failed")
	}
	logger.Debug(config)
}

package main

import (
	"flag"
	"io/ioutil"
	"os"

	. "./utils"
	"gopkg.in/yaml.v1"
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
	Tmpdirs  string
	Permdirs string
}

type EtcdConfig struct {
	Sync     bool
	Machines []string
}

type LenzConfig struct {
	Routes   string
	Forwards []string
	Stdout   bool
}

type MetricsConfig struct {
	ReportInterval int
	Statsd         string
	Rate           float32
}

type LeviConfig struct {
	Name            string
	Master          string
	PidFile         string
	TaskNum         int
	TaskInterval    int
	ReadBufferSize  int
	WriteBufferSize int

	Git     GitConfig
	Nginx   NginxConfig
	Docker  DockerConfig
	App     AppConfig
	Etcd    EtcdConfig
	Lenz    LenzConfig
	Metrics MetricsConfig
}

var config = LeviConfig{}

func LoadConfig() {
	var configPath string
	flag.BoolVar(&Logger.Mode, "DEBUG", false, "enable debug")
	flag.StringVar(&configPath, "c", "levi.yaml", "config file")
	flag.Parse()
	load(configPath)
}

func load(configPath string) {
	if _, err := os.Stat(configPath); err != nil {
		Logger.Assert(err, "config file invaild")
	}

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		Logger.Assert(err, "Read config file failed")
	}

	if err := yaml.Unmarshal(b, &config); err != nil {
		Logger.Assert(err, "Load config file failed")
	}
	Logger.Debug(config)
}

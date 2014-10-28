package main

import (
	"flag"
	"io/ioutil"
	"os"

	"./defines"
	"./logs"
	"gopkg.in/yaml.v1"
)

var config = defines.LeviConfig{}

func LoadConfig() {
	var configPath string
	flag.BoolVar(&logs.Mode, "DEBUG", false, "enable debug")
	flag.StringVar(&configPath, "c", "levi.yaml", "config file")
	flag.Parse()
	load(configPath)
}

func load(configPath string) {
	if _, err := os.Stat(configPath); err != nil {
		logs.Assert(err, "config file invaild")
	}

	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		logs.Assert(err, "Read config file failed")
	}

	if err := yaml.Unmarshal(b, &config); err != nil {
		logs.Assert(err, "Load config file failed")
	}
	logs.Debug(config)
}

package main

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"path"
)

type Env struct {
	appname string
	appuid  int
}

func GenerateConfigPath(appname string, ident string) string {
	filename := fmt.Sprintf("%s_%s.yaml", appname, ident)
	filepath := path.Join(config.App.Home, appname, filename)
	return filepath
}

func (self *Env) CreateConfigFile(job *Task) error {
	configPath := GenerateConfigPath(self.appname, job.ident)

	resp, err := Etcd.Get(path.Join("/NBE", self.appname, job.Version, "config.yaml"), false, false)
	if err != nil {
		return err
	}

	if resp.Node.Value == "" {
		ret := self.SaveFile(configPath, []byte{})
		return ret
	}

	config := map[string]interface{}{}
	yaml.Unmarshal([]byte(resp.Node.Value), &config)

	if len(config) == 0 {
		ret := self.SaveFile(configPath, []byte{})
		return ret
	}

	out, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := self.SaveFile(configPath, out); err != nil {
		return err
	}
	return nil
}

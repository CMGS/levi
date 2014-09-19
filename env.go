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

func GenerateConfigPath(appname, ident string) string {
	filename := fmt.Sprintf("%s_%s.yaml", appname, ident)
	filepath := path.Join(config.App.Home, appname, filename)
	return filepath
}

func GeneratePermdirPath(appname, ident string, test bool) string {
	if !test {
		return path.Join(config.App.Permdirs, appname)
	}
	name := fmt.Sprintf("%s_%s", appname, ident)
	return path.Join(config.App.Tmpdirs, name)
}

func (self *Env) CreateConfigFile(job *Task) error {
	return self.createConfigFile(job, "config.yaml")
}

func (self *Env) CreateTestConfigFile(job *Task) error {
	return self.createConfigFile(job, "test.yaml")
}

func (self *Env) createConfigFile(job *Task, filename string) error {
	configPath := GenerateConfigPath(self.appname, job.ident)

	resp, err := Etcd.Get(path.Join("/NBE", self.appname, job.Version, filename), false, false)
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

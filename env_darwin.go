package main

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"path"
)

func GenerateConfigPath(appname string, ident string) string {
	filename := fmt.Sprintf("%s_%s.yaml", appname, ident)
	filepath := path.Join(HomePath, appname, filename)
	return filepath
}

type Env struct {
	appname string
	appuid  int
}

func (self *Env) CreateUser() {
	logger.Info("OSX have no useradd command.")
}

func (self *Env) SaveFile(configPath string, out []byte) error {
	if err := ioutil.WriteFile(configPath, out, 0600); err != nil {
		logger.Info("Save app config failed", err)
		return err
	}
	logger.Info("OSX can not set app owner")
	return nil
}

func (self *Env) CreateConfigFile(job *Task) error {
	configPath := GenerateConfigPath(self.appname, job.ident)
	if job.Config == nil {
		ret := self.SaveFile(configPath, []byte{})
		return ret
	}
	if c, ok := job.Config.(map[string]interface{}); !ok || len(c) == 0 {
		ret := self.SaveFile(configPath, []byte{})
		return ret
	}
	out, err := yaml.Marshal(job.Config)
	if err != nil {
		logger.Info("Get app config failed", err)
		return err
	}
	if err := self.SaveFile(configPath, out); err != nil {
		return err
	}
	return nil
}

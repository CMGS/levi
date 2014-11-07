package main

import (
	"io/ioutil"

	"./defines"
	"./logs"
	"./utils"
)

func (self *Env) CreateUser() {
	logs.Info("OSX have no useradd command.")
}

func (self *Env) SaveFile(configPath string, out []byte) error {
	if err := ioutil.WriteFile(configPath, out, 0600); err != nil {
		return err
	}
	logs.Info("OSX will not chown config file.")
	return nil
}

func (self *Env) CreatePermdir(job *defines.AddTask) error {
	permdir := GeneratePermdirPath(self.appname, job.Ident, job.IsTest())
	if err := utils.MakeDir(permdir); err != nil {
		return err
	}
	logs.Info("OSX will not chown permdir.")
	return nil
}

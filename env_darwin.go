package main

import (
	"io/ioutil"
)

func (self *Env) CreateUser() {
	logger.Info("OSX have no useradd command.")
}

func (self *Env) SaveFile(configPath string, out []byte) error {
	if err := ioutil.WriteFile(configPath, out, 0600); err != nil {
		return err
	}
	logger.Info("OSX will not chown config file.")
	return nil
}

func (self *Env) CreatePermdir(job *AddTask, test bool) error {
	permdir := GeneratePermdirPath(self.appname, job.ident, test)
	if err := MakeDir(permdir); err != nil {
		return err
	}
	logger.Info("OSX will not chown permdir.")
	return nil
}

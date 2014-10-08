package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
)

func (self *Env) CheckUser() bool {
	if _, err := user.LookupId(strconv.Itoa(self.appuid)); err != nil {
		return false
	}
	return true
}

func (self *Env) CreateUser() {
	if self.CheckUser() {
		logger.Info("User", self.appname, "exist")
		return
	}
	cmd := exec.Command(
		"useradd", self.appname, "-d",
		path.Join(config.App.Home, self.appname),
		"-m", "-s", "/sbin/nologin", "-u",
		strconv.Itoa(self.appuid),
	)
	err := cmd.Run()
	if err != nil {
		logger.Info(err)
	}
}

func (self *Env) SaveFile(configPath string, out []byte) error {
	if err := ioutil.WriteFile(configPath, out, 0600); err != nil {
		return err
	}
	return os.Chown(configPath, self.appuid, self.appuid)
}

func (self *Env) CreatePermdir(job *AddTask) error {
	permdir := GeneratePermdirPath(self.appname, job.ident, job.IsTest())
	if err := MakeDir(permdir); err != nil {
		return err
	}
	return os.Chown(permdir, self.appuid, self.appuid)
}

package main

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"levi/logger"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strconv"
)

func GenerateConfigPath(appname string, apport int64) string {
	file_name := fmt.Sprintf("%s_%d.yaml", appname, apport)
	file_path := path.Join(home_path, appname, file_name)
	return file_path
}

type Env struct {
	appname string
	appuid  int
}

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
		path.Join(home_path, self.appname),
		"-m", "-s", "/sbin/nologin", "-u",
		strconv.Itoa(self.appuid),
	)
	err := cmd.Run()
	if err != nil {
		logger.Info(err)
	}
}

func (self *Env) CreateConfigFile(job *Task) error {
	file_path := GenerateConfigPath(self.appname, job.Bind)
	out, err := yaml.Marshal(job.Config)
	if err != nil {
		logger.Info("Get app config failed", err)
		return err
	}
	if err := ioutil.WriteFile(file_path, out, 0600); err != nil {
		logger.Info("Save app config failed", err)
		return err
	}
	if err := os.Chown(file_path, self.appuid, self.appuid); err != nil {
		logger.Info("Set owner as app failed", err)
		return err
	}
	return nil
}

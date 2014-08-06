package main

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
)

func GenerateConfigPath(appname string, apport int64) string {
	file_name := fmt.Sprintf("%s_%d.yaml", appname, apport)
	file_path := path.Join(DEFAULT_HOME_PATH, appname, file_name)
	return file_path
}

type Env struct {
	appname string
	appuid  int
}

func (self *Env) CreateUser() {
	cmd := exec.Command(
		"useradd", self.appname, "-d",
		path.Join(DEFAULT_HOME_PATH, self.appname),
		"-m", "-s", "/sbin/nologin", "-u",
		strconv.Itoa(self.appuid),
	)
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func (self *Env) CreateConfigFile(job *Task) error {
	file_path := GenerateConfigPath(self.appname, job.Bind)
	out, err := yaml.Marshal(job.Config)
	if err != nil {
		fmt.Println("Get app config failed", err)
		return err
	}
	if err := ioutil.WriteFile(file_path, out, 0600); err != nil {
		fmt.Println("Save app config failed", err)
		return err
	}
	if err := os.Chown(file_path, self.appuid, self.appuid); err != nil {
		fmt.Println("Set owner as app failed", err)
		return err
	}
	return nil
}

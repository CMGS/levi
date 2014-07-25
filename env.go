package main

import (
	"fmt"
	"os/exec"
	"path"
	"strconv"
)

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

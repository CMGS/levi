package main

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
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

func (self *Env) CreateConfigFile(appname string, job Task, wg *sync.WaitGroup) {
	defer wg.Done()
	file_name := strings.Join([]string{appname, strconv.Itoa(job.Bind)}, "_")
	file_name = strings.Join([]string{file_name, "yaml"}, ".")
	file_path := path.Join(DEFAULT_HOME_PATH, appname, file_name)
	out, err := yaml.Marshal(job.Config)
	if err != nil {
		fmt.Println("Get app config failed", err)
		return
	}
	err = ioutil.WriteFile(file_path, out, 0600)
}

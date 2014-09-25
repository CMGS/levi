package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/coreos/go-etcd/etcd"
)

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
}

func Test_GenerateConfigPath(t *testing.T) {
	f := func(appname, ident string) {
		if p := GenerateConfigPath(appname, ident); p != fmt.Sprintf("%s/%s/%s_%s.yaml", config.App.Home, appname, appname, ident) {
			t.Error("Not vaild")
		}
	}
	f("test", "123")
	f("test", "t_test_aaa")
	f("test", "t_daemon_bbb")
}

func Test_GeneratePermdirPath(t *testing.T) {
	f := func(appname, ident string, test bool) {
		p := GeneratePermdirPath(appname, ident, test)
		if !test && p != fmt.Sprintf("%s/%s", config.App.Permdirs, appname) {
			t.Error("Not vaild")
		}
		if test && p != fmt.Sprintf("%s/%s_%s", config.App.Tmpdirs, appname, ident) {
			t.Error("Not vaild")
		}
	}
	f("test", "123", false)
	f("test", "t_test_aaa", true)
}

func Test_CreateConfigFile(t *testing.T) {
	appname := "test"
	job := &Task{Version: "abc", ident: "xxx"}
	configPath := GenerateConfigPath(appname, job.ident)
	dir := "/tmp/levi/test"
	os.MkdirAll(dir, 0755)
	defer func() {
		os.RemoveAll(configPath)
		os.RemoveAll(dir)
	}()
	Etcd.Get = func(p string, _ bool, _ bool) (*etcd.Response, error) {
		ret := &etcd.Response{Node: &etcd.Node{Value: ""}}
		return ret, nil
	}
	env := Env{appname, 4011}
	if err := env.createConfigFile(job, "config.yaml"); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Error(err)
	}
}

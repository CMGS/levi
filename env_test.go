package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"./common"
	"./defines"
	"github.com/coreos/go-etcd/etcd"
)

func init() {
	InitTest()
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
	job := &defines.AddTask{Version: "abc", Ident: "xxx"}
	configPath := GenerateConfigPath(appname, job.Ident)
	dir := path.Join(config.App.Home, "test")
	os.MkdirAll(dir, 0755)
	defer func() {
		os.RemoveAll(configPath)
		os.RemoveAll(dir)
	}()
	common.Etcd.Get = func(p string, _ bool, _ bool) (*etcd.Response, error) {
		ret := &etcd.Response{Node: &etcd.Node{Value: ""}}
		return ret, nil
	}
	env := Env{appname, 4011}
	if err := env.CreateConfigFile(job); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Error(err)
	}
}

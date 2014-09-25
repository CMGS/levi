package main

import (
	"fmt"
	"testing"
)

func init() {
	load("levi.yaml")
	Docker = NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	Etcd = NewEtcdClient(config.Etcd.Machines)
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

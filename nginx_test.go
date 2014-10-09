package main

import (
	"fmt"
	"os"
	"path"
	"testing"
)

var nginx *Nginx

func init() {
	InitTest()
	nginx = NewNginx()
}

func Test_NginxSetUpdate(t *testing.T) {
	appname := "test"
	nginx.SetUpdate(appname)
	if _, ok := nginx.update[appname]; !ok {
		t.Error("Wrong Data")
	}
}

func Test_NginxNewStream(t *testing.T) {
	appname := "test"
	cid := "12345"
	port := "9999"
	nginx.New(appname, cid, port)
	if _, ok := nginx.upstreams[appname]; !ok {
		t.Error("Wrong Data")
	}
	if nginx.upstreams[appname].Ports[cid] != port {
		t.Error("Wrong Value")
	}
}

func Test_NginxRemoveStream(t *testing.T) {
	appname := "test2"
	cid1 := "56789"
	cid2 := "abcde"
	port1 := "123"
	port2 := "456"
	nginx.New(appname, cid1, port1)
	nginx.New(appname, cid2, port2)
	if !nginx.Remove(appname, cid1) {
		t.Error("Remove cid1 Failed")
	}
	if _, ok := nginx.upstreams[appname]; !ok {
		t.Error("Wrong Data")
	}
	if _, ok := nginx.upstreams[appname].Ports[cid2]; !ok {
		t.Error("Remove Wrong")
	}
	if !nginx.Remove(appname, cid2) {
		t.Error("Remove cid2 Failed")
	}
	if _, ok := nginx.upstreams[appname]; ok {
		t.Error("Wrong Data")
	}
}

func Test_NginxClean(t *testing.T) {
	appname := "test3"
	var configPath = path.Join(config.Nginx.Configs, fmt.Sprintf("%s.conf", appname))
	f, err := os.Create(configPath)
	if err != nil {
		t.Error(err)
	}
	f.WriteString("test")
	f.Sync()
	f.Close()
	nginx.Clear(appname)
	if _, err := os.Stat(configPath); err == nil {
		t.Error("Not clean")
	}
}

func Test_NginxSave(t *testing.T) {
	appname := "test4"
	cid := "abcde"
	port := "9999"
	nginx = NewNginx()
	nginx.SetUpdate(appname)
	nginx.New(appname, cid, port)
	var configPath = path.Join(config.Nginx.Configs, fmt.Sprintf("%s.conf", appname))
	nginx.Save()
	if _, err := os.Stat(configPath); err != nil {
		t.Error("Not Exists")
	}
	nginx.Remove(appname, cid)
	nginx.Save()
	if _, err := os.Stat(configPath); err == nil {
		t.Error("Clean Failed")
	}
}

package main

import (
	"testing"
)

func Test_MockEtcd(t *testing.T) {
	load("levi.yaml")
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
	resp, err := Etcd.Get("/test", false, false)
	if err != nil || resp != nil {
		t.Error(err)
	}
}

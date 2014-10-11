package main

import (
	"testing"

	"./defines"
)

func Test_MockEtcd(t *testing.T) {
	load("levi.yaml")
	Etcd = defines.NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	MockEtcd(Etcd)
	resp, err := Etcd.Get("/test", false, false)
	if err != nil || resp != nil {
		t.Error(err)
	}
}

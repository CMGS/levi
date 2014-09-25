package main

import (
	"testing"
)

func Test_MockEtcd(t *testing.T) {
	load("levi.yaml")
	Etcd = NewEtcd(config.Etcd.Machines)
	MockEtcd(Etcd)
}

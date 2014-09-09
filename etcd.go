package main

import (
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

var Etcd *etcd.Client

func NewEtcdClient(machines []string) (client *etcd.Client) {
	if strings.HasPrefix(machines[0], "https://") {
		logger.Assert(nil, "TLS not support")
	} else {
		client = etcd.NewClient(machines)
	}

	if config.Etcd.Sync {
		client.SyncCluster()
	}
	return client
}

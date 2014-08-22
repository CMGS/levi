package main

import (
	"github.com/coreos/go-etcd/etcd"
	"strings"
)

var Etcd *etcd.Client

func NewEtcdClient(machines []string) (client *etcd.Client) {
	// set default if not specified in env
	if len(machines) == 1 && machines[0] == "" {
		machines[0] = "http://127.0.0.1:4001"
	}
	if strings.HasPrefix(machines[0], "https://") {
		logger.Assert(nil, "TLS not support")
		client.SyncCluster()
	} else {
		client = etcd.NewClient(machines)
		client.SyncCluster()
	}
	return client
}

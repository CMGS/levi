package main

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

type EtcdWrapper struct {
	*etcd.Client
}

var Etcd *EtcdWrapper

func NewEtcd(machines []string) *EtcdWrapper {
	var client *etcd.Client
	if strings.HasPrefix(machines[0], "https://") {
		logger.Assert(nil, "TLS not support")
	} else {
		client = etcd.NewClient(machines)
	}

	if config.Etcd.Sync {
		client.SyncCluster()
	}

	e := &EtcdWrapper{Client: client}
	var makeEtcdWrapper func(*EtcdWrapper, *etcd.Client) *EtcdWrapper
	MakeWrapper(&makeEtcdWrapper)
	return makeEtcdWrapper(e, client)
}

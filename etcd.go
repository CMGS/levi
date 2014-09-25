package main

import (
	"strings"

	"github.com/coreos/go-etcd/etcd"
)

type EtcdWrapper struct {
	*etcd.Client
	Get       func(string, bool, bool) (*etcd.Response, error)
	Create    func(string, string, uint64) (*etcd.Response, error)
	CreateDir func(string, uint64) (*etcd.Response, error)
	Delete    func(string, bool) (*etcd.Response, error)
	DeleteDir func(string) (*etcd.Response, error)
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

package defines

import (
	"strings"

	"../utils"
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

func NewEtcd(machines []string, sync bool) *EtcdWrapper {
	var client *etcd.Client
	if strings.HasPrefix(machines[0], "https://") {
		utils.Logger.Assert(nil, "TLS not support")
	} else {
		client = etcd.NewClient(machines)
	}

	if sync {
		client.SyncCluster()
	}

	e := &EtcdWrapper{Client: client}
	var makeEtcdWrapper func(*EtcdWrapper, *etcd.Client) *EtcdWrapper
	utils.MakeWrapper(&makeEtcdWrapper)
	return makeEtcdWrapper(e, client)
}

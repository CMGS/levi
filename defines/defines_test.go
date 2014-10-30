package defines

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var config LeviConfig

func init() {
	config = LeviConfig{}
	config.Master = "ws://127.0.0.1:8888/"
	config.ReadBufferSize = 1024
	config.WriteBufferSize = 1024
	config.Etcd = EtcdConfig{}
	config.Etcd.Machines = []string{"udp://a", "tcp://b"}
	config.Etcd.Sync = true
	config.Docker = DockerConfig{}
	config.Docker.Endpoint = "tcp://192.168.59.103:2375"
}

func Test_MockWebSocket(t *testing.T) {
	Ws := NewWebSocket(config.Master, config.ReadBufferSize, config.WriteBufferSize)
	MockWebSocket(Ws)
	defer Ws.Close()
	if err := Ws.WriteJSON("aaa"); err != nil {
		t.Error(err)
	}
}

func Test_MockEtcd(t *testing.T) {
	Etcd := NewEtcd(config.Etcd.Machines, config.Etcd.Sync)
	MockEtcd(Etcd)
	resp, err := Etcd.Get("/test", false, false)
	if err != nil || resp != nil {
		t.Error(err)
	}
}

func Test_MockDocker(t *testing.T) {
	Docker := NewDocker(config.Docker.Endpoint)
	MockDocker(Docker)
	err := Docker.PushImage(docker.PushImageOptions{}, docker.AuthConfiguration{})
	if err != nil {
		t.Error(err)
	}
}

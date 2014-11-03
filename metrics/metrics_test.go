package metrics

import (
	"testing"

	"github.com/fsouza/go-dockerclient"

	"../common"
	"../defines"
)

var Metrics *MetricsRecorder
var Docker *defines.DockerWrapper
var config defines.MetricsConfig

func init() {
	Docker = defines.NewDocker("tcp://192.168.59.103:2375")
	defines.MockDocker(Docker)
	config = defines.MetricsConfig{10, "localhost:8083", "root", "root", "test"}
	Metrics = NewMetricsRecorder("test", config, Docker)
}

func Test_MetricData(t *testing.T) {
	data := NewMetricData("test", "app")
	if !data.isapp {
		t.Error("Wrong apptype")
	}
}

func Test_MetricReporter(t *testing.T) {
	cid := "123"
	Docker.CreateExec = func(docker.CreateExecOptions) (*docker.Exec, error) {
		return &docker.Exec{"123"}, nil
	}
	Docker.StartExec = func(id string, opt docker.StartExecOptions) error {
		opt.Success <- struct{}{}
		<-opt.Success
		return nil
	}
	Metrics.Add("test", cid, common.DEFAULT_TYPE)
	if _, ok := Metrics.apps[cid]; !ok {
		t.Error("Add Failed")
	}
	Metrics.Remove(cid)
	if _, ok := Metrics.apps[cid]; ok {
		t.Error("Remove Failed")
	}
}

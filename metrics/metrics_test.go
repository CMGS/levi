package metrics

import (
	"testing"

	"../common"
	"../defines"
)

var Metrics *MetricsRecorder
var config defines.MetricsConfig

func init() {
	config = defines.MetricsConfig{10, "localhost:8083", "root", "root", "test"}
	Metrics = NewMetricsRecorder("test", config)
}

func Test_MetricData(t *testing.T) {
	data := NewMetricData("test", "app")
	if !data.isapp {
		t.Error("Wrong apptype")
	}
}

func Test_MetricReporter(t *testing.T) {
	cid := "123"
	Metrics.Add("test", cid, common.DEFAULT_TYPE)
	if _, ok := Metrics.apps[cid]; !ok {
		t.Error("Add Failed")
	}
	Metrics.Remove(cid)
	if _, ok := Metrics.apps[cid]; ok {
		t.Error("Remove Failed")
	}
}

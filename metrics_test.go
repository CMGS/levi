package main

import (
	"testing"
)

func init() {
	InitTest()
}

func Test_MetricData(t *testing.T) {
	data := NewMetricData("test", "app")
	if !data.isapp {
		t.Error("Wrong apptype")
	}
}

func Test_MetricReporter(t *testing.T) {
	cid := "123"
	Metrics.Add("test", cid, DEFAULT_TYPE)
	if _, ok := Metrics.apps[cid]; !ok {
		t.Error("Add Failed")
	}
	Metrics.Remove(cid)
	if _, ok := Metrics.apps[cid]; ok {
		t.Error("Remove Failed")
	}
}

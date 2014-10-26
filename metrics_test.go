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

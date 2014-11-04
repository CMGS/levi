package main

import (
	"./common"
	"./defines"
	"./lenz"
	"./logs"
	"github.com/fsouza/go-dockerclient"
)

type Tester struct {
	id      string
	cid     string
	name    string
	version string
	index   int
}

func (self *Tester) Wait() {
	result := &defines.Result{
		Id:    self.id,
		Done:  true,
		Index: self.index,
		Type:  common.TEST_TASK,
		Data:  "0",
	}
	if _, err := common.Docker.WaitContainer(self.cid); err != nil {
		result.Data = err.Error()
	}
	if err := common.Ws.WriteJSON(result); err != nil {
		logs.Info(err, result)
	}
}

func (self *Tester) GetLogs() {
	result := &defines.Result{
		Id:    self.id,
		Done:  false,
		Index: self.index,
		Type:  common.TEST_TASK,
	}
	outputStream := lenz.GetBuffer(
		Lenz, result, self.name,
		self.version,
		common.TEST_TYPE,
		config.Lenz.Stdout,
	)
	opts := docker.LogsOptions{
		Container:    self.cid,
		OutputStream: outputStream,
		ErrorStream:  outputStream,
		Follow:       true,
		Stdout:       true,
		Stderr:       true,
	}
	if err := common.Docker.Logs(opts); err != nil {
		logs.Info(err)
	}
}

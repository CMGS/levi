package main

import (
	"./common"
	"./defines"
	"./lenz"
	"./logs"
	"./utils"
	"github.com/fsouza/go-dockerclient"
)

type Tester struct {
	id      string
	cid     string
	name    string
	version string
	tid     int
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
	defer func() {
		utils.RemoveContainer(self.cid, true, false)
		if err := common.Ws.WriteJSON(result); err != nil {
			logs.Info(err, result)
		}
		// clean removable flag in some case
		delete(Status.Removable, self.cid)
	}()
	if _, err := common.Docker.WaitContainer(self.cid); err != nil {
		result.Data = err.Error()
	}
}

func (self *Tester) GetLogs() {
	result := &defines.Result{
		Id:    self.id,
		Done:  false,
		Index: self.index,
		Type:  common.TEST_TASK,
	}
	fopts := &defines.ForwardOptions{
		self.tid, common.TEST_TYPE,
		self.name, self.version,
		config.Lenz.Stdout,
	}
	outputStream := lenz.GetBuffer(Lenz, result, fopts)
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

package main

import (
	"./defines"
	"./logs"
)

type Tester struct {
	id   string
	cids map[string]struct{}
}

func (self *Tester) WaitForTester() {
	var err error
	result := &defines.TaskResult{Id: self.id}
	result.Test = make(map[string]*defines.TestResult, len(self.cids))
	for cid, _ := range self.cids {
		r := &defines.TestResult{}
		if cid != "" {
			r.ExitCode, err = Docker.WaitContainer(cid)
			if err != nil {
				r.Err = err.Error()
			}
		} else {
			r.ExitCode = -1
		}
		result.Test[cid] = r
		// Remove test container
		RemoveContainer(cid, true, false)
	}

	logs.Info("Test finished", self.id)
	if err := Ws.WriteJSON(result); err != nil {
		logs.Info(err)
	}
}

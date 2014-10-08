package main

type Tester struct {
	id   string
	cids map[string]string
}

func (self *Tester) WaitForTester() {
	var err error
	result := &TaskResult{Id: self.id}
	result.Test = make(map[string]*TestResult, len(self.cids))
	for tid, cid := range self.cids {
		r := &TestResult{}
		if cid != "" {
			r.ExitCode, err = Docker.WaitContainer(cid)
			if err != nil {
				r.Err = err.Error()
			}
		} else {
			r.ExitCode = -1
		}
		result.Test[tid] = r
		// Remove test container
		RemoveContainer(cid, true, false)
	}

	logger.Info("Test finished", self.id)
	if err := Ws.WriteJSON(result); err != nil {
		logger.Info(err)
	}
}

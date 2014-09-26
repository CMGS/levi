package main

type Result struct {
	ExitCode int
	Err      string
}

type Tester struct {
	appname string
	id      string
	cids    map[string][]interface{}
}

func (self *Tester) WaitForTester() {
	var err error
	result := make(map[string][]*Result, 1)
	result[self.id] = make([]*Result, len(self.cids[self.id]))

	for index, v := range self.cids[self.id] {
		if cid, ok := v.(string); v != nil && ok && cid != "" {
			r := &Result{}
			r.ExitCode, err = Docker.WaitContainer(cid)
			r.Err = err.Error()
			result[self.id][index] = r
			RemoveContainer(cid, true)
		} else {
			result[self.id][index] = &Result{ExitCode: -1}
		}
	}

	logger.Info("Test finished", self.id)
	if err := Ws.WriteJSON(&result); err != nil {
		logger.Info(err)
	}
}

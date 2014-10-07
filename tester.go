package main

type Result struct {
	ExitCode int
	Err      string
}

type Tester struct {
	appname string
	id      string
	cids    map[string]string
}

func (self *Tester) WaitForTester() {
	var err error
	result := make(map[string]map[string]*Result, 1)
	result[self.id] = make(map[string]*Result, len(self.cids))
	for tid, cid := range self.cids {
		r := &Result{}
		if cid != "" {
			r.ExitCode, err = Docker.WaitContainer(cid)
			r.Err = err.Error()
		} else {
			r.ExitCode = -1
		}
		result[self.id][tid] = r
	}

	logger.Info("Test finished", self.id)
	if err := Ws.WriteJSON(result); err != nil {
		logger.Info(err)
	}
}

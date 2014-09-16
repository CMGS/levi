package main

import (
	"github.com/CMGS/websocket"
)

type Result struct {
	ExitCode int
	Err      interface{}
}

type Tester struct {
	appname string
	id      string
	ws      *websocket.Conn
	cids    map[string][]interface{}
}

func (self *Tester) WaitForTester() {
	result := make(map[string][]Result, 1)
	result[self.id] = make([]Result, len(self.cids[self.id]))

	for index, v := range self.cids[self.id] {
		if cid, ok := v.(string); v != nil && ok {
			r := Result{}
			r.ExitCode, r.Err = Docker.WaitContainer(cid)
			result[self.id][index] = r
			self.remove(cid, self.appname)
		}
	}

	if err := self.ws.WriteJSON(&result); err != nil {
		logger.Info(err)
	}
}

func (self *Tester) remove(id, appname string) bool {
	container := Container{
		id:      id,
		appname: appname,
	}

	if err := container.Remove(); err != nil {
		logger.Info("Remove Container", id, "failed", err)
		return false
	}

	return true
}

package main

import (
	"sync"

	"./logs"
)

type Deploy struct {
	wg    *sync.WaitGroup
	nginx *Nginx
	tasks []*AppTask
}

func (self *Deploy) doDeploy() {
	self.wg.Add(len(self.tasks))
	for _, apptask := range self.tasks {
		go func(apptask *AppTask) {
			defer self.wg.Done()
			if apptask.Tasks == nil {
				return
			}
			logs.Info("Appname", apptask.Name)
			env := &Env{apptask.Name, apptask.Uid}
			apptask.Deploy(env, self.nginx)
			apptask.Wait()
		}(apptask)
	}
}

func (self *Deploy) Deploy() {
	logs.Debug("Got tasks", len(self.tasks))
	logs.Debug(self.nginx.upstreams)
	//Do Deploy
	self.doDeploy()
	//Wait For Container Control Finish
	self.Wait()
	//Save Nginx Config
	self.nginx.Save()
	//Clean Task Queue
	self.Init()
}

func (self *Deploy) Wait() {
	self.wg.Wait()
}

func (self *Deploy) Init() {
	self.tasks = make([]*AppTask, 0, config.TaskNum)
	self.nginx = NewNginx()
}

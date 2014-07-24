package main

import (
	"fmt"
	"sync"
)

type Deploy struct {
	result map[string][]int
	tasks  *[]AppTask
	wg     *sync.WaitGroup
}

type deploy_method func(int, Task, AppTask)

func (self *Deploy) incr(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	fmt.Println(job.Image, job.Version, job.Bind)
	self.result[apptask.Id][index] = 1
}

func (self *Deploy) decr(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	fmt.Println(job.Container)
	self.result[apptask.Id][index] = 1
}

func (self *Deploy) doDeploy(method int, fn deploy_method) {
	for _, apptask := range *self.tasks {
		if apptask.Type == method {
			continue
		}
		self.wg.Add(1)
		go func(apptask AppTask) {
			defer self.wg.Done()
			fmt.Println("Appname", apptask.Name)
			self.result[apptask.Id] = make([]int, len(apptask.Tasks))
			self.wg.Add(len(apptask.Tasks))
			for index, job := range apptask.Tasks {
				go fn(index, job, apptask)
			}
		}(apptask)
	}
}

func (self *Deploy) Deploy() {
	self.doDeploy(REMOVE_CONTAINER, self.incr)
	self.doDeploy(ADD_CONTAINER, self.decr)
}

func (self *Deploy) Result() map[string][]int {
	return self.result
}

func (self *Deploy) Wait() {
	self.wg.Wait()
}

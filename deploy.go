package main

import (
	"fmt"
	"path"
	"strconv"
	"sync"
)

type Deploy struct {
	result map[string][]int
	tasks  *[]AppTask
	wg     *sync.WaitGroup
	info   map[string]map[string]string
	dst    *string
}

type deploy_method func(int, Task, AppTask)

func (self *Deploy) incr(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	//TODO PULL Image
	//TODO RUN Container
	fmt.Println("Add Container", job.Image, job.Version)
	self.result[apptask.Id][index] = 1
	if self.info[apptask.Name] == nil {
		self.info[apptask.Name] = make(map[string]string)
	}
	self.info[apptask.Name]["test_cid"+strconv.Itoa(index)] = strconv.Itoa(job.Bind)
}

func (self *Deploy) decr(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	//TODO Stop Container
	//TODO Remove Image
	fmt.Println("Remove Container", job.Container)
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

func (self *Deploy) markRemove() {
	for _, apptask := range *self.tasks {
		if apptask.Type == ADD_CONTAINER || self.info[apptask.Name] == nil {
			continue
		}
		for _, job := range apptask.Tasks {
			delete(self.info[apptask.Name], job.Container)
		}
		if len(self.info[apptask.Name]) == 0 {
			delete(self.info, apptask.Name)
		}
	}
}

func (self *Deploy) genNginxConf() {
	var upstream Upstream
	for appname, appinfo := range self.info {
		upstream = Upstream{appname, []string{}}
		for _, port := range appinfo {
			upstream.Append(port)
		}
		upstream.Execute(path.Join(*self.dst, appname+".conf"))
	}
}

func (self *Deploy) Deploy() {
	self.doDeploy(REMOVE_CONTAINER, self.incr)
	//Wait For Add Container Finish
	self.Wait()
	//Remove Contaners
	self.markRemove()
	//Update Nginx Config
	self.genNginxConf()
	self.doDeploy(ADD_CONTAINER, self.decr)
	self.Wait()
}

func (self *Deploy) Result() map[string][]int {
	return self.result
}

func (self *Deploy) Wait() {
	self.wg.Wait()
}

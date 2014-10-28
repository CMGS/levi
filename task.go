package main

import (
	"fmt"
	"sync"

	"./common"
	"./defines"
	"./logs"
)

type BuildTask struct {
	Group   string
	Name    string
	Version string
	Base    string
	Build   string
	Static  string
	Schema  string
}

type RemoveTask struct {
	Container string
	RmImage   bool
}

func (self *RemoveTask) IsRemoveImage() bool {
	return self.RmImage
}

type AddTask struct {
	Version   string
	Bind      int64
	Port      int64
	Cmd       []string
	Memory    int64
	CpuShares int64
	CpuSet    string
	Daemon    string
	Test      string
	ident     string
}

func (self *AddTask) IsDaemon() bool {
	return self.Daemon != ""
}

func (self *AddTask) IsTest() bool {
	return self.Test != ""
}

func (self *AddTask) ShouldExpose() bool {
	return self.Daemon == "" && self.Test == ""
}

func (self *AddTask) SetAsTest() {
	self.ident = fmt.Sprintf("test_%s", self.Test)
}

func (self *AddTask) SetAsDaemon() {
	self.ident = fmt.Sprintf("daemon_%s", self.Daemon)
}

func (self *AddTask) SetAsService() {
	self.ident = fmt.Sprintf("%d", self.Bind)
}

type Tasks struct {
	Build  []*BuildTask
	Add    []*AddTask
	Remove []*RemoveTask
}

type AppTask struct {
	Id     string
	Uid    int
	Name   string
	Info   bool
	Tasks  *Tasks
	wg     *sync.WaitGroup
	result *defines.TaskResult
}

func (self *AppTask) Deploy(env *Env, nginx *Nginx) {
	self.result = &defines.TaskResult{Id: self.Id}
	self.wg = &sync.WaitGroup{}
	if len(self.Tasks.Add) != 0 {
		self.wg.Add(len(self.Tasks.Add))
		self.result.Add = make([]string, len(self.Tasks.Add))
		for index, job := range self.Tasks.Add {
			switch {
			case job.IsTest():
				job.SetAsTest()
			case job.IsDaemon():
				job.SetAsDaemon()
			default:
				job.SetAsService()
			}
			go self.AddContainer(index, env, nginx)
		}
	}
	if len(self.Tasks.Remove) != 0 {
		self.wg.Add(len(self.Tasks.Remove))
		self.result.Remove = make([]bool, len(self.Tasks.Remove))
		for index, _ := range self.Tasks.Remove {
			go self.RemoveContainer(index, nginx)
		}
	}
	if len(self.Tasks.Build) != 0 {
		self.wg.Add(len(self.Tasks.Build))
		self.result.Build = make([]string, len(self.Tasks.Build))
		for index, _ := range self.Tasks.Build {
			go self.BuildImage(index)
		}
	}
}

func (self *AppTask) Wait() {
	self.wg.Wait()
	if err := Ws.WriteJSON(self.result); err != nil {
		logs.Info(err, self.result)
	}
	if len(self.Tasks.Add) != 0 {
		go func() {
			cids := map[string]struct{}{}
			for index, job := range self.Tasks.Add {
				if !job.IsTest() {
					continue
				}
				cids[self.result.Add[index]] = struct{}{}
			}
			if len(cids) != 0 {
				tester := Tester{self.Id, cids}
				tester.WaitForTester()
			}
		}()
	}
}

func (self *AppTask) storeNewContainerInfo(index int) {
	cid := self.result.Add[index]
	if cid == "" {
		return
	}
	job := self.Tasks.Add[index]
	shortID := cid[:12]
	var aid, at string
	switch {
	case job.IsTest():
		aid = job.Test
		at = common.TEST_TYPE
	case job.IsDaemon():
		aid = job.Daemon
		at = common.DAEMON_TYPE
		Status.Removable[cid] = struct{}{}
	default:
		aid = string(job.Port)
		at = common.DEFAULT_TYPE
		Status.Removable[cid] = struct{}{}
	}
	Metrics.Add(self.Name, shortID, at)
	Lenz.Attacher.Attach(shortID, self.Name, aid, at)
}

func (self *AppTask) AddContainer(index int, env *Env, nginx *Nginx) {
	defer self.wg.Done()
	job := self.Tasks.Add[index]
	defer self.storeNewContainerInfo(index)
	env.CreateUser()
	if err := env.CreateConfigFile(job); err != nil {
		logs.Info("Create app config failed", err)
		return
	}
	if err := env.CreatePermdir(job); err != nil {
		logs.Info("Create app permdir failed", err)
		return
	}
	logs.Info("Add Container", self.Name, "@", job.Version)
	image := Image{
		self.Name,
		job.Version,
		job.Bind,
	}
	if err := image.Pull(); err != nil {
		logs.Info("Pull Image", self.Name, "@", job.Version, "Failed", err)
		return
	}
	container, err := image.Run(job, self.Uid)
	if err != nil {
		logs.Info("Run Image", self.Name, "@", job.Version, "Failed", err)
		return
	}
	logs.Info("Run Image", self.Name, "@", job.Version, "Succeed", container.ID)
	if job.ShouldExpose() {
		nginx.New(self.Name, container.ID, job.ident)
		nginx.SetUpdate(self.Name)
	}
	self.result.Add[index] = container.ID
	logs.Info("Add Finished", container.ID)
}

func (self *AppTask) RemoveContainer(index int, nginx *Nginx) {
	defer self.wg.Done()
	job := self.Tasks.Remove[index]
	logs.Info("Remove Container", self.Name, job.Container)
	if _, ok := Status.Removable[job.Container]; !ok {
		logs.Info("Not Record")
		return
	}
	delete(Status.Removable, job.Container)
	defer func() {
		if !self.result.Remove[index] {
			Status.Removable[job.Container] = struct{}{}
			return
		}
	}()
	container := Container{
		id:      job.Container,
		appname: self.Name,
	}
	if err := container.Stop(); err != nil {
		logs.Info("Stop Container", job.Container, "failed", err)
		return
	}
	// Test container will be automatic removed
	if err := RemoveContainer(job.Container, false, job.IsRemoveImage()); err != nil {
		logs.Info("Remove Container", job.Container, "failed", err)
		return
	}
	if ok := nginx.Remove(self.Name, job.Container); ok {
		nginx.SetUpdate(self.Name)
	}
	self.result.Remove[index] = true
	logs.Info("Remove Finished", self.result.Remove[index])
}

func (self *AppTask) BuildImage(index int) {
	defer self.wg.Done()
	job := self.Tasks.Build[index]
	builder := NewBuilder(self.Name, job)
	if err := builder.Build(); err != nil {
		logs.Info(err)
		return
	}
	self.result.Build[index] = builder.repoTag
	logs.Info("Build Finished", self.result.Build[index])
}

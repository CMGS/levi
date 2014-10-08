package main

import (
	"fmt"
	"sync"
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
	RunEnv    string
}

func (self *RemoveTask) IsTest() bool {
	return self.RunEnv == TESTING
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
	result *TaskResult
}

func (self *AppTask) Deploy(env *Env, nginx *Nginx) {
	self.result = &TaskResult{Id: self.Id}
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
			go self.AddContainer(index, job, env, nginx)
		}
	}
	if len(self.Tasks.Remove) != 0 {
		self.wg.Add(len(self.Tasks.Remove))
		self.result.Remove = make([]bool, len(self.Tasks.Remove))
		for index, job := range self.Tasks.Remove {
			go self.RemoveContainer(index, job, nginx)
		}
	}
	if len(self.Tasks.Build) != 0 {
		self.wg.Add(len(self.Tasks.Build))
		self.result.Build = make([]string, len(self.Tasks.Build))
		for index, job := range self.Tasks.Build {
			go self.BuildImage(index, job)
		}
	}
}

func (self *AppTask) Wait() {
	self.wg.Wait()
	cids := map[string]string{}
	if len(self.Tasks.Add) == 0 {
		return
	}
	for index, job := range self.Tasks.Add {
		if !job.IsTest() {
			continue
		}
		cids[job.Test] = self.result.Add[index]
	}
	tester := Tester{self.Id, cids}
	// Block Levi
	tester.WaitForTester()
}

func (self *AppTask) AddContainer(index int, job *AddTask, env *Env, nginx *Nginx) {
	defer self.wg.Done()
	env.CreateUser()
	if err := env.CreateConfigFile(job); err != nil {
		logger.Info("Create app config failed", err)
		return
	}
	if err := env.CreatePermdir(job); err != nil {
		logger.Info("Create app permdir failed", err)
		return
	}
	logger.Info("Add Container", self.Name, "@", job.Version)
	image := Image{
		self.Name,
		job.Version,
		job.Bind,
	}
	if err := image.Pull(); err != nil {
		logger.Info("Pull Image", self.Name, "@", job.Version, "Failed", err)
		return
	}
	container, err := image.Run(job, self.Uid)
	if err != nil {
		logger.Info("Run Image", self.Name, "@", job.Version, "Failed", err)
		return
	}
	logger.Info("Run Image", self.Name, "@", job.Version, "Succeed", container.ID)
	if job.ShouldExpose() {
		nginx.New(self.Name, container.ID, job.ident)
		nginx.SetUpdate(self.Name)
	}
	self.result.Add[index] = container.ID
	// Record normal container
	if !job.IsTest() {
		Status.Removable[container.ID] = struct{}{}
	}
	logger.Info("Add Finished", container.ID)
}

func (self *AppTask) RemoveContainer(index int, job *RemoveTask, nginx *Nginx) {
	defer self.wg.Done()
	logger.Info("Remove Container", self.Name, job.Container)
	Status.Removable[job.Container] = struct{}{}
	container := Container{
		id:      job.Container,
		appname: self.Name,
	}
	if err := container.Stop(); err != nil {
		logger.Info("Stop Container", job.Container, "failed", err)
		return
	}
	if err := RemoveContainer(job.Container, job.IsTest(), job.IsRemoveImage()); err != nil {
		logger.Info("Remove Container", job.Container, "failed", err)
		return
	}
	if ok := nginx.Remove(self.Name, job.Container); ok {
		nginx.SetUpdate(self.Name)
	}
	self.result.Remove[index] = true
	logger.Info("Remove Finished", self.result.Remove[index])
}

func (self *AppTask) BuildImage(index int, job *BuildTask) {
	defer self.wg.Done()
	builder := NewBuilder(self.Name, job)
	if err := builder.Build(); err != nil {
		logger.Info(err)
		return
	}
	self.result.Build[index] = builder.repoTag
	logger.Info("Build Finished", self.result.Build[index])
}

type TestResult struct {
	ExitCode int
	Err      string
}

type TaskResult struct {
	Id     string
	Build  []string
	Add    []string
	Remove []bool
	Test   map[string]*TestResult
}

package main

import (
	"fmt"
	"sync"

	"./common"
	"./defines"
	"./logs"
	"./utils"
)

type AppTask struct {
	Id    string
	Uid   int
	Name  string
	Info  bool
	Tasks *defines.Tasks
	wg    *sync.WaitGroup
}

func (self *AppTask) Deploy(env *Env, nginx *Nginx) {
	self.wg = &sync.WaitGroup{}
	if len(self.Tasks.Add) != 0 {
		self.wg.Add(len(self.Tasks.Add))
		for index, job := range self.Tasks.Add {
			switch {
			case job.IsTest():
				job.SetAsTest()
			case job.IsDaemon():
				job.SetAsDaemon()
			default:
				job.SetAsService()
			}
			if len(self.Tasks.Remove) != 0 {
				self.AddContainer(index, env, nginx)
			} else {
				go self.AddContainer(index, env, nginx)
			}
		}
	}
	if len(self.Tasks.Remove) != 0 {
		self.wg.Add(len(self.Tasks.Remove))
		for index, _ := range self.Tasks.Remove {
			go self.RemoveContainer(index, nginx)
		}
	}
	if len(self.Tasks.Build) != 0 {
		self.wg.Add(len(self.Tasks.Build))
		for index, _ := range self.Tasks.Build {
			go self.BuildImage(index)
		}
	}
	if len(self.Tasks.Update) != 0 {
		self.wg.Add(len(self.Tasks.Update))
		for index, _ := range self.Tasks.Update {
			go self.UpdateConfig(index, env)
		}
	}
}

func (self *AppTask) Wait() {
	self.wg.Wait()
}

func (self *AppTask) storeNewContainerInfo(result *defines.Result) {
	if result.Data != "" {
		job := self.Tasks.Add[result.Index]
		cid := result.Data
		shortID := cid[:12]
		var aid, at string
		if job.IsTest() {
			result.Done = false
			self.writeBack(result)
			at = common.TEST_TYPE
			tester := Tester{
				id:      result.Id,
				cid:     cid,
				tid:     job.Id,
				name:    self.Name,
				version: job.Version,
				index:   result.Index,
			}
			tester.GetLogs()
			tester.Wait()
			return
		} else {
			switch {
			case job.IsDaemon():
				aid = job.Daemon
				at = common.DAEMON_TYPE
			default:
				aid = fmt.Sprintf("%d", job.Bind)
				at = common.DEFAULT_TYPE
			}
			Status.Removable[cid] = struct{}{}
			Lenz.Attacher.Attach(shortID, self.Name, aid, at)
		}
		Metrics.Add(self.Name, shortID, at)
	}
	self.writeBack(result)
}

func (self *AppTask) writeBack(result *defines.Result) {
	if err := common.Ws.WriteJSON(result); err != nil {
		logs.Info(err, result)
	}
}

func (self *AppTask) AddContainer(index int, env *Env, nginx *Nginx) {
	defer self.wg.Done()
	job := self.Tasks.Add[index]
	result := &defines.Result{
		Id:    self.Id,
		Done:  true,
		Index: index,
		Type:  common.ADD_TASK,
	}
	env.CreateUser()
	defer self.storeNewContainerInfo(result)
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
		nginx.New(self.Name, container.ID, job.Ident)
		nginx.SetUpdate(self.Name)
	}
	result.Data = container.ID
	logs.Info("Add Finished", container.ID)
}

func (self *AppTask) RemoveContainer(index int, nginx *Nginx) {
	defer self.wg.Done()
	job := self.Tasks.Remove[index]
	result := &defines.Result{
		Id:    self.Id,
		Done:  true,
		Index: index,
		Type:  common.REMOVE_TASK,
	}
	defer func() {
		//TODO Not Safe
		if result.Data != "" {
			delete(Status.Removable, job.Container)
		}
		self.writeBack(result)
	}()
	logs.Info("Remove Container", self.Name, job.Container)
	if _, ok := Status.Removable[job.Container]; !ok {
		logs.Info("Not Record")
		return
	}
	container := Container{
		id:      job.Container,
		appname: self.Name,
	}
	if err := container.Stop(); err != nil {
		logs.Info("Stop Container", job.Container, "failed", err)
		return
	}
	// Test container will be automatic removed
	if err := utils.RemoveContainer(job.Container, false, job.IsRemoveImage()); err != nil {
		logs.Info("Remove Container", job.Container, "failed", err)
		return
	}
	if ok := nginx.Remove(self.Name, job.Container); ok {
		nginx.SetUpdate(self.Name)
	}
	result.Data = "1"
	logs.Info("Remove Finished")
}

func (self *AppTask) BuildImage(index int) {
	defer self.wg.Done()
	job := self.Tasks.Build[index]
	result := &defines.Result{
		Id:    self.Id,
		Done:  false,
		Index: index,
		Type:  common.BUILD_TASK,
	}
	defer func() {
		result.Done = true
		self.writeBack(result)
	}()
	builder := NewBuilder(self.Name, job)
	if err := builder.Build(result); err != nil {
		logs.Info("Build Failed", err)
		return
	}
	result.Data = builder.repoTag
	logs.Info("Build Finished", builder.repoTag)
}

func (self *AppTask) UpdateConfig(index int, env *Env) {
	defer self.wg.Done()
	job := self.Tasks.Update[index]
	result := &defines.Result{
		Id:    self.Id,
		Done:  true,
		Index: index,
		Type:  common.UPDATE_TASK,
	}
	defer self.writeBack(result)
	logs.Info("Update Container Config", self.Name, job.Container)
	container := Container{
		id:      job.Container,
		appname: self.Name,
	}
	ident, err := container.GetIdent()
	if err != nil {
		logs.Info("Update Config Failed", err)
		return
	}
	env.DoCreateConfigFile(ident, common.PROD_CONFIG_FILE)
	result.Data = "1"
	logs.Info("Update Config Finished")
}

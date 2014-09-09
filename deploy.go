package main

import (
	"container/list"
	"github.com/CMGS/go-dockerclient"
	"strings"
	"sync"
)

type Deploy struct {
	result     map[string][]interface{}
	tasks      *list.List
	wg         *sync.WaitGroup
	containers *[]docker.APIContainers
	nginx      *Nginx
}

func (self *Deploy) add(index int, job Task, apptask *AppTask) string {
	logger.Info("Add Container", apptask.Name, "@", job.Version)
	image := Image{
		apptask.Name,
		job.Version,
		job.Bind,
	}
	if err := image.Pull(); err != nil {
		logger.Info("Pull Image", apptask.Name, "@", job.Version, "Failed", err)
		return ""
	}
	container, err := image.Run(&job, apptask.Uid)
	if err != nil {
		logger.Info("Run Image", apptask.Name, "@", job.Version, "Failed", err)
		return ""
	}
	logger.Info("Run Image", apptask.Name, "@", job.Version, "Succeed", container.ID)
	if !job.CheckDaemon() {
		self.nginx.New(apptask.Name, container.ID, job.ident)
		self.nginx.SetUpdate(apptask.Name)
	}
	return container.ID
}

func (self *Deploy) remove(index int, job Task, apptask *AppTask) bool {
	logger.Info("Remove Container", apptask.Name, job.Container)
	container := Container{
		id:      job.Container,
		appname: apptask.Name,
	}
	if err := container.Stop(); err != nil {
		logger.Info("Stop Container", job.Container, "failed", err)
		return false
	}
	if err := container.Remove(); err != nil {
		logger.Info("Remove Container", job.Container, "failed", err)
		return false
	}
	if ok := self.nginx.Remove(apptask.Name, job.Container); ok {
		self.nginx.SetUpdate(apptask.Name)
	}
	return true
}

func (self *Deploy) AddContainer(index int, job Task, apptask *AppTask, env *Env) {
	defer self.wg.Done()
	if err := env.CreateConfigFile(&job); err != nil {
		logger.Info("Create app config failed", err)
		return
	}
	cid := self.add(index, job, apptask)
	self.result[apptask.Id][index] = cid
	logger.Info("Add Finished", cid)
}

func (self *Deploy) RemoveContainer(index int, job Task, apptask *AppTask, _ *Env) {
	defer self.wg.Done()
	result := self.remove(index, job, apptask)
	self.result[apptask.Id][index] = result
	logger.Info("Remove Finished", result)
}

func (self *Deploy) UpdateApp(index int, job Task, apptask *AppTask, env *Env) {
	defer self.wg.Done()
	self.result[apptask.Id][index] = ""
	if result := self.remove(index, job, apptask); !result {
		return
	}
	if err := env.CreateConfigFile(&job); err != nil {
		return
	}
	cid := self.add(index, job, apptask)
	self.result[apptask.Id][index] = cid
	logger.Info("Update Finished", cid)
}

func (self *Deploy) BuildImage(index int, job Task, apptask *AppTask, _ *Env) {
	defer self.wg.Done()
	self.result[apptask.Id][index] = ""
	builder := NewBuilder(apptask.Name, &job.Build)
	if err := builder.Build(); err != nil {
		logger.Info(err)
		return
	}
	self.result[apptask.Id][index] = builder.repoTag
	logger.Info("Build Finished", builder.repoTag)
}

func (self *Deploy) DoDeploy() {
	for apptask := self.tasks.Front(); apptask != nil; apptask = apptask.Next() {
		self.wg.Add(1)
		go func(apptask AppTask) {
			defer self.wg.Done()
			logger.Info("Appname", apptask.Name)
			self.result[apptask.Id] = make([]interface{}, len(apptask.Tasks))
			self.wg.Add(len(apptask.Tasks))
			env := Env{apptask.Name, apptask.Uid}
			var f func(index int, job Task, apptask *AppTask, env *Env)
			switch apptask.Type {
			case BUILD_IMAGE:
				logger.Info("Build Task")
				f = self.BuildImage
			case ADD_CONTAINER:
				logger.Info("Add Task")
				env.CreateUser()
				f = self.AddContainer
			case UPDATE_CONTAINER:
				logger.Info("Update Task")
				env.CreateUser()
				f = self.UpdateApp
			case REMOVE_CONTAINER:
				logger.Info("Remove Task")
				f = self.RemoveContainer
			}
			for index, job := range apptask.Tasks {
				if job.IsDaemon() {
					job.SetAsDaemon()
				} else {
					job.SetAsService()
				}
				go f(index, job, &apptask, &env)
			}
		}(apptask.Value.(AppTask))
	}
}

func (self *Deploy) GenerateInfo() {
	for _, container := range *self.containers {
		var appinfo = strings.SplitN(strings.TrimLeft(container.Names[0], "/"), "_", 2)
		if strings.Contains(appinfo[1], "daemon_") {
			continue
		}
		appname, apport := appinfo[0], appinfo[1]
		self.nginx.New(appname, container.ID, apport)
	}
}

func (self *Deploy) Deploy() {
	//Generate Container Info
	self.GenerateInfo()
	//Do Deploy
	self.DoDeploy()
	//Wait For Container Control Finish
	self.Wait()
	//Save Nginx Config
	self.nginx.Save()
}

func (self *Deploy) Result() map[string][]interface{} {
	return self.result
}

func (self *Deploy) Wait() {
	self.wg.Wait()
}

func (self *Deploy) Reset() {
	self.tasks.Init()
	self.result = make(map[string][]interface{})
}

package main

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Deploy struct {
	wg    *sync.WaitGroup
	ws    *websocket.Conn
	nginx *Nginx
	tasks []*AppTask
}

func (self *Deploy) add(index int, job Task, apptask *AppTask, runenv string) string {
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
	container, err := image.Run(&job, apptask.Uid, runenv)
	if err != nil {
		logger.Info("Run Image", apptask.Name, "@", job.Version, "Failed", err)
		return ""
	}
	logger.Info("Run Image", apptask.Name, "@", job.Version, "Succeed", container.ID)
	if !job.CheckDaemon() && !job.CheckTest() {
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
	if ok := self.nginx.Remove(apptask.Name, job.Container); ok {
		self.nginx.SetUpdate(apptask.Name)
	}
	return true
}

func (self *Deploy) AddContainer(index int, job Task, apptask *AppTask, env *Env) {
	defer apptask.wg.Done()
	env.CreateUser()
	if err := env.CreateConfigFile(&job); err != nil {
		logger.Info("Create app config failed", err)
		return
	}
	if err := env.CreatePermdir(&job, false); err != nil {
		logger.Info("Create app permdir failed", err)
		return
	}
	if cid := self.add(index, job, apptask, PRODUCTION); cid != "" {
		apptask.result[apptask.Id][index] = cid
		logger.Info("Add Finished", cid)
	}
}

func (self *Deploy) RemoveContainer(index int, job Task, apptask *AppTask, _ *Env) {
	defer apptask.wg.Done()
	result := self.remove(index, job, apptask)
	apptask.result[apptask.Id][index] = result
	logger.Info("Remove Finished", result)
}

func (self *Deploy) UpdateApp(index int, job Task, apptask *AppTask, env *Env) {
	defer apptask.wg.Done()
	env.CreateUser()
	apptask.result[apptask.Id][index] = ""
	if result := self.remove(index, job, apptask); !result {
		return
	}
	if err := env.CreateConfigFile(&job); err != nil {
		return
	}
	if err := env.CreatePermdir(&job, false); err != nil {
		logger.Info("Create app permdir failed", err)
		return
	}
	if cid := self.add(index, job, apptask, PRODUCTION); cid != "" {
		apptask.result[apptask.Id][index] = cid
		logger.Info("Update Finished", cid)
	}
}

func (self *Deploy) BuildImage(index int, job Task, apptask *AppTask, _ *Env) {
	defer apptask.wg.Done()
	builder := NewBuilder(apptask.Name, &job.Build)
	if err := builder.Build(); err != nil {
		logger.Info(err)
		return
	}
	apptask.result[apptask.Id][index] = builder.repoTag
	logger.Info("Build Finished", builder.repoTag)
}

func (self *Deploy) TestImage(index int, job Task, apptask *AppTask, env *Env) {
	defer apptask.wg.Done()
	env.CreateUser()
	if err := env.CreateTestConfigFile(&job); err != nil {
		logger.Info("Create app test config failed", err)
		return
	}
	if err := env.CreatePermdir(&job, true); err != nil {
		logger.Info("Create app permdir failed", err)
		return
	}
	if cid := self.add(index, job, apptask, TESTING); cid != "" {
		apptask.result[apptask.Id][index] = cid
		logger.Info("Start Testing", cid)
	}
}

func (self *Deploy) DoDeploy() {
	self.wg.Add(len(self.tasks))
	for _, apptask := range self.tasks {
		go func(apptask *AppTask) {
			defer self.wg.Done()
			logger.Info("Appname", apptask.Name)
			apptask.result = make(map[string][]interface{}, 1)
			apptask.result[apptask.Id] = make([]interface{}, len(apptask.Tasks))
			apptask.wg = &sync.WaitGroup{}
			apptask.wg.Add(len(apptask.Tasks))
			env := Env{apptask.Name, apptask.Uid}
			var f func(index int, job Task, apptask *AppTask, env *Env)
			switch apptask.Type {
			case BUILD_IMAGE:
				logger.Info("Build Task")
				f = self.BuildImage
			case TEST_IMAGE:
				logger.Info("Test Task")
				f = self.TestImage
			case ADD_CONTAINER:
				logger.Info("Add Task")
				f = self.AddContainer
			case UPDATE_CONTAINER:
				logger.Info("Update Task")
				f = self.UpdateApp
			case REMOVE_CONTAINER:
				logger.Info("Remove Task")
				f = self.RemoveContainer
			}
			for index, job := range apptask.Tasks {
				switch {
				case job.IsTest():
					job.SetAsTest()
				case job.IsDaemon():
					job.SetAsDaemon()
				case !job.IsDaemon():
					job.SetAsService()
				}
				go f(index, job, apptask, &env)
			}
			apptask.wg.Wait()
			if err := self.ws.WriteJSON(&apptask.result); err != nil {
				logger.Info(err, apptask.result)
			}
			if apptask.Type == TEST_IMAGE {
				tester := Tester{
					appname: apptask.Name,
					id:      apptask.Id,
					cids:    apptask.result,
				}
				go tester.WaitForTester(self.ws)
			}
		}(apptask)
	}
}

func (self *Deploy) Deploy() {
	logger.Debug("Got tasks", len(self.tasks))
	logger.Debug(self.nginx.upstreams)
	//Do Deploy
	self.DoDeploy()
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

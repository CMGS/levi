package main

import (
	"github.com/CMGS/websocket"
	"sync"
)

type Deploy struct {
	tasks []*AppTask
	wg    *sync.WaitGroup
	nginx *Nginx
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
	defer apptask.wg.Done()
	if err := env.CreateConfigFile(&job); err != nil {
		logger.Info("Create app config failed", err)
		return
	}
	cid := self.add(index, job, apptask, PRODUCTION)
	apptask.result[apptask.Id][index] = cid
	logger.Info("Add Finished", cid)
}

func (self *Deploy) RemoveContainer(index int, job Task, apptask *AppTask, _ *Env) {
	defer apptask.wg.Done()
	result := self.remove(index, job, apptask)
	apptask.result[apptask.Id][index] = result
	logger.Info("Remove Finished", result)
}

func (self *Deploy) UpdateApp(index int, job Task, apptask *AppTask, env *Env) {
	defer apptask.wg.Done()
	apptask.result[apptask.Id][index] = ""
	if result := self.remove(index, job, apptask); !result {
		return
	}
	if err := env.CreateConfigFile(&job); err != nil {
		return
	}
	cid := self.add(index, job, apptask, PRODUCTION)
	apptask.result[apptask.Id][index] = cid
	logger.Info("Update Finished", cid)
}

func (self *Deploy) BuildImage(index int, job Task, apptask *AppTask, _ *Env) {
	defer apptask.wg.Done()
	apptask.result[apptask.Id][index] = ""
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
	if err := env.CreateTestConfigFile(&job); err != nil {
		logger.Info("Create app test config failed", err)
		return
	}
	cid := self.add(index, job, apptask, TESTING)
	apptask.result[apptask.Id][index] = cid
	logger.Info("Testing Start", cid)
}

func (self *Deploy) DoDeploy(ws *websocket.Conn) {
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
				env.CreateUser()
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
			if err := ws.WriteJSON(&apptask.result); err != nil {
				logger.Info(err)
			}
			if apptask.Type == TEST_IMAGE {
				tester := Tester{
					appname: apptask.Name,
					id:      apptask.Id,
					ws:      ws,
					cids:    apptask.result,
				}
				go tester.WaitForTester()
			}
		}(apptask)
	}
}

func (self *Deploy) Deploy(ws *websocket.Conn) {
	//Do Deploy
	self.DoDeploy(ws)
	//Wait For Container Control Finish
	self.Wait()
	//Save Nginx Config
	self.nginx.Save()
}

func (self *Deploy) Wait() {
	self.wg.Wait()
}

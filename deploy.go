package main

import (
	"container/list"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"strconv"
	"strings"
	"sync"
)

type Deploy struct {
	result     map[string][]interface{}
	tasks      *list.List
	wg         *sync.WaitGroup
	containers *[]docker.APIContainers
	nginx      *Nginx
	client     *docker.Client
	registry   *string
}

func (self *Deploy) add(index int, job Task, apptask AppTask) string {
	fmt.Println("Add Container", apptask.Name, "@", job.Version)
	image := Image{
		self.client,
		apptask.Name,
		job.Version,
		GenerateConfigPath(apptask.Name, job.Bind),
		job.Bind,
		self.registry,
	}
	if err := image.Pull(); err != nil {
		fmt.Println("Pull Image", apptask.Name, "@", job.Version, "Failed", err)
		return ""
	}
	container, err := image.Run(&job, apptask.User)
	if err != nil {
		fmt.Println("Run Image", apptask.Name, "@", job.Version, "Failed", err)
		return ""
	}
	fmt.Println("Run Image", apptask.Name, "@", job.Version, "Succeed", container.ID)
	self.nginx.New(apptask.Name, container.ID, strconv.Itoa(job.Bind))
	return container.ID
}

func (self *Deploy) remove(index int, job Task, apptask AppTask) bool {
	fmt.Println("Remove Container", apptask.Name, job.Container)
	container := Container{
		client:  self.client,
		id:      job.Container,
		appname: apptask.Name,
	}
	if err := container.Stop(); err != nil {
		fmt.Println("Stop Container", job.Container, "failed")
		return false
	}
	if err := container.Remove(); err != nil {
		fmt.Println("Remove Container", job.Container, "failed")
		return false
	}
	self.nginx.Remove(apptask.Name, job.Container)
	return true
}

func (self *Deploy) AddContainer(index int, job Task, apptask AppTask, env *Env) {
	defer self.wg.Done()
	if err := env.CreateConfigFile(&job); err != nil {
		return
	}
	cid := self.add(index, job, apptask)
	self.result[apptask.Id][index] = cid
}

func (self *Deploy) RemoveContainer(index int, job Task, apptask AppTask, env *Env) {
	defer self.wg.Done()
	if err := env.RemoveConfigFile(&job); err != nil {
		return
	}
	result := self.remove(index, job, apptask)
	self.result[apptask.Id][index] = result
}

func (self *Deploy) UpdateApp(index int, job Task, apptask AppTask, env *Env) {
	defer self.wg.Done()
	self.result[apptask.Id][index] = ""
	if err := env.RemoveConfigFile(&job); err != nil {
		return
	}
	if result := self.remove(index, job, apptask); !result {
		return
	}
	if err := env.CreateConfigFile(&job); err != nil {
		return
	}
	cid := self.add(index, job, apptask)
	self.result[apptask.Id][index] = cid
}

func (self *Deploy) DoDeploy() {
	for apptask := self.tasks.Front(); apptask != nil; apptask = apptask.Next() {
		self.wg.Add(1)
		go func(apptask AppTask) {
			defer self.wg.Done()
			fmt.Println("Appname", apptask.Name)
			self.result[apptask.Id] = make([]interface{}, len(apptask.Tasks))
			self.wg.Add(len(apptask.Tasks))
			env := Env{apptask.User, apptask.Uid}
			var f func(index int, job Task, apptask AppTask, env *Env)
			switch apptask.Type {
			case ADD_CONTAINER:
				env.CreateUser()
				f = self.AddContainer
			case UPDATE_CONTAINER:
				env.CreateUser()
				f = self.UpdateApp
			case REMOVE_CONTAINER:
				f = self.RemoveContainer
			}
			for index, job := range apptask.Tasks {
				go f(index, job, apptask, &env)
			}
		}(apptask.Value.(AppTask))
	}
}

func (self *Deploy) GenerateInfo() {
	for _, container := range *self.containers {
		split_names := strings.SplitN(strings.TrimLeft(container.Names[0], "/"), "_", 2)
		appname, appport := split_names[0], split_names[1]
		self.nginx.New(appname, container.ID, appport)
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
	//Restart Nginx
	self.nginx.Restart()
}

func (self *Deploy) Result() map[string][]interface{} {
	return self.result
}

func (self *Deploy) Wait() {
	self.wg.Wait()
}

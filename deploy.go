package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"strconv"
	"strings"
	"sync"
)

type Deploy struct {
	result     map[string][]interface{}
	tasks      *[]AppTask
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
	}
	if err := image.Pull(self.registry); err != nil {
		fmt.Println("Pull Image", apptask.Name, "@", job.Version, "Failed", err)
		return ""
	}
	container, err := image.Run(&job, self.registry, apptask.User)
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

func (self *Deploy) AddContainer(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	cid := self.add(index, job, apptask)
	self.result[apptask.Id][index] = cid
}

func (self *Deploy) RemoveContainer(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	result := self.remove(index, job, apptask)
	self.result[apptask.Id][index] = result
}

func (self *Deploy) UpdateApp(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	result := self.remove(index, job, apptask)
	if !result {
		self.result[apptask.Id][index] = ""
		return
	}
	cid := self.add(index, job, apptask)
	self.result[apptask.Id][index] = cid
}

func (self *Deploy) DoDeploy() {
	for _, apptask := range *self.tasks {
		self.wg.Add(1)
		go func(apptask AppTask) {
			defer self.wg.Done()
			fmt.Println("Appname", apptask.Name)
			self.result[apptask.Id] = make([]interface{}, len(apptask.Tasks))
			self.wg.Add(len(apptask.Tasks))
			switch apptask.Type {
			case ADD_CONTAINER:
				for index, job := range apptask.Tasks {
					go self.AddContainer(index, job, apptask)
				}
			case REMOVE_CONTAINER:
				for index, job := range apptask.Tasks {
					go self.RemoveContainer(index, job, apptask)
				}
			case UPDATE_CONTAINER:
				for index, job := range apptask.Tasks {
					go self.UpdateApp(index, job, apptask)
				}
			}
		}(apptask)
	}
}

func (self *Deploy) GenerateInfo() {
	for _, container := range *self.containers {
		split_names := strings.SplitN(strings.TrimLeft(container.Names[0], "/"), "_", 2)
		appname, appport := split_names[0], split_names[1]
		self.nginx.New(appname, container.ID, appport)
	}
}

func (self *Deploy) PrepareEnv() {
	for _, apptask := range *self.tasks {
		if apptask.Type == REMOVE_CONTAINER {
			continue
		}
		self.wg.Add(1)
		go func(apptask AppTask) {
			defer self.wg.Done()
			env := Env{apptask.Name, apptask.Uid}
			env.CreateUser()
			self.wg.Add(len(apptask.Tasks))
			for _, job := range apptask.Tasks {
				go env.CreateConfigFile(job, self.wg)
			}
		}(apptask)
	}
}

func (self *Deploy) Deploy() {
	//Generate Container Info
	self.GenerateInfo()
	//Prepare OS environment
	self.PrepareEnv()
	//Wait Env Prepared
	self.Wait()
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

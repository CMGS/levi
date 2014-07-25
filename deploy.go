package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"strconv"
	"strings"
	"sync"
)

type Deploy struct {
	result     map[string][]bool
	tasks      *[]AppTask
	wg         *sync.WaitGroup
	containers *[]docker.APIContainers
	nginx      *Nginx
}

func (self *Deploy) add(index int, job Task, apptask AppTask) {
	//TODO PULL Image
	//TODO GEN Servie Conf
	//TODO RUN Container
	fmt.Println("Add Container", job.Image, job.Version)
	//TODO test now, use fake cid
	self.nginx.New(apptask.Name, "test_cid"+strconv.Itoa(index), strconv.Itoa(job.Bind))
}

func (self *Deploy) remove(index int, job Task, apptask AppTask) {
	//TODO S``top Container
	//TODO Remove Servie Conf
	//TODO Remove Image
	fmt.Println("Remove Container", apptask.Name, job.Container)
	self.nginx.Remove(apptask.Name, job.Container)
}

func (self *Deploy) AddContainer(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	self.add(index, job, apptask)
	self.result[apptask.Id][index] = true
}

func (self *Deploy) RemoveContainer(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	self.remove(index, job, apptask)
	self.result[apptask.Id][index] = true
}

func (self *Deploy) UpdateApp(index int, job Task, apptask AppTask) {
	defer self.wg.Done()
	self.remove(index, job, apptask)
	self.add(index, job, apptask)
	self.result[apptask.Id][index] = true
}

func (self *Deploy) DoDeploy() {
	for _, apptask := range *self.tasks {
		self.wg.Add(1)
		go func(apptask AppTask) {
			defer self.wg.Done()
			fmt.Println("Appname", apptask.Name)
			self.result[apptask.Id] = make([]bool, len(apptask.Tasks))
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

func (self *Deploy) Result() map[string][]bool {
	return self.result
}

func (self *Deploy) Wait() {
	self.wg.Wait()
}

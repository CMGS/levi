package defines

import (
	"../logs"
	"../utils"
	"github.com/fsouza/go-dockerclient"
)

type DockerWrapper struct {
	*docker.Client
	PushImage        func(docker.PushImageOptions, docker.AuthConfiguration) error
	PullImage        func(docker.PullImageOptions, docker.AuthConfiguration) error
	CreateContainer  func(docker.CreateContainerOptions) (*docker.Container, error)
	StartContainer   func(string, *docker.HostConfig) error
	BuildImage       func(docker.BuildImageOptions) error
	KillContainer    func(docker.KillContainerOptions) error
	StopContainer    func(string, uint) error
	InspectContainer func(string) (*docker.Container, error)
	ListContainers   func(docker.ListContainersOptions) ([]docker.APIContainers, error)
	ListImages       func(bool) ([]docker.APIImages, error)
	RemoveContainer  func(docker.RemoveContainerOptions) error
	WaitContainer    func(string) (int, error)
	RemoveImage      func(string) error
}

func NewDocker(endpoint string) *DockerWrapper {
	client, err := docker.NewClient(endpoint)
	if err != nil {
		logs.Assert(err, "Docker")
	}
	d := &DockerWrapper{Client: client}
	var makeDockerWrapper func(*DockerWrapper, *docker.Client) *DockerWrapper
	utils.MakeWrapper(&makeDockerWrapper)
	return makeDockerWrapper(d, client)
}

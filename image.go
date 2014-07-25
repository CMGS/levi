package main

import (
	"github.com/fsouza/go-dockerclient"
	"os"
	"strings"
)

type Image struct {
	client      *docker.Client
	appname     string
	version     string
	config_path string
	port        int
}

func (self *Image) Pull(registry *string) error {
	url := strings.Join([]string{*registry, self.appname}, "/")
	if err := self.client.PullImage(
		docker.PullImageOptions{url, *registry, self.version, os.Stdout},
		docker.AuthConfiguration{}); err != nil {
		return err
	}
	return nil
}

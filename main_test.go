package main

import (
	"github.com/fsouza/go-dockerclient"
)

type builderTestDocker struct{ *docker.Client }

func (d builderTestDocker) BuildImage(opts docker.BuildImageOptions) error {
	return nil //FIXME
}

func (d builderTestDocker) PushImage(opts docker.PushImageOptions, auth docker.AuthConfiguration) error {
	return nil //FIXME
}

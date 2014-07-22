package main

import (
	"github.com/fsouza/go-dockerclient"
	"log"
	"strings"
)

func load_containers() []docker.APIContainers {
	docker_client, err := docker.NewClient(DOCKER)
	if err != nil {
		log.Fatal("Connect docker failed")
	}
	containers, err := docker_client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		log.Fatal("Query docker failed")
	}
	return containers
}

func process_containers(containers *[]docker.APIContainers) map[string][]string {
	ret := make(map[string][]string)
	for _, container := range *containers {
		names := container.Names[0]
		split_names := strings.SplitN(strings.TrimLeft(names, "/"), "_", 2)
		ret[split_names[0]] = append(ret[split_names[0]], split_names[1])
	}
	return ret
}

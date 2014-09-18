package main

import (
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v1"
	"path"
	"strings"
)

func getPath(name string) string {
	name = strings.TrimLeft(name, "/")
	if pos := strings.Index(name, "_daemon_"); pos > -1 {
		return path.Join("/NBE/_Apps", name[:pos], "daemons")
	}
	if pos := strings.Index(name, "_test_"); pos > -1 {
		return path.Join("/NBE/_Apps", name[:pos], "tests")
	}
	info := strings.Split(name, "_")
	appname := name[:strings.Index(name, info[len(info)-1])-1]
	return path.Join("/NBE/_Apps", appname, "apps")
}

func Start(id string) error {
	container, err := Docker.InspectContainer(id)
	if err != nil {
		return err
	}
	p := getPath(container.Name)
	logger.Debug(container.Name, p)
	resp, err := Etcd.Get(p, false, false)
	if err != nil {
		return err
	}
	record := map[string]map[string]*docker.Container{}
	if resp.Node.Value == "" {
		record[config.Name] = map[string]*docker.Container{}
		record[config.Name][id] = container
	} else {
		yaml.Unmarshal([]byte(resp.Node.Value), &record)
		record[config.Name][id] = container
	}
	logger.Debug(p, record)
	out, err := yaml.Marshal(record)
	if err != nil {
		return err
	}
	Etcd.Set(p, string(out), 0)
	return nil
}

func Die(id, name string) error {
	p := getPath(name)
	resp, err := Etcd.Get(p, false, false)
	if err != nil {
		return err
	}
	if resp.Node.Value != "" {
		record := map[string]map[string]*docker.Container{}
		yaml.Unmarshal([]byte(resp.Node.Value), &record)
		delete(record[config.Name], id)
		if len(record[config.Name]) == 0 {
			delete(record, config.Name)
		}
		logger.Debug(p, record)
		if len(record) == 0 {
			Etcd.Set(p, "", 0)
		} else {
			out, err := yaml.Marshal(record)
			if err != nil {
				return err
			}
			Etcd.Set(p, string(out), 0)
		}
	}
	return nil
}

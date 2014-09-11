package main

import (
	"github.com/CMGS/go-dockerclient"
	"gopkg.in/yaml.v1"
	"path"
	"strings"
)

func getPath(name string) string {
	appinfo := strings.SplitN(strings.TrimLeft(name, "/"), "_", 2)
	var p string
	if strings.Contains(appinfo[1], "daemon_") {
		p = path.Join("/NBE/_Apps", appinfo[0], "daemons")
	} else {
		p = path.Join("/NBE/_Apps", appinfo[0], "apps")
	}
	return p
}

func Start(id string) error {
	container, err := Docker.InspectContainer(id)
	if err != nil {
		return err
	}
	p := getPath(container.Name)
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

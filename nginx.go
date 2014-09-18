package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"
)

type Upstream struct {
	Appname string
	Ports   map[string]string
}

type Nginx struct {
	upstreams map[string]*Upstream
	update    map[string]struct{}
}

func NewNginx() *Nginx {
	nginx := &Nginx{
		make(map[string]*Upstream),
		make(map[string]struct{}),
	}
	containers, err := Docker.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		logger.Info(err)
	}
	for _, container := range containers {
		name := strings.TrimLeft(container.Names[0], "/")
		if pos := strings.Index(name, "_daemon_"); pos > -1 {
			continue
		}
		if pos := strings.Index(name, "_test_"); pos > -1 {
			continue
		}
		info := strings.Split(name, "_")
		appname := name[:strings.Index(name, info[len(info)-1])-1]
		appport := name[strings.Index(name, info[len(info)-1]):]
		nginx.New(appname, container.ID, appport)
	}
	return nginx
}

func (self *Nginx) SetUpdate(appname string) {
	_, ok := self.update[appname]
	if !ok {
		self.update[appname] = struct{}{}
	}
}

func (self *Nginx) New(appname, cid, port string) {
	if self.upstreams[appname] == nil {
		self.upstreams[appname] = &Upstream{appname, make(map[string]string)}
	}
	self.upstreams[appname].Ports[cid] = port
}

func (self *Nginx) Remove(appname, cid string) bool {
	upstream, ok := self.upstreams[appname]
	if !ok {
		return false
	}
	if len(upstream.Ports) > 0 {
		delete(upstream.Ports, cid)
	}
	if len(upstream.Ports) == 0 {
		delete(self.upstreams, appname)
	}
	return true
}

func (self *Nginx) Clear(appname string) {
	var configPath = path.Join(config.Nginx.Configs, fmt.Sprintf("%s.conf", appname))
	logger.Info("Clear config file", configPath)
	if err := os.Remove(configPath); err != nil {
		logger.Info(err)
	}
}

func (self *Nginx) Save() {
	for appname, _ := range self.update {
		upstream, ok := self.upstreams[appname]
		if !ok {
			self.Clear(appname)
			go self.DeleteStream(appname)
			continue
		}
		go self.UpdateStream(upstream)
		var configPath = path.Join(config.Nginx.Configs, fmt.Sprintf("%s.conf", appname))
		f, err := os.Create(configPath)
		defer f.Close()
		if err != nil {
			logger.Info("Create upstream config failed", err)
		}
		tmpl := template.Must(template.ParseFiles(config.Nginx.Template))
		err = tmpl.Execute(f, upstream)
		if err != nil {
			logger.Info("Generate upstream config failed", err)
		}
	}
}

func (self *Nginx) DeleteStream(appname string) {
	url := UrlJoin(config.Nginx.DyUpstream, appname)
	logger.Debug(url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		logger.Info(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Info(err)
	} else {
		defer resp.Body.Close()
	}
}

func (self *Nginx) UpdateStream(upstream *Upstream) {
	url := UrlJoin(config.Nginx.DyUpstream, upstream.Appname)
	logger.Debug("Upstream Info", upstream.Ports, url)
	var s []string = []string{}
	for _, port := range upstream.Ports {
		s = append(s, fmt.Sprintf("server 127.0.0.1:%s", port))
	}
	data := fmt.Sprintf("%s;", strings.Join(s, ";"))
	logger.Debug("Upstream Data", data)
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data))
	if err != nil {
		logger.Info(err)
	} else {
		defer resp.Body.Close()
	}
}

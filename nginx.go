package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
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
		self.Clear(appname)
	}
	return true
}

func (self *Nginx) Clear(appname string) {
	var configPath = path.Join(NgxDir, fmt.Sprintf("%s.conf", appname))
	if err := os.Remove(configPath); err != nil {
		logger.Info(err)
	}
}

func (self *Nginx) Save() {
	for appname, _ := range self.update {
		upstream, ok := self.upstreams[appname]
		if !ok {
			continue
		}
		var configPath = path.Join(NgxDir, fmt.Sprintf("%s.conf", appname))
		f, err := os.Create(configPath)
		defer f.Close()
		if err != nil {
			logger.Info("Create upstream config failed", err)
		}
		tmpl := template.Must(template.ParseFiles(NgxTmpl))
		err = tmpl.Execute(f, upstream)
		if err != nil {
			logger.Info("Generate upstream config failed", err)
		}
	}
}

func (self *Nginx) Restart() {
	if len(self.update) == 0 {
		logger.Info("No nginx config file update, nginx will not restart")
		return
	}
	logger.Debug("Reload by", self.update)
	cmd := exec.Command(NgxEndpoint, "-s", "reload")
	err := cmd.Run()
	if err != nil {
		logger.Info("Restart nginx failed", err)
	}
}

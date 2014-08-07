package main

import (
	"fmt"
	"levi/logger"
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
}

func (self *Nginx) New(appname, cid, port string) {
	if self.upstreams[appname] == nil {
		self.upstreams[appname] = &Upstream{appname, make(map[string]string)}
	}
	self.upstreams[appname].Ports[cid] = port
}

func (self *Nginx) Remove(appname, cid string) {
	upstream, ok := self.upstreams[appname]
	if !ok {
		return
	}
	if len(upstream.Ports) > 0 {
		delete(upstream.Ports, cid)
	}
	if len(upstream.Ports) == 0 {
		delete(self.upstreams, appname)
	}
}

func (self *Nginx) Save() {
	for appname, upstream := range self.upstreams {
		conf_path := path.Join(ngx_dir, fmt.Sprintf("%s.conf", appname))
		f, err := os.Create(conf_path)
		defer f.Close()
		if err != nil {
			logger.Info("Create upstream conf failed", err)
		}
		tmpl := template.Must(template.ParseFiles(ngx_tmpl))
		err = tmpl.Execute(f, upstream)
		if err != nil {
			logger.Info("Generate upstream conf failed", err)
		}
	}
}

func (self *Nginx) Restart() {
	cmd := exec.Command(ngx_endpoint, "-s", "reload")
	err := cmd.Run()
	if err != nil {
		logger.Info("Restart nginx failed", err)
	}
}

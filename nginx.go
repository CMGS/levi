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
	bin, conf_path string
	upstreams      map[string]*Upstream
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
		conf_path := path.Join(self.conf_path, fmt.Sprintf("%s.conf", appname))
		f, err := os.Create(conf_path)
		defer f.Close()
		if err != nil {
			fmt.Println("Create upstream conf failed", err)
		}
		tmpl := template.Must(template.ParseFiles(NGINX_CONF_TMPL))
		err = tmpl.Execute(f, upstream)
		if err != nil {
			fmt.Println("Generate upstream conf failed", err)
		}
	}
}

func (self *Nginx) Restart() {
	cmd := exec.Command(self.bin, "-s", "reload")
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
	}
}

package main

import (
	"fmt"
	"os"
	"text/template"
)

type Upstream struct {
	AppName string
	Ports   []string
}

func (self *Upstream) Append(port string) {
	self.Ports = append(self.Ports, port)
}

func (self *Upstream) Execute(path string) bool {
	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return false
	}
	tmpl := template.Must(template.ParseFiles(NGINX_CONF_TMPL))
	err = tmpl.Execute(f, self)
	if err != nil {
		fmt.Println("Generate upstream conf failed", err)
		return false
	}
	return true
}

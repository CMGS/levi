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

func (self *Upstream) Execute(dst string) bool {
	tmpl := template.Must(template.ParseFiles("etc/site.tmpl"))
	err := tmpl.Execute(os.Stdout, self)
	if err != nil {
		fmt.Println("Parse site.tmpl Failed", err)
		return false
	}
	return true
}

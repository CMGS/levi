package utils

import (
	"log"
)

var Logger *logger = &logger{}

type logger struct {
	Mode bool
}

func (self *logger) Info(v ...interface{}) {
	log.Println(v...)
}

func (self *logger) Debug(v ...interface{}) {
	if self.Mode {
		log.Println(v...)
	}
}

func (self *logger) Assert(err error, context string) {
	if err != nil {
		log.Fatal(context+": ", err)
	}
}

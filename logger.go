package main

import (
	"log"
)

type Logger struct {
	Mode bool
}

func (self *Logger) Info(v ...interface{}) {
	log.Println(v...)
}

func (self *Logger) Debug(v ...interface{}) {
	if self.Mode {
		log.Println(v...)
	}
}

func (self *Logger) Assert(err error, context string) {
	if err != nil {
		log.Fatal(context+": ", err)
	}
}

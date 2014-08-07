package logger

import (
	"log"
)

var DebugMode bool

func Info(v ...interface{}) {
	log.Println(v...)
}

func Debug(v ...interface{}) {
	if DebugMode {
		log.Println(v...)
	}
}

func Assert(err error, context string) {
	if err != nil {
		log.Fatal(context+": ", err)
	}
}

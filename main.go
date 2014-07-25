package main

import (
	"flag"
	"fmt"
	"github.com/CMGS/websocket"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func exit(levi *Levi, c chan os.Signal) {
	fmt.Println("Catch", <-c)
	levi.Close()
}

func main() {
	var addr = flag.String("addr", "ws://127.0.0.1:8888/", "master service address")
	var wait = flag.Int("wait", 15, "wait task time")
	var sleep = flag.Int("sleep", 15, "report sleep time")
	var num = flag.Int("num", 3, "max tasks")
	var url = flag.String("url", "unix:///var/run/docker.sock", "docker url")
	var dst = flag.String("dst", "/tmp", "nginx conf dir")
	var ngx = flag.String("nginx", "/usr/local/nginx/sbin/nginx", "nginx location")
	var c = make(chan os.Signal, 1)
	var dialer = websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	var levi = Levi{}

	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go exit(&levi, c)

	flag.Parse()
	header := http.Header{}
	ws, _, err := dialer.Dial(*addr, header)
	if err != nil {
		log.Fatal("Connect: ", err)
		return
	}
	defer ws.Close()
	levi.Connect(url)
	levi.Load()
	go levi.Report(ws, sleep)
	levi.Loop(ws, wait, num, dst, ngx)
}

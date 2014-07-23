package main

import (
	"flag"
	"github.com/CMGS/websocket"
	"log"
	"net/http"
)

func main() {
	var addr = flag.String("addr", "ws://127.0.0.1:8888/", "master service address")
	var sleep = flag.Int("sleep", 15, "merge task time")
	var num = flag.Int("num", 3, "max tasks")
	var docker_url = flag.String("url", "unix:///var/run/docker.sock", "docker url")
	var dst_dir = flag.String("dst", "/tmp", "nginx conf dir")

	var dialer = websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	flag.Parse()
	header := http.Header{}
	ws, _, err := dialer.Dial(*addr, header)
	if err != nil {
		log.Fatal("Connect: ", err)
		return
	}
	defer ws.Close()
	levi := Levi{}
	levi.Connect(docker_url)
	levi.Parse()
	levi.Loop(ws, sleep, num, dst_dir)
}

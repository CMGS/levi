package main

import (
	"flag"
	"github.com/CMGS/websocket"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func exit(levi *Levi, c chan os.Signal) {
	logger.Info("Catch", <-c)
	levi.Close()
}

func main() {
	var MasterEndpoint, DockerEndpoint string
	var TaskWait, ReportSleep, TaskNum int

	flag.StringVar(&MasterEndpoint, "addr", "ws://127.0.0.1:8888/", "master service address")
	flag.StringVar(&DockerEndpoint, "endpoint", "unix:///var/run/docker.sock", "docker endpoint")
	flag.StringVar(&NgxEndpoint, "nginx-endpoint", "/usr/local/nginx/sbin/nginx", "nginx location")
	flag.StringVar(&RegEndpoint, "registry", "127.0.0.1", "registry location")
	flag.StringVar(&NgxDir, "nginx-dir", "/tmp", "nginx conf dir")
	flag.StringVar(&NgxTmpl, "nginx-tmpl", "/etc/site.tmpl", "nginx config file template location")
	flag.StringVar(&NetworkMode, "network", "bridge", "network mode")
	flag.StringVar(&Permdirs, "permdirs", "/mnt/mfs/permdirs", "permdirs location")
	flag.StringVar(&HomePath, "home", "/tmp", "homes dir path")
	flag.BoolVar(&logger.Mode, "DEBUG", false, "enable debug")
	flag.IntVar(&TaskWait, "wait", 15, "wait task time")
	flag.IntVar(&ReportSleep, "sleep", 15, "report sleep time")
	flag.IntVar(&TaskNum, "num", 3, "max tasks")
	flag.Parse()

	var levi = Levi{}
	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go exit(&levi, c)

	var dialer = websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, _, err := dialer.Dial(MasterEndpoint, http.Header{})
	if err != nil {
		logger.Assert(err, "Master")
	}
	defer ws.Close()

	levi.Connect(DockerEndpoint)
	levi.Load()
	go levi.Report(ws, ReportSleep)
	levi.Loop(ws, TaskNum, TaskNum)
}

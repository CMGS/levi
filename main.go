package main

import (
	"flag"
	"github.com/CMGS/websocket"
	"levi/logger"
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
	var master_endpoint, docker_endpoint string
	var levi_wait, levi_sleep, levi_num int

	flag.StringVar(&master_endpoint, "addr", "ws://127.0.0.1:8888/", "master service address")
	flag.StringVar(&docker_endpoint, "endpoint", "unix:///var/run/docker.sock", "docker endpoint")
	flag.StringVar(&ngx_dir, "nginx-dir", "/tmp", "nginx conf dir")
	flag.StringVar(&ngx_tmpl, "nginx-tmpl", "/etc/site.tmpl", "nginx config file template location")
	flag.StringVar(&ngx_endpoint, "nginx-endpoint", "/usr/local/nginx/sbin/nginx", "nginx location")
	flag.StringVar(&reg_endpoint, "registry", "127.0.0.1", "registry location")
	flag.StringVar(&network_mode, "network", "bridge", "network mode")
	flag.StringVar(&permdirs, "permdirs", "/mnt/mfs/permdirs", "permdirs location")
	flag.StringVar(&home_path, "home", "/tmp", "homes dir path")
	flag.BoolVar(&logger.DebugMode, "DEBUG", false, "enable debug")
	flag.IntVar(&levi_wait, "wait", 15, "wait task time")
	flag.IntVar(&levi_sleep, "sleep", 15, "report sleep time")
	flag.IntVar(&levi_num, "num", 3, "max tasks")
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

	ws, _, err := dialer.Dial(master_endpoint, http.Header{})
	if err != nil {
		logger.Assert(err, "Master")
	}
	defer ws.Close()

	levi.Connect(docker_endpoint)
	levi.Load()
	go levi.Report(ws, levi_sleep)
	//levi.Loop(ws, wait, num, dst, ngx, registry)
	levi.Loop(ws, levi_num, levi_wait)
}

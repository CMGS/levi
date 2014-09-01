package main

import (
	"flag"
	"github.com/CMGS/websocket"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func exit(levi *Levi, c chan os.Signal) {
	logger.Info("Catch", <-c)
	levi.Close()
}

func pid(path string) {
	if err := ioutil.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0755); err != nil {
		logger.Info("Save pid file failed", err)
	}
}

func main() {
	var masterEndpoint, dockerEndpoint string
	var taskWait, reportSleep, taskNum int
	var pidFile, etcdMachines string

	flag.StringVar(&masterEndpoint, "addr", "ws://127.0.0.1:8888/", "master service address")
	flag.StringVar(&dockerEndpoint, "endpoint", "unix:///var/run/docker.sock", "docker endpoint")
	flag.StringVar(&GitEndpoint, "git", "http://git.hunantv.com", "git repos endpoint")
	flag.StringVar(&GitWorkDir, "git-work-dir", "/tmp", "where to save repos")
	flag.StringVar(&DyUpstreamUrl, "upstream-url", "http://127.0.0.1:10090/upstream", "nginx dynamic upstream url")
	flag.StringVar(&RegEndpoint, "registry", "127.0.0.1", "registry location")
	flag.StringVar(&NgxDir, "nginx-dir", "/tmp", "nginx conf dir")
	flag.StringVar(&NgxTmpl, "nginx-tmpl", "/etc/site.tmpl", "nginx config file template location")
	flag.StringVar(&NetworkMode, "network", "bridge", "network mode")
	flag.StringVar(&Permdirs, "permdirs", "/mnt/mfs/permdirs", "permdirs location")
	flag.StringVar(&HomePath, "home", "/tmp", "homes dir path")
	flag.StringVar(&pidFile, "pidfile", "/var/run/levi.pid", "pid file")
	flag.StringVar(&etcdMachines, "etcd", "http://127.0.0.1:4001", "etcd machines, seprate by comma")
	flag.BoolVar(&logger.Mode, "DEBUG", false, "enable debug")
	flag.IntVar(&taskWait, "wait", 15, "wait task time")
	flag.IntVar(&reportSleep, "sleep", 15, "report sleep time")
	flag.IntVar(&taskNum, "num", 3, "max tasks")
	flag.Parse()

	Etcd = NewEtcdClient(strings.Split(etcdMachines, ","))
	var levi = Levi{}
	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGHUP)
	go exit(&levi, c)

	var dialer = websocket.Dialer{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, _, err := dialer.Dial(masterEndpoint, http.Header{})
	if err != nil {
		logger.Assert(err, "Master")
	}
	pid(pidFile)

	defer os.Remove(pidFile)
	defer ws.Close()

	levi.Connect(dockerEndpoint)
	levi.Load()
	go levi.Report(ws, reportSleep)
	levi.Loop(ws, taskNum, taskNum)
}

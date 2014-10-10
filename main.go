package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	LoadConfig()
	Etcd = NewEtcd(config.Etcd.Machines)
	Docker = NewDocker(config.Docker.Endpoint)
	Status = NewStatus()

	defer os.Remove(config.PidFile)
	WritePid(config.PidFile)
	RunLenz()

	Ws = NewWebSocket(config.Master)
	defer Ws.Close()

	levi := NewLevi()
	go Status.Listen()
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)
		signal.Notify(c, syscall.SIGHUP)
		signal.Notify(c, syscall.SIGKILL)
		logger.Info("Catch", <-c)
		levi.Exit()
	}()
	go levi.Read()
	levi.Loop()
}

package main

import (
	"os"
	"strings"

	"./defines"
	"./lenz"
	. "./utils"
)

type Lenz struct {
	Attacher *lenz.AttachManager
	Router   *lenz.RouteManager
	Routefs  lenz.RouteFileStore
}

func NewLenz() *Lenz {
	obj := &Lenz{}
	obj.Attacher = lenz.NewAttachManager(Docker)
	obj.Router = lenz.NewRouteManager(obj.Attacher)
	obj.Routefs = lenz.RouteFileStore(config.Lenz.Routes)

	if config.Lenz.Forwards != "" {
		Logger.Info("Routing all to", config.Lenz.Forwards)
		target := defines.Target{Addrs: strings.Split(config.Lenz.Forwards, ",")}
		route := defines.Route{ID: "lenz_default", Target: &target}
		route.LoadBackends()
		obj.Router.Add(&route)
	}

	if _, err := os.Stat(config.Lenz.Routes); err == nil {
		Logger.Info("Loading and persisting routes in", config.Lenz.Routes)
		Logger.Assert(obj.Router.Load(obj.Routefs), "persistor")
	}
	return obj
}

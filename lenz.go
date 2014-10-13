package main

import (
	"os"

	"./defines"
	"./lenz"
	. "./utils"
)

type LenzForwarder struct {
	Attacher *lenz.AttachManager
	Router   *lenz.RouteManager
	Routefs  lenz.RouteFileStore
}

func NewLenz() *LenzForwarder {
	obj := &LenzForwarder{}
	obj.Attacher = lenz.NewAttachManager(Docker)
	obj.Router = lenz.NewRouteManager(obj.Attacher, config.Lenz.Stdout)
	obj.Routefs = lenz.RouteFileStore(config.Lenz.Routes)

	if len(config.Lenz.Forwards) > 0 {
		Logger.Info("Routing all to", config.Lenz.Forwards)
		target := defines.Target{Addrs: config.Lenz.Forwards}
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

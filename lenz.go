package main

import (
	"os"
	"strings"
)

type Lenz struct {
	attacher *AttachManager
	router   *RouteManager
	routefs  RouteFileStore
}

func RunLenz() {
	lenz := &Lenz{}
	lenz.attacher = NewAttachManager()
	lenz.router = NewRouteManager(lenz.attacher)
	lenz.routefs = RouteFileStore(config.Lenz.Routes)

	if config.Lenz.Forwards != "" {
		logger.Info("Routing all to", config.Lenz.Forwards)
		target := Target{Addrs: strings.Split(config.Lenz.Forwards, ",")}
		route := Route{ID: "lenz_default", Target: &target}
		route.loadBackends()
		lenz.router.Add(&route)
	}

	if _, err := os.Stat(config.Lenz.Routes); err == nil {
		logger.Info("Loading and persisting routes in", config.Lenz.Routes)
		logger.Assert(lenz.router.Load(lenz.routefs), "persistor")
	}
}

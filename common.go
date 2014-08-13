package main

const (
	ADD_CONTAINER          = 1
	REMOVE_CONTAINER       = 2
	UPDATE_CONTAINER       = 3
	CONTAINER_STOP_TIMEOUT = 15
)

var logger = Logger{}

var NgxDir, NgxTmpl, DyUpstreamUrl, RegEndpoint string
var NetworkMode, Permdirs, HomePath, PidFile string

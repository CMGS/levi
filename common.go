package main

const (
	CONTAINER_STOP_TIMEOUT = 15
	PRODUCTION             = "PROD"
	TESTING                = "TEST"
	STATUS_IDENT           = "__STATUS__"
	STATUS_DIE             = "die"
	STATUS_START           = "start"
)

var logger = Logger{}

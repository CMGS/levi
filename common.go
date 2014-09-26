package main

const (
	ADD_CONTAINER          = 1
	REMOVE_CONTAINER       = 2
	UPDATE_CONTAINER       = 3
	BUILD_IMAGE            = 4
	TEST_IMAGE             = 5
	DOCKER_INFO            = 6
	CONTAINER_STOP_TIMEOUT = 15
	PRODUCTION             = "PROD"
	TESTING                = "TEST"
	STATUS_IDENT           = "__status__"
	STATUS_DIE             = "die"
)

var logger = Logger{}

package main

const (
	ADD_CONTAINER          = 1
	REMOVE_CONTAINER       = 2
	UPDATE_CONTAINER       = 3
	BUILD_IMAGE            = 4
	CONTAINER_STOP_TIMEOUT = 15
)

var logger = Logger{}

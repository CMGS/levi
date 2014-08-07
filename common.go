package main

const (
	ADD_CONTAINER          = 1
	REMOVE_CONTAINER       = 2
	UPDATE_CONTAINER       = 3
	CONTAINER_STOP_TIMEOUT = 15
)

var ngx_dir, ngx_endpoint, reg_endpoint string
var network_mode, permdirs, ngx_tmpl, home_path string

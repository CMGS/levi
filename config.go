package main

const (
	//DOCKER = "unix:///var/run/docker.sock"
	LOCAL_DOCKER = "http://10.1.201.16"
)

var Type map[int]string = map[int]string{
	1: "ADD",
	2: "REMOVE",
	3: "UPDATE",
	4: "KEEP",
}

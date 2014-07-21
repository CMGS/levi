package main

import (
	"fmt"
	"github.com/CMGS/websocket"
	"log"
	"net"
	"time"
)

func doJobs(jobs *[]Message) {
	fmt.Println("Do jobs")
	fmt.Println(jobs)
}

func getJobs(ws *websocket.Conn, sleep *int, num *int) {
	jobs := []Message{}
	for {
		got_message := false
		message := Message{}
		ws.SetReadDeadline(time.Now().Add(time.Duration(*sleep) * time.Second))
		fmt.Println(time.Now())
		if err := ws.ReadJSON(&message); err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				log.Fatal("Read Fail:", err)
			}
		} else {
			got_message = true
		}
		switch {
		case !got_message && len(jobs) != 0:
			doJobs(&jobs)
			jobs = []Message{}
		case got_message:
			jobs = append(jobs, message)
			if len(jobs) >= *num {
				doJobs(&jobs)
				jobs = []Message{}
			}
		}
	}
}

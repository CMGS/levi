package main

import (
	"fmt"
	"github.com/CMGS/websocket"
	"log"
	"net"
	"time"
)

func checkTasks(h *Taskhub) bool {
	select {
	case tid := <-h.done:
		h.task[tid] = true
		fmt.Println(tid, "done")
	}
	for _, done := range h.task {
		if !done {
			return false
		}
	}
	return true
}

func doTasks(tasks *[]AppTask) {
	h := Taskhub{
		task: make(map[string]bool),
		done: make(chan string),
	}

	for _, task := range *tasks {
		h.task[task.Id] = false
		fmt.Println("Process", task.Name)
		go func(t AppTask) {
			//TODO add/remove/update container
			for _, job := range t.Tasks {
				fmt.Println("process", Type[job.Type], t.Name)
			}
			h.done <- t.Id
		}(task)
	}

	for {
		ret := checkTasks(&h)
		if ret {
			break
		}
	}
}

func getTasks(ws *websocket.Conn, sleep *int, num *int) {
	ws.SetPingHandler(nil)
	tasks := []AppTask{}
	for {
		got_task := false
		apptask := AppTask{}
		ws.SetReadDeadline(time.Now().Add(time.Duration(*sleep) * time.Second))
		fmt.Println(time.Now())
		if err := ws.ReadJSON(&apptask); err != nil {
			if e, ok := err.(net.Error); !ok || !e.Timeout() {
				log.Fatal("Read Fail:", err)
			}
		} else {
			got_task = true
		}
		switch {
		case !got_task && len(tasks) != 0:
			doTasks(&tasks)
			tasks = []AppTask{}
		case got_task:
			tasks = append(tasks, apptask)
			if len(tasks) >= *num {
				doTasks(&tasks)
				tasks = []AppTask{}
			}
		}
	}
}

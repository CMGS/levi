package main

import (
	"testing"
)

func Test_SetAddTaskType(t *testing.T) {
	task := AddTask{
		Bind:   9999,
		Daemon: "abc",
		Test:   "def",
	}
	task.SetAsTest()
	if !task.IsTest() || task.ident != "test_def" {
		t.Error("Test ident invaild")
	}
	task.SetAsDaemon()
	if !task.IsDaemon() || task.ident != "daemon_abc" {
		t.Error("Daemon ident invaild")
	}
	task.Daemon = ""
	task.Test = ""
	task.SetAsService()
	if !task.ShouldExpose() || task.ident != "9999" {
		t.Error("Service ident invaild")
	}
}

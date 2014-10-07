package main

import (
	"testing"
)

func Test_SetType(t *testing.T) {
	task := AddTask{
		Bind:   9999,
		Daemon: "abc",
		Test:   "def",
	}
	task.SetAsTest()
	if !task.CheckTest() || !task.IsTest() {
		t.Error("Test ident invaild")
	}
	task.SetAsDaemon()
	if !task.CheckDaemon() || !task.IsDaemon() {
		t.Error("Daemon ident invaild")
	}
	task.Daemon = ""
	task.Test = ""
	task.SetAsService()
	if !task.ShouldExpose() || task.ident != "9999" {
		t.Error("Service ident invaild")
	}
}

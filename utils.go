package main

import (
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type FakeOut struct{}

func (self FakeOut) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func UrlJoin(strs ...string) string {
	ss := make([]string, len(strs))
	for i, s := range strs {
		if i == 0 {
			ss[i] = strings.TrimRight(s, "/")
		} else {
			ss[i] = strings.TrimLeft(s, "/")
		}
	}
	return strings.Join(ss, "/")
}

func WritePid(path string) {
	if err := ioutil.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0755); err != nil {
		logger.Assert(err, "Save pid file failed")
	}
}

func GetBuffer() io.Writer {
	if logger.Mode {
		return os.Stdout
	} else {
		return FakeOut{}
	}
}

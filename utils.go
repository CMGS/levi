package main

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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

func CopyFiles(dst, src string) error {
	logger.Debug("static src: ", src)
	logger.Debug("static dst: ", dst)
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if info.IsDir() {
			e := os.Mkdir(path.Join(dst, p), info.Mode())
			return e
		} else {
			d, e := os.Create(path.Join(dst, p))
			defer d.Close()
			if e != nil {
				return e
			}

			f, e := os.Open(p)
			defer f.Close()
			if e != nil {
				return e
			}

			io.Copy(d, f)
		}
		return err
	})
}

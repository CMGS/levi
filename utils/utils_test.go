package utils

import (
	"io"
	"os"
	"testing"
)

func Test_FakeOutWrite(t *testing.T) {
	f := FakeOut{}
	if l, err := f.Write(make([]byte, 5)); err != nil {
		t.Error(err)
		if l != 5 {
			t.Error("Length invaild")
		}
	}
}

func Test_UrlJoin(t *testing.T) {
	strs := []string{"http://a.b.c", "d", "e"}
	ss := UrlJoin(strs...)
	if ss != "http://a.b.c/d/e" {
		t.Error("Join invaild")
	}
}

func Test_WritePid(t *testing.T) {
	p := "/tmp/test.pid"
	WritePid(p)
	defer os.RemoveAll(p)
	if _, err := os.Stat(p); err != nil {
		t.Error(err)
	}
}

func Test_GetBuffer(t *testing.T) {
	f := func(mod bool, o io.Writer) {
		Logger.Mode = mod
		b := GetBuffer()
		if b != o {
			t.Error("Buffer invaild")
		}
	}
	f(true, os.Stdout)
	f(false, FakeOut{})
}

func Test_CopyDir(t *testing.T) {
	defer func() {
		os.RemoveAll("/tmp/t1")
		os.RemoveAll("/tmp/t2")
	}()
	if err := MakeDir("/tmp/t1"); err != nil {
		t.Error(err)
	}
	if err := CopyDir("/tmp/t1", "/tmp/t2"); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat("/tmp/t2"); err != nil {
		t.Error(err)
	}
}

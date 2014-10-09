package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

var builder *Builder
var info *BuildTask

func init() {
	InitTest()
	info = &BuildTask{
		Group:   "platform",
		Name:    "nbetest",
		Version: "082d405",
		Base:    fmt.Sprintf("%s/nbeimage/ubuntu:python-2014.9.18", config.Docker.Registry),
		Build:   "pip install -r requirements.txt",
		Static:  "static",
	}
	builder = NewBuilder("nbetest", info)
}

func Test_FetchCode(t *testing.T) {
	if err := builder.fetchCode(); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(path.Join(builder.extendDir, info.Static)); err != nil {
		t.Error(err)
	}
}

func Test_CreateDockerFile(t *testing.T) {
	if err := builder.createDockerFile(); err != nil {
		t.Error(err)
	}
	if fi, err := os.Open(builder.dockerFilePath); err != nil {
		t.Error(err)
	} else {
		defer fi.Close()
		fd, _ := ioutil.ReadAll(fi)
		content := strings.Split(string(fd), "\n")
		if !strings.HasSuffix(content[len(content)-2], info.Build) {
			t.Error("Docker file not vaild")
		}
	}
}

func Test_CreateTar(t *testing.T) {
	if _, err := os.Stat(builder.tarPath); err == nil {
		t.Fatal("Tar exists")
	}
	if err := builder.createTar(); err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(builder.tarPath); err != nil {
		t.Error(err)
	}
}

func Test_BuildImage(t *testing.T) {
	if err := builder.buildImage(); err != nil {
		t.Error(err)
	}
}

func Test_PushImage(t *testing.T) {
	if err := builder.pushImage(); err != nil {
		t.Error(err)
	}
}

func Test_BuildClean(t *testing.T) {
	builder.clear()
	if _, err := os.Stat(builder.workDir); err == nil {
		t.Error("Clean work dir failed")
	}
	images, err := Docker.ListImages(false)
	if err != nil {
		t.Error(err)
	}
	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == "<none>:<none>" || tag == builder.repoTag {
				t.Fatal("Docker hub not clear")
			}
		}
	}
}

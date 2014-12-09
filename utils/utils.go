package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/fsouza/go-dockerclient"

	"../common"
	"../logs"
)

func RemoveContainer(id string, test bool, rmi bool) error {
	container, err := common.Docker.InspectContainer(id)
	if err != nil {
		return err
	}
	for p, rp := range container.Volumes {
		switch {
		case strings.HasSuffix(p, "/config.yaml"):
			if err := os.RemoveAll(rp); err != nil {
				return err
			}
		case test && strings.HasSuffix(p, "/permdir"):
			if err := os.RemoveAll(rp); err != nil {
				return err
			}
		}
	}
	if err := common.Docker.RemoveContainer(docker.RemoveContainerOptions{ID: id}); err != nil {
		return err
	}
	if rmi {
		if err := common.Docker.RemoveImage(container.Image); err != nil {
			return err
		}
	}
	return nil
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
		logs.Assert(err, "Save pid file failed")
	}
}

func MakeDir(p string) error {
	if err := os.MkdirAll(p, 0755); err != nil {
		return err
	}
	return nil
}

func CopyDir(source string, dest string) (err error) {
	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = CopyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return
}

func CopyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}
	return
}

func Marshal(obj interface{}) []byte {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		logs.Info("Utils Marshal:", err)
	}
	return bytes
}

func Unmarshal(input io.ReadCloser, obj interface{}) error {
	body, err := ioutil.ReadAll(input)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, obj)
	if err != nil {
		return err
	}
	return nil
}

func GetAppInfo(containerName string) (appname string, appid string, apptyp string) {
	containerName = strings.TrimLeft(containerName, "/")
	if pos := strings.LastIndex(containerName, "_daemon_"); pos > -1 {
		appname = containerName[:pos]
		appid = containerName[pos+8:]
		apptyp = common.DAEMON_TYPE
		return
	}
	if pos := strings.LastIndex(containerName, "_test_"); pos > -1 {
		appname = containerName[:pos]
		appid = containerName[pos+6:]
		apptyp = common.TEST_TYPE
		return
	}
	appinfo := strings.Split(containerName, "_")
	appname = strings.Join(appinfo[:len(appinfo)-1], "_")
	appid = appinfo[len(appinfo)-1]
	apptyp = common.DEFAULT_TYPE
	return
}

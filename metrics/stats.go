package metrics

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"../defines"
	"../logs"
	"github.com/fsouza/go-dockerclient"
)

var devDir string = ""

func getLongID(shortID string) (parentName string, longID string, err error) {
	pat := filepath.Join(devDir, "*", fmt.Sprintf("*%s*", shortID), "tasks")
	a, err := filepath.Glob(pat)
	if err != nil {
		return
	}
	if len(a) != 1 {
		return "", "", fmt.Errorf("Get Long ID Failed %s", shortID)
	}
	dir := filepath.Dir(a[0])
	longID = filepath.Base(dir)
	parentName = filepath.Base(filepath.Dir(dir))
	return
}

func GetNetStats(client *defines.DockerWrapper, cid string) (result map[string]uint64, err error) {
	var exec *docker.Exec
	exec, err = client.CreateExec(
		docker.CreateExecOptions{
			AttachStdout: true,
			Cmd: []string{
				"cat", "/proc/net/dev",
			},
			Container: cid,
		},
	)
	if err != nil {
		return
	}
	logs.Debug("Create exec id", exec.Id)
	outr, outw := io.Pipe()
	defer outr.Close()
	go func() {
		err = client.StartExec(
			exec.Id,
			docker.StartExecOptions{
				OutputStream: outw,
			},
		)
		outw.Close()
	}()
	result = map[string]uint64{}
	s := bufio.NewScanner(outr)
	var d uint64
	for i := 0; s.Scan(); {
		var name string
		var n [8]uint64
		text := s.Text()
		if strings.Index(text, ":") < 1 {
			continue
		}
		ts := strings.Split(text, ":")
		fmt.Sscanf(ts[0], "%s", &name)
		if strings.HasPrefix(name, "veth") || name == "lo" {
			continue
		}
		fmt.Sscanf(ts[1],
			"%d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d",
			&n[0], &n[1], &n[2], &n[3], &d, &d, &d, &d,
			&n[4], &n[5], &n[6], &n[7], &d, &d, &d, &d,
		)
		j := "." + strconv.Itoa(i)
		result["inbytes"+j] = n[0]
		result["inpackets"+j] = n[1]
		result["inerrs"+j] = n[2]
		result["indrop"+j] = n[3]
		result["outbytes"+j] = n[4]
		result["outpackets"+j] = n[5]
		result["outerrs"+j] = n[6]
		result["outdrop"+j] = n[7]
		i++
	}
	logs.Debug("Container net status", cid, result)
	return
}

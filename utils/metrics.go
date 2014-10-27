package utils

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/libcontainer/cgroups"
	"github.com/docker/libcontainer/cgroups/fs"
)

var devDir string = ""
var parentName string = ""

func GetCgroupStats(id string) (m *cgroups.Stats, err error) {
	if id, err = getLongID(id); err != nil {
		return
	}
	c := cgroups.Cgroup{
		Parent: parentName,
		Name:   id,
	}
	return fs.GetStats(&c)
}

func getLongID(shortID string) (longID string, err error) {
	if devDir == "" {
		devDir, err = cgroups.FindCgroupMountpoint("devices")
		if err != nil {
			return
		}
	}

	pat := filepath.Join(devDir, "*", "*"+shortID+"*", "tasks")
	a, err := filepath.Glob(pat)
	if err != nil {
		return
	}
	if len(a) != 1 {
		return "", fmt.Errorf("Get Long ID Failed %s", shortID)
	}
	dir := filepath.Dir(a[0])
	longID = filepath.Base(dir)
	parentName = filepath.Base(filepath.Dir(dir))
	return
}

func GetContainerPID(id string) (pid string, err error) {
	if devDir == "" {
		devDir, err = cgroups.FindCgroupMountpoint("devices")
		if err != nil {
			return
		}
	}

	pat := filepath.Join(devDir, "*", "*"+id+"*", "tasks")
	a, err := filepath.Glob(pat)
	if err != nil {
		return
	}
	if len(a) != 1 {
		return "", fmt.Errorf("Get Container PID Failed %s", id)
	}

	contents, err := ioutil.ReadFile(a[0])
	if err != nil {
		return
	}

	a = strings.Split(string(contents), "\n")
	return a[0], nil
}

func GetIfStats() (m map[string]interface{}, err error) {
	m = map[string]interface{}{}
	cmd := exec.Command("cat", "/proc/net/dev")
	f, err := cmd.StdoutPipe()
	//f, err := os.Open("/proc/net/dev")
	if err != nil {
		return
	}
	defer f.Close()
	err = cmd.Start()
	if err != nil {
		return
	}

	s := bufio.NewScanner(f)
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
		m["name"+j] = name
		m["inbytes"+j] = n[0]
		m["inpackets"+j] = n[1]
		m["inerrs"+j] = n[2]
		m["indrop"+j] = n[3]
		m["outbytes"+j] = n[4]
		m["outpackets"+j] = n[5]
		m["outerrs"+j] = n[6]
		m["outdrop"+j] = n[7]
		i++
	}
	err = cmd.Wait()
	return
}
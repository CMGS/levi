package metrics

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

var devDir string = ""

func getLongID(shortID string) (parentName string, longID string, pid string, err error) {
	pat := filepath.Join(devDir, "*", fmt.Sprintf("*%s*", shortID), "tasks")
	a, err := filepath.Glob(pat)
	if err != nil {
		return
	}
	if len(a) != 1 {
		return "", "", "", fmt.Errorf("Get Long ID Failed %s", shortID)
	}
	contents, err := ioutil.ReadFile(a[0])
	if err != nil {
		return
	}
	dir := filepath.Dir(a[0])
	longID = filepath.Base(dir)
	parentName = filepath.Base(filepath.Dir(dir))
	pid = strings.Split(string(contents), "\n")[0]
	return
}

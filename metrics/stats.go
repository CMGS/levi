package metrics

import (
	"fmt"
	"path/filepath"
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

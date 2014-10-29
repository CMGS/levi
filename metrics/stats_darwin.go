package metrics

import (
	"math/rand"
	"time"

	"../logs"
	"github.com/docker/libcontainer/cgroups"
)

func InitDevDir() {
	logs.Info("OSX not support cgroup mount point")
}

func GetCgroupStats(id string) (m *cgroups.Stats, err error) {
	logs.Info("OSX not support get cgroup stats")
	err = nil
	rand.Seed(time.Now().UnixNano())
	x := rand.Int63n(1e9)
	y := rand.Int63n(1e9)
	m = &cgroups.Stats{}
	m.CpuStats = cgroups.CpuStats{
		CpuUsage: cgroups.CpuUsage{
			TotalUsage:        uint64(x + y),
			UsageInUsermode:   uint64(x),
			UsageInKernelmode: uint64(y),
		},
	}
	s := map[string]uint64{}
	s["rss"] = uint64(rand.Int63n(9e10))
	m.MemoryStats = cgroups.MemoryStats{
		Usage: s["rss"] + uint64(rand.Int63n(9e10)),
		Stats: s,
	}
	return
}

func GetIfStats() (m map[string]interface{}, err error) {
	logs.Info("OSX not support get net stats")
	rand.Seed(time.Now().UnixNano())
	err = nil
	m = map[string]interface{}{}
	m["name.0"] = "darwin"
	m["inbytes.0"] = rand.Intn(10000)
	m["inpackets.0"] = rand.Int63n(100000)
	m["inerrs.0"] = 0
	m["indrop.0"] = 0
	m["outbytes.0"] = rand.Intn(10000)
	m["outpackets.0"] = rand.Int63n(100000)
	m["outerrs.0"] = 0
	m["outdrop.0"] = 0
	return
}

func NetNsSynchronize(pid string, fn func() error) (err error) {
	logs.Info("OSX not support Setns.")
	return fn()
}

func GetContainerPID(id string) (pid string, err error) {
	return "1111111", nil
}

package main

import (
	"sync"
	"time"

	. "./utils"
	"github.com/docker/libcontainer/cgroups"
)

type MetricData struct {
	cgroups.Stats
	Interfaces map[string]interface{}
}

type AppMetrics struct {
	name string
	cid  string
	typ  string
	stop chan bool
}

func (self *AppMetrics) Report() {
	defer close(self.stop)
	var finish bool = false
	for !finish {
		select {
		case <-time.After(time.Second * time.Duration(config.Metrics.ReportInterval)):
			m, err := self.generate()
			if err != nil {
				Logger.Info(err)
				continue
			}
			Logger.Debug(m)
		case finish = <-self.stop:
		}
	}
}

func (self *AppMetrics) generate() (*MetricData, error) {
	m := &MetricData{}
	pid, err := GetContainerPID(self.cid)
	if err != nil {
		return nil, err
	}
	c, err := GetCgroupStats(self.cid)
	if err != nil {
		return nil, err
	}
	m.MemoryStats = c.MemoryStats
	m.CpuStats = c.CpuStats
	m.BlkioStats = c.BlkioStats
	err = NetNsSynchronize(pid, func() (err error) {
		ifstats, err := GetIfStats()
		if err != nil {
			return err
		}
		m.Interfaces = ifstats
		return nil
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (self *AppMetrics) Stop() {
	self.stop <- true
}

type MetricsRecorder struct {
	sync.Mutex
	apps map[string]*AppMetrics
}

func NewMetricsRecorder() *MetricsRecorder {
	r := &MetricsRecorder{}
	r.apps = map[string]*AppMetrics{}
	return r
}

func (self *MetricsRecorder) Add(appname, cid, apptype string) {
	self.Lock()
	defer self.Unlock()
	if _, ok := self.apps[cid]; ok {
		return
	}
	self.apps[cid] = &AppMetrics{
		appname,
		cid,
		apptype,
		make(chan bool),
	}
	go self.apps[cid].Report()
}

func (self *MetricsRecorder) Remove(cid string) {
	self.Lock()
	defer self.Unlock()
	if _, ok := self.apps[cid]; !ok {
		return
	}
	self.apps[cid].Stop()
	delete(self.apps, cid)
}

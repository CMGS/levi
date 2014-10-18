package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	. "./utils"
	"github.com/cactus/go-statsd-client/statsd"
	"github.com/docker/libcontainer/cgroups"
)

type StatsdSender struct {
	memMax            string
	memCurrent        string
	cpuTotal          string
	interfaceInBytes  string
	interfaceOutBytes string
	client            *statsd.Client
}

func (self *StatsdSender) sendToStatsd(key string, value int64) {
	err := self.client.Timing(key, value, config.Metrics.Rate)
	if err != nil {
		Logger.Info("Sent to statsd failed", err, key, value)
	}
}

func (self *StatsdSender) Send(data *MetricData) {
	self.sendToStatsd(self.memMax, int64(data.MemoryStats.MaxUsage))
	self.sendToStatsd(self.memCurrent, int64(data.MemoryStats.Usage))
	self.sendToStatsd(self.cpuTotal, int64(data.CpuStats.CpuUsage.TotalUsage))
	iBytes, _ := strconv.ParseInt(fmt.Sprintf("%v", data.Interfaces["inbytes.0"]), 10, 64)
	oBytes, _ := strconv.ParseInt(fmt.Sprintf("%v", data.Interfaces["outbytes.0"]), 10, 64)
	self.sendToStatsd(self.interfaceInBytes, iBytes)
	self.sendToStatsd(self.interfaceOutBytes, oBytes)
}

func NewStatsdSender(appname, apptype string, client *statsd.Client) *StatsdSender {
	s := &StatsdSender{}
	s.memMax = fmt.Sprintf("%s.%s.mem.max", appname, apptype)
	s.memCurrent = fmt.Sprintf("%s.%s.mem.current", appname, apptype)
	s.cpuTotal = fmt.Sprintf("%s.%s.cpu.total", appname, apptype)
	s.interfaceInBytes = fmt.Sprintf("%s.%s.interfaces.inbytes", appname, apptype)
	s.interfaceOutBytes = fmt.Sprintf("%s.%s.interfaces.outbytes", appname, apptype)
	s.client = client
	return s
}

type MetricData struct {
	cgroups.Stats
	Interfaces map[string]interface{}
}

type AppMetrics struct {
	name string
	cid  string
	typ  string
	stop chan bool
	sync.WaitGroup
}

func (self *AppMetrics) Report(client *statsd.Client) {
	self.Add(1)
	defer self.Done()
	defer close(self.stop)
	var finish bool = false
	Logger.Info("Metrics Report", self.name, self.cid, self.typ)
	s := NewStatsdSender(self.name, self.typ, client)
	for !finish {
		select {
		case <-time.After(time.Second * time.Duration(config.Metrics.ReportInterval)):
			m, err := self.generate()
			if err != nil {
				Logger.Info(err)
				continue
			}
			s.Send(m)
		case f := <-self.stop:
			finish = f
		}
	}
	Logger.Info("Metrics Stop", self.name, self.cid, self.typ)
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
	self.Wait()
}

type MetricsRecorder struct {
	sync.Mutex
	apps   map[string]*AppMetrics
	client *statsd.Client
}

func NewMetricsRecorder() *MetricsRecorder {
	var err error
	r := &MetricsRecorder{}
	r.apps = map[string]*AppMetrics{}
	r.client, err = statsd.New(config.Metrics.Statsd, STATSD_NS)
	if err != nil {
		Logger.Assert(err, "Metrics init")
	}
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
		sync.WaitGroup{},
	}
	go self.apps[cid].Report(self.client)
}

func (self *MetricsRecorder) Stop(cid string) {
	self.Lock()
	defer self.Unlock()
	if _, ok := self.apps[cid]; !ok {
		return
	}
	self.apps[cid].Stop()
	delete(self.apps, cid)
}

func (self *MetricsRecorder) StopAll() {
	self.Lock()
	defer self.Unlock()
	defer self.client.Close()
	for _, r := range self.apps {
		r.Stop()
	}
}

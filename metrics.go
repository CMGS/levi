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

type CpuStats struct {
	user   uint64
	system uint64
}

func NewCpuStats(stats cgroups.CpuStats) CpuStats {
	c := CpuStats{}
	c.user = stats.CpuUsage.UsageInUsermode
	c.system = stats.CpuUsage.UsageInKernelmode
	return c
}

type InterfaceStats struct {
	inBytes  int64
	outBytes int64
}

func NewInterfaceStats(iStats map[string]interface{}) InterfaceStats {
	i := InterfaceStats{}
	iBytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["inbytes.0"]), 10, 64)
	oBytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["outbytes.0"]), 10, 64)
	i.inBytes = iBytes
	i.outBytes = oBytes
	return i
}

type StatsdSender struct {
	memCurrent        string
	memRss            string
	cpuUser           string
	cpuSystem         string
	interfaceInBytes  string
	interfaceOutBytes string
	client            *statsd.Client
}

func (self *StatsdSender) Gauge(key string, value int64) {
	err := self.client.Timing(key, value, config.Metrics.Rate)
	if err != nil {
		Logger.Info("Sent to statsd failed", err, key, value)
	}
}

func (self *StatsdSender) Send(data *MetricData) {
	self.Gauge(self.memCurrent, int64(data.memoryStats.Usage))
	self.Gauge(self.memRss, int64(data.memoryStats.Stats["rss"]))
	self.Gauge(self.cpuUser, int64(data.cpuStats.user))
	self.Gauge(self.cpuSystem, int64(data.cpuStats.system))

	if data.isApp {
		self.Gauge(self.interfaceInBytes, data.interfaceStats.inBytes)
		self.Gauge(self.interfaceOutBytes, data.interfaceStats.outBytes)
	}
}

func NewStatsdSender(appname, apptype string, client *statsd.Client) *StatsdSender {
	s := &StatsdSender{}
	s.memCurrent = fmt.Sprintf("%s.%s.mem.current", appname, apptype)
	s.memRss = fmt.Sprintf("%s.%s.mem.rss", appname, apptype)
	s.cpuUser = fmt.Sprintf("%s.%s.cpu.system", appname, apptype)
	s.cpuSystem = fmt.Sprintf("%s.%s.cpu.user", appname, apptype)
	s.interfaceInBytes = fmt.Sprintf("%s.%s.interfaces.inbytes", appname, apptype)
	s.interfaceOutBytes = fmt.Sprintf("%s.%s.interfaces.outbytes", appname, apptype)
	s.client = client
	return s
}

type MetricData struct {
	cpuStats       CpuStats
	memoryStats    cgroups.MemoryStats
	interfaceStats InterfaceStats
	isApp          bool
}

func NewMetricData(stats *cgroups.Stats) *MetricData {
	m := &MetricData{}
	m.memoryStats = stats.MemoryStats
	m.cpuStats = NewCpuStats(stats.CpuStats)
	return m
}

func (self *MetricData) ParseInterfaceData(iStats map[string]interface{}) {
	self.interfaceStats = NewInterfaceStats(iStats)
}

type AppMetrics struct {
	name string
	cid  string
	typ  string
	stop chan bool
	mu   *sync.Mutex
}

func (self *AppMetrics) Report(client *statsd.Client) {
	self.mu.Lock()
	defer self.mu.Unlock()
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
	c, err := GetCgroupStats(self.cid)
	if err != nil {
		return nil, err
	}
	data := NewMetricData(c)
	if self.typ == DEFAULT_TYPE {
		data.isApp = true
		pid, err := GetContainerPID(self.cid)
		if err != nil {
			return nil, err
		}
		var iStats map[string]interface{}
		err = NetNsSynchronize(pid, func() (err error) {
			var e error
			iStats, e = GetIfStats()
			if e != nil {
				return e
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		data.ParseInterfaceData(iStats)
	}
	return data, nil
}

func (self *AppMetrics) Stop() {
	self.stop <- true
	self.mu.Lock()
	defer self.mu.Unlock()
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
		&sync.Mutex{},
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

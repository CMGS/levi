package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	. "./utils"
	"github.com/docker/libcontainer/cgroups"
)

type MetricData struct {
	appname string
	apptype string
	isapp   bool
	first   bool

	mem_usage uint64
	mem_rss   uint64

	cpu_user_rate   float64
	cpu_system_rate float64
	cpu_usage_rate  float64

	net_inbytes  int64
	net_outbytes int64
	net_inerrs   int64
	net_outerrs  int64

	old_cpu_user   uint64
	old_cpu_system uint64
	old_cpu_usage  uint64

	old_net_inbytes  int64
	old_net_outbytes int64
	old_net_inerrs   int64
	old_net_outerrs  int64
}

func NewMetricData(appname, apptype string) *MetricData {
	m := &MetricData{}
	m.appname = appname
	m.apptype = apptype
	if apptype == DEFAULT_TYPE {
		m.isapp = true
	}
	m.first = true
	return m
}

func (self *MetricData) Update(cid string) bool {
	var iStats map[string]interface{}
	stats, err := GetCgroupStats(cid)
	if err != nil {
		Logger.Info("Get CPU,MEM Failed", err, cid, self.appname)
		return false
	}

	if self.isapp {
		pid, err := GetContainerPID(cid)
		if err != nil {
			Logger.Info("Get PID Failed", err, cid, self.appname)
			return false
		}
		err = NetNsSynchronize(pid, func() (err error) {
			var e error
			iStats, e = GetIfStats()
			if e != nil {
				return e
			}
			return nil
		})
		if err != nil {
			Logger.Info("Get NET Failed", err, cid, self.appname)
			return false
		}
	}

	if self.first {
		self.saveData(stats, iStats)
		self.first = false
	}

	t := float64(config.Metrics.ReportInterval * 1e9)
	self.mem_usage = stats.MemoryStats.Usage
	self.mem_rss = stats.MemoryStats.Stats["rss"]
	self.cpu_user_rate = float64((stats.CpuStats.CpuUsage.UsageInUsermode - self.old_cpu_user)) / t
	self.cpu_system_rate = float64((stats.CpuStats.CpuUsage.UsageInKernelmode - self.old_cpu_system)) / t
	self.cpu_usage_rate = float64((stats.CpuStats.CpuUsage.TotalUsage - self.old_cpu_usage)) / t

	if self.isapp {
		inbytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["inbytes.0"]), 10, 64)
		outbytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["outbytes.0"]), 10, 64)
		inerrs, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["inerrs.0"]), 10, 64)
		outerrs, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["outerrs.0"]), 10, 64)
		t := int64(config.Metrics.ReportInterval)
		self.net_inbytes = (inbytes - self.old_net_inbytes) / t
		self.net_outbytes = (outbytes - self.old_net_outbytes) / t
		self.net_inerrs = (inerrs - self.old_net_inerrs) / t
		self.net_outerrs = (outerrs - self.old_net_outbytes) / t
	}

	self.saveData(stats, iStats)
	return true
}

func (self *MetricData) saveData(stats *cgroups.Stats, iStats map[string]interface{}) {
	self.old_cpu_user = stats.CpuStats.CpuUsage.UsageInUsermode
	self.old_cpu_system = stats.CpuStats.CpuUsage.UsageInKernelmode
	self.old_cpu_usage = stats.CpuStats.CpuUsage.TotalUsage
	if self.isapp {
		self.old_net_inbytes, _ = strconv.ParseInt(fmt.Sprintf("%v", iStats["inbytes.0"]), 10, 64)
		self.old_net_outbytes, _ = strconv.ParseInt(fmt.Sprintf("%v", iStats["outbytes.0"]), 10, 64)
		self.old_net_inerrs, _ = strconv.ParseInt(fmt.Sprintf("%v", iStats["inerrs.0"]), 10, 64)
		self.old_net_outerrs, _ = strconv.ParseInt(fmt.Sprintf("%v", iStats["outerrs.0"]), 10, 64)
	}
}

type MetricsRecorder struct {
	mu     *sync.Mutex
	apps   map[string]*MetricData
	client *InfluxDBClient
	stop   chan bool
	done   chan int
}

func NewMetricsRecorder() *MetricsRecorder {
	r := &MetricsRecorder{}
	r.mu = &sync.Mutex{}
	r.apps = map[string]*MetricData{}
	r.client = NewInfluxDBClient()
	r.stop = make(chan bool)
	r.done = make(chan int)
	return r
}

func (self *MetricsRecorder) Add(appname, cid, apptype string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if _, ok := self.apps[cid]; ok {
		return
	}
	self.apps[cid] = NewMetricData(appname, apptype)
}

func (self *MetricsRecorder) Remove(cid string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if _, ok := self.apps[cid]; !ok {
		return
	}
	delete(self.apps, cid)
}

func (self *MetricsRecorder) Report() {
	defer close(self.stop)
	var finish bool = false
	for !finish {
		select {
		case <-time.After(time.Second * time.Duration(config.Metrics.ReportInterval)):
			self.Send()
		case f := <-self.stop:
			finish = f
		}
	}
	Logger.Info("Metrics Stop")
	self.done <- 1
}

func (self *MetricsRecorder) Stop() {
	self.stop <- true
	<-self.done
}

func (self *MetricsRecorder) Send() {
	self.mu.Lock()
	defer self.mu.Unlock()
	for cid, app := range self.apps {
		if app.Update(cid) {
			self.client.GenSeries(cid, app)
		}
	}
	self.client.Send()
}

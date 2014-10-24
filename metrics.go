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
	t       time.Time

	mem_usage uint64
	mem_rss   uint64

	cpu_user_rate   float64
	cpu_system_rate float64
	cpu_usage_rate  float64

	net_inbytes  float64
	net_outbytes float64
	net_inerrs   float64
	net_outerrs  float64

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
	return m
}

func (self *MetricData) InitStats(cid string) bool {
	stats, err := self.GetStats(cid)
	if err != nil {
		Logger.Info("Get Stats Failed", err)
		return false
	}
	self.old_cpu_user = stats.CpuStats.CpuUsage.UsageInUsermode
	self.old_cpu_system = stats.CpuStats.CpuUsage.UsageInKernelmode
	self.old_cpu_usage = stats.CpuStats.CpuUsage.TotalUsage

	if self.isapp {
		iStats, err := self.GetNetStats(cid)
		if err != nil {
			Logger.Info("Get Interface Stats Failed", err)
			return false
		}
		inbytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["inbytes.0"]), 10, 64)
		outbytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["outbytes.0"]), 10, 64)
		inerrs, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["inerrs.0"]), 10, 64)
		outerrs, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["outerrs.0"]), 10, 64)
		self.old_net_inbytes = inbytes
		self.old_net_outbytes = outbytes
		self.old_net_inerrs = inerrs
		self.old_net_outerrs = outerrs
	}

	self.UpdateTime()
	return true
}

func (self *MetricData) GetStats(cid string) (*cgroups.Stats, error) {
	stats, err := GetCgroupStats(cid)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (self *MetricData) GetNetStats(cid string) (map[string]interface{}, error) {
	var iStats map[string]interface{}
	pid, err := GetContainerPID(cid)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return iStats, nil
}

func (self *MetricData) UpdateTime() {
	self.t = time.Now()
}

func (self *MetricData) UpdateStats(cid string) bool {
	stats, err := self.GetStats(cid)
	if err != nil {
		Logger.Info("Get Stats Failed", err)
		return false
	}

	self.mem_usage = stats.MemoryStats.Usage
	self.mem_rss = stats.MemoryStats.Stats["rss"]

	t := float64(time.Now().Sub(self.t).Nanoseconds())
	self.cpu_user_rate = float64((stats.CpuStats.CpuUsage.UsageInUsermode - self.old_cpu_user)) / t
	self.cpu_system_rate = float64((stats.CpuStats.CpuUsage.UsageInKernelmode - self.old_cpu_system)) / t
	self.cpu_usage_rate = float64((stats.CpuStats.CpuUsage.TotalUsage - self.old_cpu_usage)) / t

	self.old_cpu_user = stats.CpuStats.CpuUsage.UsageInUsermode
	self.old_cpu_system = stats.CpuStats.CpuUsage.UsageInKernelmode
	self.old_cpu_usage = stats.CpuStats.CpuUsage.TotalUsage
	return true
}

func (self *MetricData) UpdateNetStats(cid string) bool {
	if !self.isapp {
		return true
	}
	iStats, err := self.GetNetStats(cid)
	if err != nil {
		Logger.Info("Get Interface Stats Failed", err)
		return false
	}

	t := time.Now().Sub(self.t).Seconds()
	inbytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["inbytes.0"]), 10, 64)
	outbytes, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["outbytes.0"]), 10, 64)
	inerrs, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["inerrs.0"]), 10, 64)
	outerrs, _ := strconv.ParseInt(fmt.Sprintf("%v", iStats["outerrs.0"]), 10, 64)

	self.net_inbytes = float64(inbytes-self.old_net_inbytes) / t
	self.net_outbytes = float64(outbytes-self.old_net_outbytes) / t
	self.net_inerrs = float64(inerrs-self.old_net_inerrs) / t
	self.net_outerrs = float64(outerrs-self.old_net_outerrs) / t

	self.old_net_inbytes = inbytes
	self.old_net_outbytes = outbytes
	self.old_net_inerrs = inerrs
	self.old_net_outerrs = outerrs
	return true
}

type MetricsRecorder struct {
	mu     *sync.Mutex
	apps   map[string]*MetricData
	client *InfluxDBClient
	stop   chan bool
}

func NewMetricsRecorder() *MetricsRecorder {
	r := &MetricsRecorder{}
	r.mu = &sync.Mutex{}
	r.apps = map[string]*MetricData{}
	r.client = NewInfluxDBClient()
	r.stop = make(chan bool)
	return r
}

func (self *MetricsRecorder) Add(appname, cid, apptype string) {
	self.mu.Lock()
	defer self.mu.Unlock()
	if _, ok := self.apps[cid]; ok {
		return
	}
	self.apps[cid] = NewMetricData(appname, apptype)
	self.apps[cid].InitStats(cid)
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
	self.stop <- true
}

func (self *MetricsRecorder) Stop() {
	self.stop <- true
	<-self.stop
}

func (self *MetricsRecorder) Send() {
	self.mu.Lock()
	defer self.mu.Unlock()
	for cid, app := range self.apps {
		if app.UpdateStats(cid) && app.UpdateNetStats(cid) {
			self.client.GenSeries(cid, app)
			app.UpdateTime()
		}
	}
	self.client.Send()
}

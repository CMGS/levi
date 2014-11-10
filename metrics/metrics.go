package metrics

import (
	"sync"
	"time"

	"../common"
	"../defines"
	"../logs"
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

	old_net_inbytes  uint64
	old_net_outbytes uint64
	old_net_inerrs   uint64
	old_net_outerrs  uint64
}

func NewMetricData(appname, apptype string) *MetricData {
	m := &MetricData{}
	m.appname = appname
	m.apptype = apptype
	if apptype == common.DEFAULT_TYPE {
		m.isapp = true
	}
	return m
}

func (self *MetricData) InitStats(cid string) bool {
	stats, err := GetCgroupStats(cid)
	if err != nil {
		logs.Info("Get Stats Failed", err)
		return false
	}
	self.old_cpu_user = stats.CpuStats.CpuUsage.UsageInUsermode
	self.old_cpu_system = stats.CpuStats.CpuUsage.UsageInKernelmode
	self.old_cpu_usage = stats.CpuStats.CpuUsage.TotalUsage

	if self.isapp {
		iStats, err := self.GetNetStats(cid)
		if err != nil {
			logs.Info(err)
			return false
		}
		self.old_net_inbytes = iStats["inbytes.0"]
		self.old_net_outbytes = iStats["outbytes.0"]
		self.old_net_inerrs = iStats["inerrs.0"]
		self.old_net_outerrs = iStats["outerrs.0"]
	}

	self.UpdateTime()
	return true
}

func (self *MetricData) GetNetStats(cid string) (map[string]uint64, error) {
	return GetNetStats(cid)
}

func (self *MetricData) UpdateTime() {
	self.t = time.Now()
}

func (self *MetricData) UpdateStats(cid string) bool {
	stats, err := GetCgroupStats(cid)
	if err != nil {
		logs.Info("Get Stats Failed", err)
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
		logs.Info("Get Interface Stats Failed", err)
		return false
	}

	t := time.Now().Sub(self.t).Seconds()
	self.net_inbytes = float64(iStats["inbytes.0"]-self.old_net_inbytes) / t
	self.net_outbytes = float64(iStats["outbytes.0"]-self.old_net_outbytes) / t
	self.net_inerrs = float64(iStats["inerrs.0"]-self.old_net_inerrs) / t
	self.net_outerrs = float64(iStats["outerrs.0"]-self.old_net_outerrs) / t

	self.old_net_inbytes = iStats["inbytes.0"]
	self.old_net_outbytes = iStats["outbytes.0"]
	self.old_net_inerrs = iStats["inerrs.0"]
	self.old_net_outerrs = iStats["outerrs.0"]
	return true
}

type MetricsRecorder struct {
	mu     *sync.Mutex
	apps   map[string]*MetricData
	client *InfluxDBClient
	stop   chan bool
	t      int
}

func NewMetricsRecorder(hostname string, config defines.MetricsConfig) *MetricsRecorder {
	InitDevDir()
	r := &MetricsRecorder{}
	r.mu = &sync.Mutex{}
	r.apps = map[string]*MetricData{}
	r.client = NewInfluxDBClient(hostname, config)
	r.t = config.ReportInterval
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
	var finish bool = false
	for !finish {
		select {
		case <-time.After(time.Second * time.Duration(self.t)):
			self.Send()
		case finish = <-self.stop:
			logs.Info("Metrics Stop")
		}
	}
	self.stop <- true
}

func (self *MetricsRecorder) Stop() {
	defer close(self.stop)
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

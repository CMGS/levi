package main

import (
	"net/http"

	. "./utils"
	"github.com/influxdb/influxdb/client"
)

type InfluxDBClient struct {
	client *client.Client
	series []*client.Series
}

var influxdb_columns []string = []string{"host", "apptype", "appid", "metric", "value"}

func NewInfluxDBClient() *InfluxDBClient {
	c := &client.ClientConfig{
		Host:       config.Metrics.Host,
		Username:   config.Metrics.Username,
		Password:   config.Metrics.Password,
		Database:   config.Metrics.Database,
		HttpClient: http.DefaultClient,
		IsSecure:   false,
		IsUDP:      false,
	}
	i, err := client.New(c)
	if err != nil {
		Logger.Assert(err, "InfluxDB")
	}
	return &InfluxDBClient{i, []*client.Series{}}
}

func (self *InfluxDBClient) GenSeries(cid string, app *MetricData) {
	points := [][]interface{}{
		{config.Name, app.apptype, cid, "cpu_usage", app.cpu_usage_rate},
		{config.Name, app.apptype, cid, "cpu_system", app.cpu_system_rate},
		{config.Name, app.apptype, cid, "cpu_user", app.cpu_user_rate},
		{config.Name, app.apptype, cid, "mem_usage", app.mem_usage},
		{config.Name, app.apptype, cid, "mem_rss", app.mem_rss},
	}
	if app.isapp {
		p2 := [][]interface{}{
			{config.Name, app.apptype, cid, "net_recv", app.net_inbytes},
			{config.Name, app.apptype, cid, "net_send", app.net_outbytes},
			{config.Name, app.apptype, cid, "net_recv_err", app.net_inerrs},
			{config.Name, app.apptype, cid, "net_send_err", app.net_outerrs},
		}
		for _, p := range p2 {
			points = append(points, p)
		}
	}
	series := &client.Series{
		Name:    app.appname,
		Columns: influxdb_columns,
		Points:  points,
	}
	self.series = append(self.series, series)
}

func (self *InfluxDBClient) Send() {
	if err := self.client.WriteSeries(self.series); err != nil {
		Logger.Info("Write to InfluxDB Failed", err)
	}
	self.series = []*client.Series{}
}

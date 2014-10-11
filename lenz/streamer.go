package lenz

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/syslog"
	"net"
	"net/url"
	"strings"

	"../defines"
	. "../utils"
)

func streamer(route *defines.Route, logstream chan *defines.Log) {
	var types map[string]struct{}
	if route.Source != nil {
		types = make(map[string]struct{})
		for _, t := range route.Source.Types {
			types[t] = struct{}{}
		}
	}
	for logline := range logstream {
		if types != nil {
			if _, ok := types[logline.Type]; !ok {
				continue
			}
		}
		appinfo := strings.SplitN(logline.Name, "_", 2)
		logline.Appname = appinfo[0]
		logline.Port = appinfo[1]
		logline.Tag = route.Target.AppendTag

		for offset := 0; offset < route.Backends.Len(); offset++ {
			addr, err := route.Backends.Get(logline.Appname, offset)
			if err != nil {
				Logger.Debug("Get backend failed", err)
				log.Println(logline.Appname, logline.Data)
				break
			}

			Logger.Debug(logline.Appname, addr)
			switch u, err := url.Parse(addr); {
			case err != nil:
				Logger.Debug("Lenz", err)
				route.Backends.Remove(addr)
				continue
			case u.Scheme == "udp":
				if err := udpStreamer(logline, u.Host); err != nil {
					Logger.Debug("Lenz Send to", u.Host, "by udp failed", err)
					continue
				}
			case u.Scheme == "tcp":
				if err := tcpStreamer(logline, u.Host); err != nil {
					Logger.Debug("Lenz Send to", u.Host, "by tcp failed", err)
					continue
				}
			case u.Scheme == "syslog":
				if err := syslogStreamer(logline, u.Host); err != nil {
					Logger.Debug("Lenz Sent to syslog failed", err)
					continue
				}
			}
			break
		}
	}
}

func syslogStreamer(logline *defines.Log, addr string) error {
	tag := fmt.Sprintf("%s.%s", logline.Appname, logline.Tag)
	remote, err := syslog.Dial("udp", addr, syslog.LOG_USER|syslog.LOG_INFO, tag)
	if err != nil {
		return err
	}
	io.WriteString(remote, logline.Data)
	return nil
}

func tcpStreamer(logline *defines.Log, addr string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		Logger.Debug("Resolve tcp failed", err)
		return err
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		Logger.Debug("Connect backend failed", err)
		return err
	}
	defer conn.Close()
	writeJSON(conn, logline)
	return nil
}

func udpStreamer(logline *defines.Log, addr string) error {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		Logger.Debug("Resolve udp failed", err)
		return err
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		Logger.Debug("Connect backend failed", err)
		return err
	}
	defer conn.Close()
	writeJSON(conn, logline)
	return nil
}

func writeJSON(w io.Writer, logline *defines.Log) {
	encoder := json.NewEncoder(w)
	encoder.Encode(logline)
}

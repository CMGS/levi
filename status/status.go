package status

import (
	"strings"

	"../common"
	"../defines"
	"../lenz"
	"../logs"
	"../metrics"
	"github.com/fsouza/go-dockerclient"
)

type StatusMoniter struct {
	events          chan *docker.APIEvents
	docker_registry string
	docker          *defines.DockerWrapper
	ws              *defines.WebSocketWrapper
	metrics         *metrics.MetricsRecorder
	lenz            *lenz.LenzForwarder
	Removable       map[string]struct{}
}

func NewStatus(Docker *defines.DockerWrapper, Metrics *metrics.MetricsRecorder, Lenz *lenz.LenzForwarder, Ws *defines.WebSocketWrapper, config defines.DockerConfig) *StatusMoniter {
	status := &StatusMoniter{}
	status.events = make(chan *docker.APIEvents)
	status.Removable = map[string]struct{}{}
	status.docker_registry = config.Registry
	status.docker = Docker
	status.metrics = Metrics
	status.lenz = Lenz
	status.ws = Ws
	logs.Assert(status.docker.AddEventListener(status.events), "Attacher")
	return status
}

func (self *StatusMoniter) Listen() {
	logs.Info("Status Monitor Start")
	for event := range self.events {
		logs.Debug("Status:", event.Status, event.ID, event.From)
		if event.Status == common.STATUS_DIE {
			self.metrics.Remove(event.ID[:12])
			if _, ok := self.Removable[event.ID]; ok {
				self.die(event.ID)
			}
		}
	}
}

func (self *StatusMoniter) getStatus(s string) string {
	switch {
	case strings.HasPrefix(s, "Up"):
		return common.STATUS_START
	default:
		return common.STATUS_DIE
	}
}

func (self *StatusMoniter) Report(id string) {
	containers, err := self.docker.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		logs.Info(err, "Load")
	}

	result := &defines.TaskResult{Id: id}
	result.Status = []*defines.StatusInfo{}

	logs.Info("Load container")
	for _, container := range containers {
		if !strings.HasPrefix(container.Image, self.docker_registry) {
			continue
		}
		status := self.getStatus(container.Status)
		name, aid, at := self.getAppInfo(container.Names[0])
		shortID := container.ID[:12]
		logs.Debug("Container", name, shortID, status)
		if status != common.STATUS_DIE {
			self.metrics.Add(name, shortID, at)
			self.lenz.Attacher.Attach(shortID, name, aid, at)
		}
		self.Removable[container.ID] = struct{}{}
		s := &defines.StatusInfo{status, name, container.ID}
		result.Status = append(result.Status, s)
	}
	if err := self.ws.WriteJSON(result); err != nil {
		logs.Info(err, result)
	}
}

func (self *StatusMoniter) die(id string) {
	result := &defines.TaskResult{Id: common.STATUS_IDENT}
	result.Status = make([]*defines.StatusInfo, 1)

	container, err := self.docker.InspectContainer(id)
	if err != nil {
		logs.Info("Status inspect docker failed", err)
		return
	}
	appname, _, _ := self.getAppInfo(container.Name)
	result.Status[0] = &defines.StatusInfo{common.STATUS_DIE, appname, id}
	if err := self.ws.WriteJSON(result); err != nil {
		logs.Info(err, result)
	}
}

func (self *StatusMoniter) getAppInfo(containerName string) (string, string, string) {
	containerName = strings.TrimLeft(containerName, "/")
	if pos := strings.LastIndex(containerName, "_daemon_"); pos > -1 {
		appname := containerName[:pos]
		appid := containerName[pos+8:]
		return appname, appid, common.DAEMON_TYPE
	}
	if pos := strings.LastIndex(containerName, "_test_"); pos > -1 {
		appname := containerName[:pos]
		appid := containerName[pos+6:]
		return appname, appid, common.TEST_TYPE
	}
	appinfo := strings.Split(containerName, "_")
	appname := strings.Join(appinfo[:len(appinfo)-1], "_")
	return appname, appinfo[len(appinfo)-1], common.DEFAULT_TYPE
}

package lenz

import (
	"bufio"
	"io"
	"strings"
	"sync"

	"../defines"
	. "../utils"
	"github.com/fsouza/go-dockerclient"
)

type AttachManager struct {
	sync.Mutex
	attached map[string]*LogPump
	channels map[chan *defines.AttachEvent]struct{}
	client   *defines.DockerWrapper
}

func NewAttachManager(client *defines.DockerWrapper) *AttachManager {
	m := &AttachManager{
		attached: make(map[string]*LogPump),
		channels: make(map[chan *defines.AttachEvent]struct{}),
	}
	m.client = client
	return m
}

func (m *AttachManager) Attached(id string) bool {
	_, ok := m.attached[id]
	return ok
}

func (m *AttachManager) Attach(id, name, aid, atype string) {
	// Not Thread Safe
	if m.Attached(id) {
		return
	}
	success := make(chan struct{})
	failure := make(chan error)
	outrd, outwr := io.Pipe()
	errrd, errwr := io.Pipe()
	go func() {
		err := m.client.AttachToContainer(docker.AttachToContainerOptions{
			Container:    id,
			OutputStream: outwr,
			ErrorStream:  errwr,
			Stdin:        false,
			Stdout:       true,
			Stderr:       true,
			Stream:       true,
			Success:      success,
		})
		outwr.Close()
		errwr.Close()
		Logger.Debug("Lenz Attach:", id, "finished")
		if err != nil {
			close(success)
			failure <- err
		}
		m.send(&defines.AttachEvent{Type: "detach", ID: id, Name: name, AppID: aid, AppType: atype})
		m.Lock()
		delete(m.attached, id)
		m.Unlock()
	}()
	_, ok := <-success
	if ok {
		m.Lock()
		m.attached[id] = NewLogPump(outrd, errrd, id, name, aid, atype)
		m.Unlock()
		success <- struct{}{}
		m.send(&defines.AttachEvent{Type: "attach", ID: id, Name: name, AppID: aid, AppType: atype})
		Logger.Debug("Lenz Attach:", id, "success")
		return
	}
	Logger.Debug("Lenz Attach:", id, "failure:", <-failure)
}

func (m *AttachManager) send(event *defines.AttachEvent) {
	m.Lock()
	defer m.Unlock()
	for ch, _ := range m.channels {
		// TODO: log err after timeout and continue
		ch <- event
	}
}

func (m *AttachManager) addListener(ch chan *defines.AttachEvent) {
	m.Lock()
	defer m.Unlock()
	m.channels[ch] = struct{}{}
	go func() {
		for id, pump := range m.attached {
			ch <- &defines.AttachEvent{Type: "attach", ID: id, Name: pump.Name, AppID: pump.AppID, AppType: pump.AppType}
		}
	}()
}

func (m *AttachManager) removeListener(ch chan *defines.AttachEvent) {
	m.Lock()
	defer m.Unlock()
	delete(m.channels, ch)
}

func (m *AttachManager) Get(id string) *LogPump {
	m.Lock()
	defer m.Unlock()
	return m.attached[id]
}

func (m *AttachManager) Listen(source *defines.Source, logstream chan *defines.Log, closer <-chan bool) {
	if source == nil {
		source = new(defines.Source)
	}
	events := make(chan *defines.AttachEvent)
	m.addListener(events)
	defer m.removeListener(events)
	for {
		select {
		case event := <-events:
			if event.Type == "attach" && (source.All() ||
				(source.ID != "" && strings.HasPrefix(event.ID, source.ID)) ||
				(source.Name != "" && event.Name == source.Name) ||
				(source.Filter != "" && strings.Contains(event.Name, source.Filter))) {
				pump := m.Get(event.ID)
				pump.AddListener(logstream)
				defer func() {
					if pump != nil {
						pump.RemoveListener(logstream)
					}
				}()
			} else if source.ID != "" && event.Type == "detach" &&
				strings.HasPrefix(event.ID, source.ID) {
				return
			}
		case <-closer:
			return
		}
	}
}

type LogPump struct {
	sync.Mutex
	ID       string
	Name     string
	AppID    string
	AppType  string
	channels map[chan *defines.Log]struct{}
}

func NewLogPump(stdout, stderr io.Reader, id, name, aid, atype string) *LogPump {
	obj := &LogPump{
		ID:       id,
		Name:     name,
		AppID:    aid,
		AppType:  atype,
		channels: make(map[chan *defines.Log]struct{}),
	}
	pump := func(typ string, source io.Reader) {
		buf := bufio.NewReader(source)
		for {
			data, err := buf.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					Logger.Debug("Lenz Pump:", id, typ, err)
				}
				return
			}
			obj.send(&defines.Log{
				Data:    strings.TrimSuffix(string(data), "\n"),
				ID:      id,
				Name:    name,
				AppID:   aid,
				AppType: atype,
				Type:    typ,
			})
		}
	}
	go pump("stdout", stdout)
	go pump("stderr", stderr)
	return obj
}

func (o *LogPump) send(log *defines.Log) {
	o.Lock()
	defer o.Unlock()
	for ch, _ := range o.channels {
		// TODO: log err after timeout and continue
		ch <- log
	}
}

func (o *LogPump) AddListener(ch chan *defines.Log) {
	o.Lock()
	defer o.Unlock()
	o.channels[ch] = struct{}{}
}

func (o *LogPump) RemoveListener(ch chan *defines.Log) {
	o.Lock()
	defer o.Unlock()
	delete(o.channels, ch)
}

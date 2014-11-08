package defines

import "fmt"

type Result struct {
	Id    string
	Done  bool
	Index int
	Type  int
	Data  string
}

type BuildTask struct {
	Group   string
	Name    string
	Version string
	Base    string
	Build   string
	Static  string
	Schema  string
	Bid     string
}

type RemoveTask struct {
	Container string
	RmImage   bool
}

func (self *RemoveTask) IsRemoveImage() bool {
	return self.RmImage
}

type AddTask struct {
	Version   string
	Bind      int64
	Port      int64
	Cmd       []string
	Memory    int64
	CpuShares int64
	CpuSet    string
	Daemon    string
	Test      string
	Ident     string
}

func (self *AddTask) IsDaemon() bool {
	return self.Daemon != ""
}

func (self *AddTask) IsTest() bool {
	return self.Test != ""
}

func (self *AddTask) ShouldExpose() bool {
	return self.Daemon == "" && self.Test == ""
}

func (self *AddTask) SetAsTest() {
	self.Ident = fmt.Sprintf("test_%s", self.Test)
}

func (self *AddTask) SetAsDaemon() {
	self.Ident = fmt.Sprintf("daemon_%s", self.Daemon)
}

func (self *AddTask) SetAsService() {
	self.Ident = fmt.Sprintf("%d", self.Bind)
}

type Tasks struct {
	Build  []*BuildTask
	Add    []*AddTask
	Remove []*RemoveTask
}

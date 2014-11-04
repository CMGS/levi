package defines

type TestResult struct {
	ExitCode int
	Err      string
}

type StatusInfo struct {
	Type    string
	Appname string
	Id      string
}

type TaskResult struct {
	Id     string
	Build  []string
	Add    []string
	Remove []bool
	Test   map[string]*TestResult
	Status []*StatusInfo
}

type Result struct {
	Id    string
	Done  bool
	Index int
	Type  int
	Data  string
}

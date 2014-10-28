package defines

type GitConfig struct {
	Endpoint  string
	WorkDir   string
	ExtendDir string
}

type NginxConfig struct {
	Configs    string
	Template   string
	DyUpstream string
}

type DockerConfig struct {
	Endpoint string
	Registry string
	Network  string
}

type AppConfig struct {
	Home     string
	Tmpdirs  string
	Permdirs string
}

type EtcdConfig struct {
	Sync     bool
	Machines []string
}

type LenzConfig struct {
	Routes   string
	Forwards []string
	Stdout   bool
}

type MetricsConfig struct {
	ReportInterval int
	Host           string
	Username       string
	Password       string
	Database       string
}

type LeviConfig struct {
	HostName        string
	Master          string
	PidFile         string
	TaskNum         int
	TaskInterval    int
	ReadBufferSize  int
	WriteBufferSize int

	Git     GitConfig
	Nginx   NginxConfig
	Docker  DockerConfig
	App     AppConfig
	Etcd    EtcdConfig
	Lenz    LenzConfig
	Metrics MetricsConfig
}

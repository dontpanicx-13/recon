package scanner

type ScanResult struct {
	SchemaVersion int         `json:"schema_version"`
	ScanID        string      `json:"scan_id"`
	Config        ScanConfig  `json:"config"`
	Meta          ScanMeta    `json:"meta"`
	Summary       ScanSummary `json:"summary"`
	Hosts         []Host      `json:"hosts"`
}

const (
	PortOpen     = "open"
	PortClosed   = "closed"
	PortFiltered = "filtered"
)

type Task struct {
	Host string
	Port int
}

type Result struct {
	Host string
	Port PortState
	Err  error
}

type ScanConfig struct {
	Targets        []string `json:"targets"`
	Ports          []int    `json:"ports"`
	Profile        string   `json:"profile"`
	Concurrency    int      `json:"concurrency"`
	TimeoutMS      int      `json:"timeout_ms"`
	BannerGrabbing bool     `json:"banner_grabbing"`
	TLSAnalysis    bool     `json:"tls_analysis"`
	ReverseDNS     bool     `json:"reverse_dns"`
}

type ScanMeta struct {
	Date       string `json:"date"`
	Status     string `json:"status"`
	DurationMS int64  `json:"duration_ms"`
}

type ScanSummary struct {
	HostsTotal     int `json:"hosts_total"`
	HostsFound     int `json:"hosts_found"`
	HostsCompleted int `json:"hosts_completed"`
	PortsTotal     int `json:"ports_total"`
	PortsProbed    int `json:"ports_probed"`
	OpenPorts      int `json:"open_ports"`
}

type Host struct {
	IP       string      `json:"ip"`
	Hostname string      `json:"hostname,omitempty"`
	Ports    []PortState `json:"ports"`
}

type PortState struct {
	Port         int      `json:"port"`
	State        string   `json:"state"`
	ServiceGuess string   `json:"service_guess,omitempty"`
	Banner       *string  `json:"banner"`
	TLS          *TLSInfo `json:"tls"`
}

type TLSInfo struct {
	CommonName string   `json:"cn,omitempty"`
	SAN        []string `json:"san,omitempty"`
	Issuer     string   `json:"issuer,omitempty"`
	Expires    string   `json:"expires,omitempty"`
	TLSVersion string   `json:"tls_version,omitempty"`
	Cipher     string   `json:"cipher,omitempty"`
	Note       string   `json:"note,omitempty"`
}

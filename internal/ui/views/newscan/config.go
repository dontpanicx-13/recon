package newscan

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"recon/internal/ports"
	"recon/internal/scanner"
	"recon/internal/target"
)

type StartScanMsg struct {
	Config scanner.ScanConfig
}

func (m NewScanModel) BuildScanConfig() (scanner.ScanConfig, []string, []string) {
	errs, warns := m.validate()
	if len(errs) > 0 {
		return scanner.ScanConfig{}, errs, warns
	}

	parseResult, parseErrs := target.Parse(m.targetsInput.Value(), target.Options{
		ExcludeNetworkBroadcast: true,
	})
	if len(parseErrs) > 0 {
		errs = append(errs, parseErrs...)
	}
	warns = append(warns, parseResult.Warnings...)
	if len(errs) > 0 {
		return scanner.ScanConfig{}, errs, warns
	}

	ports, portErrs := m.parsePorts()
	if len(portErrs) > 0 {
		errs = append(errs, portErrs...)
		return scanner.ScanConfig{}, errs, warns
	}

	concurrency, _ := strconv.Atoi(strings.TrimSpace(m.concurrency.Value()))
	timeoutMS, _ := strconv.Atoi(strings.TrimSpace(m.timeoutMs.Value()))

	return scanner.ScanConfig{
		Targets:        parseResult.Targets,
		Ports:          ports,
		Profile:        m.profileName(),
		Concurrency:    concurrency,
		TimeoutMS:      timeoutMS,
		BannerGrabbing: m.toggleBanner,
		TLSAnalysis:    m.toggleTLS,
		ReverseDNS:     m.toggleRDNS,
	}, errs, warns
}

func (m NewScanModel) StartScanCmd() (tea.Cmd, []string, []string) {
	cfg, errs, warns := m.BuildScanConfig()
	if len(errs) > 0 {
		return nil, errs, warns
	}
	return func() tea.Msg {
		return StartScanMsg{Config: cfg}
	}, errs, warns
}

func (m NewScanModel) profileName() string {
	switch m.profile {
	case profileQuick:
		return "quick"
	case profileFull:
		return "full"
	case profileCustom:
		return "custom"
	default:
		return "default"
	}
}

func (m NewScanModel) parsePorts() ([]int, []string) {
	switch m.portsMode {
	case portsRange:
		return parseRangePorts(m.portsRange.Value())
	case portsList:
		return parseListPorts(m.portsList.Value())
	default:
		return presetPorts(m.portsPreset), nil
	}
}

func parseRangePorts(value string) ([]int, []string) {
	parts := strings.Split(strings.TrimSpace(value), "-")
	if len(parts) != 2 {
		return nil, []string{"Ports range must be like 1-1024 within 1-65535."}
	}
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, []string{"Ports range must be like 1-1024 within 1-65535."}
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, []string{"Ports range must be like 1-1024 within 1-65535."}
	}
	if start < 1 || end > 65535 || start > end {
		return nil, []string{"Ports range must be like 1-1024 within 1-65535."}
	}
	ports := make([]int, 0, end-start+1)
	for port := start; port <= end; port++ {
		ports = append(ports, port)
	}
	return ports, nil
}

func parseListPorts(value string) ([]int, []string) {
	parts := strings.Split(value, ",")
	if len(parts) == 0 {
		return nil, []string{"Ports list must be comma-separated numbers within 1-65535."}
	}
	seen := make(map[int]struct{})
	var ports []int
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, []string{"Ports list must be comma-separated numbers within 1-65535."}
		}
		port, err := strconv.Atoi(part)
		if err != nil || port < 1 || port > 65535 {
			return nil, []string{"Ports list must be comma-separated numbers within 1-65535."}
		}
		if _, ok := seen[port]; ok {
			continue
		}
		seen[port] = struct{}{}
		ports = append(ports, port)
	}
	return ports, nil
}

func presetPorts(preset portsPresetKind) []int {
	switch preset {
	case presetTop1000:
		return ports.Top1000()
	case presetAll:
		return ports.All()
	default:
		return ports.Top100()
	}
}

func rangePorts(start, end int) []int {
	if start < 1 {
		start = 1
	}
	if end > 65535 {
		end = 65535
	}
	if start > end {
		return nil
	}
	ports := make([]int, 0, end-start+1)
	for port := start; port <= end; port++ {
		ports = append(ports, port)
	}
	return ports
}

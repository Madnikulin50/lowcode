package scancidr

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Scanner struct {
	cfg Config
}

func NewScanner(cfg Config) *Scanner {
	def := DefaultConfig()
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = def.Concurrency
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = def.Timeout
	}
	if len(cfg.Ports) == 0 {
		cfg.Ports = def.Ports
	}
	return &Scanner{cfg: cfg}
}

func (s *Scanner) Scan(ctx context.Context, cidr string, onProgress func(current, total int, ip string)) ([]Device, error) {
	cidrs := ResolveScanCIDRs(cidr)
	if len(cidrs) == 0 {
		return nil, fmt.Errorf("no scan targets for %q", cidr)
	}
	var all []Device
	var lastErr error
	for _, c := range cidrs {
		devs, err := s.scanCIDR(ctx, c, onProgress)
		if err != nil {
			lastErr = err
			log.Printf("scan %s: %v", c, err)
			continue
		}
		all = append(all, devs...)
	}
	if len(all) == 0 && lastErr != nil {
		return nil, lastErr
	}
	return all, nil
}

func (s *Scanner) scanCIDR(ctx context.Context, cidr string, onProgress func(current, total int, ip string)) ([]Device, error) {
	if !strings.Contains(cidr, "/") {
		if net.ParseIP(cidr) != nil {
			cidr += "/32"
		}
	}
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}
	ips := HostsFromCIDR(ipnet)
	if len(ips) == 0 {
		return nil, fmt.Errorf("no valid hosts in %s", cidr)
	}

	macTable := readARPTable()
	grouped := map[string][]Port{}
	seedKnownLive(ipnet, grouped, macTable)

	var mu sync.Mutex
	var results []portResult
	var wg sync.WaitGroup
	total := len(ips)
	ipSem := make(chan struct{}, s.cfg.Concurrency)
	portSem := make(chan struct{}, s.cfg.Concurrency*2)
	scanned := 0
	timeout := s.cfg.Timeout

	for _, addr := range ips {
		wg.Add(1)
		ipSem <- struct{}{}
		go func(a string) {
			defer wg.Done()
			defer func() { <-ipSem }()
			var ipMu sync.Mutex
			var ipWg sync.WaitGroup
			var ipPorts []portResult
			for _, port := range s.cfg.Ports {
				ipWg.Add(1)
				portSem <- struct{}{}
				go func(p int) {
					defer ipWg.Done()
					defer func() { <-portSem }()
					ok, banner := tryConnect(ctx, a, p, timeout)
					if ok {
						ipMu.Lock()
						ipPorts = append(ipPorts, portResult{IP: a, Port: p, Banner: banner})
						ipMu.Unlock()
					}
				}(port)
			}
			ipWg.Wait()
			mu.Lock()
			results = append(results, ipPorts...)
			scanned++
			if onProgress != nil {
				onProgress(scanned, total, a)
			}
			mu.Unlock()
		}(addr)
	}
	wg.Wait()

	for ip, ports := range groupPortResults(results) {
		grouped[ip] = append(grouped[ip], ports...)
	}
	s.markPingAndARP(ctx, ips, grouped, macTable)

	if len(grouped) == 0 {
		return nil, nil
	}

	devices := make([]Device, 0, len(grouped))
	now := time.Now().Format(time.RFC3339)
	for _, ipStr := range ips {
		ports, ok := grouped[ipStr]
		if !ok {
			continue
		}
		d := Device{
			IP:        ipStr,
			Status:    "online",
			LastSeen:  now,
			OpenPorts: ports,
			Services:  collectServices(ports),
		}
		if mac, ok := macTable[ipStr]; ok {
			d.MAC = mac
		}
		if names, err := net.LookupAddr(ipStr); err == nil && len(names) > 0 {
			d.Hostname = strings.TrimSuffix(names[0], ".")
		}
		devices = append(devices, d)
	}
	return devices, nil
}

type portResult struct {
	IP     string
	Port   int
	Banner string
}

func tryConnect(ctx context.Context, addr string, port int, timeout time.Duration) (bool, string) {
	if timeout <= 0 {
		timeout = 400 * time.Millisecond
	}
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(addr, fmt.Sprintf("%d", port)))
	if err != nil {
		return false, ""
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	return true, strings.TrimSpace(string(buf[:n]))
}

func groupPortResults(results []portResult) map[string][]Port {
	m := make(map[string][]Port)
	for _, r := range results {
		m[r.IP] = append(m[r.IP], Port{Port: r.Port, Proto: "tcp", Service: portService(r.Port), Banner: r.Banner})
	}
	return m
}

func collectServices(ports []Port) []string {
	seen := map[string]bool{}
	var out []string
	for _, p := range ports {
		s := p.Service
		if s == "" || s == "unknown" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func readARPTable() map[string]string {
	data, err := os.ReadFile("/proc/net/arp")
	if err != nil {
		return nil
	}
	table := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) >= 4 && fields[3] != "00:00:00:00:00:00" {
			table[fields[0]] = fields[3]
		}
	}
	return table
}

func portService(port int) string {
	m := map[int]string{
		21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
		80: "http", 110: "pop3", 139: "netbios-ssn", 143: "imap",
		389: "ldap", 443: "https", 445: "microsoft-ds",
		3306: "mysql", 3333: "http-service", 3389: "ms-wbt-server",
		5432: "postgresql", 5555: "adb", 6379: "redis",
		7000: "airplay", 8080: "http-proxy", 8443: "https-alt",
	}
	if s, ok := m[port]; ok {
		return s
	}
	return "unknown"
}

func seedKnownLive(ipnet *net.IPNet, grouped map[string][]Port, macTable map[string]string) {
	if ipnet == nil || grouped == nil {
		return
	}
	for ip, mac := range macTable {
		if mac == "" || mac == "00:00:00:00:00:00" {
			continue
		}
		pip := net.ParseIP(ip)
		if pip != nil && ipnet.Contains(pip) {
			if _, ok := grouped[ip]; !ok {
				grouped[ip] = nil
			}
		}
	}
	for _, lip := range localIPv4Addresses() {
		pip := net.ParseIP(lip)
		if pip != nil && ipnet.Contains(pip) {
			if _, ok := grouped[lip]; !ok {
				grouped[lip] = nil
			}
		}
	}
}

func (s *Scanner) markPingAndARP(ctx context.Context, ips []string, grouped map[string][]Port, macTable map[string]string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.cfg.Concurrency)
	var mu sync.Mutex
	for _, addr := range ips {
		if _, ok := grouped[addr]; ok {
			continue
		}
		if mac, ok := macTable[addr]; ok && mac != "" && mac != "00:00:00:00:00:00" {
			mu.Lock()
			grouped[addr] = nil
			mu.Unlock()
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(a string) {
			defer wg.Done()
			defer func() { <-sem }()
			if pingHost(ctx, a) {
				mu.Lock()
				grouped[a] = nil
				mu.Unlock()
			}
		}(addr)
	}
	wg.Wait()
}

func pingHost(ctx context.Context, addr string) bool {
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", "-W", "1", addr)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

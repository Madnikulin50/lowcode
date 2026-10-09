package agent

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/hirochachacha/go-smb2"
)

type Scanner struct {
	conc        int
	ports       []int
	dialTimeout time.Duration
}

func NewScanner(cfg Config) *Scanner {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 10
	}
	if len(cfg.ScanPorts) == 0 {
		cfg.ScanPorts = []int{22, 80, 443, 8080, 8443, 3333, 3389, 445, 139, 135, 21, 23, 25, 53, 88, 110, 143, 389, 464, 636, 993, 995, 3268, 3269, 3306, 5432, 6379, 27017, 515, 554, 548, 631, 8554, 9100, 1723, 5001, 5555, 5223, 7000}
	}
	dt := cfg.PingTimeout
	if dt <= 0 {
		dt = 400 * time.Millisecond
	}
	return &Scanner{conc: cfg.Concurrency, ports: cfg.ScanPorts, dialTimeout: dt}
}

func (s *Scanner) Scan(ctx context.Context, cidr string, onProgress func(current, total int, ip string)) ([]Device, error) {
	cidrs := ResolveScanCIDRs(cidr)
	if len(cidrs) == 0 {
		return nil, fmt.Errorf("no scan targets for %q", cidr)
	}
	log.Printf("scan: requested %q → %v", cidr, cidrs)
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
	if len(all) == 0 {
		log.Printf("scan: no live hosts on %s", strings.Join(cidrs, ", "))
	} else {
		log.Printf("scan: found %d device(s) on %s", len(all), strings.Join(cidrs, ", "))
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

	ips := hostsFromCIDR(ipnet)
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
	ipSem := make(chan struct{}, s.conc)
	portSem := make(chan struct{}, s.conc*2)
	scanned := 0
	timeout := s.dialTimeout
	if timeout <= 0 {
		timeout = 400 * time.Millisecond
	}

	for _, addr := range ips {
		wg.Add(1)
		ipSem <- struct{}{}
		go func(a string) {
			defer wg.Done()
			defer func() { <-ipSem }()
			var ipMu sync.Mutex
			var ipWg sync.WaitGroup
			var ipPorts []portResult
			for _, port := range s.ports {
				ipWg.Add(1)
				portSem <- struct{}{}
				go func(p int) {
					defer ipWg.Done()
					defer func() { <-portSem }()
					ok, banner, version := tryConnect(ctx, a, p, timeout)
					if ok {
						ipMu.Lock()
						ipPorts = append(ipPorts, portResult{IP: a, Port: p, Banner: banner, Version: version})
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
		log.Printf("scan: no live hosts on %s after probing %d IPs", cidr, len(ips))
		return nil, nil
	}

	devices := make([]Device, 0, len(grouped))
	for _, ipStr := range ips {
		ports, ok := grouped[ipStr]
		if !ok {
			continue
		}
		d := Device{
			IP:        ipStr,
			Status:    "online",
			LastSeen:  time.Now().Format(time.RFC3339),
			OpenPorts: ports,
			Services:  collectServices(ports),
			OS:        guessOS(ports),
		}
		if mac, ok := macTable[ipStr]; ok {
			d.MAC = mac
		}
		if d.Hostname == "" {
			if name := resolveDeviceName(ipStr, ports); name != "" {
				d.Hostname = name
			}
		}
		if hasSMB(ports) {
			if shares, err := smbEnumerate(ipStr); err == nil && len(shares) > 0 {
				d.Shares = shares
			}
		}
		if isDomainController(ports) {
			d.DeviceType = "server"
			if d.Domain == "" {
				d.Domain = detectDomain(ctx, ipStr)
			}
		}
		devices = append(devices, d)
	}
	s.udpEnrich(ctx, devices)
	s.mdnsEnrich(ctx, devices)
	return devices, nil
}

// udpEnrich probes the well-known UDP services (53/123/161/1900) on hosts that
// are already known to be live (found via TCP, ping or ARP). Blindly probing
// UDP on every IP of a CIDR is not feasible (silent UDP has no reliable ICMP
// close signal), so this is deliberately limited to confirmed-live targets.
func (s *Scanner) udpEnrich(ctx context.Context, devices []Device) {
	if len(devices) == 0 {
		return
	}
	timeout := s.dialTimeout
	if timeout <= 0 {
		timeout = 400 * time.Millisecond
	}
	if timeout > 1500*time.Millisecond {
		timeout = 1500 * time.Millisecond
	}
	sem := make(chan struct{}, s.conc)
	var wg sync.WaitGroup
	for i := range devices {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			ports := probeUDPPorts(ctx, devices[i].IP, timeout)
			if len(ports) == 0 {
				return
			}
			devices[i].OpenPorts = append(devices[i].OpenPorts, ports...)
			devices[i].Services = appendUnique(devices[i].Services, collectServices(ports)...)
		}(i)
	}
	wg.Wait()
}

// mdnsEnrich merges mDNS/DNS-SD signals into the discovered devices:
// announced phone-relevant services (AirPlay, HiSuite, Miracast, ...),
// the AirPlay model string, the mDNS hostname and the AirPlay deviceid (MAC).
// Devices that did not answer the multicast browse are probed unicast.
func (s *Scanner) mdnsEnrich(ctx context.Context, devices []Device) {
	browse := browseMDNS(ctx, 2*time.Second)
	for i := range devices {
		d := &devices[i]
		info, ok := browse[d.IP]
		if !ok {
			info = probeMDNSHost(ctx, d.IP, 800*time.Millisecond)
			if info.empty() {
				continue
			}
		}
		if d.Hostname == "" && info.Hostname != "" {
			d.Hostname = info.Hostname
		}
		if d.MAC == "" && info.DeviceID != "" {
			if mac := normalizeMAC(info.DeviceID); mac != "" {
				d.MAC = mac
			}
		}
		if d.Model == "" && info.Model != "" {
			d.Model = info.Model
		}
		if svcs := info.serviceList(); len(svcs) > 0 {
			d.Services = appendUnique(d.Services, svcs...)
			d.OpenPorts = appendPortOnce(d.OpenPorts, Port{Port: 5353, Proto: "udp", Service: "mdns"})
		}
	}
}

func appendUnique(in []string, add ...string) []string {
	seen := make(map[string]bool, len(in)+len(add))
	for _, s := range in {
		seen[s] = true
	}
	for _, s := range add {
		if s == "" || s == "unknown" || seen[s] {
			continue
		}
		seen[s] = true
		in = append(in, s)
	}
	return in
}

func appendPortOnce(in []Port, p Port) []Port {
	for _, q := range in {
		if q.Port == p.Port && q.Proto == p.Proto {
			return in
		}
	}
	return append(in, p)
}

type portResult struct {
	IP      string
	Port    int
	Banner  string
	Version string
}

func tryConnect(ctx context.Context, addr string, port int, timeout time.Duration) (bool, string, string) {
	if timeout <= 0 {
		timeout = 400 * time.Millisecond
	}
	target := net.JoinHostPort(addr, fmt.Sprintf("%d", port))
	dialer := net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return false, "", ""
	}
	defer conn.Close()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	rd := bufio.NewReaderSize(conn, 4096)

	// For HTTP/HTTPS ports send a GET request and extract Server header
	if port == 80 || port == 443 || port == 8080 || port == 8443 {
		_, _ = conn.Write([]byte("GET / HTTP/1.0\r\nHost: " + addr + "\r\nConnection: close\r\n\r\n"))
		var server, bannerLine string
		for {
			line, err := rd.ReadString('\n')
			if err != nil {
				break
			}
			line = strings.TrimRight(line, "\r\n")
			if bannerLine == "" && line != "" {
				bannerLine = line
			}
			if strings.HasPrefix(strings.ToLower(line), "server:") {
				server = strings.TrimSpace(line[7:])
			}
			if line == "" {
				break
			}
		}
		return true, bannerLine, server
	}

	if port == 3306 {
		// MySQL handshake is binary (starts with 0x0a); read it specially
		banner, version := readMySQLGreeting(rd)
		if banner == "" {
			return true, "", ""
		}
		return true, banner, version
	}

	raw, err := rd.ReadString('\n')
	if err != nil && raw == "" {
		return true, "", ""
	}
	banner := strings.TrimSpace(raw)
	if banner == "" {
		return true, "", ""
	}
	return true, banner, parseBannerVersion(port, banner)
}

// readMySQLGreeting reads the MySQL server handshake banner. The greeting is
// binary: the protocol version byte (0x0a) followed by a NUL-terminated server
// version string (e.g. "5.5.62-log"). The generic newline-based read would stop
// at the 0x0a byte and lose the version, so MySQL is handled separately.
func readMySQLGreeting(rd *bufio.Reader) (string, string) {
	proto := make([]byte, 1)
	if _, err := rd.Read(proto); err != nil {
		return "", ""
	}
	rest, err := rd.ReadString(0)
	if err != nil && rest == "" {
		return "", ""
	}
	rest = strings.TrimRight(rest, "\x00")
	ver := ""
	for _, r := range rest {
		if (r >= '0' && r <= '9') || r == '.' {
			ver += string(r)
		} else {
			break
		}
	}
	if ver == "" {
		return rest, ""
	}
	return rest, "mysql " + ver
}

// parseBannerVersion extracts a version string from service banners so the
// version-based vulnerability database can match. Returns "" when the banner
// does not carry a recognizable version.
func parseBannerVersion(port int, banner string) string {
	b := strings.ToLower(strings.TrimSpace(banner))
	switch port {
	case 22:
		// SSH-2.0-OpenSSH_7.2p2 Ubuntu-4ubuntu2.10
		if i := strings.Index(b, "openssh_"); i >= 0 {
			rest := b[i+len("openssh_"):]
			ver := ""
			for _, r := range rest {
				if (r >= '0' && r <= '9') || r == '.' || r == 'p' {
					ver += string(r)
				} else {
					break
				}
			}
			if ver != "" {
				return "OpenSSH_" + ver
			}
		}
	case 5432:
		// PostgreSQL 14.5 (Ubuntu ...) on x86_64...
		if strings.HasPrefix(b, "postgresql ") {
			rest := strings.TrimSpace(b[len("postgresql "):])
			for i, r := range rest {
				if r == ' ' || r == '(' {
					return "postgresql " + rest[:i]
				}
			}
			return "postgresql " + rest
		}
	case 6379:
		// Redis: "5.0.7"
		ver := ""
		for _, r := range b {
			if (r >= '0' && r <= '9') || r == '.' {
				ver += string(r)
			} else {
				break
			}
		}
		if ver != "" {
			return "redis " + ver
		}
	}
	return ""
}

func groupPortResults(results []portResult) map[string][]Port {
	m := make(map[string][]Port)
	for _, r := range results {
		svc := portService(r.Port)
		if detected := detectService(r.Port, r.Banner); detected != "" {
			svc = detected
		}
		m[r.IP] = append(m[r.IP], Port{Port: r.Port, Proto: "tcp", Service: svc, Version: r.Version, Banner: r.Banner})
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
		80: "http", 110: "pop3", 123: "ntp", 135: "msrpc", 139: "netbios-ssn", 143: "imap",
		161: "snmp", 443: "https", 445: "microsoft-ds", 993: "imaps", 995: "pop3s",
		1900: "ssdp", 3306: "mysql", 3333: "http-service", 3389: "ms-wbt-server", 5432: "postgresql", 6379: "redis",
		8080: "http-proxy", 8443: "https-alt", 27017: "mongod",
		515: "printer-lpd", 554: "rtsp", 548: "afp", 631: "ipp", 8554: "rtsp-alt", 9100: "printer-jetdirect",
		1723: "pptp", 5001: "synology-webapi",
		5555: "adb", 5223: "apns", 7000: "airplay",
	}
	if s, ok := m[port]; ok {
		return s
	}
	return "unknown"
}

func detectService(port int, banner string) string {
	b := strings.ToLower(banner)
	switch {
	case strings.Contains(b, "ssh-"):
		return "ssh"
	case strings.Contains(b, "220") && (strings.Contains(b, "ftp") || strings.Contains(b, "vsftpd")):
		return "ftp"
	case strings.HasPrefix(b, "http/") || strings.HasPrefix(b, "http"):
		return "http"
	case strings.Contains(b, "mysql") || strings.Contains(b, "mariadb"):
		return "mysql"
	case strings.HasPrefix(b, "postgresql") || strings.Contains(b, "postgresql"):
		return "postgresql"
	case strings.HasPrefix(b, "redis"):
		return "redis"
	case strings.HasPrefix(b, "rtsp"):
		return "rtsp"
	}
	return portService(port)
}

func isDomainController(ports []Port) bool {
	hasKDC := false
	hasLDAP := false
	for _, p := range ports {
		if p.Port == 88 {
			hasKDC = true
		}
		if p.Port == 389 || p.Port == 636 || p.Port == 3268 {
			hasLDAP = true
		}
	}
	return hasKDC && hasLDAP
}

func detectDomain(ctx context.Context, addr string) string {
	l, err := ldap.Dial("tcp", net.JoinHostPort(addr, "389"))
	if err != nil {
		return ""
	}
	defer l.Close()

	if err := l.UnauthenticatedBind(""); err != nil {
		return ""
	}

	search := ldap.NewSearchRequest("", ldap.ScopeBaseObject, ldap.NeverDerefAliases,
		0, 0, false, "(objectClass=*)", []string{"defaultNamingContext"}, nil)
	resp, err := l.Search(search)
	if err != nil || len(resp.Entries) == 0 {
		return ""
	}

	dn := resp.Entries[0].GetAttributeValue("defaultNamingContext")
	if dn == "" {
		return ""
	}

	parts := strings.Split(dn, ",")
	var domainParts []string
	for _, p := range parts {
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(p)), "dc=") {
			domainParts = append(domainParts, strings.TrimSpace(p)[3:])
		}
	}
	return strings.Join(domainParts, ".")
}

func resolveDeviceName(addr string, ports []Port) string {
	if name := netbiosNameQuery(addr); name != "" {
		return name
	}
	for _, p := range ports {
		if p.Port == 22 && strings.HasPrefix(p.Banner, "SSH-") {
			parts := strings.SplitN(p.Banner, " ", 2)
			if len(parts) > 1 {
				return parts[1]
			}
		}
	}
	return ""
}

func netbiosNameQuery(addr string) string {
	conn, err := net.DialTimeout("udp", net.JoinHostPort(addr, "137"), 3*time.Second)
	if err != nil {
		return ""
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	// NBSTAT request header (12 bytes) + encoded "*" name + type NBSTAT + class IN
	req := []byte{
		0x00, 0x01, // transaction ID
		0x00, 0x10, // flags: query, broadcast
		0x00, 0x01, // questions: 1
		0x00, 0x00, // answer RRs: 0
		0x00, 0x00, // authority RRs: 0
		0x00, 0x00, // additional RRs: 0
	}
	// Name encoded: "*" + 15 spaces → 32 bytes + terminator
	name := []byte("CK")
	for i := 1; i < 16; i++ {
		name = append(name, "CA"...)
	}
	req = append(req, name...)
	req = append(req, 0x00)       // name terminator
	req = append(req, 0x00, 0x21) // type NBSTAT
	req = append(req, 0x00, 0x01) // class IN

	if _, err := conn.Write(req); err != nil {
		return ""
	}

	resp := make([]byte, 1024)
	n, err := conn.Read(resp)
	if err != nil || n < 57 {
		return ""
	}

	// Response structure (after header + question):
	// - 2 bytes name flags
	// - 1 byte unit ID
	// - 18*1 byte jumper
	// - 1 byte number of names
	// For each name:
	//   - 15 bytes padded name
	//   - 1 byte type (0x00 = workstation)
	//   - 2 bytes flags
	// The computer name is usually at the first entry with type=0x00
	numNames := int(resp[56])
	if numNames == 0 || n < 57+numNames*18 {
		return ""
	}

	for i := 0; i < numNames; i++ {
		offset := 57 + i*18
		nameType := resp[offset+15]
		if nameType != 0x00 && nameType != 0x20 {
			continue
		}
		raw := strings.TrimRight(string(resp[offset:offset+15]), " ")
		return raw
	}

	return ""
}

func hasSMB(ports []Port) bool {
	for _, p := range ports {
		if p.Port == 445 || p.Port == 139 {
			return true
		}
	}
	return false
}

func smbEnumerate(addr string) ([]string, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, "445"), 3*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     "guest",
			Password: "",
		},
	}
	session, err := d.Dial(conn)
	if err != nil {
		return nil, err
	}
	defer session.Logoff()

	shares, err := session.ListSharenames()
	if err != nil {
		return nil, err
	}
	return shares, nil
}

func guessOS(ports []Port) string {
	var hasRDP, hasSMB, hasSSH, hasTelnet, hasHTTP, hasHTTPS bool
	for _, p := range ports {
		switch p.Port {
		case 3389:
			hasRDP = true
		case 445, 139:
			hasSMB = true
		case 22:
			hasSSH = true
		case 23:
			hasTelnet = true
		case 80, 8080:
			hasHTTP = true
		case 443, 8443:
			hasHTTPS = true
		}
	}
	switch {
	case hasRDP && hasSMB:
		return "Windows Server"
	case hasRDP:
		return "Windows"
	case hasSSH && !hasRDP && (hasHTTP || hasHTTPS):
		return "Linux"
	case hasSSH && hasTelnet:
		return "Network Device"
	case hasSSH:
		return "Linux/Unix"
	}
	return "unknown"
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func hostsFromCIDR(ipnet *net.IPNet) []string {
	if ipnet == nil {
		return nil
	}
	ip4 := ipnet.IP.To4()
	if ip4 == nil {
		return nil
	}
	mask := ipnet.Mask
	if len(mask) == net.IPv6len {
		mask = mask[12:]
	}
	network := ip4.Mask(mask)
	n := &net.IPNet{IP: network, Mask: mask}
	ip := make(net.IP, len(network))
	copy(ip, network)
	ones, bits := mask.Size()
	skipNetworkBroadcast := bits == 32 && bits-ones >= 2
	var ips []string
	for n.Contains(ip) {
		last := ip[len(ip)-1]
		if !ip.IsUnspecified() && (!skipNetworkBroadcast || (last != 0 && last != 255)) {
			ips = append(ips, ip.String())
		}
		incIP(ip)
	}
	return ips
}

func ResolveScanCIDRs(requested string) []string {
	requested = strings.TrimSpace(requested)
	if n := normalizeCIDR(requested); n != "" {
		requested = n
	}
	locals := LocalIPv4CIDRs()
	out := pickScanCIDRs(requested, locals)
	log.Printf("scan: requested %q resolved to %v (local physical RFC1918: %v)", requested, out, locals)
	return out
}

func pickScanCIDRs(requested string, locals []string) []string {
	requested = strings.TrimSpace(requested)
	if requested == "" || strings.EqualFold(requested, "auto") {
		if len(locals) == 0 {
			return []string{"192.168.0.0/24"}
		}
		return locals
	}
	requested = capCIDRToSlash24(requested)
	if cidrOverlapsAny(requested, locals) {
		return []string{requested}
	}
	if len(locals) > 0 {
		log.Printf("scan: %s is not on a local interface, using %v", requested, locals)
		return locals
	}
	return []string{requested}
}

func normalizeCIDR(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "auto") {
		return s
	}
	if !strings.Contains(s, "/") {
		if net.ParseIP(s) != nil {
			return s + "/32"
		}
		return s
	}
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return s
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return s
	}
	ones, _ := ipnet.Mask.Size()
	network := ip4.Mask(ipnet.Mask)
	if ones >= 24 {
		return fmt.Sprintf("%s/%d", network.String(), ones)
	}
	return capCIDRToSlash24(fmt.Sprintf("%s/%d", network.String(), ones))
}

func capCIDRToSlash24(cidr string) string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return cidr
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return cidr
	}
	ones, bits := ipnet.Mask.Size()
	if bits != 32 || ones >= 24 {
		return cidr
	}
	mask := net.CIDRMask(24, 32)
	network := ip4.Mask(mask)
	capped := fmt.Sprintf("%s/24", network.String())
	log.Printf("scan: capping %s to %s", cidr, capped)
	return capped
}

var virtualIfacePrefixes = []string{
	"docker", "br-", "veth", "virbr", "cni", "flannel", "tun", "tap",
	"wg", "lxc", "vboxnet", "vmnet", "kube-", "cali", "nodelocal",
}

func skipVirtualIface(name string, flags net.Flags) bool {
	if flags&net.FlagLoopback != 0 {
		return true
	}
	if flags&net.FlagUp == 0 {
		return true
	}
	n := strings.ToLower(name)
	for _, p := range virtualIfacePrefixes {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func LocalIPv4CIDRs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, iface := range ifaces {
		if skipVirtualIface(iface.Name, iface.Flags) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil || ip4.IsLoopback() || !isRFC1918(ip4) {
				continue
			}
			mask := ipnet.Mask
			if len(mask) == net.IPv6len {
				mask = mask[12:]
			}
			ones, bits := mask.Size()
			if bits != 32 || ones < 24 || ones > 30 {
				ones = 24
				mask = net.CIDRMask(24, 32)
			}
			network := ip4.Mask(mask)
			cidr := fmt.Sprintf("%s/%d", network.String(), ones)
			if seen[cidr] {
				continue
			}
			seen[cidr] = true
			out = append(out, cidr)
		}
	}
	return out
}

func isRFC1918(ip net.IP) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	return ip4[0] == 10 || (ip4[0] == 192 && ip4[1] == 168) || (ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31)
}

func localIPv4Addresses() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil || ip4.IsLoopback() {
				continue
			}
			out = append(out, ip4.String())
		}
	}
	return out
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

func cidrOverlapsAny(cidr string, locals []string) bool {
	_, want, err := net.ParseCIDR(cidr)
	if err != nil {
		ip := net.ParseIP(cidr)
		if ip == nil {
			return false
		}
		for _, l := range locals {
			_, n, err := net.ParseCIDR(l)
			if err == nil && n.Contains(ip) {
				return true
			}
		}
		return false
	}
	for _, l := range locals {
		_, n, err := net.ParseCIDR(l)
		if err != nil {
			continue
		}
		if n.Contains(want.IP) || want.Contains(n.IP) {
			return true
		}
	}
	return false
}

func (s *Scanner) markPingAndARP(ctx context.Context, ips []string, grouped map[string][]Port, macTable map[string]string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.conc)
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

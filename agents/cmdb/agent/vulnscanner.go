package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Vulnerability struct {
	Name        string `json:"name"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	CVE         string `json:"cve,omitempty"`
	Remediation string `json:"remediation,omitempty"`
}

type VulnScanner struct{}

func NewVulnScanner() *VulnScanner {
	return &VulnScanner{}
}

func (v *VulnScanner) Scan(ctx context.Context, device *Device) []Vulnerability {
	var mu sync.Mutex
	var vulns []Vulnerability
	var wg sync.WaitGroup

	add := func(v Vulnerability) {
		mu.Lock()
		vulns = append(vulns, v)
		mu.Unlock()
	}

	for _, p := range device.OpenPorts {
		switch p.Port {
		case 443, 8443, 636, 3269:
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				checkSSL(ctx, device.IP, port, add)
			}(p.Port)
		case 21:
			wg.Add(1)
			go func() {
				defer wg.Done()
				checkFTP(ctx, device.IP, add)
			}()
		case 22:
			wg.Add(1)
			go func() {
				defer wg.Done()
				checkSSH(ctx, device.IP, add)
			}()
		case 23:
			add(Vulnerability{
				Name:        "Telnet Enabled",
				Severity:    "HIGH",
				Description: "Telnet service is running. Telnet transmits credentials and data in cleartext.",
				Remediation: "Disable Telnet and use SSH instead.",
			})
		case 139, 445:
			wg.Add(1)
			go func() {
				defer wg.Done()
				checkSMB(ctx, device.IP, add)
			}()
		}

		if p.Service == "http" || p.Port == 80 || p.Port == 8080 {
			if !hasSecurePort(device.OpenPorts) {
				add(Vulnerability{
					Name:        "HTTP Without HTTPS",
					Severity:    "MEDIUM",
					Description: "Web server running HTTP without HTTPS. Traffic is not encrypted.",
					Remediation: "Enable HTTPS and redirect HTTP to HTTPS.",
				})
			}
		}
	}

	// version‑based checks
	for _, p := range device.OpenPorts {
		if p.Version != "" {
			for _, m := range knownVulns {
				if m.port == p.Port && versionMatch(p.Version, m.version, m.op) {
					add(Vulnerability{
						Name:        m.name,
						Severity:    m.severity,
						Description: m.description,
						CVE:         m.cve,
						Remediation: m.remediation,
					})
				}
			}
		}
	}

	wg.Wait()
	return vulns
}

func hasSecurePort(ports []Port) bool {
	for _, p := range ports {
		if p.Port == 443 || p.Port == 8443 || p.Port == 636 || p.Port == 3269 {
			return true
		}
	}
	return false
}

func checkSSL(ctx context.Context, addr string, port int, add func(Vulnerability)) {
	conn, err := tls.Dial("tcp", net.JoinHostPort(addr, fmt.Sprintf("%d", port)),
		&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer conn.Close()

	state := conn.ConnectionState()
	certs := state.PeerCertificates
	if len(certs) == 0 {
		return
	}
	cert := certs[0]

	// Expiration
	remaining := time.Until(cert.NotAfter)
	if remaining < 30*24*time.Hour {
		sev := "HIGH"
		if remaining < 0 {
			sev = "CRITICAL"
		}
		add(Vulnerability{
			Name:        "SSL Certificate Expiry",
			Severity:    sev,
			Description: fmt.Sprintf("TLS certificate expires in %.0f days (on %s).", remaining.Hours()/24, cert.NotAfter.Format("2006-01-02")),
			Remediation: "Renew the certificate before it expires.",
		})
	}

	// Self‑signed
	if cert.AuthorityKeyId == nil || string(cert.AuthorityKeyId) == string(cert.SubjectKeyId) {
		isCA := cert.IsCA
		if !isCA {
			if len(cert.Subject.CommonName) > 0 && !strings.Contains(cert.Issuer.CommonName, "CA") {
				add(Vulnerability{
					Name:        "Self-Signed SSL Certificate",
					Severity:    "MEDIUM",
					Description: fmt.Sprintf("TLS certificate is self‑signed (CN: %s).", cert.Subject.CommonName),
					Remediation: "Replace with a certificate signed by a trusted CA.",
				})
			}
		}
	}

	// Weak TLS version
	switch state.Version {
	case tls.VersionTLS10:
		add(Vulnerability{
			Name:        "TLS 1.0 Enabled",
			Severity:    "HIGH",
			Description: "Server accepts TLS 1.0 connections. This protocol version is deprecated and insecure.",
			CVE:         "CVE-2011-3389",
			Remediation: "Disable TLS 1.0 and TLS 1.1. Enable TLS 1.2 or higher.",
		})
	case tls.VersionTLS11:
		add(Vulnerability{
			Name:        "TLS 1.1 Enabled",
			Severity:    "MEDIUM",
			Description: "Server accepts TLS 1.1 connections. This protocol version is deprecated.",
			Remediation: "Disable TLS 1.1. Enable TLS 1.2 or higher.",
		})
	}
}

func checkFTP(ctx context.Context, addr string, add func(Vulnerability)) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, "21"), 3*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	banner := string(buf[:n])

	if strings.Contains(banner, "220") {
		// Try anonymous login
		_, _ = conn.Write([]byte("USER anonymous\r\n"))
		n, _ = conn.Read(buf)
		resp := string(buf[:n])
		if strings.Contains(resp, "331") {
			_, _ = conn.Write([]byte("PASS anonymous@\r\n"))
			n, _ = conn.Read(buf)
			resp = string(buf[:n])
			if strings.Contains(resp, "230") || strings.Contains(resp, "2") {
				add(Vulnerability{
					Name:        "Anonymous FTP Access",
					Severity:    "HIGH",
					Description: "FTP server allows anonymous login.",
					Remediation: "Disable anonymous FTP access or restrict permissions.",
				})
			}
		}
	}
}

func checkSSH(ctx context.Context, addr string, add func(Vulnerability)) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, "22"), 3*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	buf := make([]byte, 256)
	n, _ := conn.Read(buf)
	banner := string(buf[:n])

	// Check for SSH on non-standard port or weak banner
	if strings.Contains(banner, "dropbear") || strings.Contains(strings.ToLower(banner), "1.99") {
		add(Vulnerability{
			Name:        "Weak SSH Server",
			Severity:    "MEDIUM",
			Description: "SSH server is running a potentially vulnerable or outdated implementation.",
			Remediation: "Upgrade SSH server to latest stable version.",
		})
	}

	_ = banner
}

func checkSMB(ctx context.Context, addr string, add func(Vulnerability)) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(addr, "445"), 3*time.Second)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(3 * time.Second))

	// SMB2 negotiate request (minimal)
	negotiate := []byte{
		0x00, 0x00, 0x00, 0x00, // smb2 header
		0xfe, 0x53, 0x4d, 0x42, // protocol id
		0x40, 0x00, // structure size
		0x00, 0x00, 0x00, 0x00, // credit charge
		0x00, 0x00, // status
		0x00, 0x00, // command: negotiate
		0x00, 0x00, 0x00, 0x00, // credits
		0x00, 0x00, 0x00, 0x00, // flags
		0x00, 0x00, 0x00, 0x00, // next command
		0x00, 0x00, 0x00, 0x00, // message id
		0x00, 0x00, 0x00, 0x00, // process id
		0x00, 0x00, 0x00, 0x00, // tree id
		0x00, 0x00, 0x00, 0x00, // session id
		0x00, 0x00, 0x00, 0x00, // signature
		0x00, 0x00, 0x00, 0x00,
		// negotiate request
		0x24, 0x00, // structure size
		0x00, 0x00, // dialect count
		0x00, 0x00, 0x00, 0x00, // security mode
		0x00, 0x00, 0x00, 0x00, // reserved
		0x00, 0x00, 0x00, 0x00, // capabilities
		0x00, 0x00, 0x00, 0x00, // client guid
		0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, // negotiate context offset
		0x00, 0x00, 0x00, 0x00, // negotiate context count
		0x00, 0x00, 0x00, 0x00, // reserved
	}
	_ = negotiate
}

// version-based vulnerability DB

type vulnMatch struct {
	port                                                       int
	version, op, name, severity, description, cve, remediation string
}

func versionMatch(version, target, op string) bool {
	// simple prefix/substring matching for now
	switch op {
	case "contains":
		return strings.Contains(strings.ToLower(version), strings.ToLower(target))
	case "prefix":
		return strings.HasPrefix(strings.ToLower(version), strings.ToLower(target))
	case "lt":
		// minimal numeric compare
		return version < target
	}
	return false
}

var knownVulns = []vulnMatch{
	{22, "OpenSSH_7.2", "lt", "OpenSSH < 7.2", "HIGH", "Weak SSH key exchange algorithms allowed.", "CVE-2016-0777", "Upgrade OpenSSH to 7.2+"},
	{22, "OpenSSH_7.9", "lt", "OpenSSH < 7.9", "MEDIUM", "Vulnerable to CVE-2018-15473 user enumeration.", "CVE-2018-15473", "Upgrade OpenSSH to 7.9+"},
	{80, "Apache/2.4.49", "contains", "Apache HTTP Server 2.4.49", "CRITICAL", "Path traversal and RCE vulnerability.", "CVE-2021-41773", "Upgrade Apache to 2.4.51+"},
	{80, "Apache/2.4.50", "contains", "Apache HTTP Server 2.4.50", "CRITICAL", "Path traversal vulnerability.", "CVE-2021-42013", "Upgrade Apache to 2.4.51+"},
	{443, "Apache/2.4.49", "contains", "Apache HTTP Server 2.4.49", "CRITICAL", "Path traversal and RCE vulnerability.", "CVE-2021-41773", "Upgrade Apache to 2.4.51+"},
	{443, "Apache/2.4.50", "contains", "Apache HTTP Server 2.4.50", "CRITICAL", "Path traversal vulnerability.", "CVE-2021-42013", "Upgrade Apache to 2.4.51+"},
	{80, "nginx/1.20.0", "contains", "Nginx 1.20.0", "MEDIUM", "Known vulnerabilities in this nginx version.", "", "Upgrade nginx to latest stable."},
	{80, "nginx/1.18.0", "contains", "Nginx 1.18.0", "MEDIUM", "Known vulnerabilities in this nginx version.", "", "Upgrade nginx to latest stable."},
	{3306, "mysql 5.5", "contains", "MySQL 5.5", "HIGH", "MySQL 5.5 has reached end of life and has known vulnerabilities.", "", "Upgrade MySQL to 8.0+"},
	{3306, "mysql 5.6", "contains", "MySQL 5.6", "HIGH", "MySQL 5.6 has reached end of life and has known vulnerabilities.", "", "Upgrade MySQL to 8.0+"},
	{3306, "mysql 5.7", "contains", "MySQL 5.7", "MEDIUM", "MySQL 5.7 is approaching end of life.", "", "Upgrade MySQL to 8.0+"},
	{5432, "postgresql 9.", "contains", "PostgreSQL 9.x", "HIGH", "PostgreSQL 9.x has reached end of life.", "", "Upgrade PostgreSQL to 15+"},
	{5432, "postgresql 10.", "contains", "PostgreSQL 10.x", "HIGH", "PostgreSQL 10.x has reached end of life.", "", "Upgrade PostgreSQL to 15+"},
	{5432, "postgresql 11.", "contains", "PostgreSQL 11.x", "MEDIUM", "PostgreSQL 11.x is approaching end of life.", "", "Upgrade PostgreSQL to 15+"},
	{6379, "redis 2.", "contains", "Redis 2.x", "HIGH", "Redis 2.x has reached end of life.", "", "Upgrade Redis to 7+"},
	{6379, "redis 3.", "contains", "Redis 3.x", "HIGH", "Redis 3.x has reached end of life.", "", "Upgrade Redis to 7+"},
	{6379, "redis 4.", "contains", "Redis 4.x", "MEDIUM", "Redis 4.x has reached end of life.", "", "Upgrade Redis to 7+"},
}

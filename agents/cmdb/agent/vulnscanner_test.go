package agent

import (
	"bufio"
	"bytes"
	"testing"
)

func TestParseBannerVersion(t *testing.T) {
	cases := []struct {
		port   int
		banner string
		want   string
	}{
		{22, "SSH-2.0-OpenSSH_7.2p2 Ubuntu-4ubuntu2.10", "OpenSSH_7.2p2"},
		{22, "SSH-2.0-OpenSSH_8.9p1 Ubuntu-3ubuntu0.6", "OpenSSH_8.9p1"},
		{22, "SSH-2.0-dropbear_2018.76", ""},
		{5432, "PostgreSQL 14.5 (Ubuntu 14.5-0ubuntu0.22.04.1) on x86_64-pc-linux-gnu", "postgresql 14.5"},
		{5432, "PostgreSQL 9.4.26 on x86_64-pc-linux-gnu", "postgresql 9.4.26"},
		{6379, "5.0.7", "redis 5.0.7"},
		{6379, "2.8.24", "redis 2.8.24"},
		{80, "some random banner", ""},
	}
	for _, c := range cases {
		got := parseBannerVersion(c.port, c.banner)
		if got != c.want {
			t.Errorf("parseBannerVersion(%d, %q) = %q, want %q", c.port, c.banner, got, c.want)
		}
	}
}

func TestReadMySQLGreeting(t *testing.T) {
	// protocol version byte (0x0a) + NUL-terminated version
	banner, version := readMySQLGreeting(bufio.NewReader(bytes.NewReader([]byte{0x0a, '5', '.', '5', '.', '6', '2', '-', 'l', 'o', 'g', 0x00, 0x01, 0x02})))
	if banner != "5.5.62-log" {
		t.Errorf("banner = %q, want %q", banner, "5.5.62-log")
	}
	if version != "mysql 5.5.62" {
		t.Errorf("version = %q, want %q", version, "mysql 5.5.62")
	}
}

func TestVersionLessThan(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"OpenSSH_7.2p2", "OpenSSH_7.9", true},
		{"OpenSSH_8.0", "OpenSSH_7.9", false},
		{"OpenSSH_7.2", "OpenSSH_7.2", false},
		{"mysql 5.5.62", "mysql 5.5", false}, // 5.5.62 is not < 5.5
		{"nginx/1.18.0", "nginx/1.20.0", true},
	}
	for _, c := range cases {
		if got := versionLessThan(c.a, c.b); got != c.want {
			t.Errorf("versionLessThan(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestVersionMatch(t *testing.T) {
	cases := []struct {
		version, target, op string
		want                bool
	}{
		{"OpenSSH_7.2p2", "OpenSSH_7.9", "lt", true},
		{"OpenSSH_8.0", "OpenSSH_7.9", "lt", false},
		{"OpenSSH_7.2p2", "OpenSSH_7.2", "lt", false},
		{"mysql 5.5.62", "mysql 5.5", "contains", true},
		{"nginx/1.20.0", "nginx/1.20.0", "contains", true},
		{"redis 5.0.7", "redis 2.", "contains", false},
		{"OpenSSH_7.2p2", "OpenSSH_7", "prefix", true},
	}
	for _, c := range cases {
		if got := versionMatch(c.version, c.target, c.op); got != c.want {
			t.Errorf("versionMatch(%q, %q, %q) = %v, want %v", c.version, c.target, c.op, got, c.want)
		}
	}
}

// TestScanPopulatesCVE ensures version-based checks carry a CVE into the
// reported vulnerability.
func TestScanPopulatesCVE(t *testing.T) {
	v := NewVulnScanner()
	dev := &Device{
		IP: "127.0.0.1",
		OpenPorts: []Port{
			{Port: 22, Service: "ssh", Version: "OpenSSH_7.2p2"},
			{Port: 3306, Service: "mysql", Version: "mysql 5.5.62"},
			{Port: 6379, Service: "redis", Version: "redis 4.0.10"},
		},
	}
	vulns := v.Scan(t.Context(), dev)
	if len(vulns) == 0 {
		t.Fatal("expected version-based vulnerabilities")
	}
	for _, vul := range vulns {
		if vul.CVE == "" {
			t.Errorf("vulnerability %q has empty CVE", vul.Name)
		}
	}
}

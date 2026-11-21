package agent

import (
	"net"
	"testing"
)

func TestHostsFromCIDR_IPv4Slash24(t *testing.T) {
	_, n, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	ips := hostsFromCIDR(n)
	if len(ips) != 254 {
		t.Fatalf("got %d hosts, want 254 (To4 host expansion)", len(ips))
	}
	if ips[0] != "192.168.1.1" || ips[len(ips)-1] != "192.168.1.254" {
		t.Fatalf("range %s .. %s", ips[0], ips[len(ips)-1])
	}
}

func TestHostsFromCIDR_IPv4Slash32(t *testing.T) {
	_, n, err := net.ParseCIDR("10.0.0.5/32")
	if err != nil {
		t.Fatal(err)
	}
	ips := hostsFromCIDR(n)
	if len(ips) != 1 || ips[0] != "10.0.0.5" {
		t.Fatalf("got %v", ips)
	}
}

func TestSkipVirtualIface(t *testing.T) {
	if !skipVirtualIface("br-72de16c046e7", net.FlagUp|net.FlagRunning) {
		t.Fatal("docker bridge should be skipped")
	}
	if !skipVirtualIface("docker0", net.FlagUp) {
		t.Fatal("docker0 should be skipped")
	}
	if !skipVirtualIface("lo", net.FlagLoopback|net.FlagUp) {
		t.Fatal("loopback should be skipped")
	}
	if skipVirtualIface("wlp98s0", net.FlagUp|net.FlagRunning) {
		t.Fatal("wifi should be scanned")
	}
	if skipVirtualIface("enp3s0", net.FlagUp) {
		t.Fatal("ethernet should be scanned")
	}
}

func TestCapCIDRToSlash24(t *testing.T) {
	got := capCIDRToSlash24("172.20.0.0/16")
	if got != "172.20.0.0/24" {
		t.Fatalf("got %s", got)
	}
	got = capCIDRToSlash24("192.168.0.0/24")
	if got != "192.168.0.0/24" {
		t.Fatalf("got %s", got)
	}
}

func TestPickScanCIDRs_WrongOfficeLANFallsBackToWifi(t *testing.T) {
	locals := []string{"192.168.0.0/24"}
	got := pickScanCIDRs("192.168.1.0/24", locals)
	if len(got) != 1 || got[0] != "192.168.0.0/24" {
		t.Fatalf("got %v, want wifi /24 not docker /16", got)
	}
	got = pickScanCIDRs("10.0.0.0/24", locals)
	if len(got) != 1 || got[0] != "192.168.0.0/24" {
		t.Fatalf("got %v", got)
	}
	got = pickScanCIDRs("192.168.0.0/24", locals)
	if len(got) != 1 || got[0] != "192.168.0.0/24" {
		t.Fatalf("got %v", got)
	}
	got = pickScanCIDRs("auto", locals)
	if len(got) != 1 || got[0] != "192.168.0.0/24" {
		t.Fatalf("got %v", got)
	}
}

func TestPickScanCIDRs_DoesNotIncludeDocker16(t *testing.T) {
	locals := []string{"192.168.0.0/24"}
	got := pickScanCIDRs("192.168.1.0/24", locals)
	for _, c := range got {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			t.Fatal(err)
		}
		ones, _ := n.Mask.Size()
		if ones < 24 {
			t.Fatalf("scan CIDR too wide: %s", c)
		}
	}
}

func TestNormalizeCIDR_HostSlash24(t *testing.T) {
	got := normalizeCIDR("192.168.0.126/24")
	if got != "192.168.0.0/24" {
		t.Fatalf("got %s", got)
	}
}

func TestLocalIPv4CIDRs_NoDockerSlash16(t *testing.T) {
	for _, c := range LocalIPv4CIDRs() {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			t.Fatal(err)
		}
		ones, _ := n.Mask.Size()
		if ones < 24 {
			t.Fatalf("physical local CIDR too wide (docker /16 leak): %s", c)
		}
	}
	got := ResolveScanCIDRs("192.168.1.0/24")
	t.Logf("192.168.1.0/24 → %v (locals %v)", got, LocalIPv4CIDRs())
	for _, c := range got {
		_, n, err := net.ParseCIDR(c)
		if err != nil {
			t.Fatal(err)
		}
		ones, bits := n.Mask.Size()
		if bits == 32 && ones < 24 {
			t.Fatalf("resolved CIDR too wide: %s", c)
		}
	}
}

package scancidr

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestHostsFromCIDR_IPv4Slash24(t *testing.T) {
	_, n, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatal(err)
	}
	ips := HostsFromCIDR(n)
	if len(ips) != 254 {
		t.Fatalf("got %d hosts, want 254", len(ips))
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
	ips := HostsFromCIDR(n)
	if len(ips) != 1 || ips[0] != "10.0.0.5" {
		t.Fatalf("got %v", ips)
	}
}

func TestSkipVirtualIface(t *testing.T) {
	if !SkipVirtualIface("br-72de16c046e7", net.FlagUp|net.FlagRunning) {
		t.Fatal("docker bridge should be skipped")
	}
	if !SkipVirtualIface("docker0", net.FlagUp) {
		t.Fatal("docker0 should be skipped")
	}
	if !SkipVirtualIface("lo", net.FlagLoopback|net.FlagUp) {
		t.Fatal("loopback should be skipped")
	}
	if SkipVirtualIface("wlp98s0", net.FlagUp|net.FlagRunning) {
		t.Fatal("wifi should be scanned")
	}
}

func TestCapCIDRToSlash24(t *testing.T) {
	got := CapCIDRToSlash24("172.20.0.0/16")
	if got != "172.20.0.0/24" {
		t.Fatalf("got %s", got)
	}
	got = CapCIDRToSlash24("192.168.0.0/24")
	if got != "192.168.0.0/24" {
		t.Fatalf("got %s", got)
	}
}

func TestPickScanCIDRs_WrongOfficeLANFallsBackToWifi(t *testing.T) {
	locals := []string{"192.168.0.0/24"}
	got := PickScanCIDRs("192.168.1.0/24", locals)
	if len(got) != 1 || got[0] != "192.168.0.0/24" {
		t.Fatalf("got %v, want wifi /24", got)
	}
	got = PickScanCIDRs("auto", locals)
	if len(got) != 1 || got[0] != "192.168.0.0/24" {
		t.Fatalf("got %v", got)
	}
}

func TestNormalizeCIDR_HostSlash24(t *testing.T) {
	got := NormalizeCIDR("192.168.0.126/24")
	if got != "192.168.0.0/24" {
		t.Fatalf("got %s", got)
	}
}

func TestTryConnectLocalhost(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	go func() {
		c, err := ln.Accept()
		if err == nil {
			_, _ = c.Write([]byte("hello\n"))
			_ = c.Close()
		}
	}()
	ok, banner := tryConnect(context.Background(), "127.0.0.1", port, 400*time.Millisecond)
	if !ok {
		t.Fatal("expected connect")
	}
	if banner == "" {
		t.Fatal("expected banner")
	}
}

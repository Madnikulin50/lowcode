package agent

import (
	"strings"
	"testing"
)

func TestDedupeDevicesMergesSameHostTwoIPs(t *testing.T) {
	in := []Device{
		{IP: "192.168.0.126", Hostname: "maxim-zenbook", Status: "online"},
		{IP: "192.168.173.74", Hostname: "maxim-zenbook", Status: "online", OpenPorts: []Port{{Port: 22}}},
	}
	out := DedupeDevices(in)
	if len(out) != 1 {
		t.Fatalf("got %d devices, want 1 (same hostname, no MAC)", len(out))
	}
	if out[0].IP != "192.168.173.74" {
		t.Fatalf("merged IP = %s, want the richer/later 192.168.173.74", out[0].IP)
	}
}

func TestDedupeDevicesKeepsDistinctIPsWithoutIdentity(t *testing.T) {
	in := []Device{
		{IP: "192.168.173.10"},
		{IP: "192.168.173.11"},
	}
	out := DedupeDevices(in)
	if len(out) != 2 {
		t.Fatalf("got %d, want 2", len(out))
	}
}

func TestDedupeDevicesDoesNotMergeGenericHostnames(t *testing.T) {
	in := []Device{
		{IP: "192.168.0.1", Hostname: "_gateway", MAC: "5c:a6:e6:1f:83:4e"},
		{IP: "192.168.173.20", Hostname: "_gateway", MAC: "48:a9:8a:33:ed:dd"},
	}
	out := DedupeDevices(in)
	if len(out) != 2 {
		t.Fatalf("gateways with different MACs must stay distinct, got %d", len(out))
	}
}

func TestDedupeDevicesMergesByMAC(t *testing.T) {
	in := []Device{
		{IP: "192.168.0.10", MAC: "AA-BB-CC-DD-EE-FF", Hostname: "old"},
		{IP: "192.168.173.10", MAC: "aa:bb:cc:dd:ee:ff", Hostname: "new"},
	}
	out := DedupeDevices(in)
	if len(out) != 1 {
		t.Fatalf("got %d, want 1", len(out))
	}
	if normalizeMAC(out[0].MAC) != "aa:bb:cc:dd:ee:ff" {
		t.Fatalf("mac = %s", out[0].MAC)
	}
}

func TestNormalizeMAC(t *testing.T) {
	if got := normalizeMAC("AA-BB-CC-DD-EE-F"); got != "aa:bb:cc:dd:ee:0f" {
		t.Fatalf("got %s", got)
	}
}

func TestGenericHostname(t *testing.T) {
	if !genericHostname("_gateway") || !genericHostname("192.168.1.1") {
		t.Fatal("expected generic")
	}
	if genericHostname("maxim-zenbook") {
		t.Fatal("real hostname must be stable identity")
	}
}

func TestAPIStatusErrorUnauthorized(t *testing.T) {
	plain := []byte("Error: unauthorized\nNote: you are seeing this because system is running in development mode")
	err := apiStatusError(401, plain)
	if !isUnauthorizedAPI(err) {
		t.Fatalf("plain 401 should be unauthorized: %v", err)
	}
	if strings.Contains(err.Error(), "Note: you are seeing") {
		t.Fatalf("must not dump the development-mode note: %v", err)
	}

	jsonBody := []byte(`{"error":{"message":"Error: unauthorized"}}`)
	err = apiStatusError(401, jsonBody)
	if !isUnauthorizedAPI(err) {
		t.Fatalf("json 401 should be unauthorized: %v", err)
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("got %v", err)
	}
}

package agent

import (
	"encoding/binary"
	"testing"
)

// buildMDNSResponse assembles a minimal mDNS response packet:
//
//	header + question (PTR _airplay._tcp.local)
//	answer:  PTR "iPhone-ABCDEF._airplay._tcp.local"
//	answer:  SRV  0 0 7000 iPhone-ABCDEF.local
//	answer:  TXT  "model=iPhone14,2" "deviceid=F0:2F:74:12:34:56" "srcvers=760.40.2"
//	answer:  A    iPhone-ABCDEF.local -> 192.168.1.42
func buildMDNSResponse(t *testing.T) []byte {
	t.Helper()
	pkt := make([]byte, 0, 512)
	// header
	pkt = append(pkt, 0x00, 0x01) // ID
	pkt = append(pkt, 0x84, 0x00) // flags: response
	pkt = append(pkt, 0x00, 0x01) // QDCOUNT
	pkt = append(pkt, 0x00, 0x04) // ANCOUNT (PTR, SRV, TXT, A)
	pkt = append(pkt, 0x00, 0x00) // NSCOUNT
	pkt = append(pkt, 0x00, 0x00) // ARCOUNT
	pkt = append(pkt, 0x00)       // question name: empty root
	pkt = append(pkt, 0x00, 0x0C) // QTYPE PTR
	pkt = append(pkt, 0x00, 0x01) // QCLASS IN
	addRR := func(name string, rtype uint16, rdata []byte) {
		pkt = appendDNSName(pkt, name)
		buf := make([]byte, 10)
		binary.BigEndian.PutUint16(buf[0:], rtype)
		binary.BigEndian.PutUint16(buf[2:], 0x0001) // class IN
		binary.BigEndian.PutUint32(buf[4:], 120)    // TTL
		binary.BigEndian.PutUint16(buf[8:], uint16(len(rdata)))
		pkt = append(pkt, buf...)
		pkt = append(pkt, rdata...)
	}
	// PTR -> instance name
	instance := "iPhone-ABCDEF._airplay._tcp.local"
	ptr := make([]byte, 0, 64)
	ptr = appendDNSName(ptr, instance)
	addRR("_airplay._tcp.local", 12, ptr)
	// SRV: prio 0 weight 0 port 7000 target iPhone-ABCDEF.local
	srv := make([]byte, 0, 64)
	srv = append(srv, 0x00, 0x00, 0x00, 0x00)
	srv = binary.BigEndian.AppendUint16(srv, 7000)
	srv = appendDNSName(srv, "iPhone-ABCDEF.local")
	addRR("iPhone-ABCDEF._airplay._tcp.local", 33, srv)
	// TXT
	txt := []byte{
		16, 'm', 'o', 'd', 'e', 'l', '=', 'i', 'P', 'h', 'o', 'n', 'e', '1', '4', ',', '2',
		26, 'd', 'e', 'v', 'i', 'c', 'e', 'i', 'd', '=', 'F', '0', ':', '2', 'F', ':', '7', '4', ':', '1', '2', ':', '3', '4', ':', '5', '6',
		17, 's', 'r', 'c', 'v', 'e', 'r', 's', '=', '7', '6', '0', '.', '4', '0', '.', '2',
	}
	addRR("iPhone-ABCDEF._airplay._tcp.local", 16, txt)
	// A
	a := []byte{192, 168, 1, 42}
	addRR("iPhone-ABCDEF.local", 1, a)
	return pkt
}

func TestParseMDNSResponse(t *testing.T) {
	info, ok := parseMDNSResponse(buildMDNSResponse(t))
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	if !info.Services["airplay"] {
		t.Errorf("expected airplay service, got %v", info.Services)
	}
	if info.Model != "iPhone14,2" {
		t.Errorf("expected model iPhone14,2, got %q", info.Model)
	}
	if info.DeviceID != "F0:2F:74:12:34:56" {
		t.Errorf("expected deviceid F0:2F:74:12:34:56, got %q", info.DeviceID)
	}
	if info.Hostname != "iPhone-ABCDEF" {
		t.Errorf("expected hostname iPhone-ABCDEF, got %q", info.Hostname)
	}
	if got := info.serviceList(); len(got) != 1 || got[0] != "airplay" {
		t.Errorf("serviceList = %v, want [airplay]", got)
	}
}

func TestParseMDNSResponseNotAResponse(t *testing.T) {
	pkt := []byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if _, ok := parseMDNSResponse(pkt); ok {
		t.Error("query packet must not be parsed as a response")
	}
}

func TestServiceTypeFromName(t *testing.T) {
	cases := map[string]string{
		"iPhone._airplay._tcp.local":          "airplay",
		"_airplay._tcp.local":                 "airplay",
		"abc._googlecast._tcp.local":          "googlecast",
		"some._hisuite._tcp.local":            "hisuite",
		"_services._dns-sd._udp.local":        "",
		"plain.local":                         "",
		"whatever._companion-link._tcp.local": "companion-link",
	}
	for in, want := range cases {
		got, _ := serviceTypeFromName(in)
		if got != want {
			t.Errorf("serviceTypeFromName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMergeMDNS(t *testing.T) {
	dst := map[string]mdnsInfo{}
	mergeMDNS(dst, "10.0.0.5", mdnsInfo{
		Services: map[string]bool{"airplay": true},
		TXT:      map[string]string{"model": "iPhone14,2"},
		Hostname: "iPhone-1",
	})
	mergeMDNS(dst, "10.0.0.5", mdnsInfo{
		Services: map[string]bool{"raop": true},
		TXT:      map[string]string{"deviceid": "AA:BB:CC:DD:EE:FF"},
	})
	got := dst["10.0.0.5"]
	if !got.Services["airplay"] || !got.Services["raop"] {
		t.Errorf("expected both services, got %v", got.Services)
	}
	if got.Model != "iPhone14,2" || got.DeviceID != "AA:BB:CC:DD:EE:FF" || got.Hostname != "iPhone-1" {
		t.Errorf("merge lost fields: %+v", got)
	}
}

func TestMDNSQueryPacket(t *testing.T) {
	pkt := mdnsQueryPacket("_airplay._tcp.local", true)
	if len(pkt) < 12 {
		t.Fatal("packet too short")
	}
	if binary.BigEndian.Uint16(pkt[4:6]) != 1 {
		t.Errorf("QDCOUNT = %d, want 1", binary.BigEndian.Uint16(pkt[4:6]))
	}
	// class must have QU bit set
	cls := binary.BigEndian.Uint16(pkt[len(pkt)-2:])
	if cls&0x8000 == 0 {
		t.Errorf("unicast-response bit not set, class = %#x", cls)
	}
}

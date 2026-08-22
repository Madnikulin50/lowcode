package agent

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"testing"
	"time"
)

// udpTestServer is a minimal UDP responder for the protocol probes. Every
// request received on the socket is answered with an appropriate canned reply.
func startUDPServer(t *testing.T, handler func(req []byte) []byte) (addr string) {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	addr = pc.LocalAddr().String()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, raddr, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			reply := handler(buf[:n])
			if len(reply) == 0 {
				continue
			}
			_, _ = pc.WriteTo(reply, raddr)
		}
	}()
	return addr
}

func host(addr string) string {
	h, _, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	return h
}

func portOf(addr string) int {
	_, p, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	var n int
	if _, err := fmt.Sscanf(p, "%d", &n); err != nil {
		panic(err)
	}
	return n
}

// probeUDPPortsOn exercises the same protocol-probe path as probeUDPPorts but
// against a single explicit (random) local port, so tests do not need the
// well-known 53/123/161/1900 sockets.
func probeUDPPortsOn(ctx context.Context, addr string, port int, timeout time.Duration) []Port {
	var out []Port
	if open, _ := probeDNS(ctx, addr, port, timeout); open {
		out = append(out, Port{Port: 53, Proto: "udp", Service: "dns"})
	}
	if open, v := probeNTP(ctx, addr, port, timeout); open {
		out = append(out, Port{Port: 123, Proto: "udp", Service: "ntp", Version: v})
	}
	if open, d := probeSNMP(ctx, addr, port, timeout); open {
		out = append(out, Port{Port: 161, Proto: "udp", Service: "snmp", Version: d})
	}
	if open, s := probeSSDP(ctx, addr, port, timeout); open {
		out = append(out, Port{Port: 1900, Proto: "udp", Service: "ssdp", Banner: s})
	}
	return out
}

func TestProbeDNS(t *testing.T) {
	addr := startUDPServer(t, func(req []byte) []byte {
		if len(req) < 12 {
			return nil
		}
		resp := make([]byte, 12)
		copy(resp, req[:12]) // echo ID + header (QR set via flags below)
		resp[2] = 0x81       // QR=1, RD=1
		resp[3] = 0x80       // RA
		resp[5] = 0x01       // ANCOUNT=1
		// authority section: root NS "a.root-servers.net."
		resp = append(resp, 0x00, 0x00, 0x02, 0x00, 0x01, 0x00, 0x00, 0x0e, 0x10)
		resp = append(resp, 0x01, 'a', 0x04, 'r', 'o', 'o', 't', 0x03, 'n', 'e', 't', 0x00)
		return resp
	})
	open, _ := probeDNS(context.Background(), host(addr), portOf(addr), time.Second)
	if !open {
		t.Fatal("DNS probe should report open on a responding resolver")
	}
}

func TestProbeDNS_NoReply(t *testing.T) {
	addr := startUDPServer(t, func(req []byte) []byte { return nil })
	open, _ := probeDNS(context.Background(), host(addr), portOf(addr), 300*time.Millisecond)
	if open {
		t.Fatal("DNS probe should report closed when the peer stays silent")
	}
}

func TestProbeNTP(t *testing.T) {
	addr := startUDPServer(t, func(req []byte) []byte {
		resp := make([]byte, 48)
		resp[0] = 0x24 // LI=0, VN=4, Mode=4 (server)
		resp[1] = 2    // stratum 2
		return resp
	})
	open, version := probeNTP(context.Background(), host(addr), portOf(addr), time.Second)
	if !open {
		t.Fatal("NTP probe should report open on a responding server")
	}
	if version != "NTPv4 stratum 2" {
		t.Fatalf("got version %q", version)
	}
}

func TestProbeSNMP(t *testing.T) {
	addr := startUDPServer(t, func(req []byte) []byte {
		// Echo the version/community, respond with a GetResponse (0xA2) that
		// carries the sysDescr OCTET STRING.
		if len(req) < 20 {
			return nil
		}
		sysDescr := []byte("Linux test 5.15 #1 SMP")
		resp := []byte{0x30, 0, 0x02, 0x01, 0x00, 0x04, 0x06, 'p', 'u', 'b', 'l', 'i', 'c'}
		oid := []byte{0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00}
		vbBody := []byte{0x06, byte(len(oid))}
		vbBody = append(vbBody, oid...)
		vbBody = append(vbBody, 0x04, byte(len(sysDescr)))
		vbBody = append(vbBody, sysDescr...)
		vb := append([]byte{0x30, byte(len(vbBody))}, vbBody...)
		vbl := append([]byte{0x30, byte(len(vb))}, vb...)
		pduBody := []byte{0x02, 0x04, 0, 0, 0, 1, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00}
		pduBody = append(pduBody, vbl...)
		pdu := append([]byte{0xA2, byte(len(pduBody))}, pduBody...)
		resp = append(resp, pdu...)
		resp[1] = byte(len(resp) - 2)
		return resp
	})
	open, descr := probeSNMP(context.Background(), host(addr), portOf(addr), time.Second)
	if !open {
		t.Fatal("SNMP probe should report open on a responding agent")
	}
	if descr != "Linux test 5.15 #1 SMP" {
		t.Fatalf("got sysDescr %q", descr)
	}
}

func TestProbeSSDP(t *testing.T) {
	addr := startUDPServer(t, func(req []byte) []byte {
		if !containsAny(string(req), "M-SEARCH") {
			return nil
		}
		return []byte("HTTP/1.1 200 OK\r\nST: upnp:rootdevice\r\nSERVER: Linux/3.14 UPnP/1.0 MiniUPnPd/1.9\r\n\r\n")
	})
	open, server := probeSSDP(context.Background(), host(addr), portOf(addr), time.Second)
	if !open {
		t.Fatal("SSDP probe should report open on a responding device")
	}
	if server != "Linux/3.14 UPnP/1.0 MiniUPnPd/1.9" {
		t.Fatalf("got server %q", server)
	}
}

func TestProbeUDPPorts(t *testing.T) {
	addr := startUDPServer(t, func(req []byte) []byte {
		// Multi-service responder: NTP and SNMP share one socket.
		if len(req) >= 48 && req[0]&0x07 == 0x03 && req[40] != 0 {
			resp := make([]byte, 48)
			resp[0] = 0x24
			resp[1] = 3
			return resp
		}
		if containsAny(string(req), "public") {
			oid := []byte{0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00}
			val := []byte{0x04, 0x05, 'r', 'o', 'u', 't', 'e'}
			vb := append([]byte{0x30, byte(len(oid) + len(val) + 3)}, 0x06, byte(len(oid)))
			vb = append(vb, oid...)
			vb = append(vb, val...)
			return append([]byte{0x30, 0, 0x02, 0x01, 0x00, 0x04, 0x06, 'p', 'u', 'b', 'l', 'i', 'c', 0xA2, 0, 0x02, 0x04, 0, 0, 0, 1, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, byte(len(vb))}, vb...)
		}
		return nil
	})
	// Point the probes at the shared responder: every protocol request lands on
	// the same test socket and is answered according to its content.
	ports := probeUDPPortsOn(context.Background(), host(addr), portOf(addr), 500*time.Millisecond)
	if len(ports) == 0 {
		t.Fatal("expected NTP/SNMP services from the shared responder")
	}
	seen := map[string]bool{}
	for _, p := range ports {
		seen[p.Service] = true
		if p.Proto != "udp" {
			t.Fatalf("service %q proto = %q, want udp", p.Service, p.Proto)
		}
	}
	if !seen["ntp"] || !seen["snmp"] {
		t.Fatalf("missing services, got %v", seen)
	}
}

func TestSNMPGetRequestEncoding(t *testing.T) {
	req := snmpGetRequest("public", 0x01020304)
	if len(req) < 4 {
		t.Fatalf("request too short: %d", len(req))
	}
	if req[0] != 0x30 {
		t.Fatalf("first byte %#x, want SEQUENCE", req[0])
	}
	if req[2] != 0x02 || req[3] != 0x01 {
		t.Fatalf("expected version INTEGER, got % x", req[2:4])
	}
	// find the sysDescr OID
	if idx := indexBytes(req, sysDescrOID); idx < 0 {
		t.Fatal("sysDescr OID not present in request")
	}
}

func TestParseSNMPSysDescr(t *testing.T) {
	oid := []byte{0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00}
	msg := []byte{0x30, 0x2A, 0x02, 0x01, 0x00, 0x04, 0x06, 'p', 'u', 'b', 'l', 'i', 'c', 0xA2, 0x1C, 0x02, 0x04, 0, 0, 0, 1, 0x02, 0x01, 0x00, 0x02, 0x01, 0x00, 0x30, 0x0C, 0x06, 0x08}
	msg = append(msg, oid...)
	msg = append(msg, 0x04, 0x04, 't', 'e', 's', 't')
	if got := parseSNMPSysDescr(msg); got != "test" {
		t.Fatalf("got %q", got)
	}
	if got := parseSNMPSysDescr([]byte{0x30, 0x00}); got != "" {
		t.Fatalf("expected empty for message without OID, got %q", got)
	}
}

func TestDNSRequestEncoding(t *testing.T) {
	req := []byte{0, 0, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	binary.BigEndian.PutUint16(req[0:2], 0xBEEF)
	req = append(req, 0x00)
	req = append(req, 0x00, 0x02)
	req = append(req, 0x00, 0x01)
	if len(req) != 17 {
		t.Fatalf("got %d bytes, want 17", len(req))
	}
	if req[12] != 0x00 {
		t.Fatal("root label must be a single zero byte")
	}
}

func indexBytes(hay, needle []byte) int {
	for i := 0; i+len(needle) <= len(hay); i++ {
		match := true
		for j := range needle {
			if hay[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

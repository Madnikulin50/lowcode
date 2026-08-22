package agent

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// udpScanPorts are the UDP services probed on hosts already known to be live
// (found via TCP, ping or the ARP table). UDP is connection-less so each probe
// sends a minimal protocol request and treats any matching reply as "open".
var udpScanPorts = []int{53, 123, 161, 1900}

// probeUDPPorts probes all configured UDP ports of a host in parallel.
func probeUDPPorts(ctx context.Context, addr string, timeout time.Duration) []Port {
	var mu sync.Mutex
	var out []Port
	var wg sync.WaitGroup
	for _, p := range udpScanPorts {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			port := probeUDPPort(ctx, addr, p, timeout)
			if port == nil {
				return
			}
			mu.Lock()
			out = append(out, *port)
			mu.Unlock()
		}(p)
	}
	wg.Wait()
	return out
}

func probeUDPPort(ctx context.Context, addr string, port int, timeout time.Duration) *Port {
	switch port {
	case 53:
		open, _ := probeDNS(ctx, addr, 53, timeout)
		if !open {
			return nil
		}
		return &Port{Port: 53, Proto: "udp", Service: "dns"}
	case 123:
		open, version := probeNTP(ctx, addr, 123, timeout)
		if !open {
			return nil
		}
		return &Port{Port: 123, Proto: "udp", Service: "ntp", Version: version}
	case 161:
		open, descr := probeSNMP(ctx, addr, 161, timeout)
		if !open {
			return nil
		}
		return &Port{Port: 161, Proto: "udp", Service: "snmp", Version: descr}
	case 1900:
		open, server := probeSSDP(ctx, addr, 1900, timeout)
		if !open {
			return nil
		}
		return &Port{Port: 1900, Proto: "udp", Service: "ssdp", Banner: server}
	}
	return nil
}

// probeDNS sends a minimal DNS query (root NS) and treats any reply whose
// transaction ID matches as an open DNS port — even REFUSED/NXDOMAIN answers
// prove a resolver is listening.
func probeDNS(ctx context.Context, addr string, port int, timeout time.Duration) (bool, string) {
	conn, err := dialUDP(ctx, addr, port, timeout)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	tx := uint16(time.Now().UnixNano() & 0xFFFF)
	if tx == 0 {
		tx = 1
	}
	// Header: ID, flags RD, QDCOUNT=1 + question "." NS IN
	req := []byte{0, 0, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	binary.BigEndian.PutUint16(req[0:2], tx)
	req = append(req, 0x00)       // root label
	req = append(req, 0x00, 0x02) // QTYPE NS
	req = append(req, 0x00, 0x01) // QCLASS IN

	if _, err := conn.Write(req); err != nil {
		return false, ""
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil || n < 12 {
		return false, ""
	}
	if binary.BigEndian.Uint16(buf[0:2]) != tx {
		return false, ""
	}
	return true, ""
}

// probeNTP sends a classic 48-byte NTPv4 client request and derives the
// protocol version and stratum from the reply.
func probeNTP(ctx context.Context, addr string, port int, timeout time.Duration) (bool, string) {
	conn, err := dialUDP(ctx, addr, port, timeout)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	req := make([]byte, 48)
	req[0] = 0x23 // LI=0, VN=4, Mode=3 (client)
	sec := uint32(time.Now().Unix() + 2208988800)
	frac := uint32((time.Now().Nanosecond() * 4294967296) / 1e9)
	binary.BigEndian.PutUint32(req[40:44], sec)  // transmit timestamp seconds
	binary.BigEndian.PutUint32(req[44:48], frac) // transmit timestamp fraction

	if _, err := conn.Write(req); err != nil {
		return false, ""
	}
	buf := make([]byte, 48)
	n, err := conn.Read(buf)
	if err != nil || n < 48 {
		return false, ""
	}
	vn := int(buf[0]>>3) & 0x7
	if vn == 0 {
		vn = 4
	}
	stratum := int(buf[1])
	version := fmt.Sprintf("NTPv%d", vn)
	if stratum > 0 && stratum < 16 {
		version += fmt.Sprintf(" stratum %d", stratum)
	}
	return true, version
}

// probeSNMP sends an SNMPv1 GET of sysDescr (community "public") and, on any
// response, reports the port open with the parsed sysDescr string.
func probeSNMP(ctx context.Context, addr string, port int, timeout time.Duration) (bool, string) {
	conn, err := dialUDP(ctx, addr, port, timeout)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	reqID := uint32(time.Now().UnixNano() & 0xFFFFFFFF)
	if reqID == 0 {
		reqID = 1
	}
	if _, err := conn.Write(snmpGetRequest("public", reqID)); err != nil {
		return false, ""
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n < 2 {
		return false, ""
	}
	return true, parseSNMPSysDescr(buf[:n])
}

// probeSSDP unicast M-SEARCH: many UPnP devices answer unicast discovery.
func probeSSDP(ctx context.Context, addr string, port int, timeout time.Duration) (bool, string) {
	conn, err := dialUDP(ctx, addr, port, timeout)
	if err != nil {
		return false, ""
	}
	defer conn.Close()

	msg := "M-SEARCH * HTTP/1.1\r\nHOST: 239.255.255.250:1900\r\nMAN: \"ssdp:discover\"\r\nMX: 1\r\nST: ssdp:all\r\n\r\n"
	if _, err := conn.Write([]byte(msg)); err != nil {
		return false, ""
	}
	buf := make([]byte, 2048)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return false, ""
	}
	text := string(buf[:n])
	if !strings.Contains(text, "200 OK") && !strings.Contains(strings.ToLower(text), "notify") {
		return false, ""
	}
	server := ""
	for _, line := range strings.Split(text, "\r\n") {
		if strings.HasPrefix(strings.ToLower(line), "server:") {
			server = strings.TrimSpace(line[7:])
			break
		}
	}
	return true, server
}

func dialUDP(ctx context.Context, addr string, port int, timeout time.Duration) (*net.UDPConn, error) {
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "udp", net.JoinHostPort(addr, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, err
	}
	uc := conn.(*net.UDPConn)
	_ = uc.SetDeadline(time.Now().Add(timeout))
	return uc, nil
}

// sysDescrOID is 1.3.6.1.2.1.1.1.0 (first two sub-ids folded into 0x2B).
var sysDescrOID = []byte{0x2B, 0x06, 0x01, 0x02, 0x01, 0x01, 0x01, 0x00}

// snmpGetRequest builds an SNMPv1 GET-PDU for sysDescr with the given community.
func snmpGetRequest(community string, reqID uint32) []byte {
	comm := []byte(community)

	vbBody := []byte{0x06, byte(len(sysDescrOID))}
	vbBody = append(vbBody, sysDescrOID...)
	vbBody = append(vbBody, 0x05, 0x00) // value NULL
	vb := append([]byte{0x30, byte(len(vbBody))}, vbBody...)

	vbl := append([]byte{0x30, byte(len(vb))}, vb...)

	pduBody := []byte{0x02, 0x04, 0, 0, 0, 0}
	binary.BigEndian.PutUint32(pduBody[2:6], reqID)
	pduBody = append(pduBody, 0x02, 0x01, 0x00) // error-status 0
	pduBody = append(pduBody, 0x02, 0x01, 0x00) // error-index 0
	pduBody = append(pduBody, vbl...)
	pdu := append([]byte{0xA0, byte(len(pduBody))}, pduBody...)

	msgBody := []byte{0x02, 0x01, 0x00} // version v1
	msgBody = append(msgBody, 0x04, byte(len(comm)))
	msgBody = append(msgBody, comm...)
	msgBody = append(msgBody, pdu...)
	return append([]byte{0x30, byte(len(msgBody))}, msgBody...)
}

// parseSNMPSysDescr extracts the sysDescr value (OCTET STRING after the
// sysDescr OID) from an SNMP GetResponse message. Empty if not present.
func parseSNMPSysDescr(b []byte) string {
	idx := bytes.Index(b, sysDescrOID)
	if idx < 0 {
		return ""
	}
	rest := b[idx+len(sysDescrOID):]
	if len(rest) < 2 || rest[0] != 0x04 {
		return ""
	}
	l := int(rest[1])
	if l <= 0 || l > len(rest)-2 || l > 1024 {
		return ""
	}
	return strings.TrimSpace(string(rest[2 : 2+l]))
}

package agent

import (
	"context"
	"encoding/binary"
	"log"
	"net"
	"sort"
	"strings"
	"time"
)

// mdnsInfo aggregates the DNS-SD signals a host announced via mDNS
// (Apple AirPlay, Google Cast, Miracast, HiSuite, ...).
type mdnsInfo struct {
	Services map[string]bool
	TXT      map[string]string
	Hostname string
	DeviceID string
	Model    string
}

func (m mdnsInfo) serviceList() []string {
	var out []string
	for s := range m.Services {
		if mdnsPhoneServiceTypes[s] || s == "mdns" {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

func (m mdnsInfo) empty() bool {
	return len(m.Services) == 0 && m.Hostname == "" && m.Model == "" && m.DeviceID == ""
}

// mdnsPhoneServiceTypes are DNS-SD service types that hint at a mobile device
// (or a device with phone/tablet-like mDNS behaviour). They are stored into
// Device.Services and feed the classifier.
var mdnsPhoneServiceTypes = map[string]bool{
	"airplay": true, "raop": true, "companion-link": true, "sleep-proxy": true,
	"touch-remote": true, "mediaremotetv": true, "appletv-v2": true, "airtunes": true,
	"googlecast": true, "miracast": true, "hisuite": true, "mipcs": true,
	"androidtvremote": true, "androidtvremote2": true, "smartsync": true,
	"ncast": true, "samsung": true, "samsung-smartthings": true,
	"apple-mobdev2": true, "dacp": true, "music": true, "companion": true,
	"mdns": true,
}

// mdnsBrowseQueries are sent as PTR queries on every mDNS browse/probe.
var mdnsBrowseQueries = []string{
	"_services._dns-sd._udp.local",
	"_airplay._tcp.local",
	"_raop._tcp.local",
	"_companion-link._tcp.local",
	"_sleep-proxy._udp.local",
	"_googlecast._tcp.local",
	"_miracast._tcp.local",
	"_hisuite._tcp.local",
	"_mipcs._tcp.local",
	"_androidtvremote2._tcp.local",
	"_smartsync._tcp.local",
	"_ncast._tcp.local",
	"_touch-remote._tcp.local",
	"_mediaremotetv._tcp.local",
}

const mdnsGroupAddr = "224.0.0.251:5353"

// browseMDNS sends the phone-relevant PTR queries to the mDNS multicast group
// and collects the responses from all hosts on the local link.
func browseMDNS(ctx context.Context, wait time.Duration) map[string]mdnsInfo {
	if wait <= 0 {
		wait = 2 * time.Second
	}
	group := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
	conn, err := net.ListenMulticastUDP("udp4", nil, group)
	if err != nil {
		log.Printf("mdns browse: cannot listen multicast: %v", err)
		return nil
	}
	defer conn.Close()
	_ = conn.SetReadBuffer(1 << 20)

	out := map[string]mdnsInfo{}
	if err := mdnsExchange(ctx, conn, group, wait, func(src *net.UDPAddr, pkt []byte) {
		info, ok := parseMDNSResponse(pkt)
		if !ok {
			return
		}
		mergeMDNS(out, src.IP.String(), info)
	}); err != nil {
		return out
	}
	return out
}

// probeMDNSHost sends the same PTR queries unicast to a single host and
// collects its answers (used for hosts that did not answer the multicast
// browse — mDNS responders answer unicast queries with the QU bit set).
func probeMDNSHost(ctx context.Context, addr string, wait time.Duration) mdnsInfo {
	if wait <= 0 {
		wait = 800 * time.Millisecond
	}
	raddr := &net.UDPAddr{IP: net.ParseIP(addr), Port: 5353}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return mdnsInfo{}
	}
	defer conn.Close()

	var out mdnsInfo
	_ = mdnsExchange(ctx, conn, raddr, wait, func(src *net.UDPAddr, pkt []byte) {
		info, ok := parseMDNSResponse(pkt)
		if !ok {
			return
		}
		tmp := map[string]mdnsInfo{}
		mergeMDNS(tmp, src.IP.String(), out)
		mergeMDNS(tmp, src.IP.String(), info)
		out = tmp[src.IP.String()]
	})
	return out
}

// mdnsExchange sends all browse queries to dst and reads responses until
// the deadline; handler is called for every received datagram.
func mdnsExchange(ctx context.Context, conn *net.UDPConn, dst net.Addr, wait time.Duration, handler func(*net.UDPAddr, []byte)) error {
	deadline := time.Now().Add(wait)
	_ = conn.SetDeadline(deadline)

	for _, q := range mdnsBrowseQueries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		pkt := mdnsQueryPacket(q, true)
		if _, err := conn.WriteTo(pkt, dst); err != nil {
			return err
		}
	}

	buf := make([]byte, 65535)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			return nil
		}
		handler(src, buf[:n])
	}
}

// mdnsQueryPacket builds a standard DNS query for a PTR record with the
// unicast-response bit (QU) set when requested.
func mdnsQueryPacket(qname string, unicastResp bool) []byte {
	var pkt []byte
	pkt = append(pkt, 0x00, 0x00) // transaction ID
	pkt = append(pkt, 0x00, 0x00) // flags: standard query
	pkt = append(pkt, 0x00, 0x01) // QDCOUNT = 1
	pkt = append(pkt, 0x00, 0x00) // ANCOUNT
	pkt = append(pkt, 0x00, 0x00) // NSCOUNT
	pkt = append(pkt, 0x00, 0x00) // ARCOUNT
	pkt = appendDNSName(pkt, qname)
	pkt = append(pkt, 0x00, 0x0C) // QTYPE: PTR
	cls := uint16(0x0001)         // class IN
	if unicastResp {
		cls |= 0x8000 // QU: unicast response
	}
	pkt = append(pkt, byte(cls>>8), byte(cls))
	return pkt
}

func appendDNSName(pkt []byte, name string) []byte {
	for _, label := range strings.Split(name, ".") {
		if label == "" {
			continue
		}
		l := len(label)
		if l > 63 {
			l = 63
		}
		pkt = append(pkt, byte(l))
		pkt = append(pkt, label[:l]...)
	}
	return append(pkt, 0x00)
}

// parseMDNSResponse parses a DNS wire response and extracts the mDNS signals:
// announced service types (PTR), TXT key/values (AirPlay model, deviceid),
// and the hostname (from A records).
func parseMDNSResponse(pkt []byte) (mdnsInfo, bool) {
	var info mdnsInfo
	if len(pkt) < 12 {
		return info, false
	}
	// Must be a response (QR bit)
	if binary.BigEndian.Uint16(pkt[2:4])&0x8000 == 0 {
		return info, false
	}
	qd := int(binary.BigEndian.Uint16(pkt[4:6]))
	an := int(binary.BigEndian.Uint16(pkt[6:8]))
	ns := int(binary.BigEndian.Uint16(pkt[8:10]))
	ar := int(binary.BigEndian.Uint16(pkt[10:12]))

	off := 12
	for i := 0; i < qd; i++ {
		name, n, ok := readDNSName(pkt, off)
		if !ok {
			return info, false
		}
		_ = name
		off = n + 4 // skip QTYPE + QCLASS
	}
	if off > len(pkt) {
		return info, false
	}

	for _, count := range []int{an, ns, ar} {
		for i := 0; i < count; i++ {
			name, n, ok := readDNSName(pkt, off)
			if !ok {
				return info, false
			}
			off = n
			if off+10 > len(pkt) {
				return info, false
			}
			rtype := int(binary.BigEndian.Uint16(pkt[off : off+2]))
			rdlen := int(binary.BigEndian.Uint16(pkt[off+8 : off+10]))
			off += 10
			if off+rdlen > len(pkt) {
				return info, false
			}
			rdata := pkt[off : off+rdlen]
			off += rdlen

			switch rtype {
			case 12: // PTR: instance/service name
				svc, _ := serviceTypeFromName(name)
				if s, _, ok := readDNSName(pkt, off-rdlen); ok {
					if s2, ok2 := serviceTypeFromName(s); ok2 {
						svc = s2
					}
				}
				if svc != "" {
					if info.Services == nil {
						info.Services = map[string]bool{}
					}
					info.Services[svc] = true
				}
			case 16: // TXT: length-prefixed key=value strings
				for _, kv := range parseTXT(rdata) {
					if info.TXT == nil {
						info.TXT = map[string]string{}
					}
					info.TXT[kv[0]] = kv[1]
					if kv[0] == "model" && info.Model == "" {
						info.Model = kv[1]
					}
					if kv[0] == "deviceid" && info.DeviceID == "" {
						info.DeviceID = kv[1]
					}
				}
			case 1: // A: name -> IPv4 (hostname announcement)
				if info.Hostname == "" && rdlen == 4 {
					info.Hostname = stripMDNSName(name)
				}
			case 33: // SRV: target hostname
				if rdlen >= 6 {
					if s, _, ok := readDNSName(pkt, off-rdlen+6); ok {
						if info.Hostname == "" {
							info.Hostname = stripMDNSName(s)
						}
					}
				}
			}
		}
	}
	return info, true
}

// serviceTypeFromName extracts the DNS-SD service type from a name like
// "iPhone._airplay._tcp.local" or "_airplay._tcp.local".
func serviceTypeFromName(name string) (string, bool) {
	name = strings.ToLower(strings.TrimSuffix(name, "."))
	parts := strings.Split(name, ".")
	for _, p := range parts {
		if strings.HasPrefix(p, "_") {
			t := strings.TrimPrefix(p, "_")
			if t == "tcp" || t == "udp" || t == "local" || t == "dns-sd" || t == "services" {
				continue
			}
			return t, true
		}
	}
	return "", false
}

// stripMDNSName converts "iPhone-9B1C.local" into "iPhone-9B1C".
func stripMDNSName(name string) string {
	name = strings.TrimSuffix(name, ".")
	name = strings.TrimSuffix(name, ".local")
	return name
}

func parseTXT(rdata []byte) [][2]string {
	var out [][2]string
	for i := 0; i < len(rdata); {
		l := int(rdata[i])
		i++
		if l == 0 || i+l > len(rdata) {
			break
		}
		s := string(rdata[i : i+l])
		i += l
		kv := strings.SplitN(s, "=", 2)
		k := strings.ToLower(strings.TrimSpace(kv[0]))
		v := ""
		if len(kv) > 1 {
			v = kv[1]
		}
		if k != "" {
			out = append(out, [2]string{k, v})
		}
	}
	return out
}

// readDNSName decodes a DNS name (with compression pointers) at off.
func readDNSName(pkt []byte, off int) (string, int, bool) {
	var labels []string
	pos := off
	end := -1
	hops := 0
	for {
		if pos >= len(pkt) {
			return "", 0, false
		}
		l := int(pkt[pos])
		switch {
		case l == 0:
			if end < 0 {
				end = pos + 1
			}
			return strings.Join(labels, "."), end, true
		case l&0xC0 == 0xC0:
			if pos+1 >= len(pkt) {
				return "", 0, false
			}
			ptr := int(binary.BigEndian.Uint16(pkt[pos:pos+2]) & 0x3FFF)
			if end < 0 {
				end = pos + 2
			}
			sub, _, ok := readDNSName(pkt, ptr)
			if !ok {
				return "", 0, false
			}
			labels = append(labels, sub)
			return strings.Join(labels, "."), end, true
		default:
			if l > 63 || pos+1+l > len(pkt) {
				return "", 0, false
			}
			labels = append(labels, string(pkt[pos+1:pos+1+l]))
			pos += 1 + l
		}
		hops++
		if hops > 32 {
			return "", 0, false
		}
	}
}

func mergeMDNS(dst map[string]mdnsInfo, ip string, info mdnsInfo) {
	cur, ok := dst[ip]
	if !ok {
		dst[ip] = info
		return
	}
	if cur.Services == nil {
		cur.Services = map[string]bool{}
	}
	for s := range info.Services {
		cur.Services[s] = true
	}
	if cur.TXT == nil {
		cur.TXT = map[string]string{}
	}
	for k, v := range info.TXT {
		if cur.TXT[k] == "" || v != "" {
			cur.TXT[k] = v
		}
	}
	if cur.Hostname == "" {
		cur.Hostname = info.Hostname
	}
	if cur.DeviceID == "" {
		cur.DeviceID = info.DeviceID
	}
	if cur.Model == "" {
		cur.Model = info.Model
	}
	// Derive the convenience fields from the merged TXT map (a responder may
	// announce model/deviceid in a later packet than the one that set them).
	if cur.Model == "" {
		cur.Model = cur.TXT["model"]
	}
	if cur.DeviceID == "" {
		cur.DeviceID = cur.TXT["deviceid"]
	}
	dst[ip] = cur
}

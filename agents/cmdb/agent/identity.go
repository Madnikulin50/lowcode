package agent

import (
	"errors"
	"net"
	"strings"
)

// errDeviceNotFound is returned by Storage.Find* when no matching row exists.
// Other errors (API, DB) must not be treated as a miss — that always INSERTs duplicates.
var errDeviceNotFound = errors.New("device not found")
var errUnauthorizedAPI = errors.New("unauthorized")

func isDeviceNotFound(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, errDeviceNotFound) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "device not found")
}

func isUnauthorizedAPI(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, errUnauthorizedAPI) {
		return true
	}
	low := strings.ToLower(err.Error())
	return strings.Contains(low, "unauthorized") || strings.Contains(low, "api error 401") || strings.Contains(low, "api error 403")
}

// genericHostnames are PTR/default names shared by many distinct devices (gateways).
// Matching on them would merge unrelated hosts across subnets.
func genericHostname(h string) bool {
	s := strings.ToLower(strings.TrimSpace(h))
	if s == "" {
		return true
	}
	switch s {
	case "_gateway", "gateway", "localhost", "unknown", "none", "null",
		"router", "dsldevice", "dsl-gateway", "home.gateway":
		return true
	}
	if net.ParseIP(s) != nil {
		return true
	}
	return false
}

func normalizeMAC(mac string) string {
	s := strings.ToLower(strings.TrimSpace(mac))
	if s == "" {
		return ""
	}
	s = strings.ReplaceAll(s, "-", ":")
	s = strings.ReplaceAll(s, ".", ":")
	parts := strings.Split(s, ":")
	if len(parts) != 6 {
		return s
	}
	for i, p := range parts {
		if len(p) == 1 {
			parts[i] = "0" + p
		}
	}
	return strings.Join(parts, ":")
}

func normalizeIP(ip string) string {
	s := strings.TrimSpace(ip)
	if s == "" {
		return ""
	}
	if parsed := net.ParseIP(s); parsed != nil {
		if v4 := parsed.To4(); v4 != nil {
			return v4.String()
		}
		return parsed.String()
	}
	return s
}

func stableHostname(h string) string {
	s := strings.TrimSpace(h)
	if genericHostname(s) {
		return ""
	}
	return strings.ToLower(s)
}

func qlLiteral(v string) string {
	return "'" + strings.ReplaceAll(v, "'", `\'`) + "'"
}

// DedupeDevices collapses one scan batch onto a stable identity so persistAll
// does not INSERT the same host twice (two NIC IPs, overlapping CIDRs, or the
// classify+vuln second persist).
//
// Match order: MAC, then non-generic hostname, then IP.
func DedupeDevices(in []Device) []Device {
	if len(in) < 2 {
		return in
	}
	index := make(map[string]int, len(in))
	out := make([]Device, 0, len(in))
	for _, d := range in {
		key := deviceIdentityKey(d)
		if key == "" {
			out = append(out, d)
			continue
		}
		if i, ok := index[key]; ok {
			out[i] = mergeDevice(out[i], d)
			continue
		}
		index[key] = len(out)
		out = append(out, d)
	}
	return out
}

func deviceIdentityKey(d Device) string {
	if mac := normalizeMAC(d.MAC); mac != "" {
		return "mac:" + mac
	}
	if host := stableHostname(d.Hostname); host != "" {
		return "host:" + host
	}
	if ip := normalizeIP(d.IP); ip != "" {
		return "ip:" + ip
	}
	return ""
}

func mergeDevice(a, b Device) Device {
	out := a
	if normalizeMAC(out.MAC) == "" {
		out.MAC = b.MAC
	}
	if strings.TrimSpace(out.Hostname) == "" || genericHostname(out.Hostname) {
		if strings.TrimSpace(b.Hostname) != "" {
			out.Hostname = b.Hostname
		}
	}
	if strings.TrimSpace(out.Vendor) == "" {
		out.Vendor = b.Vendor
	}
	if strings.TrimSpace(out.Model) == "" {
		out.Model = b.Model
	}
	if strings.TrimSpace(out.OS) == "" {
		out.OS = b.OS
	}
	if strings.TrimSpace(out.Domain) == "" {
		out.Domain = b.Domain
	}
	if out.DeviceType == "" || out.DeviceType == "unknown" {
		if b.DeviceType != "" {
			out.DeviceType = b.DeviceType
		}
	}
	if b.LastSeen != "" {
		out.LastSeen = b.LastSeen
	}
	if b.Status != "" {
		out.Status = b.Status
	}
	if len(b.OpenPorts) > len(out.OpenPorts) {
		out.OpenPorts = b.OpenPorts
	}
	if len(b.Services) > len(out.Services) {
		out.Services = b.Services
	}
	if len(b.Shares) > len(out.Shares) {
		out.Shares = b.Shares
	}
	if len(b.Vulnerabilities) > 0 {
		out.Vulnerabilities = b.Vulnerabilities
	}
	// Prefer the IP from the richer (or later) observation so DHCP/subnet
	// moves update the existing row instead of keeping a stale address.
	if b.IP != "" && (len(b.OpenPorts) >= len(a.OpenPorts) || a.IP == "") {
		out.IP = b.IP
	}
	return out
}

package scancidr

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func HostsFromCIDR(ipnet *net.IPNet) []string {
	if ipnet == nil {
		return nil
	}
	ip4 := ipnet.IP.To4()
	if ip4 == nil {
		return nil
	}
	mask := ipnet.Mask
	if len(mask) == net.IPv6len {
		mask = mask[12:]
	}
	network := ip4.Mask(mask)
	n := &net.IPNet{IP: network, Mask: mask}
	ip := make(net.IP, len(network))
	copy(ip, network)
	ones, bits := mask.Size()
	skipNetworkBroadcast := bits == 32 && bits-ones >= 2
	var ips []string
	for n.Contains(ip) {
		last := ip[len(ip)-1]
		if !ip.IsUnspecified() && (!skipNetworkBroadcast || (last != 0 && last != 255)) {
			ips = append(ips, ip.String())
		}
		incIP(ip)
	}
	return ips
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func ResolveScanCIDRs(requested string) []string {
	requested = strings.TrimSpace(requested)
	if n := NormalizeCIDR(requested); n != "" {
		requested = n
	}
	locals := LocalIPv4CIDRs()
	out := PickScanCIDRs(requested, locals)
	log.Printf("scan: requested %q resolved to %v (local physical RFC1918: %v)", requested, out, locals)
	return out
}

func PickScanCIDRs(requested string, locals []string) []string {
	requested = strings.TrimSpace(requested)
	if requested == "" || strings.EqualFold(requested, "auto") {
		if len(locals) == 0 {
			return []string{"192.168.0.0/24"}
		}
		return locals
	}
	requested = CapCIDRToSlash24(requested)
	if cidrOverlapsAny(requested, locals) {
		return []string{requested}
	}
	if len(locals) > 0 {
		log.Printf("scan: %s is not on a local interface, using %v", requested, locals)
		return locals
	}
	return []string{requested}
}

func NormalizeCIDR(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "auto") {
		return s
	}
	if !strings.Contains(s, "/") {
		if net.ParseIP(s) != nil {
			return s + "/32"
		}
		return s
	}
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return s
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return s
	}
	ones, _ := ipnet.Mask.Size()
	network := ip4.Mask(ipnet.Mask)
	if ones >= 24 {
		return fmt.Sprintf("%s/%d", network.String(), ones)
	}
	return CapCIDRToSlash24(fmt.Sprintf("%s/%d", network.String(), ones))
}

func CapCIDRToSlash24(cidr string) string {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return cidr
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return cidr
	}
	ones, bits := ipnet.Mask.Size()
	if bits != 32 || ones >= 24 {
		return cidr
	}
	mask := net.CIDRMask(24, 32)
	network := ip4.Mask(mask)
	capped := fmt.Sprintf("%s/24", network.String())
	log.Printf("scan: capping %s to %s", cidr, capped)
	return capped
}

var virtualIfacePrefixes = []string{
	"docker", "br-", "veth", "virbr", "cni", "flannel", "tun", "tap",
	"wg", "lxc", "vboxnet", "vmnet", "kube-", "cali", "nodelocal",
}

func SkipVirtualIface(name string, flags net.Flags) bool {
	if flags&net.FlagLoopback != 0 {
		return true
	}
	if flags&net.FlagUp == 0 {
		return true
	}
	n := strings.ToLower(name)
	for _, p := range virtualIfacePrefixes {
		if strings.HasPrefix(n, p) {
			return true
		}
	}
	return false
}

func LocalIPv4CIDRs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, iface := range ifaces {
		if SkipVirtualIface(iface.Name, iface.Flags) {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil || ip4.IsLoopback() || !IsRFC1918(ip4) {
				continue
			}
			mask := ipnet.Mask
			if len(mask) == net.IPv6len {
				mask = mask[12:]
			}
			ones, bits := mask.Size()
			if bits != 32 || ones < 24 || ones > 30 {
				ones = 24
				mask = net.CIDRMask(24, 32)
			}
			network := ip4.Mask(mask)
			cidr := fmt.Sprintf("%s/%d", network.String(), ones)
			if seen[cidr] {
				continue
			}
			seen[cidr] = true
			out = append(out, cidr)
		}
	}
	return out
}

func IsRFC1918(ip net.IP) bool {
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}
	return ip4[0] == 10 || (ip4[0] == 192 && ip4[1] == 168) || (ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31)
}

func localIPv4Addresses() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipnet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipnet.IP.To4()
			if ip4 == nil || ip4.IsLoopback() {
				continue
			}
			out = append(out, ip4.String())
		}
	}
	return out
}

func cidrOverlapsAny(cidr string, locals []string) bool {
	_, want, err := net.ParseCIDR(cidr)
	if err != nil {
		ip := net.ParseIP(cidr)
		if ip == nil {
			return false
		}
		for _, l := range locals {
			_, n, err := net.ParseCIDR(l)
			if err == nil && n.Contains(ip) {
				return true
			}
		}
		return false
	}
	for _, l := range locals {
		_, n, err := net.ParseCIDR(l)
		if err != nil {
			continue
		}
		if n.Contains(want.IP) || want.Contains(n.IP) {
			return true
		}
	}
	return false
}

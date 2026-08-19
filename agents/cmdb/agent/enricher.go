package agent

import (
	"context"
	"net"
	"strings"
)

type Enricher struct{}

func NewEnricher() *Enricher {
	return &Enricher{}
}

func (e *Enricher) Enrich(ctx context.Context, devices []Device) []Device {
	for i := range devices {
		d := &devices[i]
		if d.Hostname == "" && d.IP != "" {
			if names, err := net.LookupAddr(d.IP); err == nil && len(names) > 0 {
				d.Hostname = strings.TrimSuffix(names[0], ".")
			}
		}
		if d.MAC != "" && d.Vendor == "" {
			d.Vendor = lookupOUI(d.MAC)
		}
	}
	return devices
}

var ouiDB = map[string]string{
	"00:50:56": "VMware",
	"00:0C:29": "VMware",
	"00:05:69": "VMware",
	"00:1C:42": "VMware",
	"08:00:27": "Oracle VirtualBox",
	"00:15:5D": "Microsoft Hyper-V",
	"52:54:00": "QEMU/KVM",
	"8C:85:90": "Cisco",
	"00:1B:0C": "Cisco",
	"00:21:D8": "Cisco",
	"00:26:0B": "Cisco",
	"00:14:5E": "Cisco",
	"28:6E:D4": "Huawei",
	"00:1A:A1": "Huawei",
	"00:05:86": "Juniper",
	"00:12:1E": "Juniper",
	"00:19:E2": "Juniper",
	"00:23:9C": "Juniper",
	"00:26:88": "Juniper",
	"00:0F:EA": "Dell",
	"00:14:22": "Dell",
	"00:1E:4F": "Dell",
	"00:21:9B": "Dell",
	"00:24:E8": "Dell",
	"00:04:AC": "HP",
	"00:17:A4": "HP",
	"00:1A:4B": "HP",
	"00:21:5A": "HP",
	"00:23:7D": "HP",
	"00:0B:82": "Apple",
	"00:1B:63": "Apple",
	"00:1E:C2": "Apple",
	"00:23:32": "Apple",
	"00:25:00": "Apple",
	"00:26:08": "Apple",
	"00:26:B0": "Apple",
	"00:50:E4": "Apple",
	"04:0C:CE": "Apple",
	"04:26:65": "Apple",
	"04:54:53": "Apple",
	"08:66:98": "Apple",
	"0C:30:21": "Apple",
	"0C:74:C2": "Apple",
	"10:40:F3": "Apple",
	"00:0A:95": "Apple",
	"00:0D:93": "Apple",
	"00:1F:45": "Synology",
	"00:11:32": "Synology",
	"00:0C:6E": "Raspberry Pi",
	"B8:27:EB": "Raspberry Pi",
	"DC:A6:32": "Raspberry Pi",
	"E4:5F:01": "Raspberry Pi",
	"00:0E:C6": "TP-Link",
	"00:1A:3F": "TP-Link",
	"00:27:19": "TP-Link",
	"50:C7:BF": "TP-Link",
	"54:E6:FC": "TP-Link",
	"64:70:02": "TP-Link",
	"C0:4A:00": "TP-Link",
	"EC:08:6B": "TP-Link",
	"00:1D:0F": "Xiaomi",
	"04:CF:8C": "Xiaomi",
	"18:FE:34": "Xiaomi",
	"48:8C:5A": "Xiaomi",
	"9C:F4:8E": "Xiaomi",
	"00:0C:43": "Roku",
	"00:0D:4F": "Roku",
	"00:22:A2": "Roku",
	"00:17:31": "Netgear",
	"00:1C:12": "Netgear",
	"00:22:3F": "Netgear",
	"2C:B0:5D": "Netgear",
	"6C:B0:CE": "Netgear",
	"84:1B:5E": "Netgear",
	"B0:39:56": "Netgear",
	"00:0F:B5": "ASUS",
	"00:1A:92": "ASUS",
	"00:1B:FC": "ASUS",
	"10:BF:48": "ASUS",
	"1C:B7:2C": "ASUS",
	"00:0F:66": "Intel",
	"00:1B:21": "Intel",
	"00:1E:67": "Intel",
	"00:21:6A": "Intel",
	"00:24:D6": "Intel",
	"00:1B:77": "Linksys",
	"00:0C:41": "Linksys",
	"00:14:BF": "Linksys",
	"00:1A:70": "Linksys",
	"E0:46:9A": "Hikvision",
	"00:0B:DB": "D-Link",
	"00:1C:F0": "D-Link",
	"28:10:7B": "D-Link",
	"38:63:BB": "D-Link",
	"5C:D9:98": "D-Link",
	"B0:C5:54": "D-Link",
	"00:08:9B": "Panasonic",
	"00:1B:9E": "Panasonic",
	"00:04:F0": "Sony",
	"00:12:7B": "Sony",
	"00:1A:80": "Sony",
	"00:21:3C": "Sony",
	"00:24:BE": "Sony",
	"00:0E:58": "ZTE",
	"00:1D:C9": "ZTE",
	"2C:0E:3F": "Ubiquiti",
	"00:27:22": "Ubiquiti",
	"44:D9:E7": "Ubiquiti",
	"74:83:C2": "Ubiquiti",
	"78:8A:20": "Ubiquiti",
	"CC:2D:E0": "Ubiquiti",
	"00:11:50": "MikroTik",
	"4C:5E:0C": "MikroTik",
	"64:D1:54": "MikroTik",
	"DC:2C:6E": "MikroTik",
	"E4:8D:8C": "MikroTik",
}

func lookupOUI(mac string) string {
	mac = strings.ToUpper(strings.ReplaceAll(mac, "-", ":"))
	for prefix, vendor := range ouiDB {
		if strings.HasPrefix(mac, prefix) {
			return vendor
		}
	}
	return ""
}

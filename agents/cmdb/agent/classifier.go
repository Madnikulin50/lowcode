package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Classifier struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewClassifier(baseURL, model string) *Classifier {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "deepseek-v2"
	}
	return &Classifier{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (cl *Classifier) Classify(ctx context.Context, d Device) Device {
	// 1. Strong deterministic rules — always win
	if t := strongDeviceType(d); t != "" {
		d.DeviceType = t
		return d
	}

	// 2. Heuristic fallback
	if t := guessDeviceType(d); t != "" && t != "unknown" {
		d.DeviceType = t
		return d
	}

	// 3. LLM for ambiguous cases
	if t := cl.classifyWithLLM(ctx, d); t != "" {
		d.DeviceType = t
		return d
	}

	return d
}

// strongDeviceType returns a definitive type based on unambiguous signals.
// Empty string means "not sure, use heuristic/LLM".
func strongDeviceType(d Device) string {
	portSet := make(map[int]bool)
	svcSet := make(map[string]bool)
	for _, p := range d.OpenPorts {
		portSet[p.Port] = true
		if p.Service != "" {
			svcSet[strings.ToLower(p.Service)] = true
		}
	}
	// mDNS services (airplay, hisuite, ...) are announced via DNS-SD and
	// stored in Device.Services, not as port banners.
	for _, s := range d.Services {
		if s != "" {
			svcSet[strings.ToLower(s)] = true
		}
	}
	vendor := strings.ToLower(d.Vendor)
	host := strings.ToLower(d.Hostname)
	os := strings.ToLower(d.OS)
	model := strings.ToLower(d.Model)

	// Domain controllers are unambiguous
	if portSet[88] && (portSet[389] || portSet[636] || portSet[3268]) {
		return "domain-controller"
	}

	// Database servers
	if portSet[3306] || portSet[5432] || portSet[6379] || portSet[27017] || svcSet["mysql"] || svcSet["postgresql"] || svcSet["redis"] || svcSet["mongod"] {
		return "server"
	}

	// Printer by ports / vendor
	if portSet[515] || portSet[9100] || portSet[631] {
		return "printer"
	}
	if containsAny(vendor, "hp", "canon", "epson", "brother", "kyocera", "xerox") && (portSet[80] || portSet[9100] || portSet[515]) {
		return "printer"
	}

	// Camera by RTSP / vendor
	if portSet[554] || portSet[8554] || svcSet["rtsp"] {
		return "camera"
	}
	if strings.Contains(vendor, "hikvision") || strings.Contains(vendor, "dahua") {
		return "camera"
	}

	// Mobile devices (phones / tablets)
	// Exact model from AirPlay mDNS TXT is the strongest signal.
	if strings.Contains(model, "iphone") || strings.Contains(model, "ipod") {
		return "phone"
	}
	if strings.Contains(model, "ipad") {
		return "tablet"
	}
	// Android Debug Bridge (Wi-Fi debugging) — an Android device.
	if portSet[5555] {
		if containsAny(host, "sm-t", "tab", "ipad") {
			return "tablet"
		}
		return "phone"
	}
	// Huawei / Xiaomi PC-suite announcements.
	if svcSet["hisuite"] || svcSet["mipcs"] {
		return "phone"
	}
	// Apple devices announcing AirPlay/RAOP/Companion/Sleep-proxy.
	if svcSet["airplay"] || svcSet["raop"] || svcSet["companion-link"] || svcSet["sleep-proxy"] {
		if containsAny(model, "ipad") || containsAny(host, "ipad") {
			return "tablet"
		}
		if containsAny(model, "iphone", "ipod") || containsAny(host, "iphone", "ipod") {
			return "phone"
		}
		if containsAny(model, "appletv") || containsAny(host, "appletv") {
			return "iot"
		}
		// Apple device without model detail (could be a Mac) — fall through
	}

	// Network devices (switch/router/firewall) by vendor
	if strings.Contains(vendor, "mikrotik") {
		return "router"
	}
	if strings.Contains(vendor, "ubiquiti") || strings.Contains(vendor, "ubiquiti networks") {
		return "router"
	}
	if strings.Contains(vendor, "cisco") || strings.Contains(vendor, "juniper") || strings.Contains(vendor, "huawei") {
		// Vendor alone is not enough — Huawei/Cisco also make phones, APs, cameras.
		if portSet[23] || portSet[22] || portSet[161] {
			return "switch"
		}
	}

	// Virtualization hosts
	if containsAny(vendor, "vmware", "virtualbox", "hyper-v", "qemu", "xen") {
		return "server"
	}

	// Hostname hints
	if host != "" {
		switch {
		case strings.HasPrefix(host, "ipad") || strings.HasPrefix(host, "sm-t"):
			return "tablet"
		case strings.HasPrefix(host, "iphone") || strings.HasPrefix(host, "ipod") || strings.HasPrefix(host, "android-") || strings.HasPrefix(host, "android_"):
			return "phone"
		case strings.HasPrefix(host, "sm-g") || strings.HasPrefix(host, "sm-a") || strings.HasPrefix(host, "sm-n") || strings.HasPrefix(host, "sm-s"):
			return "phone"
		case containsAny(host, "redmi", "poco", "pixel", "galaxy", "huawei", "honor", "moto", "lm-g", "lm-v", "nokia"):
			return "phone"
		case strings.HasPrefix(host, "rm-"):
			return "phone"
		case strings.Contains(host, "printer") || strings.Contains(host, "print-") || strings.Contains(host, "hp-"):
			return "printer"
		case strings.Contains(host, "camera") || strings.Contains(host, "cam-") || strings.Contains(host, "ipcam"):
			return "camera"
		case strings.Contains(host, "switch") || strings.Contains(host, "core-") || strings.Contains(host, "access-"):
			return "switch"
		case strings.Contains(host, "router") || strings.Contains(host, "gw") || strings.Contains(host, "gateway"):
			return "router"
		case strings.Contains(host, "firewall") || strings.Contains(host, "fw"):
			return "firewall"
		case strings.Contains(host, "server") || strings.Contains(host, "srv") || strings.Contains(host, "dc-"):
			return "server"
		case strings.Contains(host, "nas"):
			return "nas"
		}
	}

	// OS hints
	if strings.Contains(os, "windows") && portSet[3389] && !portSet[445] {
		return "workstation"
	}

	return ""
}

func guessDeviceType(d Device) string {
	portSet := make(map[int]bool)
	for _, p := range d.OpenPorts {
		portSet[p.Port] = true
	}
	vendor := strings.ToLower(d.Vendor)
	host := strings.ToLower(d.Hostname)

	// Printer ports
	if portSet[515] || portSet[9100] || portSet[631] {
		return "printer"
	}
	// Camera ports
	if portSet[554] || portSet[8554] {
		return "camera"
	}
	// Network gear (telnet + SNMP)
	if portSet[23] && (portSet[161] || portSet[162]) {
		return "switch"
	}
	if portSet[22] && portSet[161] {
		return "switch"
	}
	// Firewall
	if portSet[443] && portSet[23] && !portSet[80] {
		return "firewall"
	}
	// Domain controller
	if portSet[88] && (portSet[389] || portSet[636]) {
		return "domain-controller"
	}
	// Workstation
	if portSet[3389] {
		return "workstation"
	}
	// Server: SSH + web, or SSH + SMB
	if portSet[22] && (portSet[80] || portSet[443] || portSet[8080]) {
		return "server"
	}
	if portSet[22] && (portSet[445] || portSet[3306] || portSet[5432] || portSet[6379]) {
		return "server"
	}
	// NAS
	if portSet[80] && (portSet[445] || portSet[139]) && portSet[548] {
		return "nas"
	}

	// Vendor based
	if containsAny(vendor, "cisco", "juniper", "huawei", "mikrotik", "ubiquiti", "zyxel", "tp-link", "netgear", "asus") {
		if portSet[80] || portSet[443] || portSet[23] || portSet[22] || portSet[161] {
			return "router"
		}
	}
	// Mobile vendor with a completely closed host (typical for phones: no
	// services exposed). Apple is excluded here — a closed Apple device is
	// ambiguous (Mac/iPad/iPhone); only a clear phone hostname decides.
	if len(d.OpenPorts) == 0 {
		if m := mobileVendor(d.MAC); m != "" {
			ml := strings.ToLower(m)
			if !strings.Contains(ml, "apple") {
				return "phone"
			}
			if containsAny(host, "iphone", "ipod") {
				return "phone"
			}
			if containsAny(host, "ipad") {
				return "tablet"
			}
		}
	}
	if containsAny(vendor, "apple") && len(d.OpenPorts) > 0 {
		return "workstation"
	}
	if containsAny(vendor, "vmware", "virtualbox", "hyper-v", "qemu", "xen") {
		return "server"
	}

	// Hostname hints
	if host != "" {
		switch {
		case strings.Contains(host, "printer") || strings.Contains(host, "hp-"):
			return "printer"
		case strings.Contains(host, "camera") || strings.Contains(host, "cam-"):
			return "camera"
		case strings.Contains(host, "switch") || strings.Contains(host, "core-"):
			return "switch"
		case strings.Contains(host, "router") || strings.Contains(host, "gw") || strings.Contains(host, "gateway"):
			return "router"
		case strings.Contains(host, "firewall") || strings.Contains(host, "fw"):
			return "firewall"
		case strings.Contains(host, "server") || strings.Contains(host, "srv") || strings.Contains(host, "dc-"):
			return "server"
		case strings.Contains(host, "nas"):
			return "nas"
		}
	}

	return "unknown"
}

func (cl *Classifier) classifyWithLLM(ctx context.Context, d Device) string {
	prompt := fmt.Sprintf(`You are a network inventory classifier. Determine the device type from the evidence.
Rules:
- Use ports AND services AND vendor AND hostname together.
- A device with RDP(3389) is likely a Windows workstation, not a router.
- A device with MySQL/PostgreSQL/Redis is a database server.
- A printer usually has ports 515, 9100 or 631 and a printer vendor (HP, Canon, Epson, Brother).
- A network camera usually has RTSP (554) and vendors like Hikvision, Dahua.
- An AP has a MAC of a wireless vendor (Ubiquiti, Ruckus, Cisco Aironet, Aruba) and usually only port 80/443.
- A switch usually has telnet(23) or SSH(22) plus SNMP(161) and is a vendor like Cisco, Juniper, Huawei.
- A router usually has port 80/443 admin and vendor like MikroTik, Ubiquiti, TP-Link, Zyxel, but NOT web-servers like nginx/Apache.
- Do NOT call something a router just because it has port 80. Web servers (nginx, Apache) mean "server".
- A phone or tablet: ADB (5555), mDNS services _airplay/_raop/_companion-link/_hisuite/_mipcs, AirPlay model like "iPhone14,2"/"iPad13,4", hostnames like iphone/ipad/android-/SM-G9xxx/Redmi/Pixel/Galaxy, or a mobile vendor (Apple, Samsung, Xiaomi, Huawei, Google, OnePlus, Oppo, Vivo) with no open ports.
- Apple vendors alone are ambiguous: a Mac is "workstation", an iPhone/iPad is phone/tablet.
- "unknown" is a valid answer.

IP: %s
MAC: %s
Vendor: %s
Hostname: %s
OS: %s
Model: %s
Open ports: %s
Services: %s

Respond with raw JSON only:
{"deviceType":"server|workstation|router|switch|firewall|printer|camera|nas|ap|iot|phone|tablet|domain-controller|unknown","reason":"short reason"}`,
		d.IP, d.MAC, d.Vendor, d.Hostname, d.OS, d.Model, formatPorts(d.OpenPorts), formatServices(d.OpenPorts))

	body := map[string]interface{}{
		"model":  cl.model,
		"prompt": prompt,
		"stream": false,
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", cl.baseURL+"/api/generate", bytes.NewReader(b))
	if err != nil {
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cl.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	var result struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(raw, &result); err != nil || result.Response == "" {
		return ""
	}

	cleaned := cleanJSON(result.Response)
	var classification struct {
		DeviceType string `json:"deviceType"`
	}
	if err := json.Unmarshal([]byte(cleaned), &classification); err != nil {
		return ""
	}
	t := strings.ToLower(strings.TrimSpace(classification.DeviceType))
	switch t {
	case "server", "workstation", "router", "switch", "firewall", "printer", "camera", "nas", "ap", "iot", "phone", "tablet", "domain-controller", "unknown":
		return t
	}
	return ""
}

func formatPorts(ports []Port) string {
	var ps []string
	for _, p := range ports {
		ps = append(ps, fmt.Sprintf("%d/%s", p.Port, p.Proto))
	}
	return strings.Join(ps, ",")
}

func formatServices(ports []Port) string {
	var ss []string
	seen := map[string]bool{}
	for _, p := range ports {
		s := p.Service
		if s != "" && !seen[s] {
			ss = append(ss, s)
			seen[s] = true
		}
	}
	return strings.Join(ss, ",")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

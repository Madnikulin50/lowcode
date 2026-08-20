package agent

import "testing"

func TestClassifyHeuristics(t *testing.T) {
	cases := []struct {
		name string
		dev  Device
		want string
	}{
		{
			name: "domain controller",
			dev:  Device{OpenPorts: []Port{{Port: 88}, {Port: 389}, {Port: 445}}},
			want: "domain-controller",
		},
		{
			name: "db server with mysql",
			dev:  Device{OpenPorts: []Port{{Port: 3306}, {Port: 22}}, Vendor: "Dell Inc."},
			want: "server",
		},
		{
			name: "postgres server",
			dev:  Device{OpenPorts: []Port{{Port: 5432}, {Port: 22}}},
			want: "server",
		},
		{
			name: "printer by jetdirect",
			dev:  Device{OpenPorts: []Port{{Port: 9100}, {Port: 515}}, Vendor: "HP"},
			want: "printer",
		},
		{
			name: "printer by vendor",
			dev:  Device{OpenPorts: []Port{{Port: 80}}, Vendor: "Brother Industries, LTD."},
			want: "printer",
		},
		{
			name: "camera by rtsp",
			dev:  Device{OpenPorts: []Port{{Port: 554}}, Vendor: "Hikvision"},
			want: "camera",
		},
		{
			name: "router mikrotik winbox",
			dev:  Device{OpenPorts: []Port{{Port: 8291}, {Port: 80}}, Vendor: "MikroTik"},
			want: "router",
		},
		{
			name: "switch cisco telnet snmp",
			dev:  Device{OpenPorts: []Port{{Port: 23}, {Port: 161}}, Vendor: "Cisco Systems, Inc."},
			want: "switch",
		},
		{
			name: "windows workstation rdp",
			dev:  Device{OpenPorts: []Port{{Port: 3389}, {Port: 135}}, OS: "Windows"},
			want: "workstation",
		},
		{
			name: "linux server ssh web",
			dev:  Device{OpenPorts: []Port{{Port: 22}, {Port: 80}, {Port: 443}}, OS: "Linux", Vendor: "Dell Inc."},
			want: "server",
		},
		{
			name: "hostname printer",
			dev:  Device{OpenPorts: []Port{{Port: 80}}, Hostname: "hp-laserjet-4250"},
			want: "printer",
		},
		{
			name: "hostname server",
			dev:  Device{OpenPorts: []Port{{Port: 22}, {Port: 80}}, Hostname: "srv-web-01"},
			want: "server",
		},
		{
			name: "nas",
			dev:  Device{OpenPorts: []Port{{Port: 80}, {Port: 445}, {Port: 548}}, Vendor: "Synology"},
			want: "nas",
		},
		{
			name: "unknown empty",
			dev:  Device{OpenPorts: []Port{{Port: 8080}}},
			want: "unknown",
		},
		{
			name: "iphone by airplay model",
			dev:  Device{Model: "iPhone14,2", Services: []string{"airplay"}, Hostname: "iPhone-ABCDEF"},
			want: "phone",
		},
		{
			name: "ipad by airplay model",
			dev:  Device{Model: "iPad13,4", Services: []string{"airplay", "raop"}},
			want: "tablet",
		},
		{
			name: "iphone by hostname",
			dev:  Device{Hostname: "iPhone-Максима"},
			want: "phone",
		},
		{
			name: "ipad by hostname",
			dev:  Device{Hostname: "iPad-Marya"},
			want: "tablet",
		},
		{
			name: "android by adb port",
			dev:  Device{OpenPorts: []Port{{Port: 5555, Service: "adb"}}},
			want: "phone",
		},
		{
			name: "huawei phone by hisuite",
			dev:  Device{Services: []string{"hisuite"}, Vendor: "Huawei Technologies Co.,Ltd"},
			want: "phone",
		},
		{
			name: "xiaomi phone by mipcs",
			dev:  Device{Services: []string{"mipcs"}},
			want: "phone",
		},
		{
			name: "samsung phone by hostname",
			dev:  Device{Hostname: "SM-G991B"},
			want: "phone",
		},
		{
			name: "samsung tablet by hostname",
			dev:  Device{Hostname: "SM-T870"},
			want: "tablet",
		},
		{
			name: "android default hostname",
			dev:  Device{Hostname: "android-3f2a1b"},
			want: "phone",
		},
		{
			name: "closed samsung host",
			dev:  Device{MAC: "00:12:47:12:34:56"},
			want: "phone",
		},
		{
			name: "huawei phone not switch without ports",
			dev:  Device{OpenPorts: nil, Vendor: "Huawei Technologies Co.,Ltd", Hostname: "HUAWEI-P30"},
			want: "phone",
		},
		{
			name: "huawei switch with telnet",
			dev:  Device{OpenPorts: []Port{{Port: 23}, {Port: 161}}, Vendor: "Huawei Technologies Co.,Ltd"},
			want: "switch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := strongDeviceType(tc.dev)
			if got == "" {
				got = guessDeviceType(tc.dev)
			}
			if got == "" && tc.want != "" {
				got = "(none)"
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

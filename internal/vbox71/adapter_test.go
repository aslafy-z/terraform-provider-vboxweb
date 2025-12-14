package vbox71

import (
	"testing"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxapi"
)

func TestParseNATRedirect71(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    vboxapi.NATRedirect
		wantErr bool
	}{
		{
			name:  "valid TCP rule",
			input: "ssh,1,127.0.0.1,2222,10.0.2.15,22",
			want: vboxapi.NATRedirect{
				Name:      "ssh",
				Protocol:  vboxapi.NATProtocolTCP,
				HostIP:    "127.0.0.1",
				HostPort:  2222,
				GuestIP:   "10.0.2.15",
				GuestPort: 22,
			},
		},
		{
			name:  "valid UDP rule",
			input: "dns,0,,53,,53",
			want: vboxapi.NATRedirect{
				Name:      "dns",
				Protocol:  vboxapi.NATProtocolUDP,
				HostIP:    "",
				HostPort:  53,
				GuestIP:   "",
				GuestPort: 53,
			},
		},
		{
			name:  "empty host and guest IP",
			input: "web,1,,8080,,80",
			want: vboxapi.NATRedirect{
				Name:      "web",
				Protocol:  vboxapi.NATProtocolTCP,
				HostIP:    "",
				HostPort:  8080,
				GuestIP:   "",
				GuestPort: 80,
			},
		},
		{
			name:    "invalid format - too few fields",
			input:   "ssh,1,127.0.0.1,2222",
			wantErr: true,
		},
		{
			name:    "invalid format - too many fields",
			input:   "ssh,1,127.0.0.1,2222,10.0.2.15,22,extra",
			wantErr: true,
		},
		{
			name:    "invalid protocol",
			input:   "ssh,2,127.0.0.1,2222,10.0.2.15,22",
			wantErr: true,
		},
		{
			name:    "invalid host port",
			input:   "ssh,1,127.0.0.1,invalid,10.0.2.15,22",
			wantErr: true,
		},
		{
			name:    "invalid guest port",
			input:   "ssh,1,127.0.0.1,2222,10.0.2.15,invalid",
			wantErr: true,
		},
		{
			name:    "host port out of range",
			input:   "ssh,1,127.0.0.1,99999,10.0.2.15,22",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNATRedirect71(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNATRedirect71() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Protocol != tt.want.Protocol {
				t.Errorf("Protocol = %v, want %v", got.Protocol, tt.want.Protocol)
			}
			if got.HostIP != tt.want.HostIP {
				t.Errorf("HostIP = %v, want %v", got.HostIP, tt.want.HostIP)
			}
			if got.HostPort != tt.want.HostPort {
				t.Errorf("HostPort = %v, want %v", got.HostPort, tt.want.HostPort)
			}
			if got.GuestIP != tt.want.GuestIP {
				t.Errorf("GuestIP = %v, want %v", got.GuestIP, tt.want.GuestIP)
			}
			if got.GuestPort != tt.want.GuestPort {
				t.Errorf("GuestPort = %v, want %v", got.GuestPort, tt.want.GuestPort)
			}
		})
	}
}

func TestParseNATNetworkRule71(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    vboxapi.NATRedirect
		wantErr bool
	}{
		{
			name:  "valid TCP rule",
			input: "ssh:tcp:127.0.0.1:2222:10.0.2.15:22",
			want: vboxapi.NATRedirect{
				Name:      "ssh",
				Protocol:  vboxapi.NATProtocolTCP,
				HostIP:    "127.0.0.1",
				HostPort:  2222,
				GuestIP:   "10.0.2.15",
				GuestPort: 22,
			},
		},
		{
			name:  "valid UDP rule uppercase",
			input: "dns:UDP::53::53",
			want: vboxapi.NATRedirect{
				Name:      "dns",
				Protocol:  vboxapi.NATProtocolUDP,
				HostIP:    "",
				HostPort:  53,
				GuestIP:   "",
				GuestPort: 53,
			},
		},
		{
			name:  "valid UDP rule lowercase",
			input: "dns:udp::53::53",
			want: vboxapi.NATRedirect{
				Name:      "dns",
				Protocol:  vboxapi.NATProtocolUDP,
				HostIP:    "",
				HostPort:  53,
				GuestIP:   "",
				GuestPort: 53,
			},
		},
		{
			name:    "invalid format - too few fields",
			input:   "ssh:tcp:127.0.0.1:2222",
			wantErr: true,
		},
		{
			name:    "invalid protocol",
			input:   "ssh:icmp:127.0.0.1:2222:10.0.2.15:22",
			wantErr: true,
		},
		{
			name:    "invalid host port",
			input:   "ssh:tcp:127.0.0.1:invalid:10.0.2.15:22",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNATNetworkRule71(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNATNetworkRule71() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
			if got.Protocol != tt.want.Protocol {
				t.Errorf("Protocol = %v, want %v", got.Protocol, tt.want.Protocol)
			}
			if got.HostIP != tt.want.HostIP {
				t.Errorf("HostIP = %v, want %v", got.HostIP, tt.want.HostIP)
			}
			if got.HostPort != tt.want.HostPort {
				t.Errorf("HostPort = %v, want %v", got.HostPort, tt.want.HostPort)
			}
		})
	}
}

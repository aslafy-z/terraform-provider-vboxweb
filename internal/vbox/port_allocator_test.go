package vbox

import (
	"testing"
)

func TestSelectAvailablePort(t *testing.T) {
	tests := []struct {
		name      string
		usedPorts []UsedPort
		opts      PortAllocatorOptions
		want      uint16
		wantErr   bool
	}{
		{
			name:      "no ports used - picks minimum",
			usedPorts: []UsedPort{},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				Scope:   HostIPScopeAny,
			},
			want: 20000,
		},
		{
			name: "first port used - picks second",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: ""},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				Scope:   HostIPScopeAny,
			},
			want: 20001,
		},
		{
			name: "multiple ports used - picks first available",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: ""},
				{Port: 20001, HostIP: ""},
				{Port: 20002, HostIP: ""},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				Scope:   HostIPScopeAny,
			},
			want: 20003,
		},
		{
			name: "gap in used ports - picks lowest gap",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: ""},
				{Port: 20002, HostIP: ""},
				{Port: 20003, HostIP: ""},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				Scope:   HostIPScopeAny,
			},
			want: 20001,
		},
		{
			name: "all ports used - error",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: ""},
				{Port: 20001, HostIP: ""},
				{Port: 20002, HostIP: ""},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20002,
				Scope:   HostIPScopeAny,
			},
			wantErr: true,
		},
		{
			name: "exact scope - different IP not conflicting",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: "192.168.1.1"},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				HostIP:  "127.0.0.1",
				Scope:   HostIPScopeExact,
			},
			want: 20000, // Same port OK because different host IP
		},
		{
			name: "exact scope - same IP conflicts",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: "127.0.0.1"},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				HostIP:  "127.0.0.1",
				Scope:   HostIPScopeExact,
			},
			want: 20001, // Must pick next port
		},
		{
			name: "exact scope - any IP (empty) always conflicts",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: ""}, // Empty = any
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				HostIP:  "127.0.0.1",
				Scope:   HostIPScopeExact,
			},
			want: 20001, // Empty conflicts with everything
		},
		{
			name: "any scope - all IPs conflict",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: "192.168.1.1"},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20010,
				HostIP:  "127.0.0.1",
				Scope:   HostIPScopeAny,
			},
			want: 20001, // Any scope treats all as conflicting
		},
		{
			name:      "invalid range - min > max",
			usedPorts: []UsedPort{},
			opts: PortAllocatorOptions{
				MinPort: 20010,
				MaxPort: 20000,
				Scope:   HostIPScopeAny,
			},
			wantErr: true,
		},
		{
			name:      "single port range - available",
			usedPorts: []UsedPort{},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20000,
				Scope:   HostIPScopeAny,
			},
			want: 20000,
		},
		{
			name: "single port range - used",
			usedPorts: []UsedPort{
				{Port: 20000, HostIP: ""},
			},
			opts: PortAllocatorOptions{
				MinPort: 20000,
				MaxPort: 20000,
				Scope:   HostIPScopeAny,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SelectAvailablePort(tt.usedPorts, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("SelectAvailablePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("SelectAvailablePort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultPortAllocatorOptions(t *testing.T) {
	opts := DefaultPortAllocatorOptions()

	if opts.MinPort != 20000 {
		t.Errorf("MinPort = %v, want 20000", opts.MinPort)
	}
	if opts.MaxPort != 40000 {
		t.Errorf("MaxPort = %v, want 40000", opts.MaxPort)
	}
	if opts.Scope != HostIPScopeAny {
		t.Errorf("Scope = %v, want %v", opts.Scope, HostIPScopeAny)
	}
	if !opts.IncludeNATNetworks {
		t.Errorf("IncludeNATNetworks = false, want true")
	}
}

func TestUsedPortsByPort(t *testing.T) {
	tests := []struct {
		name      string
		usedPorts []UsedPort
		want      []uint16
	}{
		{
			name:      "empty",
			usedPorts: []UsedPort{},
			want:      nil,
		},
		{
			name: "single port",
			usedPorts: []UsedPort{
				{Port: 2222, HostIP: ""},
			},
			want: []uint16{2222},
		},
		{
			name: "multiple ports - sorted",
			usedPorts: []UsedPort{
				{Port: 8080, HostIP: ""},
				{Port: 2222, HostIP: ""},
				{Port: 3389, HostIP: ""},
			},
			want: []uint16{2222, 3389, 8080},
		},
		{
			name: "duplicate ports - deduplicated",
			usedPorts: []UsedPort{
				{Port: 2222, HostIP: "127.0.0.1"},
				{Port: 2222, HostIP: "192.168.1.1"},
				{Port: 8080, HostIP: ""},
			},
			want: []uint16{2222, 8080},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UsedPortsByPort(tt.usedPorts)
			if len(got) != len(tt.want) {
				t.Errorf("UsedPortsByPort() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("UsedPortsByPort()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestHostIPScopeConstants(t *testing.T) {
	// Verify the scope constants have expected values
	if HostIPScopeAny != "any" {
		t.Errorf("HostIPScopeAny = %q, want %q", HostIPScopeAny, "any")
	}
	if HostIPScopeExact != "exact" {
		t.Errorf("HostIPScopeExact = %q, want %q", HostIPScopeExact, "exact")
	}
}

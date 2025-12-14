package vbox

import (
	"testing"
)

func TestNormalizeHostIP(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"0.0.0.0", ""},
		{"127.0.0.1", "127.0.0.1"},
		{"192.168.1.1", "192.168.1.1"},
		{"  0.0.0.0  ", ""},
		{"  127.0.0.1  ", "127.0.0.1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizeHostIP(tt.input); got != tt.want {
				t.Errorf("NormalizeHostIP(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHostIPConflicts(t *testing.T) {
	tests := []struct {
		name string
		ip1  string
		ip2  string
		want bool
	}{
		{"both empty", "", "", true},
		{"both 0.0.0.0", "0.0.0.0", "0.0.0.0", true},
		{"empty vs 0.0.0.0", "", "0.0.0.0", true},
		{"any vs specific", "", "127.0.0.1", true},
		{"specific vs any", "127.0.0.1", "", true},
		{"same specific", "127.0.0.1", "127.0.0.1", true},
		{"different specific", "127.0.0.1", "192.168.1.1", false},
		{"0.0.0.0 vs specific", "0.0.0.0", "127.0.0.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HostIPConflicts(tt.ip1, tt.ip2); got != tt.want {
				t.Errorf("HostIPConflicts(%q, %q) = %v, want %v", tt.ip1, tt.ip2, got, tt.want)
			}
		})
	}
}

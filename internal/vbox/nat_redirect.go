// Package vbox provides high-level VirtualBox operations.
package vbox

import "strings"

// NormalizeHostIP normalizes a host IP address for comparison.
// Empty string and "0.0.0.0" are treated equivalently as "any".
func NormalizeHostIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" || ip == "0.0.0.0" {
		return ""
	}
	return ip
}

// HostIPConflicts checks if two host IPs conflict.
// Two IPs conflict if:
// - Both are "any" (empty or 0.0.0.0)
// - Either is "any"
// - They are exactly the same
func HostIPConflicts(ip1, ip2 string) bool {
	n1 := NormalizeHostIP(ip1)
	n2 := NormalizeHostIP(ip2)

	// If either is "any", they conflict
	if n1 == "" || n2 == "" {
		return true
	}

	// Otherwise, only conflict if identical
	return n1 == n2
}

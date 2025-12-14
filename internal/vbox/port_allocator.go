package vbox

import (
	"context"
	"fmt"
	"sort"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxapi"
)

// HostIPScope determines how host IP addresses are considered when checking for port conflicts.
type HostIPScope string

const (
	// HostIPScopeAny treats all host IP bindings as conflicting (conservative).
	HostIPScopeAny HostIPScope = "any"
	// HostIPScopeExact only considers rules with the same host_ip as conflicting.
	HostIPScopeExact HostIPScope = "exact"
)

// PortAllocatorOptions configures the auto host port selection.
type PortAllocatorOptions struct {
	// MinPort is the minimum port in the allocation range (inclusive).
	MinPort uint16
	// MaxPort is the maximum port in the allocation range (inclusive).
	MaxPort uint16
	// HostIP is the host IP address that the new rule will bind to.
	HostIP string
	// Scope determines how host IP addresses are considered for conflicts.
	Scope HostIPScope
	// IncludeNATNetworks includes NAT Network port forward rules in conflict detection.
	IncludeNATNetworks bool
}

// DefaultPortAllocatorOptions returns default options for port allocation.
func DefaultPortAllocatorOptions() PortAllocatorOptions {
	return PortAllocatorOptions{
		MinPort:            20000,
		MaxPort:            40000,
		HostIP:             "",
		Scope:              HostIPScopeAny,
		IncludeNATNetworks: true,
	}
}

// UsedPort represents a port that is in use, along with its binding info.
type UsedPort struct {
	Port   uint16
	HostIP string
}

// CollectUsedPorts enumerates all NAT port forwarding rules across all VMs (and optionally
// NAT Networks) and returns the set of used host ports.
func CollectUsedPorts(ctx context.Context, api vboxapi.VBoxAPI, session string, includeNATNetworks bool) ([]UsedPort, error) {
	var usedPorts []UsedPort

	// Get all machines
	machineRefs, err := api.GetMachines(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate machines: %w", err)
	}

	// For each machine, check all network adapter slots (0-7)
	for _, machineRef := range machineRefs {
		for slot := uint32(0); slot <= 7; slot++ {
			adapterRef, err := api.GetNetworkAdapter(ctx, machineRef, slot)
			if err != nil {
				// Adapter might not exist or not accessible, skip
				continue
			}

			natEngineRef, err := api.GetNATEngine(ctx, adapterRef)
			if err != nil {
				// NAT engine might not be available (different attachment type)
				continue
			}

			redirects, err := api.GetNATRedirects(ctx, natEngineRef)
			if err != nil {
				continue
			}

			for _, r := range redirects {
				usedPorts = append(usedPorts, UsedPort{
					Port:   r.HostPort,
					HostIP: r.HostIP,
				})
			}
		}
	}

	// Optionally include NAT Network rules
	if includeNATNetworks {
		natNetworkRefs, err := api.GetNATNetworks(ctx, session)
		if err == nil { // Ignore errors - NAT Networks might not be available
			for _, natNetRef := range natNetworkRefs {
				rules, err := api.GetNATNetworkPortForwardRules4(ctx, natNetRef)
				if err != nil {
					continue
				}

				for _, r := range rules {
					usedPorts = append(usedPorts, UsedPort{
						Port:   r.HostPort,
						HostIP: r.HostIP,
					})
				}
			}
		}
	}

	return usedPorts, nil
}

// SelectAvailablePort selects an available port from the given range that does not
// conflict with any used ports.
func SelectAvailablePort(usedPorts []UsedPort, opts PortAllocatorOptions) (uint16, error) {
	if opts.MinPort > opts.MaxPort {
		return 0, fmt.Errorf("invalid port range: min %d > max %d", opts.MinPort, opts.MaxPort)
	}

	// Build a set of ports that are considered "used" based on the scope
	usedSet := make(map[uint16]bool)
	for _, up := range usedPorts {
		conflicting := false
		if opts.Scope == HostIPScopeAny {
			// All ports are conflicting regardless of host IP
			conflicting = true
		} else {
			// Only conflict if host IPs actually conflict
			conflicting = HostIPConflicts(opts.HostIP, up.HostIP)
		}
		if conflicting {
			usedSet[up.Port] = true
		}
	}

	// Find the lowest available port in the range
	for port := opts.MinPort; port <= opts.MaxPort; port++ {
		if !usedSet[port] {
			return port, nil
		}
	}

	// No available ports
	rangeSize := int(opts.MaxPort) - int(opts.MinPort) + 1
	usedInRange := 0
	for port := opts.MinPort; port <= opts.MaxPort; port++ {
		if usedSet[port] {
			usedInRange++
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d: %d of %d ports are in use by other VirtualBox NAT rules",
		opts.MinPort, opts.MaxPort, usedInRange, rangeSize)
}

// AllocatePort is a convenience function that collects used ports and selects an available one.
func AllocatePort(ctx context.Context, api vboxapi.VBoxAPI, session string, opts PortAllocatorOptions) (uint16, error) {
	usedPorts, err := CollectUsedPorts(ctx, api, session, opts.IncludeNATNetworks)
	if err != nil {
		return 0, err
	}
	return SelectAvailablePort(usedPorts, opts)
}

// UsedPortsByPort returns a sorted list of unique ports that are in use.
func UsedPortsByPort(usedPorts []UsedPort) []uint16 {
	seen := make(map[uint16]bool)
	var result []uint16
	for _, up := range usedPorts {
		if !seen[up.Port] {
			seen[up.Port] = true
			result = append(result, up.Port)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

// Package vboxapi defines the interface for VirtualBox SOAP operations.
// This package has no dependencies on other internal packages to avoid import cycles.
package vboxapi

import "context"

// VBoxAPI defines the interface for VirtualBox SOAP operations.
// This abstraction allows supporting multiple VirtualBox versions with
// version-specific implementations.
type VBoxAPI interface {
	// Session management
	Logon(ctx context.Context, username, password string) (session string, err error)
	Logoff(ctx context.Context, session string) error
	GetSessionObject(ctx context.Context, session string) (sessionObj string, err error)

	// Machine lookup and enumeration
	FindMachine(ctx context.Context, session, nameOrID string) (machineRef string, err error)
	GetMachines(ctx context.Context, session string) (machineRefs []string, err error)

	// Machine creation and registration
	CreateMachine(ctx context.Context, session, name, osTypeId, sourceMachineRef string) (machineRef string, err error)
	RegisterMachine(ctx context.Context, session, machineRef string) error
	UnregisterMachine(ctx context.Context, machineRef string) (mediaRefs []string, err error)
	DeleteConfig(ctx context.Context, machineRef string, mediaRefs []string) (progressRef string, err error)

	// Machine properties
	GetMachineId(ctx context.Context, machineRef string) (uuid string, err error)
	GetMachineName(ctx context.Context, machineRef string) (name string, err error)
	GetMachineState(ctx context.Context, machineRef string) (state string, err error)
	GetOSTypeId(ctx context.Context, machineRef string) (osTypeId string, err error)

	// Clone
	CloneTo(ctx context.Context, srcMachineRef, targetMachineRef, mode string, options []string) (progressRef string, err error)

	// Power management
	LaunchVMProcess(ctx context.Context, machineRef, sessionObj, sessionType string) (progressRef string, err error)
	LockMachine(ctx context.Context, machineRef, sessionObj string, shared bool) error
	UnlockSession(ctx context.Context, sessionObj string) error
	GetConsole(ctx context.Context, sessionObj string) (consoleRef string, err error)
	PowerDown(ctx context.Context, consoleRef string) (progressRef string, err error)

	// Progress monitoring
	GetProgressCompleted(ctx context.Context, progressRef string) (completed bool, err error)
	GetProgressResultCode(ctx context.Context, progressRef string) (resultCode int32, err error)
	GetProgressErrorText(ctx context.Context, progressRef string) (errorText string, err error)

	// Network adapters and NAT engine
	GetNetworkAdapter(ctx context.Context, machineRef string, slot uint32) (adapterRef string, err error)
	GetNATEngine(ctx context.Context, adapterRef string) (natEngineRef string, err error)
	GetNATRedirects(ctx context.Context, natEngineRef string) ([]NATRedirect, error)
	AddNATRedirect(ctx context.Context, natEngineRef, name string, proto NATProtocol, hostIP string, hostPort uint16, guestIP string, guestPort uint16) error
	RemoveNATRedirect(ctx context.Context, natEngineRef, name string) error

	// NAT Networks (for port conflict detection across NAT networks)
	GetNATNetworks(ctx context.Context, session string) (natNetworkRefs []string, err error)
	GetNATNetworkPortForwardRules4(ctx context.Context, natNetworkRef string) ([]NATRedirect, error)

	// Mutable machine operations (require lock)
	GetMutableMachine(ctx context.Context, sessionObj string) (mutableMachineRef string, err error)
	SaveSettings(ctx context.Context, machineRef string) error

	// Version info
	GetAPIVersion(ctx context.Context, session string) (version string, err error)
}

// NATProtocol represents the protocol for NAT port forwarding.
type NATProtocol string

const (
	NATProtocolTCP NATProtocol = "TCP"
	NATProtocolUDP NATProtocol = "UDP"
)

// NATRedirect represents a parsed NAT port forwarding rule.
// The adapter is responsible for parsing the version-specific format.
type NATRedirect struct {
	Name      string
	Protocol  NATProtocol
	HostIP    string
	HostPort  uint16
	GuestIP   string
	GuestPort uint16
}

// MachineState constants normalized across versions.
const (
	MachineStateNull       = "Null"
	MachineStatePoweredOff = "PoweredOff"
	MachineStateRunning    = "Running"
	MachineStateSaved      = "Saved"
	MachineStatePaused     = "Paused"
)

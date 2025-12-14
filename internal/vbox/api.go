// Package vbox provides a client for interacting with VirtualBox via vboxwebsrv SOAP API.
// It supports multiple VirtualBox versions through version-specific adapters.
package vbox

import "context"

// VBoxAPI defines the interface for VirtualBox SOAP operations.
// This abstraction allows supporting multiple VirtualBox versions with
// version-specific implementations.
type VBoxAPI interface {
	// Session management
	Logon(ctx context.Context, username, password string) (session string, err error)
	Logoff(ctx context.Context, session string) error
	GetSessionObject(ctx context.Context, session string) (sessionObj string, err error)

	// Machine lookup
	FindMachine(ctx context.Context, session, nameOrID string) (machineRef string, err error)

	// Machine creation and registration
	CreateMachine(ctx context.Context, session, name, osTypeId, sourceMachineRef string) (machineRef string, err error)
	RegisterMachine(ctx context.Context, session, machineRef string) error
	UnregisterMachine(ctx context.Context, machineRef string) (mediaRefs []string, err error)
	DeleteConfig(ctx context.Context, machineRef string, mediaRefs []string) (progressRef string, err error)

	// Machine properties
	GetMachineId(ctx context.Context, machineRef string) (uuid string, err error)
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

	// Version info
	GetAPIVersion(ctx context.Context, session string) (version string, err error)
}

// MachineState constants normalized across versions.
const (
	MachineStateNull       = "Null"
	MachineStatePoweredOff = "PoweredOff"
	MachineStateRunning    = "Running"
	MachineStateSaved      = "Saved"
	MachineStatePaused     = "Paused"
)

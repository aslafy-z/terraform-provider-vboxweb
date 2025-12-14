package vbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox71"
	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxapi"
)

// Client provides high-level operations for VirtualBox management.
type Client struct {
	endpoint string
	username string
	password string
}

// NewClient creates a new VirtualBox client.
func NewClient(endpoint, username, password string) *Client {
	return &Client{endpoint: endpoint, username: username, password: password}
}

// CloneRequest describes a VM clone operation.
type CloneRequest struct {
	Name         string
	Source       string
	CloneMode    string
	CloneOptions []string
	DesiredState string // started|stopped
	SessionType  string // headless|gui
	Timeout      time.Duration
}

var errNotFound = errors.New("not found")

// IsNotFound returns true if the error indicates a resource was not found.
func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

// newAdapter creates a version-appropriate adapter.
// Currently only supports VBox 7.1, but designed for future version support.
func newAdapter(endpoint string) vboxapi.VBoxAPI {
	// TODO: In the future, could auto-detect version and return appropriate adapter
	return vbox71.NewAdapter(endpoint)
}

func (c *Client) withSession(ctx context.Context, fn func(ctx context.Context, api vboxapi.VBoxAPI, session string) error) error {
	api := newAdapter(c.endpoint)

	session, err := api.Logon(ctx, c.username, c.password)
	if err != nil {
		return err
	}

	// Always try to logoff.
	defer func() {
		_ = api.Logoff(context.Background(), session)
	}()

	return fn(ctx, api, session)
}

// CloneAndConverge creates a new VM by cloning and sets its power state.
func (c *Client) CloneAndConverge(ctx context.Context, req CloneRequest) (uuid string, currentState string, err error) {
	if strings.TrimSpace(req.Name) == "" {
		return "", "", fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Source) == "" {
		return "", "", fmt.Errorf("source is required")
	}
	if req.Timeout <= 0 {
		req.Timeout = 20 * time.Minute
	}
	if req.SessionType == "" {
		req.SessionType = "headless"
	}
	if req.CloneMode == "" {
		req.CloneMode = "MachineState"
	}
	if req.DesiredState == "" {
		req.DesiredState = "stopped"
	}

	err = c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		srcRef, err := findMachine(ctx, api, session, req.Source)
		if err != nil {
			return err
		}

		// Get source osTypeId for the new machine
		osTypeId, err := api.GetOSTypeId(ctx, srcRef)
		if err != nil {
			return fmt.Errorf("failed to get source OS type: %w", err)
		}

		targetRef, err := api.CreateMachine(ctx, session, req.Name, osTypeId, srcRef)
		if err != nil {
			return err
		}

		progressRef, err := api.CloneTo(ctx, srcRef, targetRef, req.CloneMode, req.CloneOptions)
		if err != nil {
			return err
		}
		if err := waitProgress(ctx, api, progressRef, req.Timeout); err != nil {
			return err
		}

		if err := api.RegisterMachine(ctx, session, targetRef); err != nil {
			return err
		}

		uuid, err = api.GetMachineId(ctx, targetRef)
		if err != nil {
			return err
		}

		// Converge state
		currentState, err = convergeState(ctx, api, session, targetRef, req.DesiredState, req.SessionType, req.Timeout)
		if err != nil {
			return err
		}
		return nil
	})

	return uuid, currentState, err
}

// MachineInfo contains basic information about a VirtualBox machine.
type MachineInfo struct {
	ID    string
	Name  string
	State string
}

// GetMachineInfoByID returns basic information about a VM by its UUID.
func (c *Client) GetMachineInfoByID(ctx context.Context, id string) (*MachineInfo, error) {
	var info MachineInfo
	err := c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		mRef, err := findMachine(ctx, api, session, id)
		if err != nil {
			return err
		}
		info.ID, err = api.GetMachineId(ctx, mRef)
		if err != nil {
			return err
		}
		info.Name, err = api.GetMachineName(ctx, mRef)
		if err != nil {
			return err
		}
		info.State, err = api.GetMachineState(ctx, mRef)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// GetStateByID returns the current state of a VM by its UUID.
func (c *Client) GetStateByID(ctx context.Context, id string) (string, error) {
	var out string
	err := c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		mRef, err := findMachine(ctx, api, session, id)
		if err != nil {
			return err
		}
		st, err := api.GetMachineState(ctx, mRef)
		if err != nil {
			return err
		}
		out = st
		return nil
	})
	return out, err
}

// ConvergeStateByID changes a VM's power state.
func (c *Client) ConvergeStateByID(ctx context.Context, id, desiredState, sessionType string, timeout time.Duration) (string, error) {
	var out string
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}
	if sessionType == "" {
		sessionType = "headless"
	}
	desiredState = strings.ToLower(strings.TrimSpace(desiredState))
	if desiredState != "started" && desiredState != "stopped" {
		return "", fmt.Errorf("invalid desired state: %s", desiredState)
	}

	err := c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		mRef, err := findMachine(ctx, api, session, id)
		if err != nil {
			return err
		}
		out, err = convergeState(ctx, api, session, mRef, desiredState, sessionType, timeout)
		return err
	})
	return out, err
}

// DeleteByID deletes a VM by its UUID.
func (c *Client) DeleteByID(ctx context.Context, id string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}

	return c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		mRef, err := findMachine(ctx, api, session, id)
		if err != nil {
			return err
		}

		// Ensure powered off (best-effort).
		_ = ensurePoweredOff(ctx, api, session, mRef, timeout)

		mediaRefs, err := api.UnregisterMachine(ctx, mRef)
		if err != nil {
			return err
		}

		progressRef, err := api.DeleteConfig(ctx, mRef, mediaRefs)
		if err != nil {
			return err
		}
		if err := waitProgress(ctx, api, progressRef, timeout); err != nil {
			return err
		}

		return nil
	})
}

// ---- helpers ----

func findMachine(ctx context.Context, api vboxapi.VBoxAPI, session, nameOrID string) (string, error) {
	machineRef, err := api.FindMachine(ctx, session, nameOrID)
	if err != nil {
		// Best-effort mapping to not found.
		errLower := strings.ToLower(err.Error())
		if strings.Contains(errLower, "could not find") || strings.Contains(errLower, "object not found") {
			return "", fmt.Errorf("%w: machine %s", errNotFound, nameOrID)
		}
		return "", err
	}
	if strings.TrimSpace(machineRef) == "" {
		return "", fmt.Errorf("%w: machine %s", errNotFound, nameOrID)
	}
	return machineRef, nil
}

func waitProgress(ctx context.Context, api vboxapi.VBoxAPI, progressRef string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}
	deadline := time.Now().Add(timeout)
	pollInterval := 2 * time.Second

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if we've exceeded deadline
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for progress after %v", timeout)
		}

		// Check if completed
		completed, err := api.GetProgressCompleted(ctx, progressRef)
		if err != nil {
			return fmt.Errorf("failed to get progress completion status: %w", err)
		}

		if completed {
			// Operation completed, check result
			rc, err := api.GetProgressResultCode(ctx, progressRef)
			if err != nil {
				return fmt.Errorf("failed to get progress result code: %w", err)
			}
			if rc != 0 {
				// Try to fetch an error message.
				errText, _ := api.GetProgressErrorText(ctx, progressRef)
				if errText != "" {
					return fmt.Errorf("progress failed (resultCode=%d): %s", rc, errText)
				}
				return fmt.Errorf("progress failed (resultCode=%d)", rc)
			}
			return nil
		}

		// Not completed yet, wait and poll again
		time.Sleep(pollInterval)
	}
}

func convergeState(ctx context.Context, api vboxapi.VBoxAPI, vboxSession string, machineRef, desiredState, sessionType string, timeout time.Duration) (string, error) {
	st, err := api.GetMachineState(ctx, machineRef)
	if err != nil {
		return "", err
	}

	want := strings.ToLower(desiredState)
	if want == "started" {
		if st == vboxapi.MachineStateRunning {
			return st, nil
		}
		if err := ensureRunning(ctx, api, vboxSession, machineRef, sessionType, timeout); err != nil {
			return "", err
		}
	} else if want == "stopped" {
		if st == vboxapi.MachineStatePoweredOff {
			return st, nil
		}
		if err := ensurePoweredOff(ctx, api, vboxSession, machineRef, timeout); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("invalid desired state: %s", desiredState)
	}

	st, err = api.GetMachineState(ctx, machineRef)
	if err != nil {
		return "", err
	}
	return st, nil
}

func ensureRunning(ctx context.Context, api vboxapi.VBoxAPI, vboxSession, machineRef, sessionType string, timeout time.Duration) error {
	sessObj, err := api.GetSessionObject(ctx, vboxSession)
	if err != nil {
		return err
	}

	progressRef, err := api.LaunchVMProcess(ctx, machineRef, sessObj, sessionType)
	if err != nil {
		return err
	}

	if err := waitProgress(ctx, api, progressRef, timeout); err != nil {
		return err
	}

	// Always unlock.
	_ = api.UnlockSession(context.Background(), sessObj)
	return nil
}

func ensurePoweredOff(ctx context.Context, api vboxapi.VBoxAPI, vboxSession, machineRef string, timeout time.Duration) error {
	sessObj, err := api.GetSessionObject(ctx, vboxSession)
	if err != nil {
		return err
	}

	err = api.LockMachine(ctx, machineRef, sessObj, true)
	if err != nil {
		// If already powered off or not lockable, bubble up.
		return err
	}

	consoleRef, err := api.GetConsole(ctx, sessObj)
	if err != nil {
		return err
	}

	progressRef, err := api.PowerDown(ctx, consoleRef)
	if err != nil {
		return err
	}

	if err := waitProgress(ctx, api, progressRef, timeout); err != nil {
		return err
	}

	_ = api.UnlockSession(context.Background(), sessObj)
	return nil
}

// NATPortForwardRule represents a NAT port forwarding rule.
type NATPortForwardRule struct {
	MachineID   string
	AdapterSlot uint32
	Name        string
	Protocol    vboxapi.NATProtocol
	HostIP      string
	HostPort    uint16
	GuestIP     string
	GuestPort   uint16
}

// CreateNATPortForward creates a new NAT port forwarding rule on a VM's adapter.
// The VM must be powered off or the adapter settings must allow hot changes.
func (c *Client) CreateNATPortForward(ctx context.Context, rule NATPortForwardRule) error {
	return c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		// Find the machine
		machineRef, err := findMachine(ctx, api, session, rule.MachineID)
		if err != nil {
			return err
		}

		// Get a session object to lock the machine
		sessObj, err := api.GetSessionObject(ctx, session)
		if err != nil {
			return fmt.Errorf("failed to get session object: %w", err)
		}

		// Lock the machine with shared lock (allows modifying settings while VM is running)
		if err := api.LockMachine(ctx, machineRef, sessObj, true); err != nil {
			return fmt.Errorf("failed to lock machine: %w", err)
		}
		defer func() { _ = api.UnlockSession(context.Background(), sessObj) }()

		// Get the mutable machine reference
		mutableMachineRef, err := api.GetMutableMachine(ctx, sessObj)
		if err != nil {
			return fmt.Errorf("failed to get mutable machine: %w", err)
		}

		// Get the network adapter
		adapterRef, err := api.GetNetworkAdapter(ctx, mutableMachineRef, rule.AdapterSlot)
		if err != nil {
			return fmt.Errorf("failed to get network adapter slot %d: %w", rule.AdapterSlot, err)
		}

		// Get the NAT engine
		natEngineRef, err := api.GetNATEngine(ctx, adapterRef)
		if err != nil {
			return fmt.Errorf("failed to get NAT engine: %w", err)
		}

		// Add the redirect
		if err := api.AddNATRedirect(ctx, natEngineRef, rule.Name, rule.Protocol, rule.HostIP, rule.HostPort, rule.GuestIP, rule.GuestPort); err != nil {
			return fmt.Errorf("failed to add NAT redirect: %w", err)
		}

		// Save settings
		if err := api.SaveSettings(ctx, mutableMachineRef); err != nil {
			return fmt.Errorf("failed to save machine settings: %w", err)
		}

		return nil
	})
}

// ReadNATPortForward reads a NAT port forwarding rule by name.
// Returns nil, nil if the rule does not exist.
func (c *Client) ReadNATPortForward(ctx context.Context, machineID string, adapterSlot uint32, name string) (*NATPortForwardRule, error) {
	var result *NATPortForwardRule
	err := c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		// Find the machine
		machineRef, err := findMachine(ctx, api, session, machineID)
		if err != nil {
			return err
		}

		// Get the network adapter
		adapterRef, err := api.GetNetworkAdapter(ctx, machineRef, adapterSlot)
		if err != nil {
			return fmt.Errorf("failed to get network adapter slot %d: %w", adapterSlot, err)
		}

		// Get the NAT engine
		natEngineRef, err := api.GetNATEngine(ctx, adapterRef)
		if err != nil {
			return fmt.Errorf("failed to get NAT engine: %w", err)
		}

		redirects, err := api.GetNATRedirects(ctx, natEngineRef)
		if err != nil {
			return fmt.Errorf("failed to get NAT redirects: %w", err)
		}

		// Find the rule by name
		for _, r := range redirects {
			if r.Name == name {
				result = &NATPortForwardRule{
					MachineID:   machineID,
					AdapterSlot: adapterSlot,
					Name:        r.Name,
					Protocol:    r.Protocol,
					HostIP:      r.HostIP,
					HostPort:    r.HostPort,
					GuestIP:     r.GuestIP,
					GuestPort:   r.GuestPort,
				}
				break
			}
		}
		return nil
	})
	return result, err
}

// DeleteNATPortForward removes a NAT port forwarding rule.
// Returns nil if the rule does not exist (idempotent).
func (c *Client) DeleteNATPortForward(ctx context.Context, machineID string, adapterSlot uint32, name string) error {
	return c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		// Find the machine
		machineRef, err := findMachine(ctx, api, session, machineID)
		if err != nil {
			// If machine doesn't exist, rule is already gone
			if IsNotFound(err) {
				return nil
			}
			return err
		}

		// Get a session object to lock the machine
		sessObj, err := api.GetSessionObject(ctx, session)
		if err != nil {
			return fmt.Errorf("failed to get session object: %w", err)
		}

		// Lock the machine with shared lock (allows modifying settings while VM is running)
		if err := api.LockMachine(ctx, machineRef, sessObj, true); err != nil {
			return fmt.Errorf("failed to lock machine: %w", err)
		}
		defer func() { _ = api.UnlockSession(context.Background(), sessObj) }()

		// Get the mutable machine reference
		mutableMachineRef, err := api.GetMutableMachine(ctx, sessObj)
		if err != nil {
			return fmt.Errorf("failed to get mutable machine: %w", err)
		}

		// Get the network adapter
		adapterRef, err := api.GetNetworkAdapter(ctx, mutableMachineRef, adapterSlot)
		if err != nil {
			return fmt.Errorf("failed to get network adapter slot %d: %w", adapterSlot, err)
		}

		// Get the NAT engine
		natEngineRef, err := api.GetNATEngine(ctx, adapterRef)
		if err != nil {
			return fmt.Errorf("failed to get NAT engine: %w", err)
		}

		// Remove the redirect (ignore error if rule doesn't exist)
		if err := api.RemoveNATRedirect(ctx, natEngineRef, name); err != nil {
			// Best-effort: if the error indicates rule not found, ignore
			errLower := strings.ToLower(err.Error())
			if !strings.Contains(errLower, "not found") && !strings.Contains(errLower, "does not exist") {
				return fmt.Errorf("failed to remove NAT redirect: %w", err)
			}
		}

		// Save settings
		if err := api.SaveSettings(ctx, mutableMachineRef); err != nil {
			return fmt.Errorf("failed to save machine settings: %w", err)
		}

		return nil
	})
}

// AllocateNATHostPort finds an available host port for a new NAT port forwarding rule.
func (c *Client) AllocateNATHostPort(ctx context.Context, opts PortAllocatorOptions) (uint16, error) {
	var port uint16
	err := c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		var err error
		port, err = AllocatePort(ctx, api, session, opts)
		return err
	})
	return port, err
}

// GetAllNATRedirects returns all NAT redirects for a specific machine and adapter slot.
func (c *Client) GetAllNATRedirects(ctx context.Context, machineID string, adapterSlot uint32) ([]vboxapi.NATRedirect, error) {
	var result []vboxapi.NATRedirect
	err := c.withSession(ctx, func(ctx context.Context, api vboxapi.VBoxAPI, session string) error {
		// Find the machine
		machineRef, err := findMachine(ctx, api, session, machineID)
		if err != nil {
			return err
		}

		// Get the network adapter
		adapterRef, err := api.GetNetworkAdapter(ctx, machineRef, adapterSlot)
		if err != nil {
			return fmt.Errorf("failed to get network adapter slot %d: %w", adapterSlot, err)
		}

		// Get the NAT engine
		natEngineRef, err := api.GetNATEngine(ctx, adapterRef)
		if err != nil {
			return fmt.Errorf("failed to get NAT engine: %w", err)
		}

		result, err = api.GetNATRedirects(ctx, natEngineRef)
		if err != nil {
			return fmt.Errorf("failed to get NAT redirects: %w", err)
		}

		return nil
	})
	return result, err
}

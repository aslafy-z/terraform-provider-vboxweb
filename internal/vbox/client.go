package vbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
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
func newAdapter(endpoint string) VBoxAPI {
	// TODO: In the future, could auto-detect version and return appropriate adapter
	return NewAdapter71(endpoint)
}

func (c *Client) withSession(ctx context.Context, fn func(ctx context.Context, api VBoxAPI, session string) error) error {
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

	err = c.withSession(ctx, func(ctx context.Context, api VBoxAPI, session string) error {
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

// GetStateByID returns the current state of a VM by its UUID.
func (c *Client) GetStateByID(ctx context.Context, id string) (string, error) {
	var out string
	err := c.withSession(ctx, func(ctx context.Context, api VBoxAPI, session string) error {
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

	err := c.withSession(ctx, func(ctx context.Context, api VBoxAPI, session string) error {
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

	return c.withSession(ctx, func(ctx context.Context, api VBoxAPI, session string) error {
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

func findMachine(ctx context.Context, api VBoxAPI, session, nameOrID string) (string, error) {
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

func waitProgress(ctx context.Context, api VBoxAPI, progressRef string, timeout time.Duration) error {
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

func convergeState(ctx context.Context, api VBoxAPI, vboxSession string, machineRef, desiredState, sessionType string, timeout time.Duration) (string, error) {
	st, err := api.GetMachineState(ctx, machineRef)
	if err != nil {
		return "", err
	}

	want := strings.ToLower(desiredState)
	if want == "started" {
		if st == MachineStateRunning {
			return st, nil
		}
		if err := ensureRunning(ctx, api, vboxSession, machineRef, sessionType, timeout); err != nil {
			return "", err
		}
	} else if want == "stopped" {
		if st == MachineStatePoweredOff {
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

func ensureRunning(ctx context.Context, api VBoxAPI, vboxSession, machineRef, sessionType string, timeout time.Duration) error {
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

func ensurePoweredOff(ctx context.Context, api VBoxAPI, vboxSession, machineRef string, timeout time.Duration) error {
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

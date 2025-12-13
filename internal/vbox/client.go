package vbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hooklift/gowsdl/soap"
	vb "github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox71"
)

type Client struct {
	endpoint string
	username string
	password string
}

func NewClient(endpoint, username, password string) *Client {
	return &Client{endpoint: endpoint, username: username, password: password}
}

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

func IsNotFound(err error) bool {
	return errors.Is(err, errNotFound)
}

func (c *Client) withSession(ctx context.Context, fn func(ctx context.Context, svc vb.VboxPortType, session string) error) error {
	soapClient := soap.NewClient(c.endpoint)
	svc := vb.NewVboxPortType(soapClient)

	logonResp, err := svc.IWebsessionManager_logonContext(ctx, &vb.IWebsessionManager_logon{Username: c.username, Password: c.password})
	if err != nil {
		return err
	}
	session := logonResp.Returnval

	// Always try to logoff.
	defer func() {
		_, _ = svc.IWebsessionManager_logoffContext(context.Background(), &vb.IWebsessionManager_logoff{RefIVirtualBox: session})
	}()

	return fn(ctx, svc, session)
}

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

	err = c.withSession(ctx, func(ctx context.Context, svc vb.VboxPortType, session string) error {
		srcRef, err := findMachine(ctx, svc, session, req.Source)
		if err != nil {
			return err
		}

		// Get source osTypeId and platform architecture for the new machine
		osTypeId, err := getOSTypeId(ctx, svc, srcRef)
		if err != nil {
			return fmt.Errorf("failed to get source OS type: %w", err)
		}

		platformArch, err := getPlatformArchitecture(ctx, svc, srcRef)
		if err != nil {
			return fmt.Errorf("failed to get source platform architecture: %w", err)
		}

		targetRef, err := createMachine(ctx, svc, session, req.Name, osTypeId, platformArch)
		if err != nil {
			return err
		}

		progressRef, err := cloneTo(ctx, svc, srcRef, targetRef, req.CloneMode, req.CloneOptions)
		if err != nil {
			return err
		}
		if err := waitProgress(ctx, svc, progressRef, req.Timeout); err != nil {
			return err
		}

		if err := registerMachine(ctx, svc, session, targetRef); err != nil {
			return err
		}

		uuid, err = machineUUID(ctx, svc, targetRef)
		if err != nil {
			return err
		}

		// Converge state
		currentState, err = convergeState(ctx, svc, session, targetRef, req.DesiredState, req.SessionType, req.Timeout)
		if err != nil {
			return err
		}
		return nil
	})

	return uuid, currentState, err
}

func (c *Client) GetStateByID(ctx context.Context, id string) (string, error) {
	var out string
	err := c.withSession(ctx, func(ctx context.Context, svc vb.VboxPortType, session string) error {
		mRef, err := findMachine(ctx, svc, session, id)
		if err != nil {
			return err
		}
		st, err := machineState(ctx, svc, mRef)
		if err != nil {
			return err
		}
		out = string(st)
		return nil
	})
	return out, err
}

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

	err := c.withSession(ctx, func(ctx context.Context, svc vb.VboxPortType, session string) error {
		mRef, err := findMachine(ctx, svc, session, id)
		if err != nil {
			return err
		}
		out, err = convergeState(ctx, svc, session, mRef, desiredState, sessionType, timeout)
		return err
	})
	return out, err
}

func (c *Client) DeleteByID(ctx context.Context, id string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 20 * time.Minute
	}

	return c.withSession(ctx, func(ctx context.Context, svc vb.VboxPortType, session string) error {
		mRef, err := findMachine(ctx, svc, session, id)
		if err != nil {
			return err
		}

		// Ensure powered off (best-effort).
		_ = ensurePoweredOff(ctx, svc, session, mRef, timeout)

		cm := vb.CleanupModeFull
		unreg, err := svc.IMachine_unregisterContext(ctx, &vb.IMachine_unregister{This: mRef, CleanupMode: &cm})
		if err != nil {
			return err
		}

		prog, err := svc.IMachine_deleteConfigContext(ctx, &vb.IMachine_deleteConfig{This: mRef, Media: unreg.Returnval})
		if err != nil {
			return err
		}
		if err := waitProgress(ctx, svc, prog.Returnval, timeout); err != nil {
			return err
		}

		return nil
	})
}

// ---- helpers ----

func findMachine(ctx context.Context, svc vb.VboxPortType, session, nameOrID string) (string, error) {
	resp, err := svc.IVirtualBox_findMachineContext(ctx, &vb.IVirtualBox_findMachine{This: session, NameOrId: nameOrID})
	if err != nil {
		// Best-effort mapping to not found.
		if strings.Contains(strings.ToLower(err.Error()), "could not find") || strings.Contains(strings.ToLower(err.Error()), "object not found") {
			return "", fmt.Errorf("%w: machine %s", errNotFound, nameOrID)
		}
		return "", err
	}
	if strings.TrimSpace(resp.Returnval) == "" {
		return "", fmt.Errorf("%w: machine %s", errNotFound, nameOrID)
	}
	return resp.Returnval, nil
}

func getOSTypeId(ctx context.Context, svc vb.VboxPortType, machineRef string) (string, error) {
	resp, err := svc.IMachine_getOSTypeIdContext(ctx, &vb.IMachine_getOSTypeId{This: machineRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func getPlatformArchitecture(ctx context.Context, svc vb.VboxPortType, machineRef string) (vb.PlatformArchitecture, error) {
	// Get the IPlatform interface from the machine
	platformResp, err := svc.IMachine_getPlatformContext(ctx, &vb.IMachine_getPlatform{This: machineRef})
	if err != nil {
		return vb.PlatformArchitectureX86, err // default to x86 if can't get
	}

	// Get architecture from the platform
	archResp, err := svc.IPlatform_getArchitectureContext(ctx, &vb.IPlatform_getArchitecture{This: platformResp.Returnval})
	if err != nil {
		return vb.PlatformArchitectureX86, err // default to x86 if can't get
	}

	if archResp.Returnval == nil {
		return vb.PlatformArchitectureX86, nil
	}
	return *archResp.Returnval, nil
}

func createMachine(ctx context.Context, svc vb.VboxPortType, session, name, osTypeId string, platform vb.PlatformArchitecture) (string, error) {
	resp, err := svc.IVirtualBox_createMachineContext(ctx, &vb.IVirtualBox_createMachine{
		This:     session,
		Name:     name,
		Platform: &platform,
		OsTypeId: osTypeId,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func registerMachine(ctx context.Context, svc vb.VboxPortType, session, machineRef string) error {
	_, err := svc.IVirtualBox_registerMachineContext(ctx, &vb.IVirtualBox_registerMachine{This: session, Machine: machineRef})
	return err
}

func cloneTo(ctx context.Context, svc vb.VboxPortType, srcRef, targetRef, mode string, options []string) (string, error) {
	m := vb.CloneMode(mode)

	var optPtrs []*vb.CloneOptions
	for _, o := range options {
		oo := vb.CloneOptions(o)
		optPtrs = append(optPtrs, &oo)
	}

	resp, err := svc.IMachine_cloneToContext(ctx, &vb.IMachine_cloneTo{This: srcRef, Target: targetRef, Mode: &m, Options: optPtrs})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func waitProgress(ctx context.Context, svc vb.VboxPortType, progressRef string, timeout time.Duration) error {
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
		completedResp, err := svc.IProgress_getCompletedContext(ctx, &vb.IProgress_getCompleted{This: progressRef})
		if err != nil {
			return fmt.Errorf("failed to get progress completion status: %w", err)
		}

		if completedResp.Returnval {
			// Operation completed, check result
			rc, err := svc.IProgress_getResultCodeContext(ctx, &vb.IProgress_getResultCode{This: progressRef})
			if err != nil {
				return fmt.Errorf("failed to get progress result code: %w", err)
			}
			if rc.Returnval != 0 {
				// Try to fetch an error message.
				ei, eiErr := svc.IProgress_getErrorInfoContext(ctx, &vb.IProgress_getErrorInfo{This: progressRef})
				if eiErr == nil && strings.TrimSpace(ei.Returnval) != "" {
					txt, txtErr := svc.IVirtualBoxErrorInfo_getTextContext(ctx, &vb.IVirtualBoxErrorInfo_getText{This: ei.Returnval})
					if txtErr == nil && strings.TrimSpace(txt.Returnval) != "" {
						return fmt.Errorf("progress failed (resultCode=%d): %s", rc.Returnval, txt.Returnval)
					}
				}
				return fmt.Errorf("progress failed (resultCode=%d)", rc.Returnval)
			}
			return nil
		}

		// Not completed yet, wait and poll again
		time.Sleep(pollInterval)
	}
}

func machineUUID(ctx context.Context, svc vb.VboxPortType, machineRef string) (string, error) {
	resp, err := svc.IMachine_getIdContext(ctx, &vb.IMachine_getId{This: machineRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func machineState(ctx context.Context, svc vb.VboxPortType, machineRef string) (vb.MachineState, error) {
	resp, err := svc.IMachine_getStateContext(ctx, &vb.IMachine_getState{This: machineRef})
	if err != nil {
		return vb.MachineStateNull, err
	}
	if resp.Returnval == nil {
		return vb.MachineStateNull, nil
	}
	return *resp.Returnval, nil
}

func convergeState(ctx context.Context, svc vb.VboxPortType, vboxSession string, machineRef, desiredState, sessionType string, timeout time.Duration) (string, error) {
	st, err := machineState(ctx, svc, machineRef)
	if err != nil {
		return "", err
	}

	want := strings.ToLower(desiredState)
	if want == "started" {
		if st == vb.MachineStateRunning {
			return string(st), nil
		}
		if err := ensureRunning(ctx, svc, vboxSession, machineRef, sessionType, timeout); err != nil {
			return "", err
		}
	} else if want == "stopped" {
		if st == vb.MachineStatePoweredOff {
			return string(st), nil
		}
		if err := ensurePoweredOff(ctx, svc, vboxSession, machineRef, timeout); err != nil {
			return "", err
		}
	} else {
		return "", fmt.Errorf("invalid desired state: %s", desiredState)
	}

	st, err = machineState(ctx, svc, machineRef)
	if err != nil {
		return "", err
	}
	return string(st), nil
}

func ensureRunning(ctx context.Context, svc vb.VboxPortType, vboxSession, machineRef, sessionType string, timeout time.Duration) error {
	sessObj, err := svc.IWebsessionManager_getSessionObjectContext(ctx, &vb.IWebsessionManager_getSessionObject{RefIVirtualBox: vboxSession})
	if err != nil {
		return err
	}

	// launchVMProcess returns IProgress
	prog, err := svc.IMachine_launchVMProcessContext(ctx, &vb.IMachine_launchVMProcess{This: machineRef, Session: sessObj.Returnval, Name: sessionType})
	if err != nil {
		return err
	}

	if err := waitProgress(ctx, svc, prog.Returnval, timeout); err != nil {
		return err
	}

	// Always unlock.
	_, _ = svc.ISession_unlockMachineContext(context.Background(), &vb.ISession_unlockMachine{This: sessObj.Returnval})
	return nil
}

func ensurePoweredOff(ctx context.Context, svc vb.VboxPortType, vboxSession, machineRef string, timeout time.Duration) error {
	sessObj, err := svc.IWebsessionManager_getSessionObjectContext(ctx, &vb.IWebsessionManager_getSessionObject{RefIVirtualBox: vboxSession})
	if err != nil {
		return err
	}
	lock := vb.LockTypeShared

	_, err = svc.IMachine_lockMachineContext(ctx, &vb.IMachine_lockMachine{This: machineRef, Session: sessObj.Returnval, LockType: &lock})
	if err != nil {
		// If already powered off or not lockable, bubble up.
		return err
	}

	console, err := svc.ISession_getConsoleContext(ctx, &vb.ISession_getConsole{This: sessObj.Returnval})
	if err != nil {
		return err
	}

	prog, err := svc.IConsole_powerDownContext(ctx, &vb.IConsole_powerDown{This: console.Returnval})
	if err != nil {
		return err
	}

	if err := waitProgress(ctx, svc, prog.Returnval, timeout); err != nil {
		return err
	}

	_, _ = svc.ISession_unlockMachineContext(context.Background(), &vb.ISession_unlockMachine{This: sessObj.Returnval})
	return nil
}

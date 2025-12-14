package vbox

import (
	"context"
	"strings"

	vb "github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox71"
	"github.com/hooklift/gowsdl/soap"
)

// Adapter71 implements VBoxAPI for VirtualBox 7.1.
type Adapter71 struct {
	svc vb.VboxPortType
}

// NewAdapter71 creates a new adapter for VirtualBox 7.1.
func NewAdapter71(endpoint string) *Adapter71 {
	soapClient := soap.NewClient(endpoint)
	return &Adapter71{svc: vb.NewVboxPortType(soapClient)}
}

func (a *Adapter71) Logon(ctx context.Context, username, password string) (string, error) {
	resp, err := a.svc.IWebsessionManager_logonContext(ctx, &vb.IWebsessionManager_logon{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) Logoff(ctx context.Context, session string) error {
	_, err := a.svc.IWebsessionManager_logoffContext(ctx, &vb.IWebsessionManager_logoff{
		RefIVirtualBox: session,
	})
	return err
}

func (a *Adapter71) GetSessionObject(ctx context.Context, session string) (string, error) {
	resp, err := a.svc.IWebsessionManager_getSessionObjectContext(ctx, &vb.IWebsessionManager_getSessionObject{
		RefIVirtualBox: session,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) FindMachine(ctx context.Context, session, nameOrID string) (string, error) {
	resp, err := a.svc.IVirtualBox_findMachineContext(ctx, &vb.IVirtualBox_findMachine{
		This:     session,
		NameOrId: nameOrID,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) CreateMachine(ctx context.Context, session, name, osTypeId, sourceMachineRef string) (string, error) {
	// VBox 7.1 requires platform architecture
	platformArch, err := a.getPlatformArchitecture(ctx, sourceMachineRef)
	if err != nil {
		// Default to x86 if we can't determine
		arch := vb.PlatformArchitectureX86
		platformArch = &arch
	}

	resp, err := a.svc.IVirtualBox_createMachineContext(ctx, &vb.IVirtualBox_createMachine{
		This:     session,
		Name:     name,
		Platform: platformArch,
		OsTypeId: osTypeId,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) getPlatformArchitecture(ctx context.Context, machineRef string) (*vb.PlatformArchitecture, error) {
	platformResp, err := a.svc.IMachine_getPlatformContext(ctx, &vb.IMachine_getPlatform{This: machineRef})
	if err != nil {
		return nil, err
	}

	archResp, err := a.svc.IPlatform_getArchitectureContext(ctx, &vb.IPlatform_getArchitecture{This: platformResp.Returnval})
	if err != nil {
		return nil, err
	}

	return archResp.Returnval, nil
}

func (a *Adapter71) RegisterMachine(ctx context.Context, session, machineRef string) error {
	_, err := a.svc.IVirtualBox_registerMachineContext(ctx, &vb.IVirtualBox_registerMachine{
		This:    session,
		Machine: machineRef,
	})
	return err
}

func (a *Adapter71) UnregisterMachine(ctx context.Context, machineRef string) ([]string, error) {
	cm := vb.CleanupModeFull
	resp, err := a.svc.IMachine_unregisterContext(ctx, &vb.IMachine_unregister{
		This:        machineRef,
		CleanupMode: &cm,
	})
	if err != nil {
		return nil, err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) DeleteConfig(ctx context.Context, machineRef string, mediaRefs []string) (string, error) {
	resp, err := a.svc.IMachine_deleteConfigContext(ctx, &vb.IMachine_deleteConfig{
		This:  machineRef,
		Media: mediaRefs,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) GetMachineId(ctx context.Context, machineRef string) (string, error) {
	resp, err := a.svc.IMachine_getIdContext(ctx, &vb.IMachine_getId{This: machineRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) GetMachineState(ctx context.Context, machineRef string) (string, error) {
	resp, err := a.svc.IMachine_getStateContext(ctx, &vb.IMachine_getState{This: machineRef})
	if err != nil {
		return "", err
	}
	if resp.Returnval == nil {
		return MachineStateNull, nil
	}
	return string(*resp.Returnval), nil
}

func (a *Adapter71) GetOSTypeId(ctx context.Context, machineRef string) (string, error) {
	resp, err := a.svc.IMachine_getOSTypeIdContext(ctx, &vb.IMachine_getOSTypeId{This: machineRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) CloneTo(ctx context.Context, srcMachineRef, targetMachineRef, mode string, options []string) (string, error) {
	m := vb.CloneMode(mode)

	var optPtrs []*vb.CloneOptions
	for _, o := range options {
		oo := vb.CloneOptions(o)
		optPtrs = append(optPtrs, &oo)
	}

	resp, err := a.svc.IMachine_cloneToContext(ctx, &vb.IMachine_cloneTo{
		This:    srcMachineRef,
		Target:  targetMachineRef,
		Mode:    &m,
		Options: optPtrs,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) LaunchVMProcess(ctx context.Context, machineRef, sessionObj, sessionType string) (string, error) {
	resp, err := a.svc.IMachine_launchVMProcessContext(ctx, &vb.IMachine_launchVMProcess{
		This:    machineRef,
		Session: sessionObj,
		Name:    sessionType,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) LockMachine(ctx context.Context, machineRef, sessionObj string, shared bool) error {
	lockType := vb.LockTypeWrite
	if shared {
		lockType = vb.LockTypeShared
	}
	_, err := a.svc.IMachine_lockMachineContext(ctx, &vb.IMachine_lockMachine{
		This:     machineRef,
		Session:  sessionObj,
		LockType: &lockType,
	})
	return err
}

func (a *Adapter71) UnlockSession(ctx context.Context, sessionObj string) error {
	_, err := a.svc.ISession_unlockMachineContext(ctx, &vb.ISession_unlockMachine{This: sessionObj})
	return err
}

func (a *Adapter71) GetConsole(ctx context.Context, sessionObj string) (string, error) {
	resp, err := a.svc.ISession_getConsoleContext(ctx, &vb.ISession_getConsole{This: sessionObj})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) PowerDown(ctx context.Context, consoleRef string) (string, error) {
	resp, err := a.svc.IConsole_powerDownContext(ctx, &vb.IConsole_powerDown{This: consoleRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) GetProgressCompleted(ctx context.Context, progressRef string) (bool, error) {
	resp, err := a.svc.IProgress_getCompletedContext(ctx, &vb.IProgress_getCompleted{This: progressRef})
	if err != nil {
		return false, err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) GetProgressResultCode(ctx context.Context, progressRef string) (int32, error) {
	resp, err := a.svc.IProgress_getResultCodeContext(ctx, &vb.IProgress_getResultCode{This: progressRef})
	if err != nil {
		return -1, err
	}
	return resp.Returnval, nil
}

func (a *Adapter71) GetProgressErrorText(ctx context.Context, progressRef string) (string, error) {
	ei, err := a.svc.IProgress_getErrorInfoContext(ctx, &vb.IProgress_getErrorInfo{This: progressRef})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(ei.Returnval) == "" {
		return "", nil
	}

	txt, err := a.svc.IVirtualBoxErrorInfo_getTextContext(ctx, &vb.IVirtualBoxErrorInfo_getText{This: ei.Returnval})
	if err != nil {
		return "", err
	}
	return txt.Returnval, nil
}

func (a *Adapter71) GetAPIVersion(ctx context.Context, session string) (string, error) {
	resp, err := a.svc.IVirtualBox_getAPIVersionContext(ctx, &vb.IVirtualBox_getAPIVersion{This: session})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

// Package vbox71 provides VirtualBox 7.1 specific implementation of the VBoxAPI interface.
package vbox71

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox71/generated"
	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxapi"
	"github.com/hooklift/gowsdl/soap"
)

// Adapter implements vboxapi.VBoxAPI for VirtualBox 7.1.
type Adapter struct {
	svc generated.VboxPortType
}

// NewAdapter creates a new adapter for VirtualBox 7.1.
func NewAdapter(endpoint string) *Adapter {
	soapClient := soap.NewClient(endpoint)
	return &Adapter{svc: generated.NewVboxPortType(soapClient)}
}

func (a *Adapter) Logon(ctx context.Context, username, password string) (string, error) {
	resp, err := a.svc.IWebsessionManager_logonContext(ctx, &generated.IWebsessionManager_logon{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) Logoff(ctx context.Context, session string) error {
	_, err := a.svc.IWebsessionManager_logoffContext(ctx, &generated.IWebsessionManager_logoff{
		RefIVirtualBox: session,
	})
	return err
}

func (a *Adapter) GetSessionObject(ctx context.Context, session string) (string, error) {
	resp, err := a.svc.IWebsessionManager_getSessionObjectContext(ctx, &generated.IWebsessionManager_getSessionObject{
		RefIVirtualBox: session,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) FindMachine(ctx context.Context, session, nameOrID string) (string, error) {
	resp, err := a.svc.IVirtualBox_findMachineContext(ctx, &generated.IVirtualBox_findMachine{
		This:     session,
		NameOrId: nameOrID,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) CreateMachine(ctx context.Context, session, name, osTypeId, sourceMachineRef string) (string, error) {
	// VBox 7.1 requires platform architecture
	platformArch, err := a.getPlatformArchitecture(ctx, sourceMachineRef)
	if err != nil {
		// Default to x86 if we can't determine
		arch := generated.PlatformArchitectureX86
		platformArch = &arch
	}

	resp, err := a.svc.IVirtualBox_createMachineContext(ctx, &generated.IVirtualBox_createMachine{
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

func (a *Adapter) getPlatformArchitecture(ctx context.Context, machineRef string) (*generated.PlatformArchitecture, error) {
	platformResp, err := a.svc.IMachine_getPlatformContext(ctx, &generated.IMachine_getPlatform{This: machineRef})
	if err != nil {
		return nil, err
	}

	archResp, err := a.svc.IPlatform_getArchitectureContext(ctx, &generated.IPlatform_getArchitecture{This: platformResp.Returnval})
	if err != nil {
		return nil, err
	}

	return archResp.Returnval, nil
}

func (a *Adapter) RegisterMachine(ctx context.Context, session, machineRef string) error {
	_, err := a.svc.IVirtualBox_registerMachineContext(ctx, &generated.IVirtualBox_registerMachine{
		This:    session,
		Machine: machineRef,
	})
	return err
}

func (a *Adapter) UnregisterMachine(ctx context.Context, machineRef string) ([]string, error) {
	cm := generated.CleanupModeFull
	resp, err := a.svc.IMachine_unregisterContext(ctx, &generated.IMachine_unregister{
		This:        machineRef,
		CleanupMode: &cm,
	})
	if err != nil {
		return nil, err
	}
	return resp.Returnval, nil
}

func (a *Adapter) DeleteConfig(ctx context.Context, machineRef string, mediaRefs []string) (string, error) {
	resp, err := a.svc.IMachine_deleteConfigContext(ctx, &generated.IMachine_deleteConfig{
		This:  machineRef,
		Media: mediaRefs,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetMachineId(ctx context.Context, machineRef string) (string, error) {
	resp, err := a.svc.IMachine_getIdContext(ctx, &generated.IMachine_getId{This: machineRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetMachineName(ctx context.Context, machineRef string) (string, error) {
	resp, err := a.svc.IMachine_getNameContext(ctx, &generated.IMachine_getName{This: machineRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetMachineState(ctx context.Context, machineRef string) (string, error) {
	resp, err := a.svc.IMachine_getStateContext(ctx, &generated.IMachine_getState{This: machineRef})
	if err != nil {
		return "", err
	}
	if resp.Returnval == nil {
		return vboxapi.MachineStateNull, nil
	}
	return string(*resp.Returnval), nil
}

func (a *Adapter) GetOSTypeId(ctx context.Context, machineRef string) (string, error) {
	resp, err := a.svc.IMachine_getOSTypeIdContext(ctx, &generated.IMachine_getOSTypeId{This: machineRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) CloneTo(ctx context.Context, srcMachineRef, targetMachineRef, mode string, options []string) (string, error) {
	m := generated.CloneMode(mode)

	var optPtrs []*generated.CloneOptions
	for _, o := range options {
		oo := generated.CloneOptions(o)
		optPtrs = append(optPtrs, &oo)
	}

	resp, err := a.svc.IMachine_cloneToContext(ctx, &generated.IMachine_cloneTo{
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

func (a *Adapter) LaunchVMProcess(ctx context.Context, machineRef, sessionObj, sessionType string) (string, error) {
	resp, err := a.svc.IMachine_launchVMProcessContext(ctx, &generated.IMachine_launchVMProcess{
		This:    machineRef,
		Session: sessionObj,
		Name:    sessionType,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) LockMachine(ctx context.Context, machineRef, sessionObj string, shared bool) error {
	lockType := generated.LockTypeWrite
	if shared {
		lockType = generated.LockTypeShared
	}
	_, err := a.svc.IMachine_lockMachineContext(ctx, &generated.IMachine_lockMachine{
		This:     machineRef,
		Session:  sessionObj,
		LockType: &lockType,
	})
	return err
}

func (a *Adapter) UnlockSession(ctx context.Context, sessionObj string) error {
	_, err := a.svc.ISession_unlockMachineContext(ctx, &generated.ISession_unlockMachine{This: sessionObj})
	return err
}

func (a *Adapter) GetConsole(ctx context.Context, sessionObj string) (string, error) {
	resp, err := a.svc.ISession_getConsoleContext(ctx, &generated.ISession_getConsole{This: sessionObj})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) PowerDown(ctx context.Context, consoleRef string) (string, error) {
	resp, err := a.svc.IConsole_powerDownContext(ctx, &generated.IConsole_powerDown{This: consoleRef})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetProgressCompleted(ctx context.Context, progressRef string) (bool, error) {
	resp, err := a.svc.IProgress_getCompletedContext(ctx, &generated.IProgress_getCompleted{This: progressRef})
	if err != nil {
		return false, err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetProgressResultCode(ctx context.Context, progressRef string) (int32, error) {
	resp, err := a.svc.IProgress_getResultCodeContext(ctx, &generated.IProgress_getResultCode{This: progressRef})
	if err != nil {
		return -1, err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetProgressErrorText(ctx context.Context, progressRef string) (string, error) {
	ei, err := a.svc.IProgress_getErrorInfoContext(ctx, &generated.IProgress_getErrorInfo{This: progressRef})
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(ei.Returnval) == "" {
		return "", nil
	}

	txt, err := a.svc.IVirtualBoxErrorInfo_getTextContext(ctx, &generated.IVirtualBoxErrorInfo_getText{This: ei.Returnval})
	if err != nil {
		return "", err
	}
	return txt.Returnval, nil
}

func (a *Adapter) GetAPIVersion(ctx context.Context, session string) (string, error) {
	resp, err := a.svc.IVirtualBox_getAPIVersionContext(ctx, &generated.IVirtualBox_getAPIVersion{This: session})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetMachines(ctx context.Context, session string) ([]string, error) {
	resp, err := a.svc.IVirtualBox_getMachinesContext(ctx, &generated.IVirtualBox_getMachines{This: session})
	if err != nil {
		return nil, err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetNetworkAdapter(ctx context.Context, machineRef string, slot uint32) (string, error) {
	resp, err := a.svc.IMachine_getNetworkAdapterContext(ctx, &generated.IMachine_getNetworkAdapter{
		This: machineRef,
		Slot: slot,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetNATEngine(ctx context.Context, adapterRef string) (string, error) {
	resp, err := a.svc.INetworkAdapter_getNATEngineContext(ctx, &generated.INetworkAdapter_getNATEngine{
		This: adapterRef,
	})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetNATRedirects(ctx context.Context, natEngineRef string) ([]vboxapi.NATRedirect, error) {
	resp, err := a.svc.INATEngine_getRedirectsContext(ctx, &generated.INATEngine_getRedirects{
		This: natEngineRef,
	})
	if err != nil {
		return nil, err
	}

	// VBox 7.1 format: "name,proto,hostIP,hostPort,guestIP,guestPort"
	// proto: 0=UDP, 1=TCP
	var redirects []vboxapi.NATRedirect
	for _, raw := range resp.Returnval {
		r, err := parseNATRedirect71(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse NAT redirect %q: %w", raw, err)
		}
		redirects = append(redirects, r)
	}
	return redirects, nil
}

// parseNATRedirect71 parses VBox 7.1 NAT redirect format.
// Format: "name,proto,hostIP,hostPort,guestIP,guestPort"
// proto: 0=UDP, 1=TCP
func parseNATRedirect71(raw string) (vboxapi.NATRedirect, error) {
	parts := strings.Split(raw, ",")
	if len(parts) != 6 {
		return vboxapi.NATRedirect{}, fmt.Errorf("expected 6 comma-separated fields, got %d", len(parts))
	}

	protoNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return vboxapi.NATRedirect{}, fmt.Errorf("invalid protocol value %q: %w", parts[1], err)
	}

	var proto vboxapi.NATProtocol
	switch protoNum {
	case 0:
		proto = vboxapi.NATProtocolUDP
	case 1:
		proto = vboxapi.NATProtocolTCP
	default:
		return vboxapi.NATRedirect{}, fmt.Errorf("unknown protocol number %d", protoNum)
	}

	hostPort, err := strconv.ParseUint(parts[3], 10, 16)
	if err != nil {
		return vboxapi.NATRedirect{}, fmt.Errorf("invalid host port %q: %w", parts[3], err)
	}

	guestPort, err := strconv.ParseUint(parts[5], 10, 16)
	if err != nil {
		return vboxapi.NATRedirect{}, fmt.Errorf("invalid guest port %q: %w", parts[5], err)
	}

	return vboxapi.NATRedirect{
		Name:      parts[0],
		Protocol:  proto,
		HostIP:    parts[2],
		HostPort:  uint16(hostPort),
		GuestIP:   parts[4],
		GuestPort: uint16(guestPort),
	}, nil
}

func (a *Adapter) AddNATRedirect(ctx context.Context, natEngineRef, name string, proto vboxapi.NATProtocol, hostIP string, hostPort uint16, guestIP string, guestPort uint16) error {
	var vbProto *generated.NATProtocol
	if proto == vboxapi.NATProtocolTCP {
		p := generated.NATProtocolTCP
		vbProto = &p
	} else {
		p := generated.NATProtocolUDP
		vbProto = &p
	}

	_, err := a.svc.INATEngine_addRedirectContext(ctx, &generated.INATEngine_addRedirect{
		This:      natEngineRef,
		Name:      name,
		Proto:     vbProto,
		HostIP:    hostIP,
		HostPort:  hostPort,
		GuestIP:   guestIP,
		GuestPort: guestPort,
	})
	return err
}

func (a *Adapter) RemoveNATRedirect(ctx context.Context, natEngineRef, name string) error {
	_, err := a.svc.INATEngine_removeRedirectContext(ctx, &generated.INATEngine_removeRedirect{
		This: natEngineRef,
		Name: name,
	})
	return err
}

func (a *Adapter) GetNATNetworks(ctx context.Context, session string) ([]string, error) {
	resp, err := a.svc.IVirtualBox_getNATNetworksContext(ctx, &generated.IVirtualBox_getNATNetworks{This: session})
	if err != nil {
		return nil, err
	}
	return resp.Returnval, nil
}

func (a *Adapter) GetNATNetworkPortForwardRules4(ctx context.Context, natNetworkRef string) ([]vboxapi.NATRedirect, error) {
	resp, err := a.svc.INATNetwork_getPortForwardRules4Context(ctx, &generated.INATNetwork_getPortForwardRules4{This: natNetworkRef})
	if err != nil {
		return nil, err
	}

	// VBox 7.1 NAT Network format: "name:proto:hostIP:hostPort:guestIP:guestPort"
	// proto: tcp or udp (lowercase string)
	var redirects []vboxapi.NATRedirect
	for _, raw := range resp.Returnval {
		r, err := parseNATNetworkRule71(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse NAT network rule %q: %w", raw, err)
		}
		redirects = append(redirects, r)
	}
	return redirects, nil
}

// parseNATNetworkRule71 parses VBox 7.1 NAT Network port forward format.
// Format: "name:proto:hostIP:hostPort:guestIP:guestPort"
// proto: "tcp" or "udp"
func parseNATNetworkRule71(raw string) (vboxapi.NATRedirect, error) {
	parts := strings.Split(raw, ":")
	if len(parts) != 6 {
		return vboxapi.NATRedirect{}, fmt.Errorf("expected 6 colon-separated fields, got %d", len(parts))
	}

	var proto vboxapi.NATProtocol
	switch strings.ToLower(parts[1]) {
	case "tcp":
		proto = vboxapi.NATProtocolTCP
	case "udp":
		proto = vboxapi.NATProtocolUDP
	default:
		return vboxapi.NATRedirect{}, fmt.Errorf("unknown protocol %q", parts[1])
	}

	hostPort, err := strconv.ParseUint(parts[3], 10, 16)
	if err != nil {
		return vboxapi.NATRedirect{}, fmt.Errorf("invalid host port %q: %w", parts[3], err)
	}

	guestPort, err := strconv.ParseUint(parts[5], 10, 16)
	if err != nil {
		return vboxapi.NATRedirect{}, fmt.Errorf("invalid guest port %q: %w", parts[5], err)
	}

	return vboxapi.NATRedirect{
		Name:      parts[0],
		Protocol:  proto,
		HostIP:    parts[2],
		HostPort:  uint16(hostPort),
		GuestIP:   parts[4],
		GuestPort: uint16(guestPort),
	}, nil
}

func (a *Adapter) GetMutableMachine(ctx context.Context, sessionObj string) (string, error) {
	resp, err := a.svc.ISession_getMachineContext(ctx, &generated.ISession_getMachine{This: sessionObj})
	if err != nil {
		return "", err
	}
	return resp.Returnval, nil
}

func (a *Adapter) SaveSettings(ctx context.Context, machineRef string) error {
	_, err := a.svc.IMachine_saveSettingsContext(ctx, &generated.IMachine_saveSettings{This: machineRef})
	return err
}

// Compile-time check that Adapter implements vboxapi.VBoxAPI
var _ vboxapi.VBoxAPI = (*Adapter)(nil)

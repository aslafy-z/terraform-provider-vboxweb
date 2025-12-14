package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox"
	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxapi"
)

type natPortForwardResource struct {
	client *vbox.Client
}

type natPortForwardModel struct {
	// Identity fields
	MachineID   types.String `tfsdk:"machine_id"`
	AdapterSlot types.Int64  `tfsdk:"adapter_slot"`
	Name        types.String `tfsdk:"name"`

	// Rule configuration
	Protocol  types.String `tfsdk:"protocol"`
	HostIP    types.String `tfsdk:"host_ip"`
	HostPort  types.Int64  `tfsdk:"host_port"`
	GuestIP   types.String `tfsdk:"guest_ip"`
	GuestPort types.Int64  `tfsdk:"guest_port"`

	// Auto host port configuration
	AutoHostPort    types.Bool   `tfsdk:"auto_host_port"`
	AutoHostPortMin types.Int64  `tfsdk:"auto_host_port_min"`
	AutoHostPortMax types.Int64  `tfsdk:"auto_host_port_max"`
	AutoHostIPScope types.String `tfsdk:"auto_host_ip_scope"`

	// Computed
	EffectiveHostPort types.Int64  `tfsdk:"effective_host_port"`
	ID                types.String `tfsdk:"id"`
}

func NewNatPortForwardResource() resource.Resource {
	return &natPortForwardResource{}
}

func (r *natPortForwardResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_port_forward"
}

func (r *natPortForwardResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*vbox.Client)
}

func (r *natPortForwardResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Manages a NAT port forwarding rule on a VirtualBox VM network adapter.

This resource creates a single NAT port forwarding rule on a VM's NAT-attached network adapter.
It supports an optional "auto host port" mode that automatically selects an available host port
from a configured range, avoiding conflicts with other VirtualBox NAT port forwarding rules.

**Important guarantees and limitations:**
- When using auto_host_port, the selected port is guaranteed not to conflict with any other
  VirtualBox NAT port forwarding rule on the same VirtualBox instance at apply time.
- This does NOT guarantee the port is not used by other (non-VirtualBox) processes on the host.
- VirtualBox may not surface runtime bind failures if the port is already in use.
- Changes to any rule attribute (except auto_host_port settings) will trigger rule replacement.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for this resource (machine_id:adapter_slot:name).",
			},
			"machine_id": schema.StringAttribute{
				Required:    true,
				Description: "VirtualBox machine ID (UUID) that owns the NAT adapter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"adapter_slot": schema.Int64Attribute{
				Required:    true,
				Description: "Network adapter slot number (0-7, corresponding to nic1-nic8).",
				Validators: []validator.Int64{
					int64validator.Between(0, 7),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the NAT port forwarding rule. Must be unique within the adapter's NAT engine.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "Protocol for the port forwarding rule: 'tcp' or 'udp'.",
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("tcp", "udp"),
				},
			},
			"host_ip": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Host IP address to bind to. Empty string or '0.0.0.0' means all interfaces.",
			},
			"host_port": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Host port number. If omitted or 0 and auto_host_port is true, a port will be automatically selected.",
				Validators: []validator.Int64{
					int64validator.Between(0, 65535),
				},
			},
			"guest_ip": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Guest IP address. Empty string is typically fine for most use cases.",
			},
			"guest_port": schema.Int64Attribute{
				Required:    true,
				Description: "Guest port number (1-65535).",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"auto_host_port": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "If true and host_port is not set (or is 0), automatically select an available host port.",
			},
			"auto_host_port_min": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(20000),
				Description: "Minimum port for auto-selection range (inclusive). Default: 20000.",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"auto_host_port_max": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(40000),
				Description: "Maximum port for auto-selection range (inclusive). Default: 40000.",
				Validators: []validator.Int64{
					int64validator.Between(1, 65535),
				},
			},
			"auto_host_ip_scope": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("any"),
				Description: "How to handle host IP when checking for port conflicts: 'any' (all bindings conflict) or 'exact' (only same host_ip conflicts). Default: 'any'.",
				Validators: []validator.String{
					stringvalidator.OneOf("any", "exact"),
				},
			},
			"effective_host_port": schema.Int64Attribute{
				Computed:    true,
				Description: "The actual host port in use. This equals host_port when explicitly set, or the auto-selected port when using auto_host_port.",
			},
		},
	}
}

func (r *natPortForwardResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan natPortForwardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the host port to use
	hostPort := uint16(plan.HostPort.ValueInt64())

	// If auto_host_port is enabled and host_port is not set (or is 0), allocate a port
	if plan.AutoHostPort.ValueBool() && hostPort == 0 {
		opts := vbox.PortAllocatorOptions{
			MinPort:            uint16(plan.AutoHostPortMin.ValueInt64()),
			MaxPort:            uint16(plan.AutoHostPortMax.ValueInt64()),
			HostIP:             plan.HostIP.ValueString(),
			Scope:              vbox.HostIPScope(plan.AutoHostIPScope.ValueString()),
			IncludeNATNetworks: true,
		}

		allocatedPort, err := r.client.AllocateNATHostPort(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError("Failed to allocate host port", err.Error())
			return
		}
		hostPort = allocatedPort
	}

	// Validate that we have a valid host port
	if hostPort == 0 {
		resp.Diagnostics.AddError(
			"Invalid host port",
			"host_port must be specified or auto_host_port must be enabled to automatically select a port",
		)
		return
	}

	// Parse protocol
	proto := vboxapi.NATProtocolTCP
	if strings.EqualFold(plan.Protocol.ValueString(), "udp") {
		proto = vboxapi.NATProtocolUDP
	}

	// Create the rule
	rule := vbox.NATPortForwardRule{
		MachineID:   plan.MachineID.ValueString(),
		AdapterSlot: uint32(plan.AdapterSlot.ValueInt64()),
		Name:        plan.Name.ValueString(),
		Protocol:    proto,
		HostIP:      plan.HostIP.ValueString(),
		HostPort:    hostPort,
		GuestIP:     plan.GuestIP.ValueString(),
		GuestPort:   uint16(plan.GuestPort.ValueInt64()),
	}

	if err := r.client.CreateNATPortForward(ctx, rule); err != nil {
		resp.Diagnostics.AddError("Failed to create NAT port forward rule", err.Error())
		return
	}

	// Read back to confirm
	readRule, err := r.client.ReadNATPortForward(ctx, rule.MachineID, rule.AdapterSlot, rule.Name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to verify NAT port forward rule", err.Error())
		return
	}
	if readRule == nil {
		resp.Diagnostics.AddError("NAT port forward rule not found after creation", "The rule was created but could not be read back")
		return
	}

	// Update state
	plan.ID = types.StringValue(fmt.Sprintf("%s:%d:%s", rule.MachineID, rule.AdapterSlot, rule.Name))
	plan.HostPort = types.Int64Value(int64(hostPort))
	plan.EffectiveHostPort = types.Int64Value(int64(readRule.HostPort))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *natPortForwardResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state natPortForwardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the rule
	rule, err := r.client.ReadNATPortForward(
		ctx,
		state.MachineID.ValueString(),
		uint32(state.AdapterSlot.ValueInt64()),
		state.Name.ValueString(),
	)
	if err != nil {
		// If the machine doesn't exist, remove from state
		if vbox.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read NAT port forward rule", err.Error())
		return
	}

	// If rule doesn't exist, remove from state
	if rule == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update state with actual values
	state.EffectiveHostPort = types.Int64Value(int64(rule.HostPort))

	// Update protocol to match actual
	if rule.Protocol == vboxapi.NATProtocolTCP {
		state.Protocol = types.StringValue("tcp")
	} else {
		state.Protocol = types.StringValue("udp")
	}

	state.HostIP = types.StringValue(rule.HostIP)
	state.GuestIP = types.StringValue(rule.GuestIP)
	state.GuestPort = types.Int64Value(int64(rule.GuestPort))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *natPortForwardResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan natPortForwardModel
	var state natPortForwardModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// NAT port forward rules don't support in-place updates - we need to delete and recreate
	// This is because VirtualBox API doesn't have an "update" operation for redirects

	// Delete the old rule
	err := r.client.DeleteNATPortForward(
		ctx,
		state.MachineID.ValueString(),
		uint32(state.AdapterSlot.ValueInt64()),
		state.Name.ValueString(),
	)
	if err != nil && !vbox.IsNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete old NAT port forward rule", err.Error())
		return
	}

	// Determine the host port to use
	hostPort := uint16(plan.HostPort.ValueInt64())

	// If auto_host_port is enabled and host_port is not set (or is 0), allocate a port
	if plan.AutoHostPort.ValueBool() && hostPort == 0 {
		opts := vbox.PortAllocatorOptions{
			MinPort:            uint16(plan.AutoHostPortMin.ValueInt64()),
			MaxPort:            uint16(plan.AutoHostPortMax.ValueInt64()),
			HostIP:             plan.HostIP.ValueString(),
			Scope:              vbox.HostIPScope(plan.AutoHostIPScope.ValueString()),
			IncludeNATNetworks: true,
		}

		allocatedPort, err := r.client.AllocateNATHostPort(ctx, opts)
		if err != nil {
			resp.Diagnostics.AddError("Failed to allocate host port", err.Error())
			return
		}
		hostPort = allocatedPort
	}

	// Validate that we have a valid host port
	if hostPort == 0 {
		resp.Diagnostics.AddError(
			"Invalid host port",
			"host_port must be specified or auto_host_port must be enabled to automatically select a port",
		)
		return
	}

	// Parse protocol
	proto := vboxapi.NATProtocolTCP
	if strings.EqualFold(plan.Protocol.ValueString(), "udp") {
		proto = vboxapi.NATProtocolUDP
	}

	// Create the new rule
	rule := vbox.NATPortForwardRule{
		MachineID:   plan.MachineID.ValueString(),
		AdapterSlot: uint32(plan.AdapterSlot.ValueInt64()),
		Name:        plan.Name.ValueString(),
		Protocol:    proto,
		HostIP:      plan.HostIP.ValueString(),
		HostPort:    hostPort,
		GuestIP:     plan.GuestIP.ValueString(),
		GuestPort:   uint16(plan.GuestPort.ValueInt64()),
	}

	if err := r.client.CreateNATPortForward(ctx, rule); err != nil {
		resp.Diagnostics.AddError("Failed to create NAT port forward rule", err.Error())
		return
	}

	// Read back to confirm
	readRule, err := r.client.ReadNATPortForward(ctx, rule.MachineID, rule.AdapterSlot, rule.Name)
	if err != nil {
		resp.Diagnostics.AddError("Failed to verify NAT port forward rule", err.Error())
		return
	}
	if readRule == nil {
		resp.Diagnostics.AddError("NAT port forward rule not found after creation", "The rule was created but could not be read back")
		return
	}

	// Update state
	plan.ID = types.StringValue(fmt.Sprintf("%s:%d:%s", rule.MachineID, rule.AdapterSlot, rule.Name))
	plan.HostPort = types.Int64Value(int64(hostPort))
	plan.EffectiveHostPort = types.Int64Value(int64(readRule.HostPort))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *natPortForwardResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state natPortForwardModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteNATPortForward(
		ctx,
		state.MachineID.ValueString(),
		uint32(state.AdapterSlot.ValueInt64()),
		state.Name.ValueString(),
	)
	if err != nil {
		// Ignore not found errors - rule is already gone
		if !vbox.IsNotFound(err) {
			resp.Diagnostics.AddError("Failed to delete NAT port forward rule", err.Error())
			return
		}
	}
}

// ImportState implements resource.ResourceWithImportState
func (r *natPortForwardResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expected import ID format: machine_id:adapter_slot:name
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID format: machine_id:adapter_slot:name, got: %s", req.ID),
		)
		return
	}

	machineID := parts[0]
	adapterSlotStr := parts[1]
	name := parts[2]

	// Parse adapter slot
	var adapterSlot int64
	_, err := fmt.Sscanf(adapterSlotStr, "%d", &adapterSlot)
	if err != nil || adapterSlot < 0 || adapterSlot > 7 {
		resp.Diagnostics.AddError(
			"Invalid adapter slot",
			fmt.Sprintf("Adapter slot must be a number between 0 and 7, got: %s", adapterSlotStr),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("machine_id"), machineID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("adapter_slot"), adapterSlot)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

// Ensure the resource implements the ResourceWithImportState interface
var _ resource.ResourceWithImportState = &natPortForwardResource{}

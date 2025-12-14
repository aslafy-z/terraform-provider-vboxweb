package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox"
)

type machineCloneResource struct {
	client *vbox.Client
}

type machineCloneModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Source       types.String `tfsdk:"source"`
	CloneMode    types.String `tfsdk:"clone_mode"`
	CloneOptions types.List   `tfsdk:"clone_options"`

	DesiredState types.String `tfsdk:"state"`
	SessionType  types.String `tfsdk:"session_type"`
	WaitTimeout  types.String `tfsdk:"wait_timeout"`

	CurrentState types.String `tfsdk:"current_state"`
}

func NewMachineCloneResource() resource.Resource {
	return &machineCloneResource{}
}

func (r *machineCloneResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (r *machineCloneResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*vbox.Client)
}

func (r *machineCloneResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Clones an existing VirtualBox VM and optionally starts/stops it.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Machine UUID.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the new cloned VM.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"source": schema.StringAttribute{
				Required:    true,
				Description: "Source VM name or UUID to clone from.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"clone_mode": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Clone mode: MachineState, MachineAndChildStates, AllStates. Default: MachineState.",
				Validators: []validator.String{
					stringvalidator.OneOf("MachineState", "MachineAndChildStates", "AllStates"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"clone_options": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Clone options: Link, KeepAllMACs, KeepNATMACs, KeepDiskNames, KeepHwUUIDs.",
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf(
						"Link",
						"KeepAllMACs",
						"KeepNATMACs",
						"KeepDiskNames",
						"KeepHwUUIDs",
					)),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},

			"state": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Desired state: started or stopped. Default: stopped.",
				Validators: []validator.String{
					stringvalidator.OneOf("started", "stopped"),
				},
			},
			"session_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Session type used when starting a VM: headless or gui. Default: headless.",
			},
			"wait_timeout": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How long to wait for long operations (clone/start/stop/deleteConfig). Default: 20m.",
			},
			"current_state": schema.StringAttribute{
				Computed:    true,
				Description: "Observed VirtualBox machine state (best-effort).",
			},
		},
	}
}

func normalizeDesiredState(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "started", "running", "on":
		return "started"
	case "stopped", "poweredoff", "powered_off", "off":
		return "stopped"
	default:
		return s
	}
}

func parseTimeout(s string) time.Duration {
	if strings.TrimSpace(s) == "" {
		return 20 * time.Minute
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 20 * time.Minute
	}
	if d <= 0 {
		return 20 * time.Minute
	}
	return d
}

func (r *machineCloneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan machineCloneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.CloneMode.IsNull() || plan.CloneMode.ValueString() == "" {
		plan.CloneMode = types.StringValue("MachineState")
	}
	if plan.DesiredState.IsNull() || plan.DesiredState.ValueString() == "" {
		plan.DesiredState = types.StringValue("stopped")
	}
	if plan.SessionType.IsNull() || plan.SessionType.ValueString() == "" {
		plan.SessionType = types.StringValue("headless")
	}
	if plan.WaitTimeout.IsNull() || plan.WaitTimeout.ValueString() == "" {
		plan.WaitTimeout = types.StringValue("20m")
	}

	desired := normalizeDesiredState(plan.DesiredState.ValueString())
	timeout := parseTimeout(plan.WaitTimeout.ValueString())

	uuid, curState, err := r.client.CloneAndConverge(ctx, vbox.CloneRequest{
		Name:         plan.Name.ValueString(),
		Source:       plan.Source.ValueString(),
		CloneMode:    plan.CloneMode.ValueString(),
		CloneOptions: vbox.ListToStrings(plan.CloneOptions),
		DesiredState: desired,
		SessionType:  plan.SessionType.ValueString(),
		Timeout:      timeout,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to clone VM", err.Error())
		return
	}

	plan.ID = types.StringValue(uuid)
	plan.CurrentState = types.StringValue(curState)
	plan.DesiredState = types.StringValue(desired)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *machineCloneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state machineCloneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.ValueString() == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	cur, err := r.client.GetStateByID(ctx, state.ID.ValueString())
	if err != nil {
		// If it was deleted out of band, drop from state.
		if vbox.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read VM state", err.Error())
		return
	}

	state.CurrentState = types.StringValue(cur)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *machineCloneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan machineCloneModel
	var prior machineCloneModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &prior)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.ID.IsNull() || plan.ID.ValueString() == "" {
		resp.Diagnostics.AddError("Missing id", "Resource state is missing id")
		return
	}

	if plan.DesiredState.IsNull() || plan.DesiredState.ValueString() == "" {
		plan.DesiredState = types.StringValue("stopped")
	}
	if plan.SessionType.IsNull() || plan.SessionType.ValueString() == "" {
		plan.SessionType = types.StringValue("headless")
	}
	if plan.WaitTimeout.IsNull() || plan.WaitTimeout.ValueString() == "" {
		plan.WaitTimeout = types.StringValue("20m")
	}

	desired := normalizeDesiredState(plan.DesiredState.ValueString())
	timeout := parseTimeout(plan.WaitTimeout.ValueString())

	cur, err := r.client.ConvergeStateByID(ctx, plan.ID.ValueString(), desired, plan.SessionType.ValueString(), timeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to change VM state", err.Error())
		return
	}

	plan.CurrentState = types.StringValue(cur)
	plan.DesiredState = types.StringValue(desired)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *machineCloneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state machineCloneModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ID.IsNull() || state.ID.ValueString() == "" {
		return
	}

	timeout := 20 * time.Minute
	if !state.WaitTimeout.IsNull() {
		timeout = parseTimeout(state.WaitTimeout.ValueString())
	}

	if err := r.client.DeleteByID(ctx, state.ID.ValueString(), timeout); err != nil {
		if vbox.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to delete VM", err.Error())
		return
	}
}

// ImportState implements resource.ResourceWithImportState.
// Import ID format: machine UUID or name
func (r *machineCloneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The import ID can be either a machine UUID or name
	machineInfo, err := r.client.GetMachineInfoByID(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to import machine",
			fmt.Sprintf("Could not find machine with ID or name %q: %s", req.ID, err.Error()),
		)
		return
	}

	// Set the ID (UUID)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), machineInfo.ID)...)

	// Set the name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), machineInfo.Name)...)

	// Set current state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("current_state"), machineInfo.State)...)

	// Set defaults for fields that can't be determined from existing machine
	// source is unknown for imported machines - set to empty string (will require manual update)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("source"), "")...)

	// Set sensible defaults for clone options
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("clone_mode"), "MachineState")...)

	// Determine desired state based on current state
	desiredState := "stopped"
	if machineInfo.State == "Running" {
		desiredState = "started"
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("state"), desiredState)...)

	// Set default session type and timeout
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("session_type"), "headless")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("wait_timeout"), "20m")...)
}

// Ensure the resource implements the ResourceWithImportState interface
var _ resource.ResourceWithImportState = &machineCloneResource{}

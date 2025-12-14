---
page_title: "vboxweb_machine Resource - terraform-provider-vboxweb"
subcategory: ""
description: |-
  Manages a VirtualBox virtual machine by cloning from an existing template.
---

# vboxweb_machine (Resource)

Creates and manages a VirtualBox virtual machine by cloning an existing VM template.

The resource automatically handles:

- Cloning the source VM with specified options
- Starting or stopping the VM based on desired state
- Cleaning up all associated files and media on destroy

## Example Usage

### Basic Clone

```hcl
resource "vboxweb_machine" "example" {
  name   = "my-vm"
  source = "ubuntu-template"
}
```

### Clone with Power Management

```hcl
resource "vboxweb_machine" "web_server" {
  name         = "web-server"
  source       = "ubuntu-22.04-base"
  state        = "started"
  session_type = "headless"
}
```

### Linked Clone (Faster, Uses Less Disk)

```hcl
resource "vboxweb_machine" "linked" {
  name          = "linked-clone"
  source        = "ubuntu-template"
  clone_mode    = "MachineState"
  clone_options = ["Link"]
  state         = "started"
}
```

### Full Clone with All Options

```hcl
resource "vboxweb_machine" "full" {
  name          = "full-clone"
  source        = "golden-image"
  clone_mode    = "MachineAndChildStates"
  clone_options = ["KeepAllMACs", "KeepDiskNames"]
  state         = "started"
  session_type  = "gui"
  wait_timeout  = "30m"
}
```

## Argument Reference

### Required

- `name` (String) Name of the new cloned VM. Changing this forces a new resource.
- `source` (String) Source VM name or UUID to clone from. Changing this forces a new resource.

### Optional

- `clone_mode` (String) Clone mode. Valid values: `MachineState` (default), `MachineAndChildStates`, `AllStates`. Changing this forces a new resource.
- `clone_options` (List of String) Clone options. Valid values: `Link`, `KeepAllMACs`, `KeepNATMACs`, `KeepDiskNames`, `KeepHwUUIDs`. Changing this forces a new resource.
- `state` (String) Desired power state. Valid values: `started`, `stopped` (default).
- `session_type` (String) Session type when starting the VM. Valid values: `headless` (default), `gui`.
- `wait_timeout` (String) Timeout for long-running operations (clone, start, stop, delete). Default: `20m`. Format: Go duration string (e.g., `30m`, `1h`).

## Attribute Reference

- `id` (String) The UUID of the virtual machine.
- `current_state` (String) The observed VirtualBox machine state (e.g., `Running`, `PoweredOff`, `Saved`).

## Import

Existing VirtualBox machines can be imported into Terraform using their UUID or name.

```shell
terraform import vboxweb_machine.example <machine_uuid_or_name>
```

### Import by UUID

```shell
terraform import vboxweb_machine.example "550e8400-e29b-41d4-a716-446655440000"
```

### Import by Name

```shell
terraform import vboxweb_machine.example "my-existing-vm"
```

~> **Note:** When importing an existing machine, the `source` attribute will be set to an empty string since the original source VM cannot be determined. After import, you should update your Terraform configuration to set `source` to an appropriate value (or any placeholder value, as it is only used during initial clone). The `clone_mode` and `clone_options` attributes will also be set to defaults.

## Lifecycle Behavior

### Create

1. Finds the source VM by name or UUID
2. Creates a new VM definition with the same platform architecture and OS type
3. Clones the source VM to the new VM
4. Registers the cloned VM with VirtualBox
5. Starts or stops the VM based on the `state` attribute

### Update

Only the `state` attribute can be updated in-place. Changes to `name`, `source`, `clone_mode`, or `clone_options` will force recreation of the resource.

### Delete

1. Powers off the VM (best-effort)
2. Unregisters the VM
3. Deletes all associated files and media using `CleanupModeFull`

## Timeouts

All long-running operations respect the `wait_timeout` attribute:

- Clone operation
- Start operation  
- Stop operation
- Delete operation

The default timeout is 20 minutes. Increase this for large VMs or slow storage.

---
page_title: "vboxweb_nat_port_forward Resource - terraform-provider-vboxweb"
subcategory: ""
description: |-
  Manages a NAT port forwarding rule on a VirtualBox VM network adapter.
---

# vboxweb_nat_port_forward (Resource)

Manages a NAT port forwarding rule on a VirtualBox VM network adapter.

This resource creates a single NAT port forwarding rule on a VM's NAT-attached network adapter.
It supports an optional "auto host port" mode that automatically selects an available host port
from a configured range, avoiding conflicts with other VirtualBox NAT port forwarding rules.

## Important Guarantees and Limitations

- When using `auto_host_port`, the selected port is guaranteed not to conflict with any other 
  VirtualBox NAT port forwarding rule on the same VirtualBox instance at apply time.
- This does **NOT** guarantee the port is not used by other (non-VirtualBox) processes on the host.
- VirtualBox may not surface runtime bind failures if the port is already in use by another process.
- Changes to any rule attribute (except auto_host_port settings) will trigger rule replacement.

## Example Usage

### Explicit Host Port

```hcl
resource "vboxweb_nat_port_forward" "ssh" {
  machine_id   = vboxweb_machine.vm.id
  adapter_slot = 0
  name         = "ssh"
  protocol     = "tcp"
  host_ip      = "127.0.0.1"
  host_port    = 2222
  guest_port   = 22
}
```

### Auto Host Port Selection

```hcl
resource "vboxweb_nat_port_forward" "ssh" {
  machine_id   = vboxweb_machine.vm.id
  adapter_slot = 0
  name         = "ssh"
  protocol     = "tcp"
  host_ip      = "127.0.0.1"
  guest_port   = 22

  auto_host_port     = true
  auto_host_port_min = 20000
  auto_host_port_max = 40000
  auto_host_ip_scope = "any"
}

# Use the auto-selected port
output "ssh_port" {
  value = vboxweb_nat_port_forward.ssh.effective_host_port
}
```

### Multiple Port Forwards

```hcl
resource "vboxweb_nat_port_forward" "ssh" {
  machine_id   = vboxweb_machine.vm.id
  adapter_slot = 0
  name         = "ssh"
  protocol     = "tcp"
  guest_port   = 22

  auto_host_port = true
}

resource "vboxweb_nat_port_forward" "http" {
  machine_id   = vboxweb_machine.vm.id
  adapter_slot = 0
  name         = "http"
  protocol     = "tcp"
  guest_port   = 80

  auto_host_port = true
}

resource "vboxweb_nat_port_forward" "https" {
  machine_id   = vboxweb_machine.vm.id
  adapter_slot = 0
  name         = "https"
  protocol     = "tcp"
  guest_port   = 443

  auto_host_port = true
}
```

### UDP Port Forward

```hcl
resource "vboxweb_nat_port_forward" "dns" {
  machine_id   = vboxweb_machine.vm.id
  adapter_slot = 0
  name         = "dns"
  protocol     = "udp"
  host_port    = 5353
  guest_port   = 53
}
```

## Schema

### Required

- `machine_id` (String) - VirtualBox machine ID (UUID) that owns the NAT adapter.
- `adapter_slot` (Number) - Network adapter slot number (0-7, corresponding to nic1-nic8).
- `name` (String) - Name of the NAT port forwarding rule. Must be unique within the adapter's NAT engine.
- `protocol` (String) - Protocol for the port forwarding rule: `tcp` or `udp`.
- `guest_port` (Number) - Guest port number (1-65535).

### Optional

- `host_ip` (String) - Host IP address to bind to. Empty string or `0.0.0.0` means all interfaces. Default: `""`.
- `host_port` (Number) - Host port number. If omitted or 0 and `auto_host_port` is true, a port will be automatically selected.
- `guest_ip` (String) - Guest IP address. Empty string is typically fine for most use cases. Default: `""`.
- `auto_host_port` (Boolean) - If true and `host_port` is not set (or is 0), automatically select an available host port. Default: `false`.
- `auto_host_port_min` (Number) - Minimum port for auto-selection range (inclusive). Default: `20000`.
- `auto_host_port_max` (Number) - Maximum port for auto-selection range (inclusive). Default: `40000`.
- `auto_host_ip_scope` (String) - How to handle host IP when checking for port conflicts: `any` (all bindings conflict) or `exact` (only same host_ip conflicts). Default: `any`.

### Read-Only

- `id` (String) - Unique identifier for this resource (format: `machine_id:adapter_slot:name`).
- `effective_host_port` (Number) - The actual host port in use. This equals `host_port` when explicitly set, or the auto-selected port when using `auto_host_port`.

## Auto Host Port Algorithm

When `auto_host_port` is enabled and no explicit `host_port` is provided:

1. The provider enumerates all NAT port forwarding rules across all VMs in the VirtualBox instance.
2. Optionally, it also includes NAT Network port forwarding rules.
3. Based on the `auto_host_ip_scope` setting:
   - `any`: All host ports in use are considered conflicting, regardless of their host IP binding.
   - `exact`: Only ports with the same host IP (or "any" binding) are considered conflicting.
4. The provider selects the lowest available port in the configured range.
5. If no ports are available, an error is returned with diagnostic information.

### Concurrency Note

The auto-port selection avoids collisions with existing rules at the time of apply. However, if two Terraform applies run in parallel and both attempt to auto-select ports, they may select the same port. This is a documented limitation.

## Import

Import is supported using the ID format `machine_id:adapter_slot:name`:

```shell
terraform import vboxweb_nat_port_forward.ssh "12345678-1234-1234-1234-123456789abc:0:ssh"
```

## Prerequisites

- The target network adapter must be configured with NAT attachment type.
- The VM should typically be powered off when creating or modifying NAT port forwarding rules, although VirtualBox may allow hot changes in some cases.

---
page_title: "Multi-Version VirtualBox Support - Architecture Guide"
subcategory: ""
description: |-
  Guide to the provider's multi-version VirtualBox architecture and how to add support for new versions.
---

# Multi-Version VirtualBox Support

This document describes the architecture for supporting multiple VirtualBox versions and how to add support for new versions.

## Architecture Overview

The provider uses an **adapter pattern** to support multiple VirtualBox versions:

```
internal/
├── vboxapi/                 # Interface package (no external dependencies)
│   └── api.go              # VBoxAPI interface + common types
│
├── vbox/                    # Version-agnostic client layer
│   ├── client.go           # High-level client (uses VBoxAPI)
│   ├── helpers.go          # Terraform type utilities
│   ├── nat_redirect.go     # NAT redirect parsing (common format)
│   └── port_allocator.go   # Port allocation logic
│
├── vbox71/                  # VirtualBox 7.1 specific
│   ├── adapter.go          # VBoxAPI implementation for 7.1
│   └── generated/          # Generated WSDL bindings
│       └── vbox_service.go
│
└── vboxXX/                  # Future: VirtualBox X.X specific
    ├── adapter.go
    └── generated/
        └── vbox_service.go
```

### Key Principles

1. **`internal/vboxapi/api.go`**: Defines the `VBoxAPI` interface and common types - no dependencies to avoid import cycles
2. **`internal/vbox/client.go`**: High-level operations that work with any VBoxAPI implementation
3. **`internal/vboxXX/adapter.go`**: Version-specific adapter implementing VBoxAPI
4. **`internal/vboxXX/generated/`**: Generated WSDL bindings (version-specific)

### What Goes Where

| File Location | Contents | Version-Specific? |
|---------------|----------|-------------------|
| `vboxapi/api.go` | VBoxAPI interface, NATProtocol, MachineState constants | No |
| `vbox/client.go` | High-level operations (Clone, Delete, NAT port forwarding) | No |
| `vbox/nat_redirect.go` | Redirect string parsing (format is common across versions) | No |
| `vbox/port_allocator.go` | Port allocation algorithm | No |
| `vbox/helpers.go` | Terraform type utilities | No |
| `vbox71/adapter.go` | VBox 7.1 VBoxAPI implementation | **Yes** |
| `vbox71/generated/vbox_service.go` | Generated SOAP types for 7.1 | **Yes** |

## Adding Support for a New VirtualBox Version

### Step 1: Generate WSDL Bindings

```bash
# Install gowsdl
go install github.com/hooklift/gowsdl/cmd/gowsdl@latest

# Create version-specific directory with generated subpackage
mkdir -p internal/vboxXX/generated

# Generate bindings (replace XX with version, e.g., 80 for 8.0)
cd internal/vboxXX/generated
gowsdl -p generated -o vbox_service.go "http://localhost:18083/?wsdl"

# Fix the unexported field issue
sed -i 's/\t_this string/\tThis string/g' vbox_service.go
```

### Step 2: Create the Adapter

Create `internal/vboxXX/adapter.go`:

```go
package vboxXX

import (
    "context"
    "github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxXX/generated"
    "github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxapi"
    "github.com/hooklift/gowsdl/soap"
)

// Adapter implements vboxapi.VBoxAPI for VirtualBox X.X.
type Adapter struct {
    svc generated.VboxPortType
}

// NewAdapter creates a new adapter for VirtualBox X.X.
func NewAdapter(endpoint string) *Adapter {
    soapClient := soap.NewClient(endpoint)
    return &Adapter{svc: generated.NewVboxPortType(soapClient)}
}

// Implement all VBoxAPI methods...
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

// ... implement all other VBoxAPI methods

// Compile-time check that Adapter implements vboxapi.VBoxAPI
var _ vboxapi.VBoxAPI = (*Adapter)(nil)
```

### Step 3: Update Client to Support New Version

Update `internal/vbox/client.go` to detect and use the new adapter:

```go
import (
    "github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox71"
    "github.com/aslafy-z/terraform-provider-vboxweb/internal/vbox80"
    "github.com/aslafy-z/terraform-provider-vboxweb/internal/vboxapi"
)

func newAdapter(endpoint string, preferredVersion string) vboxapi.VBoxAPI {
    switch preferredVersion {
    case "7.1":
        return vbox71.NewAdapter(endpoint)
    case "8.0":
        return vbox80.NewAdapter(endpoint)
    default:
        // Auto-detect or default to latest
        return vbox71.NewAdapter(endpoint)
    }
}
```

### Step 4: Add Version Detection (Optional)

For auto-detection, you could:

1. Try to connect and call `GetAPIVersion()`
2. Parse the version and select the appropriate adapter
3. Fall back to a default if detection fails

## Version Compatibility Matrix

| Provider Version | VBox 7.1 | VBox 8.0 |
|------------------|----------|----------|
| 0.1.x            | ✅       | ❓       |

## API Differences Between Versions

When adding a new version, be aware of potential API differences:

### VBox 7.1 Specifics

- `IMachine_createMachine` requires `PlatformArchitecture` parameter
- NAT engine methods use `INATEngine_*` interface

### Common Patterns

- Redirect string format: `name,proto,hostIP,hostPort,guestIP,guestPort` (consistent)
- Protocol encoding: `0=UDP, 1=TCP` (consistent)
- NAT Network rules format: `name:proto:hostIP:hostPort:guestIP:guestPort` (consistent)

## Testing Multiple Versions

To test against different VirtualBox versions:

1. Set up VMs with different VBox versions
2. Run `vboxwebsrv` on each
3. Update test configuration to point to different endpoints
4. Run acceptance tests against each version

```bash
# Test against VBox 7.1
VBOX_ENDPOINT=http://vbox71-host:18083/ go test ./...

# Test against VBox 8.0
VBOX_ENDPOINT=http://vbox80-host:18083/ go test ./...
```

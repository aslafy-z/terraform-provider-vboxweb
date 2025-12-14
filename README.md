# Terraform Provider VBoxWeb

[![Test](https://github.com/aslafy-z/terraform-provider-vboxweb/actions/workflows/test.yml/badge.svg)](https://github.com/aslafy-z/terraform-provider-vboxweb/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/aslafy-z/terraform-provider-vboxweb)](https://goreportcard.com/report/github.com/aslafy-z/terraform-provider-vboxweb)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Terraform provider for managing VirtualBox virtual machines via the `vboxwebsrv` SOAP API.

## Features

- 🖥️ **Clone VMs** from existing templates with full/linked clone support
- ⚡ **Power Management** - Start and stop VMs with configurable session types
- 🌐 **NAT Port Forwarding** - Configure port forwarding rules with automatic port allocation
- 📦 **Import Existing VMs** - Import VMs into Terraform state by UUID or name
- 🧹 **Clean Lifecycle** - Automatic cleanup of VM files and attached media on destroy
- 🔌 **Multi-Version Architecture** - Designed to support multiple VirtualBox versions (7.1+ currently)

## Requirements

- [Terraform](https://www.terraform.io/downloads) 1.0+
- [VirtualBox](https://www.virtualbox.org/) 7.1+
- `vboxwebsrv` running and accessible

## Quick Start

### 1. Start VirtualBox Web Service

```bash
# Development mode (no authentication)
vboxwebsrv --host localhost --authentication null

# Production mode (with authentication)
vboxwebsrv --host localhost
```

### 2. Configure Provider

```hcl
terraform {
  required_providers {
    vboxweb = {
      source = "aslafy-z/vboxweb"
    }
  }
}

provider "vboxweb" {
  endpoint = "http://localhost:18083/"
  username = ""  # Empty if using --authentication null
  password = ""
}
```

### 3. Create a VM

```hcl
# Clone a VM from an existing template
resource "vboxweb_machine" "web" {
  name   = "web-server"
  source = "ubuntu-template"  # Name or UUID of source VM
  state  = "started"
}

# Add NAT port forwarding
resource "vboxweb_nat_port_forward" "ssh" {
  machine_id   = vboxweb_machine.web.id
  adapter_slot = 0
  name         = "ssh"
  protocol     = "tcp"
  guest_port   = 22

  auto_host_port     = true
  auto_host_port_min = 20000
  auto_host_port_max = 30000
}

output "ssh_command" {
  value = "ssh user@localhost -p ${vboxweb_nat_port_forward.ssh.effective_host_port}"
}
```

> **Note:** The source VM must be powered off before cloning.

## Resources

| Resource | Description |
|----------|-------------|
| [`vboxweb_machine`](docs/resources/machine.md) | Manages VirtualBox VMs via cloning |
| [`vboxweb_nat_port_forward`](docs/resources/nat_port_forward.md) | Manages NAT port forwarding rules |

## Documentation

- **[Getting Started Guide](docs/guides/getting-started.md)** - Full setup walkthrough
- **[Provider Configuration](docs/index.md)** - Provider schema and authentication
- **[Multi-Version Architecture](docs/guides/multi-version-support.md)** - Adding support for new VirtualBox versions

## Development

### Prerequisites

- Go 1.21+
- VirtualBox 7.1+ with `vboxwebsrv` running

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Local Development

Create a `playground/` directory (gitignored) with a `.terraformrc` file pointing to your local build:

```hcl
# playground/.terraformrc
provider_installation {
  dev_overrides {
    "registry.terraform.io/aslafy-z/vboxweb" = "/path/to/terraform-provider-vboxweb"
  }
  direct {}
}
```

Then run Terraform with the config file:

```bash
cd playground
TF_CLI_CONFIG_FILE=.terraformrc terraform apply
```

### Generating Documentation

```bash
make docs        # Generate docs from templates
make docs-check  # Verify docs are up-to-date (CI uses this)
```

### Regenerating WSDL Bindings

When adding support for a new VirtualBox version:

```bash
# Install gowsdl
go install github.com/hooklift/gowsdl/cmd/gowsdl@latest

# Start vboxwebsrv for your VirtualBox version
vboxwebsrv --host localhost --authentication null

# Generate bindings
mkdir -p internal/vboxXX/generated
cd internal/vboxXX/generated
gowsdl -p generated -o vbox_service.go "http://localhost:18083/?wsdl"

# Fix unexported field
sed -i 's/\t_this string/\tThis string/g' vbox_service.go
```

See the [Multi-Version Support Guide](docs/guides/multi-version-support.md) for full details.

## Acknowledgments

This project builds upon several open source projects:

- **[go-vbox-api](https://github.com/0n0sendai/go-vbox-api)** - Go bindings for the VirtualBox SOAP API, which served as inspiration for this provider
- **[gowsdl](https://github.com/hooklift/gowsdl)** - WSDL to Go code generator, used to generate the SOAP client from `vboxwebsrv` WSDL definitions

## License

[MIT](LICENSE)

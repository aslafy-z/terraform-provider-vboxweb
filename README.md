# terraform-provider-vboxweb

A Terraform provider for VirtualBox using the `vboxwebsrv` SOAP API. Supports VirtualBox 7.1+.

## Features

- Clone VMs from existing templates
- Manage VM power state (started/stopped)
- NAT port forwarding with auto host port allocation
- Automatic cleanup on destroy

## Quick Start

```hcl
provider "vboxweb" {
  endpoint = "http://localhost:18083/"
  username = ""
  password = ""
}

resource "vboxweb_machine" "example" {
  name   = "my-clone"
  source = "ubuntu-template"
  state  = "started"
}
```

## Documentation

- **[Getting Started](docs/guides/getting-started.md)** - Setup guide and prerequisites
- **[Provider Configuration](docs/index.md)** - Provider schema and authentication
- **[Multi-Version Support](docs/guides/multi-version-support.md)** - Architecture and adding new VBox versions

- **[vboxweb_machine Resource](docs/resources/machine.md)** - VM cloning resource
- **[vboxweb_nat_port_forward Resource](docs/resources/nat_port_forward.md)** - NAT port forwarding

## Development

```bash
# Build
go build -o terraform-provider-vboxweb

# Test
go test ./...

# Local testing with dev overrides
cd examples && cat README.md
```

## License

[MIT](LICENSE)

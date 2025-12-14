# terraform-provider-vboxweb

A Terraform provider for VirtualBox using the `vboxwebsrv` SOAP API. Supports VirtualBox 7.1+.

## Features

- Clone VMs from existing templates
- Manage VM power state (started/stopped)
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
- **[vboxweb_machine Resource](docs/resources/machine.md)** - Full resource documentation

## Development

```bash
# Build
go build -o terraform-provider-vboxweb

# Test
go test ./...

# Local testing with dev overrides
cd examples && cat README.md
```

## Architecture

The provider uses an adapter pattern for multi-version VirtualBox support:

```
internal/vbox/
├── api.go          # VBoxAPI interface (version-agnostic)
├── client.go       # High-level client using the interface
├── adapter71.go    # VBox 7.1 implementation
└── helpers.go      # Terraform type utilities

internal/vbox71/    # Generated WSDL bindings for VBox 7.1
```

## Adding Support for New VirtualBox Versions

1. Generate WSDL bindings:
```bash
go install github.com/hooklift/gowsdl/cmd/gowsdl@latest
mkdir -p internal/vboxXX
cd internal/vboxXX
gowsdl -p vboxXX -o vbox_service.go "http://localhost:18083/?wsdl"
# Fix unexported field: gowsdl generates "_this" but Go needs exported "This"
sed -i 's/\t_this string/\tThis string/g' vbox_service.go
```

2. Create a new adapter (`internal/vbox/adapterXX.go`) implementing `VBoxAPI`

3. Update `newAdapter()` in `client.go` to detect and use the new version

## License

[MIT](LICENSE)

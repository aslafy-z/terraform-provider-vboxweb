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

## Regenerating WSDL Bindings

The provider includes bindings for VirtualBox 7.1. For newer VirtualBox versions with API changes:

```bash
go install github.com/hooklift/gowsdl/cmd/gowsdl@latest
mkdir -p internal/vboxXX  # Replace XX with version
cd internal/vboxXX
gowsdl -p vboxXX -o vbox_service.go "http://localhost:18083/?wsdl"
sed -i 's/_this string/This string/g' vbox_service.go
# Update internal/vbox/client.go to import the new package
```

## License

[MIT](LICENSE)

# Development Testing

This directory is for local development testing of the provider.

## Quick Start

```bash
# Build the provider (from project root)
go build -o terraform-provider-vboxweb .

# Generate terraformrc with correct absolute path
cat > examples/.terraformrc <<EOF
provider_installation {
  dev_overrides {
    "registry.terraform.io/aslafy-z/vboxweb" = "$(pwd)"
  }
  direct {}
}
EOF

# Set dev overrides and run terraform
export TF_CLI_CONFIG_FILE="$(pwd)/examples/.terraformrc"
cd examples
terraform plan
terraform apply
```

## Configuration

Edit `main.tf` to set your source VM name and desired settings.

## Documentation

See [docs/guides/getting-started.md](../docs/guides/getting-started.md) for full documentation.

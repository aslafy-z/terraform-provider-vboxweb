# terraform-provider-vboxweb (vboxwebsrv)

A Terraform provider for VirtualBox using the `vboxwebsrv` SOAP API. Supports VirtualBox 7.1+.

## What is supported

Resource:

- `vboxweb_machine`
  - Create a new VM by cloning an existing VM ("Create From Clone")
  - Choose the desired power state on create: `started` or `stopped`
  - Update only the power state after creation (`started` <-> `stopped`)

On delete, the provider powers off the VM (best-effort), unregisters it, and deletes its files/media using `CleanupModeFull`.

## Requirements

- **VirtualBox 7.1+** with `vboxwebsrv` running
- **Terraform** >= 1.0

## Provider configuration

```hcl
provider "vboxweb" {
  endpoint = "http://vbox-host:18083/"
  username = var.vbox_user
  password = var.vbox_pass
}
```

## Resource usage

```hcl
resource "vboxweb_machine" "clone" {
  name   = "my-clone"
  source = "golden-template" # name or UUID

  # Optional (defaults shown)
  clone_mode    = "MachineState" # or MachineAndChildStates, AllStates
  clone_options = ["KeepAllMACs"]

  state        = "started"  # or stopped
  session_type = "headless" # or gui

  # How long to wait for clone/start/stop/delete
  wait_timeout = "20m"
}
```

Attributes:

- `id` (computed): Machine UUID
- `current_state` (computed): Observed VirtualBox state (e.g. `Running`, `PoweredOff`)

## Build

```bash
go build -o terraform-provider-vboxweb
```

## Development / Local Testing

The `examples/` directory contains a ready-to-use test setup:

```bash
# 1. Start vboxwebsrv (disable auth for testing)
vboxwebsrv --host localhost --authentication null

# 2. Build the provider
go build -o terraform-provider-vboxweb

# 3. Run terraform with dev overrides (skip init)
cd examples
export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"

# Edit main.tf - set "source" to your existing VM name

terraform plan
terraform apply
terraform destroy
```

With `dev_overrides`, Terraform uses the local binary directly - no need to run `terraform init`.

## Install for Production Use

Terraform discovers providers from `~/.terraform.d/plugins/<HOSTNAME>/<NAMESPACE>/<NAME>/<VERSION>/<OS>_<ARCH>/`.

Example (Linux AMD64):

```bash
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/aslafy-z/vboxweb/0.1.0/linux_amd64
cp terraform-provider-vboxweb ~/.terraform.d/plugins/registry.terraform.io/aslafy-z/vboxweb/0.1.0/linux_amd64/
```

Then in your `terraform` block:

```hcl
terraform {
  required_providers {
    vboxweb = {
      source  = "registry.terraform.io/aslafy-z/vboxweb"
      version = "0.1.0"
    }
  }
}
```

## Notes

- This provider creates a fresh VirtualBox webservice session per Terraform operation and logs off at the end.
- `name`, `source`, `clone_mode`, and `clone_options` are "ForceNew" (changing them will replace the VM).
- Power operations use `launchVMProcess` (start) and `lockMachine + powerDown` (stop).
- WSDL bindings in `internal/vbox71/` are generated from VBox 7.1 using `gowsdl`.

## Regenerating WSDL Bindings

If you need to update the bindings for a different VBox version:

```bash
go install github.com/hooklift/gowsdl/cmd/gowsdl@latest
cd internal/vbox71
gowsdl -p vbox71 -o vbox_service.go "http://localhost:18083/?wsdl"
sed -i 's/_this string/This string/g' vbox_service.go
```

## Troubleshooting

### Connection refused
Make sure `vboxwebsrv` is running:
```bash
vboxwebsrv --host localhost --authentication null
```

### "Must specify a valid platform architecture"
This error occurs with VirtualBox 7.1+ if using outdated WSDL bindings. Regenerate the bindings from your running `vboxwebsrv` as shown above.

### "Machine settings file already exists"
A previous failed clone left partial files. Clean up manually:
```bash
rm -rf ~/VirtualBox\ VMs/<vm-name>/
VBoxManage list hdds  # Find orphaned disks
VBoxManage closemedium disk <UUID> --delete
```

### Authentication errors
If authentication is required, configure `vboxwebsrv` accordingly and provide credentials in the provider config.

## License

See [LICENSE](LICENSE) file.

---
page_title: "Getting Started with VBoxWeb Provider"
description: |-
  Learn how to set up the VBoxWeb provider and create your first virtual machine.
---

# Getting Started with VBoxWeb Provider

This guide walks you through setting up the VBoxWeb provider and creating your first VirtualBox virtual machine with Terraform.

## Prerequisites

Before you begin, ensure you have:

1. **VirtualBox 7.1+** installed on your system
2. **Terraform 1.0+** installed
3. An existing VM template to clone (you need at least one VM in VirtualBox)

## Step 1: Start the VirtualBox Web Service

The VBoxWeb provider communicates with VirtualBox through its SOAP web service. Start the service:

```bash
# For development/testing (no authentication)
vboxwebsrv --host localhost --authentication null

# For production (with authentication)
vboxwebsrv --host localhost
```

The service runs on port 18083 by default.

## Step 2: Create a Terraform Configuration

Create a new directory for your Terraform configuration:

```bash
mkdir my-vbox-infra
cd my-vbox-infra
```

Create `main.tf`:

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

# Create a VM by cloning an existing template
resource "vboxweb_machine" "my_vm" {
  name   = "terraform-managed-vm"
  source = "Ubuntu-Template"  # Replace with your existing VM name
  state  = "started"
}

output "vm_id" {
  value = vboxweb_machine.my_vm.id
}

output "vm_state" {
  value = vboxweb_machine.my_vm.current_state
}
```

## Step 3: Initialize and Apply

```bash
# Initialize Terraform (downloads the provider)
terraform init

# Preview the changes
terraform plan

# Create the VM
terraform apply
```

After a successful apply, you should see:

```
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

vm_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
vm_state = "Running"
```

## Step 4: Verify in VirtualBox

Open VirtualBox Manager and you should see your new VM running.

## Step 5: Clean Up

When you're done, destroy the VM:

```bash
terraform destroy
```

This will:

1. Power off the VM
2. Unregister it from VirtualBox
3. Delete all associated files and media

## Local Installation (Without Registry)

If the provider isn't published to a registry yet, you can install it locally.

Terraform discovers providers from `~/.terraform.d/plugins/<HOSTNAME>/<NAMESPACE>/<NAME>/<VERSION>/<OS>_<ARCH>/`.

Example for Linux AMD64:

```bash
# Build the provider
go build -o terraform-provider-vboxweb

# Install to plugins directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/aslafy-z/vboxweb/0.1.0/linux_amd64
cp terraform-provider-vboxweb ~/.terraform.d/plugins/registry.terraform.io/aslafy-z/vboxweb/0.1.0/linux_amd64/
```

Then in your Terraform configuration:

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

## Troubleshooting

### "Connection refused" error

The vboxwebsrv is not running. Start it with:

```bash
vboxwebsrv --host localhost --authentication null
```

### "Could not find machine" error

The source VM name doesn't exist. Check available VMs:

```bash
VBoxManage list vms
```

### "Must specify a valid platform architecture" error

This error occurs with VirtualBox 7.1+ if using outdated WSDL bindings. The provider must be built with VBox 7.1 WSDL bindings. See the [development guide](https://github.com/aslafy-z/terraform-provider-vboxweb#regenerating-wsdl-bindings).

### "Machine settings file already exists" error

A previous failed clone left partial files. Clean up manually:

```bash
rm -rf ~/VirtualBox\ VMs/<vm-name>/
VBoxManage list hdds  # Find orphaned disks
VBoxManage closemedium disk <UUID> --delete
```

### Authentication errors

If authentication is required, configure `vboxwebsrv` accordingly and provide credentials in the provider config.

# Testing the VBoxWeb Terraform Provider

## Prerequisites

1. **VirtualBox** installed with a VM ready to be cloned
2. **vboxwebsrv** running (typically on `http://localhost:18083/`)
3. **Terraform** installed

## Setup

### 1. Start vboxwebsrv

```bash
vboxwebsrv -H localhost -p 18083
```

Or in the background:

```bash
vboxwebsrv -H localhost -p 18083 &
```

### 2. Configure the Provider Override

The `.terraformrc` file in this directory tells Terraform to use the locally built provider binary instead of downloading from the registry.

**Option A**: Copy to your home directory:
```bash
cp .terraformrc ~/.terraformrc
```

**Option B**: Set the environment variable:
```bash
export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"
```

### 3. Edit main.tf

Update the `main.tf` file:

1. Set the `source` attribute to the name (or UUID) of your existing VM to clone
2. Adjust `endpoint`, `username`, and `password` if needed (empty strings work if vboxwebsrv has no auth)

### 4. Run Terraform

```bash
# With dev_overrides, skip init and go directly to plan/apply
# Preview changes
terraform plan

# Apply (creates the clone)
terraform apply

# When done, destroy the clone
terraform destroy
```

## Example Output

```
vboxweb_machine.test: Creating...
vboxweb_machine.test: Creation complete after 15s [id=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

machine_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
machine_state = "PoweredOff"
```

## Troubleshooting

- **Connection refused**: Make sure vboxwebsrv is running
- **Clone fails**: Ensure the source VM exists and is not running
- **Permission denied**: Check vboxwebsrv authentication settings
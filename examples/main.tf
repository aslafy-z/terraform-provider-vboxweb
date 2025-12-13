terraform {
  required_providers {
    vboxweb = {
      source = "registry.terraform.io/aslafy-z/vboxweb"
    }
  }
}

provider "vboxweb" {
  endpoint = "http://localhost:18083/"
  username = ""  # vboxwebsrv username (can be empty if auth disabled)
  password = ""  # vboxwebsrv password (can be empty if auth disabled)
}

resource "vboxweb_machine" "test" {
  name   = "terraform-test-clone"
  source = "Ubuntu Base"  # Name or UUID of the VM to clone

  # Optional: clone_mode can be "MachineState", "MachineAndChildStates", or "AllStates"
  clone_mode = "MachineState"

  # Optional: clone_options can include "Link", "KeepAllMACs", "KeepNATMACs", "KeepDiskNames", "KeepHwUUIDs"
  # clone_options = ["Link"]

  # Optional: desired state - "started" or "stopped" (default: "stopped")
  state = "started"

  # Optional: session type for starting - "headless" or "gui" (default: "headless")
  session_type = "headless"

  # Optional: timeout for long operations (default: "20m")
  wait_timeout = "20m"
}

output "machine_id" {
  value = vboxweb_machine.test.id
}

output "machine_state" {
  value = vboxweb_machine.test.current_state
}
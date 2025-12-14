resource "vboxweb_nat_port_forward" "ssh_auto" {
  machine_id   = vboxweb_machine.example.id
  adapter_slot = 0
  name         = "ssh"
  protocol     = "tcp"
  guest_port   = 22

  # Automatically select an available host port
  auto_host_port     = true
  auto_host_port_min = 20000
  auto_host_port_max = 30000
}

# Access the automatically selected port
output "ssh_port" {
  value = vboxweb_nat_port_forward.ssh_auto.effective_host_port
}
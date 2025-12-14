resource "vboxweb_nat_port_forward" "ssh" {
  machine_id   = vboxweb_machine.example.id
  adapter_slot = 0
  name         = "ssh"
  protocol     = "tcp"
  host_port    = 2222
  guest_port   = 22
}
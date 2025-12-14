resource "vboxweb_nat_port_forward" "ssh" {
  machine_id   = vboxweb_machine.web.id
  adapter_slot = 0
  name         = "ssh"
  protocol     = "tcp"
  host_port    = 2222
  guest_port   = 22
}

resource "vboxweb_nat_port_forward" "http" {
  machine_id   = vboxweb_machine.web.id
  adapter_slot = 0
  name         = "http"
  protocol     = "tcp"
  host_port    = 8080
  guest_port   = 80
}

resource "vboxweb_nat_port_forward" "https" {
  machine_id   = vboxweb_machine.web.id
  adapter_slot = 0
  name         = "https"
  protocol     = "tcp"
  host_port    = 8443
  guest_port   = 443
}
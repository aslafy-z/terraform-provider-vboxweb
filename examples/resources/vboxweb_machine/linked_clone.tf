resource "vboxweb_machine" "linked" {
  name          = "linked-clone"
  source        = "ubuntu-template"
  clone_mode    = "MachineState"
  clone_options = ["Link"]
  state         = "started"
}
resource "vboxweb_machine" "full" {
  name          = "full-clone"
  source        = "golden-image"
  clone_mode    = "MachineAndChildStates"
  clone_options = ["KeepAllMACs", "KeepDiskNames"]
  state         = "started"
  session_type  = "gui"
  wait_timeout  = "30m"
}
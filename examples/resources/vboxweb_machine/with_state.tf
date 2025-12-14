resource "vboxweb_machine" "web_server" {
  name         = "web-server"
  source       = "ubuntu-22.04-base"
  state        = "started"
  session_type = "headless"
}
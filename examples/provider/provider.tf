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

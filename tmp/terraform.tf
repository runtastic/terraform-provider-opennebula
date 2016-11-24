# Set up a new service names "blobs" on everything
variable "one_endpoint" {
  default = "https://rtpain-opennebula.runtastic.com:8633/RPC2"
}

variable "one_user" {}
variable "one_password" {}

variable "service_name" {
  description = "The name of the service for which templates, images, networks, ... will be created"
  default = "test"
}

provider "opennebula" {
  endpoint = "${var.one_endpoint}"
  user = "${var.one_user}"
  password = "${var.one_password}"
}

resource "opennebula_template" "service" {
  name = "donidoni"

  sched_requirements = "CLUSTER_ID=\"100\""

  cpu = "0.1"
  memory = "1700"
  vcpu = "2"

  context = {
    # dns_hostname = "yes",
    network = "YES",
    ssh_public_key = "$USER[SSH_PUBLIC_KEY]",
    # token = "YES",
    # username = "vagrant"
  }

  #graphics = {
  #  # keymap = "de",
  #  listen = "0.0.0.0",
  #  type = "VNC"
  #}
#
  #nic = {
  #  network = "br-lnz-stg",
  #  network_uname = "oneadmin"
  #  # model = "virtio"
  #}

}

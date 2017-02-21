provider "digitalocean" {
  # See README.md for setup instructions.
}

# Tags to label and organize droplets:
resource "digitalocean_tag" "name" {
  name = "${var.name}"
}

resource "digitalocean_tag" "app" {
  name = "${var.app}"
}

resource "digitalocean_tag" "terraform" {
  name = "terraform"
}

resource "digitalocean_droplet" "tf_test_vm" {
  ssh_keys = ["${var.do_public_key_id}"]
  image    = "${var.do_os}"
  region   = "${var.do_dc}"
  size     = "${var.do_size}"
  name     = "${var.name}-${count.index}"
  count    = "${var.num_hosts}"

  tags = [
    "${var.app}",
    "${var.name}",
    "terraform",
  ]

  # Wait for machine to be SSH-able:
  provisioner "remote-exec" {
    inline = ["exit"]

    connection {
      type        = "ssh"
      user        = "${var.do_username}"
      private_key = "${file("${var.do_private_key_path}")}"
    }
  }
}

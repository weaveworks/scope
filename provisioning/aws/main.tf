# Specify the provider and access details
provider "aws" {
  # Access key, secret key and region are sourced from environment variables or input arguments -- see README.md
  region = "${var.aws_dc}"
}

resource "aws_security_group" "allow_ssh" {
  name        = "${var.name}_allow_ssh"
  description = "AWS security group to allow SSH-ing onto AWS EC2 instances (created using Terraform)."

  # Open TCP port for SSH:
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${var.client_ip}/32"]
  }

  tags {
    Name      = "${var.name}_allow_ssh"
    App       = "${var.app}"
    CreatedBy = "terraform"
  }
}

resource "aws_security_group" "allow_docker" {
  name        = "${var.name}_allow_docker"
  description = "AWS security group to allow communication with Docker on AWS EC2 instances (created using Terraform)."

  # Open TCP port for Docker:
  ingress {
    from_port   = 2375
    to_port     = 2375
    protocol    = "tcp"
    cidr_blocks = ["${var.client_ip}/32"]
  }

  tags {
    Name      = "${var.name}_allow_docker"
    App       = "${var.app}"
    CreatedBy = "terraform"
  }
}

resource "aws_security_group" "allow_weave" {
  name        = "${var.name}_allow_weave"
  description = "AWS security group to allow communication with Weave on AWS EC2 instances (created using Terraform)."

  # Open TCP port for Weave:
  ingress {
    from_port   = 12375
    to_port     = 12375
    protocol    = "tcp"
    cidr_blocks = ["${var.client_ip}/32"]
  }

  tags {
    Name      = "${var.name}_allow_weave"
    App       = "${var.app}"
    CreatedBy = "terraform"
  }
}

resource "aws_security_group" "allow_private_ingress" {
  name        = "${var.name}_allow_private_ingress"
  description = "AWS security group to allow all private ingress traffic on AWS EC2 instances (created using Terraform)."

  # Full inbound local network access on both TCP and UDP
  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["${var.aws_vpc_cidr_block}"]
  }

  tags {
    Name      = "${var.name}_allow_private_ingress"
    App       = "${var.app}"
    CreatedBy = "terraform"
  }
}

resource "aws_security_group" "allow_all_egress" {
  name        = "${var.name}_allow_all_egress"
  description = "AWS security group to allow all egress traffic on AWS EC2 instances (created using Terraform)."

  # Full outbound internet access on both TCP and UDP
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags {
    Name      = "${var.name}_allow_all_egress"
    App       = "${var.app}"
    CreatedBy = "terraform"
  }
}

resource "aws_instance" "tf_test_vm" {
  instance_type = "${var.aws_size}"
  count         = "${var.num_hosts}"

  # Lookup the correct AMI based on the region we specified
  ami = "${lookup(var.aws_amis, var.aws_dc)}"

  key_name = "${var.aws_public_key_name}"

  security_groups = [
    "${aws_security_group.allow_ssh.name}",
    "${aws_security_group.allow_docker.name}",
    "${aws_security_group.allow_weave.name}",
    "${aws_security_group.allow_private_ingress.name}",
    "${aws_security_group.allow_all_egress.name}",
  ]

  # Wait for machine to be SSH-able:
  provisioner "remote-exec" {
    inline = ["exit"]

    connection {
      type = "ssh"

      # Lookup the correct username based on the AMI we specified
      user        = "${lookup(var.aws_usernames, "${lookup(var.aws_amis, var.aws_dc)}")}"
      private_key = "${file("${var.aws_private_key_path}")}"
    }
  }

  tags {
    Name      = "${var.name}-${count.index}"
    App       = "${var.app}"
    CreatedBy = "terraform"
  }
}

variable "client_ip" {
  description = "IP address of the client machine"
}

variable "app" {
  description = "Name of the application using the created EC2 instance(s)."
  default     = "default"
}

variable "name" {
  description = "Name of the EC2 instance(s)."
  default     = "test"
}

variable "num_hosts" {
  description = "Number of EC2 instance(s)."
  default     = 1
}

variable "aws_vpc_cidr_block" {
  description = "AWS VPC CIDR block to use to attribute private IP addresses."
  default     = "172.31.0.0/16"
}

variable "aws_public_key_name" {
  description = "Name of the SSH keypair to use in AWS."
}

variable "aws_private_key_path" {
  description = "Path to file containing private key"
  default     = "~/.ssh/id_rsa"
}

variable "aws_dc" {
  description = "The AWS region to create things in."
  default     = "us-east-1"
}

variable "aws_amis" {
  default = {
    # Ubuntu Server 16.04 LTS (HVM), SSD Volume Type:
    "us-east-1" = "ami-40d28157"
    "eu-west-2" = "ami-23d0da47"

    # Red Hat Enterprise Linux 7.3 (HVM), SSD Volume Type:

    #"us-east-1" = "ami-b63769a1"
  }
}

variable "aws_usernames" {
  description = "User to SSH as into the AWS instance."

  default = {
    "ami-40d28157" = "ubuntu"   # Ubuntu Server 16.04 LTS (HVM)
    "ami-b63769a1" = "ec2-user" # Red Hat Enterprise Linux 7.3 (HVM)
  }
}

variable "aws_size" {
  description = "AWS' selected machine size"
  default     = "t2.medium"                  # Instance with 2 cores & 4 GB memory
}

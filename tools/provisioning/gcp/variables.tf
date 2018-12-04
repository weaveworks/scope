variable "gcp_username" {
  description = "Google Cloud Platform SSH username"
}

variable "app" {
  description = "Name of the application using the created Compute Engine instance(s)."
  default     = "default"
}

variable "name" {
  description = "Name of the Compute Engine instance(s)."
  default     = "test"
}

variable "num_hosts" {
  description = "Number of Compute Engine instance(s)."
  default     = 1
}

variable "client_ip" {
  description = "IP address of the client machine"
}

variable "gcp_public_key_path" {
  description = "Path to file containing public key"
  default     = "~/.ssh/id_rsa.pub"
}

variable "gcp_private_key_path" {
  description = "Path to file containing private key"
  default     = "~/.ssh/id_rsa"
}

variable "gcp_project" {
  description = "Google Cloud Platform project"
  default     = "weave-net-tests"
}

variable "gcp_image" {
  # See also: https://cloud.google.com/compute/docs/images
  # For example:
  # - "ubuntu-os-cloud/ubuntu-1604-lts"
  # - "debian-cloud/debian-8"
  # - "centos-cloud/centos-7"
  # - "rhel-cloud/rhel7"
  description = "Google Cloud Platform OS"

  default = "ubuntu-os-cloud/ubuntu-1604-lts"
}

variable "gcp_size" {
  # See also: 
  #   $ gcloud compute machine-types list
  description = "Google Cloud Platform's selected machine size"

  default = "n1-standard-1"
}

variable "gcp_region" {
  description = "Google Cloud Platform's selected region"
  default     = "us-central1"
}

variable "gcp_zone" {
  description = "Google Cloud Platform's selected zone"
  default     = "us-central1-a"
}

variable "gcp_network" {
  description = "Google Cloud Platform's selected network"
  default     = "test"
}

variable "gcp_network_global_cidr" {
  description = "CIDR covering all regions for the selected Google Cloud Platform network"
  default     = "10.128.0.0/9"
}

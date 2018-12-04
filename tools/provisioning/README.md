# Weaveworks provisioning

## Introduction

This project allows you to get hold of some machine either locally or on one of the below cloud providers:

* Amazon Web Services
* Digital Ocean
* Google Cloud Platform

You can then use these machines as is or run various Ansible playbooks from `../config_management` to set up Weave Net, Kubernetes, etc.

## Set up

* You will need [Vagrant](https://www.vagrantup.com) installed on your machine and added to your `PATH` in order to be able to provision local (virtual) machines automatically.

  * On macOS: `brew install vagrant`
  * On Linux (via Aptitude): `sudo apt install vagrant`
  * For other platforms or more details, see [here](https://www.vagrantup.com/docs/installation/)

* You will need [Terraform](https://www.terraform.io) installed on your machine and added to your `PATH` in order to be able to provision cloud-hosted machines automatically.

  * On macOS: `brew install terraform`
  * On Linux (via Aptitude): `sudo apt install terraform`
  * If you need a specific version:

          curl -fsS https://releases.hashicorp.com/terraform/x.y.z/terraform_x.y.z_linux_amd64.zip | gunzip > terraform && chmod +x terraform && sudo mv terraform /usr/bin
  * For other platforms or more details, see [here](https://www.terraform.io/intro/getting-started/install.html)

* Depending on the cloud provider, you may have to create an account, manually onboard, create and register SSH keys, etc. 
  Please refer to the `README.md` in each sub-folder for more details.

## Usage in scripts

Source `setup.sh`, set the `SECRET_KEY` environment variable, and depending on the cloud provider you want to use, call either:

* `gcp_on` / `gcp_off`
* `do_on` / `do_off`
* `aws_on` / `aws_off`

## Usage in shell

Source `setup.sh`, set the `SECRET_KEY` environment variable, and depending on the cloud provider you want to use, call either:

* `gcp_on` / `gcp_off`
* `do_on` / `do_off`
* `aws_on` / `aws_off`

Indeed, the functions defined in `setup.sh` are also exported as aliases, so you can call them from your shell directly.

Other aliases are also defined, in order to make your life easier:

* `tf_ssh`: to ease SSH-ing into the virtual machines, reading the username and IP address to use from Terraform, as well as setting default SSH options.
* `tf_ansi`: to ease applying an Ansible playbook to a set of virtual machines, dynamically creating the inventory, as well as setting default SSH options.

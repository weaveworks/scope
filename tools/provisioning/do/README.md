# Digital Ocean

## Introduction

This project allows you to get hold of some machine on Digital Ocean.
You can then use these machines as is or run various Ansible playbooks from `../config_management` to set up Weave Net, Kubernetes, etc.

## Setup

* Log in [cloud.digitalocean.com](https://cloud.digitalocean.com) with your account.

* Go to `Settings` > `Security` > `SSH keys` > `Add SSH Key`.
  Enter your SSH public key and the name for it, and click `Add SSH Key`.
  Set the path to your private key as an environment variable:

```
export DIGITALOCEAN_SSH_KEY_NAME=<your Digital Ocean SSH key name>
export TF_VAR_do_private_key_path="$HOME/.ssh/id_rsa"
```

* Go to `API` > `Tokens` > `Personal access tokens` > `Generate New Token`
  Enter your token name and click `Generate Token` to get your 64-characters-long API token.
  Set these as environment variables:

```
export DIGITALOCEAN_TOKEN_NAME="<your Digital Ocean API token name>"
export DIGITALOCEAN_TOKEN=<your Digital Ocean API token>
```

* Run the following command to get the Digital Ocean ID for your SSH public key (e.g. `1234567`) and set it as an environment variable:

```
$ export TF_VAR_do_public_key_id=$(curl -s -X GET -H "Content-Type: application/json" \
-H "Authorization: Bearer $DIGITALOCEAN_TOKEN" "https://api.digitalocean.com/v2/account/keys" \
| jq -c --arg key_name "$DIGITALOCEAN_SSH_KEY_NAME" '.ssh_keys | .[] | select(.name==$key_name) | .id')
```

  or pass it as a Terraform variable:

```
$ terraform <command> \
-var 'do_private_key_path=<path to your SSH private key>' \
-var 'do_public_key_id=<ID of your SSH public key>'
```

### Bash aliases

You can set the above variables temporarily in your current shell, permanently in your `~/.bashrc` file, or define aliases to activate/deactivate them at will with one single command by adding the below to your `~/.bashrc` file:

```
function _do_on() {
  export DIGITALOCEAN_TOKEN_NAME="<your_token_name>"        # Replace with appropriate value.
  export DIGITALOCEAN_TOKEN=<your_token>                    # Replace with appropriate value.
  export DIGITALOCEAN_SSH_KEY_NAME="<your_ssh_key_name>"    # Replace with appropriate value.
  export TF_VAR_do_private_key_path="$HOME/.ssh/id_rsa"     # Replace with appropriate value.
  export TF_VAR_do_public_key_path="$HOME/.ssh/id_rsa.pub"  # Replace with appropriate value.
  export TF_VAR_do_public_key_id=<your_ssh_key_id>          # Replace with appropriate value.
}
alias _do_on='_do_on'
function _do_off() {
  unset DIGITALOCEAN_TOKEN_NAME
  unset DIGITALOCEAN_TOKEN
  unset DIGITALOCEAN_SSH_KEY_NAME
  unset TF_VAR_do_private_key_path
  unset TF_VAR_do_public_key_path
  unset TF_VAR_do_public_key_id
}
alias _do_off='_do_off'
```

N.B.: 

* sourcing `../setup.sh` defines aliases called `do_on` and `do_off`, similarly to the above (however, notice no `_` in front of the name, as opposed to the ones above);
* `../setup.sh`'s `do_on` alias needs the `SECRET_KEY` environment variable to be set in order to decrypt sensitive information.

## Usage

* Create the machine: `terraform apply`
* Show the machine's status: `terraform show`
* Stop and destroy the machine: `terraform destroy`
* SSH into the newly-created machine:

```
$ ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no `terraform output username`@`terraform output public_ips`
```

or

```
source ../setup.sh
tf_ssh 1  # Or the nth machine, if multiple VMs are provisioned.
``` 

## Resources

* [https://www.terraform.io/docs/providers/do/](https://www.terraform.io/docs/providers/do/)
* [https://www.terraform.io/docs/providers/do/r/droplet.html](https://www.terraform.io/docs/providers/do/r/droplet.html)
* [Terraform variables](https://www.terraform.io/intro/getting-started/variables.html)

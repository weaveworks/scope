# Google Cloud Platform

## Introduction

This project allows you to get hold of some machine on Google Cloud Platform.
You can then use these machines as is or run various Ansible playbooks from `../config_management` to set up Weave Net, Kubernetes, etc.

## Setup

* Log in [console.cloud.google.com](https://console.cloud.google.com) with your Google account.

* Go to `API Manager` > `Credentials` > `Create credentials` > `Service account key`, 
  in `Service account`, select `Compute Engine default service account`,
  in `Key type`, select `JSON`, and then click `Create`.

* This will download a JSON file to your machine. Place this file wherever you want and then create the following environment variables:

```
$ export GOOGLE_CREDENTIALS_FILE="path/to/your.json"
$ export GOOGLE_CREDENTIALS=$(cat "$GOOGLE_CREDENTIALS_FILE")
```

* Go to `Compute Engine` > `Metadata` > `SSH keys` and add your username and SSH public key;
  or
  set it up using `gcloud compute project-info add-metadata --metadata-from-file sshKeys=~/.ssh/id_rsa.pub`.
  If you used your default SSH key (i.e. `~/.ssh/id_rsa.pub`), then you do not have anything to do.
  Otherwise, you will have to either define the below environment variable:

``` 
$ export TF_VAR_gcp_public_key_path=<path to your SSH public key>
$ export TF_VAR_gcp_private_key_path=<path to your SSH private key>
```

  or to pass these as Terraform variables:

```
$ terraform <command> \
-var 'gcp_public_key_path=<path to your SSH public key>' \
-var 'gcp_private_key_path=<path to your SSH private key>'
```

* Set the username in your public key as an environment variable.
  This will be used as the username of the Linux account created on the machine, which you will need to SSH into it later on.

  N.B.: 
  * GCP already has the username set from the SSH public key you uploaded in the previous step.
  * If your username is an email address, e.g. `name@domain.com`, then GCP uses `name` as the username.

```
export TF_VAR_gcp_username=<your SSH public key username>
```

* Set your current IP address as an environment variable:

```
export TF_VAR_client_ip=$(curl -s -X GET http://checkip.amazonaws.com/)
```

  or pass it as a Terraform variable:

```
$ terraform <command> -var 'client_ip=$(curl -s -X GET http://checkip.amazonaws.com/)'
```

* Set your project as an environment variable:

```
export TF_VAR_gcp_project=weave-net-tests
```

  or pass it as a Terraform variable:

```
$ terraform <command> -var 'gcp_project=weave-net-tests'
```

### Bash aliases

You can set the above variables temporarily in your current shell, permanently in your `~/.bashrc` file, or define aliases to activate/deactivate them at will with one single command by adding the below to your `~/.bashrc` file:

```
function _gcp_on() {
  export GOOGLE_CREDENTIALS_FILE="<path/to/your/json/credentials/file.json"
  export GOOGLE_CREDENTIALS=$(cat "$GOOGLE_CREDENTIALS_FILE")
  export TF_VAR_gcp_private_key_path="$HOME/.ssh/id_rsa"     # Replace with appropriate value.
  export TF_VAR_gcp_public_key_path="$HOME/.ssh/id_rsa.pub"  # Replace with appropriate value.
  export TF_VAR_gcp_username=$(cat "$TF_VAR_gcp_public_key_path" | cut -d' ' -f3 | cut -d'@' -f1)
}
alias _gcp_on='_gcp_on'
function _gcp_off() {
  unset GOOGLE_CREDENTIALS_FILE
  unset GOOGLE_CREDENTIALS
  unset TF_VAR_gcp_private_key_path
  unset TF_VAR_gcp_public_key_path
  unset TF_VAR_gcp_username
}
```

N.B.: 

* sourcing `../setup.sh` defines aliases called `gcp_on` and `gcp_off`, similarly to the above (however, notice no `_` in front of the name, as opposed to the ones above);
* `../setup.sh`'s `gcp_on` alias needs the `SECRET_KEY` environment variable to be set in order to decrypt sensitive information.

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

* [https://www.terraform.io/docs/providers/google/](https://www.terraform.io/docs/providers/google/)
* [https://www.terraform.io/docs/providers/google/r/compute_instance.html](https://www.terraform.io/docs/providers/google/r/compute_instance.html)
* [Terraform variables](https://www.terraform.io/intro/getting-started/variables.html)

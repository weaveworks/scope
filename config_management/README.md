# Weaveworks configuration management

## Introduction

This project allows you to configure a machine with:

* Docker and Weave Net for development: `setup_weave-net_dev.yml`
* Docker and Weave Net for testing: `setup_weave-net_test.yml`
* Docker, Kubernetes and Weave Kube (CNI plugin): `setup_weave-kube.yml`

You can then use these environments for development, testing and debugging.

## Set up

You will need [Python](https://www.python.org/downloads/) and [Ansible 2.+](http://docs.ansible.com/ansible/intro_installation.html) installed on your machine and added to your `PATH` in order to be able to configure environments automatically.

* On any platform, if you have Python installed: `pip install ansible`
* On macOS: `brew install ansible`
* On Linux (via Aptitude): `sudo apt install ansible`
* On Linux (via YUM): `sudo yum install ansible`
* For other platforms or more details, see [here](http://docs.ansible.com/ansible/intro_installation.html)

Frequent errors during installation are:

* `fatal error: Python.h: No such file or directory`: install `python-dev`
* `fatal error: ffi.h: No such file or directory`: install `libffi-dev`
* `fatal error: openssl/opensslv.h: No such file or directory`: install `libssl-dev`

Full steps for a blank Ubuntu/Debian Linux machine:

    sudo apt-get install -qq -y python-pip python-dev libffi-dev libssl-dev
    sudo pip install -U cffi
    sudo pip install ansible

## Tags

These can be used to selectively run (`--tags "tag1,tag2"`) or skip (`--skip-tags "tag1,tag2"`) tasks.

  * `output`: print potentially useful output from hosts (e.g. output of `kubectl get pods --all-namespaces`)

## Usage

### Local machine

```
ansible-playbook -u <username> -i "localhost", -c local setup_weave-kube.yml
```

### Vagrant

Provision your local VM using Vagrant:

```
cd $(mktemp -d -t XXX)
vagrant init ubuntu/xenial64  # or, e.g. centos/7
vagrant up
```

then set the following environment variables by extracting the output of `vagrant ssh-config`:

```
eval $(vagrant ssh-config | sed \
-ne 's/\ *HostName /vagrant_ssh_host=/p' \
-ne 's/\ *User /vagrant_ssh_user=/p' \
-ne 's/\ *Port /vagrant_ssh_port=/p' \
-ne 's/\ *IdentityFile /vagrant_ssh_id_file=/p')
```

and finally run:

```
ansible-playbook --private-key=$vagrant_ssh_id_file -u $vagrant_ssh_user \
--ssh-extra-args="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
-i "$vagrant_ssh_host:$vagrant_ssh_port," setup_weave-kube.yml
```

or, for specific versions of Kubernetes and Docker:

```
ansible-playbook --private-key=$vagrant_ssh_id_file -u $vagrant_ssh_user \
--ssh-extra-args="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
-i "$vagrant_ssh_host:$vagrant_ssh_port," setup_weave-kube.yml \
--extra-vars "docker_version=1.12.3 kubernetes_version=1.4.4"
```

NOTE: Kubernetes APT repo includes only the latest version, so currently
retrieving an older version will fail.

### Terraform

Provision your machine using the Terraform scripts from `../provisioning`, then run:

```
terraform output ansible_inventory > /tmp/ansible_inventory
```

and

```
ansible-playbook \
    --private-key="$(terraform output private_key_path)" \
    -u "$(terraform output username)" \
    -i /tmp/ansible_inventory \
    --ssh-extra-args="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
    ../../config_management/setup_weave-kube.yml

```

To specify versions of Kubernetes and Docker see Vagrant examples above.

N.B.: `--ssh-extra-args` is used to provide:

* `StrictHostKeyChecking=no`: as VMs come and go, the same IP can be used by a different machine, so checking the host's SSH key may fail. Note that this introduces a risk of a man-in-the-middle attack.
* `UserKnownHostsFile=/dev/null`: if you previously connected a VM with the same IP but a different public key, and added it to `~/.ssh/known_hosts`, SSH may still fail to connect, hence we use `/dev/null` instead of `~/.ssh/known_hosts`.

## Resources

* [https://www.vagrantup.com/docs/provisioning/ansible.html](https://www.vagrantup.com/docs/provisioning/ansible.html)
* [http://docs.ansible.com/ansible/guide_vagrant.html](http://docs.ansible.com/ansible/guide_vagrant.html)

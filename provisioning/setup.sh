#!/bin/bash
#
# Description:
#   Helper functions to programmatically provision (e.g. for CIT).
#   Aliases on these functions are also created so that this script can be
#   sourced in your shell, in your ~/.bashrc file, etc. and directly called.
#
# Usage:
#   Source this file and call the relevant functions.
#

function ssh_public_key() {
    echo -e "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDZBgLQts30PYXEMJnCU21QC+1ZE0Sv/Ry48Au3nYXn1KNoW/7C2qQ3KO2ZnpZRHCstFiU8QIlB9edi0cgcAoDWBkCiFBZEORxMvohWtrRQzf+x59o48lVjA/Fn7G+9hmavhLaDf6Qe7OhH8XUshNtnIQIUvNEWXKE75k32wUbuF8ibhJNpOOYKL4tVXK6IIKg6jR88BwGKPY/NZCl/HbhjnDJY0zCU1pZSprN6o/S953y/XXVozkh1772fCNeu4USfbt0oZOEJ57j6EWwEYIJhoeAEMAoD8ELt/bc/5iex8cuarM4Uib2JHO6WPWbBQ0NlrARIOKLrxkjjfGWarOLWBAgvwQn5zLg1pKb7aI4+jbA+ZSrII5B2HuYE9MDlU8NPL4pHrRfapGLkG/Fe9zNPvScXh+9iSWfD6G5ZoISutjiJO/iVYN0QSuj9QEIj9tl20czFz3Dhnq4sPPl5hoLunyQfajY7C/ipv6ilJyrEc0V6Z9FdPhpEI+HOgJr2vDQTFscQuyfWuzGJDZf6zPdZWo2pBql9E7piARuNAjakylGar/ebkCgfy28XQoDbDT0P0VYp+E8W5EYacx+zc5MuNhRTvbsO12fydT8V61MtA78wM/b0059feph+0zTykEHk670mYVoE3erZX+U1/BVBLSV9QzopO6/Pgx2ryriJfQ== weaveworks-cit"
}

function decrypt() {
    if [ -z "$1" ]; then
        echo >&2 "Failed to decode and decrypt $2: no secret key was provided."
        return 1
    fi
    echo "$3" | openssl base64 -d | openssl enc -d -aes256 -pass "pass:$1"
}

function ssh_private_key() {
    # The private key has been AES256-encrypted and then Base64-encoded using the following command:
    #   $ openssl enc -in /tmp/weaveworks_cit_id_rsa -e -aes256 -pass stdin | openssl base64 > /tmp/weaveworks_cit_id_rsa.aes.b64
    # The below command does the reverse, i.e. base64-decode and AES-decrypt the file, and prints it to stdout.
    # N.B.: Ask the password to Marc, or otherwise re-generate the SSH key using:
    #   $ ssh-keygen -t rsa -b 4096 -C "weaveworks-cit"
    decrypt "$1" "SSH private key" "$(
        cat <<EOF
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
EOF
    )"
}

function set_up_ssh_private_key() {
    if [ -z "$1" ]; then
        echo >&2 "Failed to decode and decrypt SSH private key: no secret key was provided."
        return 1
    fi
    local ssh_private_key_path="$HOME/.ssh/weaveworks_cit_id_rsa"
    [ -e "$ssh_private_key_path" ] && rm -f "$ssh_private_key_path"
    ssh_private_key "$1" >"$ssh_private_key_path"
    chmod 400 "$ssh_private_key_path"
    echo "$ssh_private_key_path"
}

function gcp_credentials() {
    # The below GCP service account JSON credentials have been AES256-encrypted and then Base64-encoded using the following command:
    #   $ openssl enc -in ~/.ssh/weaveworks-cit.json -e -aes256 -pass stdin | openssl base64 > /tmp/weaveworks-cit.json.aes.b64
    # The below command does the reverse, i.e. base64-decode and AES-decrypt the file, and prints it to stdout.
    # N.B.: Ask the password to Marc, or otherwise re-generate the credentials for GCP, as per ../tools/provisioning/gcp/README.md.
    decrypt "$1" "JSON credentials" "$(
        cat <<EOF
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
EOF
    )"
}

# shellcheck disable=2155
function do_on() {
    # Set up everything required to run tests on Digital Ocean.
    # Steps from ../tools/provisioning/do/README.md have been followed.
    # All sensitive files have been encrypted, see respective functions.
    if [ -z "$SECRET_KEY" ]; then
        echo >&2 "Failed to configure for Digital Ocean: no value for the SECRET_KEY environment variable."
        return 1
    fi

    # SSH public key:
    export TF_VAR_do_public_key_path="$HOME/.ssh/weaveworks_cit_id_rsa.pub"
    ssh_public_key >"$TF_VAR_do_public_key_path"
    export DIGITALOCEAN_SSH_KEY_NAME="weaveworks-cit"
    export TF_VAR_do_public_key_id=5228799

    # SSH private key:
    export TF_VAR_do_private_key_path=$(set_up_ssh_private_key "$SECRET_KEY")

    # API token:
    # The below Digital Ocean token has been AES256-encrypted and then Base64-encoded using the following command:
    #   $ openssl enc -in /tmp/digital_ocean_token.txt -e -aes256 -pass stdin | openssl base64 > /tmp/digital_ocean_token.txt.aes.b64
    # The below command does the reverse, i.e. base64-decode and AES-decrypt the file, and prints it to stdout.
    # N.B.: Ask the password to Marc, or otherwise re-generate the token for Digital Ocean, as per ../tools/provisioning/do/README.md.
    export DIGITALOCEAN_TOKEN=$(decrypt "$SECRET_KEY" "Digital Ocean token" "U2FsdGVkX1/Gq5Rj9dDDraME8xK30JOyJ9dhfQzPBaaePJHqDPIG6of71DdJW0UyFUyRtbRflCPaZ8Um1pDJpU5LoNWQk4uCApC8+xciltT73uQtttLBG8FqgFBvYIHS")
    export DIGITALOCEAN_TOKEN_NAME="weaveworks-cit"
    export TF_VAR_client_ip=$(curl -s -X GET http://checkip.amazonaws.com/)
}
alias do_on='do_on'

function do_off() {
    unset TF_VAR_do_public_key_path
    unset DIGITALOCEAN_SSH_KEY_NAME
    unset TF_VAR_do_public_key_id
    unset TF_VAR_do_private_key_path
    unset DIGITALOCEAN_TOKEN
    unset DIGITALOCEAN_TOKEN_NAME
    unset TF_VAR_client_ip
}
alias do_off='do_off'

# shellcheck disable=2155
function gcp_on() {
    # Set up everything required to run tests on GCP.
    # Steps from ../tools/provisioning/gcp/README.md have been followed.
    # All sensitive files have been encrypted, see respective functions.
    if [ -z "$SECRET_KEY" ]; then
        echo >&2 "Failed to configure for Google Cloud Platform: no value for the SECRET_KEY environment variable."
        return 1
    fi

    # SSH public key and SSH username:
    export TF_VAR_gcp_public_key_path="$HOME/.ssh/weaveworks_cit_id_rsa.pub"
    ssh_public_key >"$TF_VAR_gcp_public_key_path"
    export TF_VAR_gcp_username=$(cut -d' ' -f3 "$TF_VAR_gcp_public_key_path" | cut -d'@' -f1)

    # SSH private key:
    export TF_VAR_gcp_private_key_path=$(set_up_ssh_private_key "$SECRET_KEY")

    # JSON credentials:
    export GOOGLE_CREDENTIALS_FILE="$HOME/.ssh/weaveworks-cit.json"
    [ -e "$GOOGLE_CREDENTIALS_FILE" ] && rm -f "$GOOGLE_CREDENTIALS_FILE"
    gcp_credentials "$SECRET_KEY" >"$GOOGLE_CREDENTIALS_FILE"
    chmod 400 "$GOOGLE_CREDENTIALS_FILE"
    export GOOGLE_CREDENTIALS=$(cat "$GOOGLE_CREDENTIALS_FILE")

    export TF_VAR_client_ip=$(curl -s -X GET http://checkip.amazonaws.com/)
    export TF_VAR_gcp_project="${PROJECT:-"weave-net-tests"}"
    # shellcheck disable=2015
    [ -z "$PROJECT" ] && echo >&2 "WARNING: no value provided for PROJECT environment variable: defaulted it to $TF_VAR_gcp_project." || true
}
alias gcp_on='gcp_on'

function gcp_off() {
    unset TF_VAR_gcp_public_key_path
    unset TF_VAR_gcp_username
    unset TF_VAR_gcp_private_key_path
    unset GOOGLE_CREDENTIALS_FILE
    unset GOOGLE_CREDENTIALS
    unset TF_VAR_client_ip
    unset TF_VAR_gcp_project
}
alias gcp_off='gcp_off'

# shellcheck disable=2155
function aws_on() {
    # Set up everything required to run tests on Amazon Web Services.
    # Steps from ../tools/provisioning/aws/README.md have been followed.
    # All sensitive files have been encrypted, see respective functions.
    if [ -z "$SECRET_KEY" ]; then
        echo >&2 "Failed to configure for Amazon Web Services: no value for the SECRET_KEY environment variable."
        return 1
    fi

    # SSH public key:
    export TF_VAR_aws_public_key_name="weaveworks_cit_id_rsa"

    # SSH private key:
    export TF_VAR_aws_private_key_path=$(set_up_ssh_private_key "$SECRET_KEY")

    # The below AWS access key ID and secret access key have been AES256-encrypted and then Base64-encoded using the following commands:
    #   $ openssl enc -in /tmp/aws_access_key_id.txt     -e -aes256 -pass stdin | openssl base64 > /tmp/aws_access_key_id.txt.aes.b64
    #   $ openssl enc -in /tmp/aws_secret_access_key.txt -e -aes256 -pass stdin | openssl base64 > /tmp/aws_secret_access_key.txt.aes.b64
    # The below commands do the reverse, i.e. base64-decode and AES-decrypt the encrypted and encoded strings, and print it to stdout.
    # N.B.: Ask the password to Marc, or otherwise re-generate the AWS access key ID and secret access key, as per ../tools/provisioning/aws/README.md.
    export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID:-$(decrypt "$SECRET_KEY" "AWS access key ID" "U2FsdGVkX1+MLsvG53ZVSmFhjvQtWio0pXQpG5Ua+5JaoizuZKtJZFJxrSSyx0jb")}
    export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY:-$(decrypt "$SECRET_KEY" "AWS secret access key" "U2FsdGVkX1+VNjgWv5iGKRqBYP7o8MpOIMnd3BOYPiEho1Mjosx++9CknaZJbeR59vSuz4UdgTS6ezH2dnq2Fw==")}
    export TF_VAR_client_ip=$(curl -s -X GET http://checkip.amazonaws.com/)
}
alias aws_on='aws_on'

function aws_off() {
    unset TF_VAR_aws_public_key_name
    unset TF_VAR_aws_private_key_path
    unset AWS_ACCESS_KEY_ID
    unset AWS_SECRET_ACCESS_KEY
    unset TF_VAR_client_ip
}
alias aws_off='aws_off'

function tf_ssh_usage() {
    cat >&2 <<-EOF
ERROR: $1

Usage:
  $ tf_ssh <host ID (1-based)> [OPTION]...
Examples:
  $ tf_ssh 1
  $ tf_ssh 1 -o LogLevel VERBOSE
Available machines:
EOF
    cat -n >&2 <<<"$(terraform output public_etc_hosts)"
}

# shellcheck disable=SC2155
function tf_ssh() {
    [ -z "$1" ] && tf_ssh_usage "No host ID provided." && return 1
    local ip="$(sed "$1q;d" <<<"$(terraform output public_etc_hosts)" | cut -d ' ' -f 1)"
    shift # Drop the first argument, corresponding to the machine ID, to allow passing other arguments to SSH using "$@" -- see below.
    [ -z "$ip" ] && tf_ssh_usage "Invalid host ID provided." && return 1
    # shellcheck disable=SC2029
    ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no "$@" "$(terraform output username)@$ip"
}
alias tf_ssh='tf_ssh'

function tf_ansi_usage() {
    cat >&2 <<-EOF
ERROR: $1

Usage:
  $ tf_ansi <playbook or playbook ID (1-based)> [OPTION]...
Examples:
  $ tf_ansi setup_weave-net_dev
  $ tf_ansi 1
  $ tf_ansi 1 -vvv
Available playbooks:
EOF
    cat -n >&2 <<<"$(for file in "$(dirname "${BASH_SOURCE[0]}")"/../../config_management/*.yml; do basename "$file" | sed 's/.yml//'; done)"
}

# shellcheck disable=SC2155,SC2064
function tf_ansi() {
    [ -z "$1" ] && tf_ansi_usage "No Ansible playbook provided." && return 1
    local id="$1"
    shift # Drop the first argument to allow passing other arguments to Ansible using "$@" -- see below.
    if [[ "$id" =~ ^[0-9]+$ ]]; then
        local playbooks=(../../config_management/*.yml)
        local path="${playbooks[(($id-1))]}" # Select the ith entry in the list of playbooks (0-based).
    else
        local path="$(dirname "${BASH_SOURCE[0]}")/../../config_management/$id.yml"
    fi
    local inventory="$(mktemp /tmp/ansible_inventory_XXX)"
    trap 'rm -f $inventory' SIGINT SIGTERM RETURN
    echo -e "$(terraform output ansible_inventory)" >"$inventory"
    [ ! -r "$path" ] && tf_ansi_usage "Ansible playbook not found: $path" && return 1
    ansible-playbook "$@" -u "$(terraform output username)" -i "$inventory" --ssh-extra-args="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" "$path"
}
alias tf_ansi='tf_ansi'

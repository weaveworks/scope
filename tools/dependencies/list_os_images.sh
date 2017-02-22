#!/bin/bash

function usage() {
    cat <<EOF
Description: 
  Script to list OS images, sorted and in a Terraform-friendly format.
Dependencies:
  - gcloud, Google Cloud Platform's CLI
  - aws,
Usage: 
  $ ./$(basename "$0") PROVIDER OS
  PROVIDER={gcp}
  OS={ubuntu|debian|centos}
Example: 
  $ ./$(basename "$0") gcp ubuntu
  ubuntu-os-cloud/ubuntu-1204-lts
  ubuntu-os-cloud/ubuntu-1404-lts
  ubuntu-os-cloud/ubuntu-1604-lts
  ubuntu-os-cloud/ubuntu-1610
EOF
}

function find_aws_owner_id() {
    local os_owner_ids=(
        "ubuntu:099720109477"
        "debian:379101102735"
        "centos:679593333241"
    )
    for os_owner_id in "${os_owner_ids[@]}"; do
        os=${os_owner_id%%:*}
        owner_id=${os_owner_id#*:}
        if [ "$os" == "$1" ]; then
            echo "$owner_id"
            return 0
        fi
    done
    echo >&2 "No AWS owner ID for $1."
    exit 1
}

if [ -z "$1" ]; then
    echo >&2 "No specified provider."
    usage
    exit 1
fi

if [ -z "$2" ]; then
    if [ "$1" == "help" ]; then
        usage
        exit 0
    else
        echo >&2 "No specified operating system."
        usage
        exit 1
    fi
fi

case "$1" in
    'gcp')
        gcloud compute images list --standard-images --regexp=".*?$2.*" \
            --format="csv[no-heading][separator=/](selfLink.map().scope(projects).segment(0),family)" \
            | sort -d
        ;;
    'aws')
        aws --region "${3:-us-east-1}" ec2 describe-images \
            --owners "$(find_aws_owner_id "$2")" \
            --filters "Name=name,Values=$2*" \
            --query 'Images[*].{name:Name,id:ImageId}'
        # Other examples:
        # - CentOS: aws --region us-east-1 ec2 describe-images --owners aws-marketplace --filters Name=product-code,Values=aw0evgkw8e5c1q413zgy5pjce
        # - Debian: aws --region us-east-1 ec2 describe-images --owners 379101102735 --filters "Name=architecture,Values=x86_64" "Name=name,Values=debian-jessie-*" "Name=root-device-type,Values=ebs" "Name=virtualization-type,Values=hvm"
        ;;
    'do')
        curl -s -X GET \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
            "https://api.digitalocean.com/v2/images?page=1&per_page=999999" \
            | jq --raw-output ".images | .[] | .slug" | grep "$2" | sort -d
        ;;
    *)
        echo >&2 "Unknown provider [$1]."
        usage
        exit 1
        ;;
esac

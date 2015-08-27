#!/bin/bash

export SSH_AUTH_SOCK=

remote() {
  host=$1
  shift
  echo "[$host] Running: $@" >&2
  gcloud compute ssh $host --command "$@"
}

for h in $(gcloud compute instances list | grep -v NAME | awk '{print $1}') ; do
  cat ~/work/weave/repos/scope/scope.tar | remote $h "sudo docker load"
  cat ~/work/weave/repos/scope/scope | remote $h "sudo tee /usr/local/bin/scope >/dev/null; sudo chmod a+x /usr/local/bin/scope"
  if $(echo $h | grep -q "master") ; then
    remote $h "sudo scope stop ; sudo DOCKER_BRIDGE=cbr0 scope launch --probe.docker.bridge cbr0 --probe.kubernetes true"
  else
    remote $h "sudo scope stop ; sudo DOCKER_BRIDGE=cbr0 scope launch --no-app --probe.docker.bridge cbr0 kubernetes-master"
  fi
done

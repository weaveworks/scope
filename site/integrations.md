---
title: Integrations
menu_order: 60
---


Weave Scope integrates with the following orchestrators and cloud environments: 

 * [Launching Scope as a Service with Docker Compose](#launching-scope-as-a-service-and-docker-compose)
 * [Visualizing Kubernetes Clusters with Scope](/site/integrations/kubernetes.md)
 * [AWS Cloud Formation](/site/integrations/aws-cloud-formation.md)
 * [Installing Scope as a DC/OS Package](/site/dc-os.md)
 * [Weave Scope and Rancher Catalog](/site/integrations/rancher-scope.md)
 

###<a name="launching-scope-as-a-service-and-docker-compose"></a>Launching Scope as a Service with Docker Compose

To use Scope with Docker Compose, obtain your `SCOPE_SERVICE_TOKEN`. This is found in the console after you've logged in to the Scope service.

Version 2 of this YAML file supports networks and volumes as defined by any of the plugins you might be using. See [Compose File Reference](https://docs.docker.com/compose/compose-file/) for more information. 

Launch Docker Compose with one of the two fragments below and set the value of the Scope as a Service token in an environment variable:


    SCOPE_SERVICE_TOKEN=abcdef_my_token  docker-compose up -d


###Docker Compose Format Version 1:

    probe:
      image: weaveworks/scope:0.15
      net: "host"
      pid: "host"
      privileged: true
      labels:
        - "works.weave.role=system"
      volumes:
        - "/var/run/docker.sock:/var/run/docker.sock:rw"
      command:
        - "--probe.docker"
        - "true"
        - "--service-token"
        - "${SCOPE_SERVICE_TOKEN}"

##Docker Compose Format Version 2:

    version: '2'
    services:
      probe:
        image: weaveworks/scope:0.15
        network_mode: "host"
        pid: "host"
        privileged: true
        labels:
          - "works.weave.role=system"
        volumes:
          - "/var/run/docker.sock:/var/run/docker.sock:rw"
        command:
          - "--probe.docker"
          - "true"
          - "--service-token"
          - "${SCOPE_SERVICE_TOKEN}"

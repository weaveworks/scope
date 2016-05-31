---
title: Installing Weave Scope
menu_order: 20
---


Weave Scope consists of three parts: the probe, the app and the user interface. It can be deployed either as a standalone configuration, or if you don’t want to manage the administration yourself, you can sign up for Weave Scope as a service.

The following topics are discussed:

 * [Installing Scope on Docker](#docker)
   * [In Service Mode](#docker-service)
   * [Using Weave Net](#docker-weave)
   * [Using Docker Compose](#docker-compose)
   * [Using Docker Compose in Service Mode](#docker-compose-sercice)
 * [Installing Scope on Kubernetes](#k8s)
 * [Installing Scope on Amazon ECS](#ecs)
 * [Installing Scope on minimesos](#minimesos)
 * [Installing Scope on Mesosphere DC/OS](#dcos)

##<a name="docker"></a>Installing Scope on Docker

To install Scope on your local Docker machine in standalone mode, run the following commands:

    sudo wget -O /usr/local/bin/scope https://git.io/scope
    sudo chmod a+x /usr/local/bin/scope
    sudo scope launch

This script downloads and runs a recent Scope image from the Docker Hub. Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your browser to `http://localhost:4040`.

>**Note:** If you're using boot2docker, replace localhost with the output of boot2docker IP.

>  Scope allows anyone with access to the UI control over your containers: as such, the Scope app endpoint (port 4040) should not be made accessible on the Internet.  Additionally traffic between the app and the probe is currently insecure and should also not traverse the Internet.

###<a name="docker-service"></a>In Service Mode

To install Scope on your local Docker machine in Service Mode, run the following commands:

    sudo wget -O /usr/local/bin/scope https://git.io/scope
    sudo chmod a+x /usr/local/bin/scope
    sudo scope launch --service-token=<token>

Where `--service-token=<token>`  is the token that was sent to you when you signed up.

This script downloads and runs a recent Scope docker image from the Docker Hub. Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your web browser to [scope.weave.works](https://scope.weave.works) and login using your email.  Click on 'My Scope' in the top right-hand corner to see the Scope UI.

###<a name="docker-weave"></a> Using Weave Net

If Scope is running on the same machine as the Weave Network, then the probe uses weaveDNS to automatically discover any other apps on the network. Scope does this by registering itself under the address scope.weave.local.

Each probe sends its reports to every app registered at this address. If you have weaveDNS set up and running, no further steps are necessary.

If you don’t want to use weaveDNS, then Scope can be instructed to cluster with other Scope instances on the command line. Hostnames and IP addresses are acceptable, both with and without ports, for example:

    # scope launch scope1:4030 192.168.0.12 192.168.0.11:4030

Hostnames will be regularly resolved as A records, and each answer used as a target.

###<a name="docker-compose"></a>Using Docker Compose

To install Scope on your local Docker machine in Standalone Mode using Docker Compose, run the following commands using one of the two fragments below.

    docker-compose up -d

Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your browser to `http://localhost:4040`.

**Docker Compose Format Version 1:**

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

**Docker Compose Format Version 2:**

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

Version 2 of this YAML file supports networks and volumes as defined by any plugins you might be using. See [Compose File Reference](https://docs.docker.com/compose/compose-file/) for more information.

###<a name="docker-compose"></a>Using Docker Compose in Service Mode

To install Scope on your local Docker machine in Service Mode using Docker Compose, run the following commands using one of the two fragments below.

    SCOPE_SERVICE_TOKEN=<token>  docker-compose up -d

Where `SCOPE_SERVICE_TOKEN=<token>` is the token that was sent to you when you signed up.

Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your web browser to [scope.weave.works](https://scope.weave.works) and login using your email.  Click on 'My Scope' in the top right-hand corner to see the Scope UI.

**Docker Compose Format Version 1:**

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

**Docker Compose Format Version 2:**

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

Version 2 of this YAML file supports networks and volumes as defined by any plugins you might be using. See [Compose File Reference](https://docs.docker.com/compose/compose-file/) for more information.

##<a name="k8s"></a>Installing Scope on Kubernetes

To install Scope on a Kubernetes cluster in Standalone Mode, follow these instructions:

**Before You Begin**

Ensure that the cluster allows privileged pods - this is required by the Scope probes. By default, privileged pods are allowed from Kubernetes 1.1 and up. If you are running an earlier version or a non-default configuration, ensure that your API Server and all of your Kubelets are launched with the flag `--allow_privileged`.

Your cluster must also support [DaemonSets](https://github.com/kubernetes/kubernetes/blob/master/docs/design/daemon.md).  DaemonSets are necessary to ensure that each Kubernetes node can run a Scope Probe.

To enable DaemonSets in an existing cluster, add the `--runtime-config=extensions/v1beta1/daemonsets=true` argument to the [apiserver](https://github.com/kubernetes/kubernetes/blob/master/docs/admin/kube-apiserver.md)'s configuration. This is normally found in the `/etc/kubernetes/manifest/kube-apiserver.manifest`file after a restart of [the apiserver and controller manager](https://github.com/kubernetes/kubernetes/issues/18656) has occurred.  If you are creating a new cluster, set `KUBE_ENABLE_DAEMONSETS=true` in your cluster configuration.

>**Note:** Prior to Kubernetes version 1.2 DaemonSets would fail to schedule pods on unschedulable nodes (typically the master).  This will result in the probe not running on that node.  See [#1030](https://github.com/weaveworks/scope/issues/1030) for more information.  It is advised that you use Kubernetes version 1.2 or higher.

**Install Scope on Your Cluster**

It is recommended that you run Scope natively in your Kubernetes cluster using `kubectl` with launch URL as show below.

The simplest way to get the latest release of Scope deployed onto a Kubernetes cluster is the following:

    kubectl create -f 'https://scope.weave.works/launch/k8s/weavescope.yaml' --validate=false

> The `--validate=false` flag is currently required due to a bug in Kubernetes (see
[kubernetes/kubernetes#24089](https://github.com/kubernetes/kubernetes/issues/24089) for more details

Allowable parameters for the launcher URL:

- `v` - Weave Scope version or tag, e.g. `latest` or `0.15.0`, current release is the default
- `service-token` - Weave Scope Cloud Service token
- `k8s-service-type` - Kubernetes service type (for running Scope in Standalone mode), can be either
`LoadBalancer` or `NodePort`, by default this is unspecified (only internal access)

To download and read the Scope manifest:

    curl --silent --remote-name https://scope.weave.works/launch/k8s/weavescope.yaml

This runs a recent Scope image from the Docker Hub and will launch a probe onto every node as well as a single app. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

**Open Scope in Your Browser**

    kubectl port-forward $(kubectl get pod --selector=weavescope-component=weavescope-app -o jsonpath={.items..metadata.name}) 4040

Open http://localhost:4040 in your browser. This allows you to access the Scope UI securely, without opening it to the Internet.

##<a name="ecs"></a>Installing Scope on Amazon ECS

There are currently three options for launching Weave Scope in ECS:

* A [CloudFormation template](https://www.weave.works/deploy-weave-aws-cloudformation-template/) to launch and easily evaluate Scope directly from your browser.
* An [Amazon Machine Image (AMI)](https://github.com/weaveworks/integrations/tree/master/aws/ecs#weaves-ecs-amis) for each ECS region.
* [A simple way to tailor the AMIs to your needs](https://github.com/weaveworks/integrations/tree/master/aws/ecs#creating-your-own-customized-weave-ecs-ami).

The AWS CloudFormation template is the easiest way to get started with Weave Net and Weave Scope. CloudFormation templates provide developers and systems administrators a simple way to create a collection or a stack of related AWS resources, and it provisions and updates them in an orderly and predictable fashion.

Use this specially created Weaveworks CloudFormation template to create an EC2 instance with all of the resources you need, including Weave Net and Weave Scope.

Before launching the stack:

* [Set up an Amazon Account](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/get-set-up-for-amazon-ec2.html)
* [Create the Key Pairs]("http://docs.aws.amazon.com/gettingstarted/latest/wah/getting-started-prereq.html). You will need to reference the name of the key pairs when you create the stack.

The link below will launch a sample app using a Cloudformation template, but you can swap out the IdentiOrca app and use your own app instead.

**Ready to launch a stack?  Click here to launch a stack to AWS:**

[![](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/new?templateURL=https:%2F%2Fs3.amazonaws.com%2Fweaveworks-cfn-public%2Fintegrations%2Fecs-identiorca.json)

For step by step instructions on how to configure the stack, see: [Install Weave to AWS with One-Click](https://www.weave.works/deploy-weave-aws-cloudformation-template/)

##<a name="minimesos"></a>Installing Scope on minimesos

The [minimesos](https://github.com/ContainerSolutions/minimesos) project enables you to run an Apache Mesos cluster on a single
machine, which makes it very easy to develop Mesos frameworks.

Weave Scope is part of minimesos cluster by default, and can be accessed at `http://172.17.0.1:4040/`.

If Weave Scope is removed from your minimesos cluster, you can add it back with the following command:

```
minimesos install --marathonFile https://raw.githubusercontent.com/weaveworks/scope/master/examples/mesos/minimesos.json
```

##<a name="dcos"></a>Installing Scope as a DC/OS Package

Scope can be installed as a DC/OS Package through the open Universe.

DC/OS is short for Datacenter Operating System, a distributed operating system using Apache Mesos as its kernel. The easiest way to get start with DC/OS in the public-cloud is to [deploy it on Amazon Web Services (AWS)](https://mesosphere.com/amazon/).

For more information see, [Deploying Weave Scope on DC/OS](https://www.weave.works/guides/deploy-weave-scope-dcos/)

**See Also**

 * [Understanding Weave Scope](/site/how-it-works.md)

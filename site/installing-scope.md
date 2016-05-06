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
   * [In Service Mode](#k8s-service)
 * [Installing Scope on Amazon ECS](#ecs)
   * [In Service Mode](#ecs-service)
 * [Installing Scope on Mesosphere DC/OS](#dcos)
   * [In Service Mode](#dcos-service)

##<a name="docker"></a>Installing Scope on Docker

To install Scope on your local Docker machine in the Standalone Mode, run the following commands:

    sudo wget -O /usr/local/bin/scope https://git.io/scope
    sudo chmod a+x /usr/local/bin/scope
    sudo scope launch

This script will download and run a recent Scope image from the Docker Hub. Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your browser to `http://localhost:4040`.

>**Note:** If you're using boot2docker, replace localhost with the output of boot2docker IP.

>  Scope allows anyone with access to the UI control over your containers: as such, the Scope app endpoint (port 4040) should not be made accessible on the Internet.  Additionally traffic between the app and the probe is currently insecure and should also not traverse the Internet.

###<a name="docker-service"></a>In Service Mode

To install Scope on your local Docker machine in Service Mode, run the following commands:

    sudo wget -O /usr/local/bin/scope https://git.io/scope
    sudo chmod a+x /usr/local/bin/scope
    sudo scope launch --service-token=SCOPE_SERVICE_TOKEN

Where `--service-token=<token>`  is the token that was sent to you when you signed up.

This script will download and run a recent Scope image from the Docker Hub. Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your web browser to [scope.weave.works](https://scope.weave.works) and login using your email.  Click on 'My Scope' in the top right-hand corner to see the Scope UI.

###<a name="docker-weave">Using Weave Net

If Scope is running on the same machine as the Weave Network, then the probe uses weaveDNS to automatically discover any other apps on the network. Scope does this by registering itself under the address scope.weave.local.

Each probe sends its reports to every app registered at this address. If you have weaveDNS set up and running, no further steps are necessary.

If you don’t want to use weaveDNS, then Scope can be instructed to cluster with other Scope instances on the command line. Hostnames and IP addresses are acceptable, both with and without ports, for example:

    # scope launch scope1:4030 192.168.0.12 192.168.0.11:4030

Hostnames will be regularly resolved as A records, and each answer used as a target.

###<a name="docker-compose">Using Docker Compose

To install Scope on your local Docker machine using Docker Compose, run the following commands using one of the two fragments below.

    docker-compose up -d

Docker Compose Format Version 1:

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

Docker Compose Format Version 2:

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

To use Scope with Docker Compose, obtain your `SCOPE_SERVICE_TOKEN`. This is found in the console after you've logged in to the Scope service.

Launch Docker Compose using one of the two fragments below and then set the value of the token as an environment variable.

    SCOPE_SERVICE_TOKEN=abcdef_my_token  docker-compose up -d

Docker Compose Format Version 1:

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

Docker Compose Format Version 2:

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

Before visualizing Kubernetes clusters with Scope, it is recommended that you run Scope natively in your Kubernetes cluster using [these resource definitions](https://github.com/TheNewNormal/kube-charts/tree/master/weavescope/manifests).

**1. Ensure that the cluster allows privileged pods.**

This is required by the Scope probes. By default, privileged pods are allowed from Kubernetes 1.1 and up. If you are running an earlier version or a non-default configuration, ensure that your API Server and all of your Kubelets are launched with the flag `--allow_privileged`.

**2. Your cluster must support [DaemonSets](https://github.com/kubernetes/kubernetes/blob/master/docs/design/daemon.md).**

DaemonSets are necessary to ensure that each Kubernetes node can run a Scope Probe:

To enable them in an existing cluster, add the `--runtime-config=extensions/v1beta1/daemonsets=true` argument to the [apiserver](https://github.com/kubernetes/kubernetes/blob/master/docs/admin/kube-apiserver.md)'s configuration. This is normally found in the `/etc/kubernetes/manifest/kube-apiserver.manifest`file after a restart of [the apiserver and controller manager](https://github.com/kubernetes/kubernetes/issues/18656) has occurred.

If you are creating a new cluster, set `KUBE_ENABLE_DAEMONSETS=true` in your cluster configuration.

**3. Download the resource definitions:**

    for I in app-rc app-svc probe-ds; do
     curl -s -L https://raw.githubusercontent.com/TheNewNormal/kube-charts/master/weavescope/manifests/scope-$I.yaml -o       scope-$I.yaml
    done

**4. Tweak the Scope probe configuration in `scope-probe-ds.yaml`, namely:**

If you have an account at (scope.weave.works)[https://scope.weave.works] and want to use Scope in Cloud Service Mode, uncomment the `--probe.token=foo` argument, substitute `foo` by the token found in your account page, and comment out the `$(WEAVE_SCOPE_APP_SERVICE_HOST):$(WEAVE_SCOPE_APP_SERVICE_PORT)` argument.

**5. Finally, install Scope in your cluster (order is important):**

    kubectl create -f scope-app-rc.yaml  # Only if you want to run Scope in Standalone Mode
    kubectl create -f scope-app-svc.yaml # Only if you want to run Scope in Standalone Mode
    kubectl create -f scope-probe-ds.yaml

###<a name="k8s-service">In Service Mode

##<a name="ecs"></a>Installing Scope on Amazon ECS

We currently provide three options for launching Weave Scope in ECS:

* A [CloudFormation template](https://www.weave.works/deploy-weave-aws-cloudformation-template/) to launch and easily evaluate Scope directly from your browser.
* An [Amazon Machine Image (AMI)](https://github.com/weaveworks/integrations/tree/master/aws/ecs#weaves-ecs-amis) for each ECS region.
* [A simple way to tailor the AMIs to your needs](https://github.com/weaveworks/integrations/tree/master/aws/ecs#creating-your-own-customized-weave-ecs-ami).

The AWS CloudFormation template is the easiest way to get started with [Weave Net] and [Weave Scope]. CloudFormation templates provide developers and systems administrators a simple way to create a collection or a stack of related AWS resources, and it provisions and updates them in an orderly and predictable fashion.

Use this specially created Weaveworks CloudFormation template to create an EC2 instance with all of the resources you need, including Weave Net and Weave Scope.

Before launching the stack:

* [Set up an Amazon Account](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/get-set-up-for-amazon-ec2.html)
* [Create the Key Pairs]("http://docs.aws.amazon.com/gettingstarted/latest/wah/getting-started-prereq.html). You will need to reference the name of the key pairs when you create the stack.

The link below will launch a sample app using a Cloudformation template, but you can swap out the IdentiOrca app and use your own app instead.

**Ready to launch a stack?  Click here to launch a stack to AWS:**

[![](images/cloudformation-launch-stack.png)](https://console.aws.amazon.com/cloudformation/home#/stacks/new?templateURL=https:%2F%2Fs3.amazonaws.com%2Fweaveworks-cfn-public%2Fintegrations%2Fecs-identiorca.json)


For step by step instructions on how to configure the stack, see: [Install Weave to AWS with One-Click](https://www.weave.works/deploy-weave-aws-cloudformation-template/)

###<a name="ecs-service">In Service Mode

##<a name="installing-scope-as-a-dc/os-package"></a>Installing Scope as a DC/OS Package

Scope provides a simple live map of your application without requiring any instrumentation or coding, and can be installed as a DC/OS Package through the open Universe.

DC/OS is short for Datacenter Operating System, a distributed operating system using Apache Mesos as its kernel. The easiest way to get start with DC/OS in the public-cloud is to [deploy it on Amazon Web Services (AWS)](https://mesosphere.com/amazon/).

How to install the Weave Scope packages on DC/OS and where it can be useful to you while developing a distributed application that targets DC/OS is described.

For more information see, [Deploying Weave Scope on DC/OS](https://www.weave.works/guides/deploy-weave-scope-dcos/)

###<a name="dcos-service">In Service Mode

**See Also**

 * [Understanding Weave Scope](/site/how-it-works.md)

---
title: Installing Weave Scope
menu_order: 20
---


Weave Scope consists of three parts: the probe, the app and the user interface.  Scope can be deployed in either a standalone configuration, where you run everything yourself, or by using Weave Cloud.

Weave Cloud is the recommended option if:

 * You are deploying to larger clusters.
 * You require secure remote access.
 * You wish to share access with your coworkers.

The following topics are discussed:

 * [Installing Scope on Docker](#docker)
   * [Using Weave Cloud](#docker-weave-cloud)
   * [On A Local Cluster](#local-cluster)
   * [Weave Net and Scope](#net-scope)
   * [Using Docker Compose](#docker-compose)
   * [Using Docker Compose in Weave Cloud](#docker-compose-cloud)
 * [Installing Scope on Kubernetes](#k8s)
   * [Using Weave Cloud](#k8s-weave-cloud)
 * [Installing Scope on Amazon ECS](#ecs)
 * [Installing Scope on minimesos](#minimesos)
 * [Installing Scope on Mesosphere DC/OS](#dcos)

##<a name="docker"></a>Installing Scope on Docker

To install Scope onto your PC with docker installed in standalone mode, run the following commands:

    sudo curl -L git.io/scope -o /usr/local/bin/scope
    sudo chmod a+x /usr/local/bin/scope
    scope launch

This script downloads and runs a recent Scope image from the Docker Hub. Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your browser to `http://localhost:4040`. The URLs are also displayed to you in the terminal window after Weave Scope was launched. 

If you are using docker-machine, you can find the IP by running, `docker-machine ip <VM name>`.

Where, 

 * `<VM name>` is the name you gave to your virtual machine with docker-machine.

#### NB: Scope allows anyone with access to the UI control over your containers, and the hosts running them!

Therefore, the Scope app endpoint (port 4040) should not be made accessible on the Internet.  Additionally, traffic between the app and the probe is currently insecure and should also not traverse the Internet, meaning that you should either use the private / internal IP addresses of your nodes when setting it up, or route this traffic through weave net.  To put scope behind a very simple password, you can use [Caddy](https://github.com/mholt/caddy) to protect the endpoint by making the port 4040 available to localhost and using caddy to proxy it.... or, just use weave cloud.  

###<a name="docker-weave-cloud"></a>Using Weave Cloud

First, obtain a service-token from Weave Cloud by signing up at [https://cloud.weave.works](https://cloud.weave.works/).

Then, install Scope onto your local Docker machine and start using Weave Cloud:

    sudo curl -L git.io/scope -o /usr/local/bin/scope
    sudo chmod a+x /usr/local/bin/scope
    scope launch --service-token=<token>

Where,

* `--service-token=<token>` is the token you obtained after you signed up for Weave Cloud.

This script downloads and runs a recent Scope docker image from the Docker Hub. Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After Scope has been launched, open your web browser to [https://cloud.weave.works](https://cloud.weave.works) and login. Click 'View Instance' in the top right-hand corner to see the Scope user interface.

###<a name="net-scope"></a> Local-Cluster without WeaveNet / WeaveDNS
These are the directions for "Any Random Cluster," with no dependency on Weave Net.  Suppose I have a cluster like this:

192.168.100.16
192.168.100.17
192.168.100.18
192.168.100.19
192.168.100.20

We'll assume that no special hostnames or DNS settings have been fiddled with, so we're going to use IP addresses only to configure scope on this cluster:

**On Each Node, first run:**
```
    sudo curl -L git.io/scope -o /usr/local/bin/scope
    sudo chmod a+x /usr/local/bin/scope
```
**Then Run on the first node**
```
    scope launch 192.168.100.18 192.168.100.19 192.168.100.20
```
**Second node**
```
    scope launch 192.168.100.17 192.168.100.20 192.168.100.21
```
**Third Node**
```
    scope launch 192.168.100.17 192.168.100.18 192.168.100.21
```
**Fourth Node**
```
    scope launch 192.168.100.17 192.198.100.19 192.168.100.20
```
Key point here being that you need to make each node aware of the others.  You can also configure a single "target" nodee that you point all the others at, and once again, the key is that they know each other exist.  

###<a name="net-scope"></a> Weave Net and Scope

If Scope is running on the same machine as the Weave Network, then the probe uses weaveDNS to automatically discover any other apps on the network. Scope does this by registering itself under the address `scope.weave.local`.

Each probe sends its reports to every app registered at this address. If you have weaveDNS set up and running, no further steps are required.

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

###<a name="docker-compose-cloud"></a>Using Docker Compose with Weave Cloud

First, obtain a service-token from Weave Cloud by signing up at [https://cloud.weave.works](https://cloud.weave.works/).

To install Scope on your local Docker machine with Weave Cloud and use Docker Compose, run the following commands, using one of the two fragments below.

    SCOPE_SERVICE_TOKEN=<token>  docker-compose up -d

Where,

* `SCOPE_SERVICE_TOKEN=<token>` is the token you obtained after you signed up for Weave Cloud.

Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your web browser to [https://cloud.weave.works](https://cloud.weave.works) and login.  Click 'View Instance' in the top right-hand corner to see the Scope user interface.

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

After Scope has been launched, open your web browser to [https://cloud.weave.works](https://cloud.weave.works) and login.  Click 'View Instance' in the top right-hand corner to see the Scope user interface.

##<a name="k8s"></a>Installing Scope on Kubernetes

To install Scope onto a Kubernetes cluster in Standalone Mode, follow these instructions.

**Before You Begin**

* Ensure that the cluster allows privileged pods - this is required by the Scope probes. By default, privileged pods are allowed from Kubernetes 1.1 and up. If you are running an earlier version or a non-default configuration, ensure that your API Server and all of your Kubelets are launched with the flag `--allow_privileged`.

* The version of [kubectl](http://kubernetes.io/docs/user-guide/kubectl-overview/) must match the cluster version that you are using.

* The cluster must support [DaemonSets](https://github.com/kubernetes/kubernetes/blob/master/docs/design/daemon.md).  DaemonSets are necessary to ensure that each Kubernetes node can run a Scope Probe.

To enable DaemonSets in an existing cluster, add the `--runtime-config=extensions/v1beta1/daemonsets=true` argument to the [apiserver](https://github.com/kubernetes/kubernetes/blob/master/docs/admin/kube-apiserver.md)'s configuration. This is normally found in the `/etc/kubernetes/manifest/kube-apiserver.manifest`file after a restart of [the apiserver and controller manager](https://github.com/kubernetes/kubernetes/issues/18656) has occurred.  If you are creating a new cluster, set `KUBE_ENABLE_DAEMONSETS=true` in your cluster configuration.

> **Note:** Prior to Kubernetes version 1.2 DaemonSets would fail to schedule pods on unschedulable nodes (typically the master).  This will result in the probe not running on that node.  See [#1030](https://github.com/weaveworks/scope/issues/1030) for more information.  It is advised that you use Kubernetes version 1.2 or higher.

**Install Scope on Your Cluster**

You can run Scope natively in your Kubernetes cluster using `kubectl` with a launch URL as show below.

The simplest way to get the latest release of Scope deployed onto a Kubernetes cluster is by running the following:

    kubectl create -f 'https://cloud.weave.works/launch/k8s/weavescope.yaml' --validate=false

This runs a recent Scope image from the Docker Hub and launches a probe onto every node as well as a single app. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

Allowable parameters for the launcher URL:

- `v` - Weave Scope version or tag, e.g. `latest` current release is the default
- `k8s-service-type` - Kubernetes service type (for running Scope in Standalone mode), can be either
`LoadBalancer` or `NodePort`, by default this is unspecified (only internal access)

> **Note:**The `--validate=false` flag is currently required due to a bug in Kubernetes (see
[kubernetes/kubernetes#24089](https://github.com/kubernetes/kubernetes/issues/24089) for more details

To download and read the Scope manifest run:

    curl --silent --remote-name https://cloud.weave.works/launch/k8s/weavescope.yaml

**Open Scope in Your Browser**

    kubectl port-forward $(kubectl get pod --selector=weavescope-component=weavescope-app -o jsonpath='{.items..metadata.name}') 4040

Open http://localhost:4040 in your browser. This allows you to access the Scope UI securely, without opening it to the Internet.

###<a name="k8s-weave-cloud"></a>Using Weave Cloud

**Before You Begin**

* All the prerequisites from above apply.

* Obtain a service-token from Weave Cloud by signing up at [https://cloud.weave.works](https://cloud.weave.works/).

**Install Scope on Your Cluster**

To get the latest release of Scope deployed onto a Kubernetes cluster, run the following:

    kubectl create -f 'https://cloud.weave.works/launch/k8s/weavescope.yaml?servie-token=<token>' --validate=false

Where,

* `service-token=<token>` is the token you obtained after you signed up for Weave Cloud.

> **Note:**The `--validate=false` flag is currently required due to a bug in Kubernetes (see
[kubernetes/kubernetes#24089](https://github.com/kubernetes/kubernetes/issues/24089) for details.

This runs a recent Scope image from the Docker Hub and launches a probe onto every node, pointing them at Weave Cloud. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

**Open Scope in Your Browser**

After Scope has been launched, open your web browser to [https://cloud.weave.works](https://cloud.weave.works) and login.  Click 'View Instance' in the top right-hand corner to see the Scope user interface.

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

The [minimesos](https://github.com/ContainerSolutions/minimesos) project enables you to run an Apache Mesos cluster on a single machine, which makes it very easy to develop Mesos frameworks.

By default, Weave Scope is included in the minimesos cluster, and can be accessed at `http://172.17.0.1:4040/`.

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

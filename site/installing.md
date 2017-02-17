---
title: Installing Weave Scope
menu_order: 20
---


Weave Scope consists of three parts: the probe, the app and the user interface.  Scope can be deployed in either a standalone configuration, where you run everything yourself, or you can use Weave Cloud.

Weave Cloud is the recommended option if:

 * You are deploying to larger clusters.
 * You require secure remote access.
 * You want to share access with your coworkers.
 * You want to minimize Weave Scope memory and CPU usage.

The following topics are discussed:

 * [Installing Scope on Docker](#docker)
   * [Using Weave Cloud](#docker-weave-cloud)
   * [Installing Scope on a Local Cluster Without Weave Net](#cluster-no-net)
   * [Weave Net and Scope](#net-scope)
   * [Using Docker Compose](#docker-compose)
   * [Using Docker Compose in Weave Cloud](#docker-compose-cloud)
 * [Installing Scope on Kubernetes](#k8s)
   * [With Weave Cloud](#k8s-weave-cloud)
   * [Without Weave Cloud](#k8s-standalone)   
 * [Installing Scope on Amazon ECS](#ecs)
 * [Installing Scope on minimesos](#minimesos)
 * [Installing Scope on Mesosphere DC/OS](#dcos)

##<a name="docker"></a>Installing Scope on Docker

To install Scope in stand-alone mode, run the following commands:

    sudo curl -L git.io/scope -o /usr/local/bin/scope
    sudo chmod a+x /usr/local/bin/scope
    scope launch

This script downloads and runs a recent Scope image from Docker Hub. Scope needs to be installed onto every machine that you want to monitor.

After Scope is installed, open your browser to `http://localhost:4040`.  

If you are using docker-machine, you can find the IP by running, `docker-machine ip <VM name>`.

Where,

 * `<VM name>` is the name you gave to your virtual machine with docker-machine.

>>**Note:** Scope allows anyone with access to the user interface, control over your containers. As such, the Scope app endpoint (port 4040) should not be made accessible on the Internet.  Also traffic between the app and the probe is insecure and should not traverse the Internet. This means that you should either use the private / internal IP addresses of your nodes when setting it up, or route this traffic through Weave Net.  Put Scope behind a password, by using somthing like [Caddy](https://github.com/mholt/caddy) to protect the endpoint and make port 4040 available to localhost with Caddy proxying it. Or you can skip these steps, and just use Weave Cloud to manage the security for you.

###<a name="docker-weave-cloud"></a>Using Weave Cloud

First, obtain a Weave Cloud token by signing up at [https://cloud.weave.works](https://cloud.weave.works/).

!['Weave Cloud Token'](images/weave-cloud-token.png)

Launch Scope onto your nodes:

    sudo curl -L git.io/scope -o /usr/local/bin/scope
    sudo chmod a+x /usr/local/bin/scope
    scope launch --service-token=<token>

Where,

* `--service-token=<token>` is the token you obtained from Weave Cloud.

This script downloads and runs a recent Scope Docker image from the Docker Hub. Scope needs to be installed onto every machine that you want to monitor.

Open your web browser to [https://cloud.weave.works](https://cloud.weave.works) and login. Click 'View Instance' to see the Scope user interface.


###<a name="cluster-no-net"></a>Installing Scope on a Local Cluster Without Weave Net

This example assumes a local cluster that is not networked with Weave Net, and also has no special hostnames or DNS settings. You will launch Scope with the IP addresses of all of the nodes in the cluster.

Suppose you have the following nodes in a cluster:

 192.168.100.16
 192.168.100.17
 192.168.100.18
 192.168.100.19
 192.168.100.20

 Using the above IP addresses, you will manually peer each node with all of the other nodes during Scope launch:

**1. To begin run the following on each node:**

    sudo curl -L git.io/scope -o /usr/local/bin/scope
    sudo chmod a+x /usr/local/bin/scope

 **2. Then on the first node launch scope with:**

    scope launch 192.168.100.18 192.168.100.19 192.168.100.20

 **3. And do the same for all of the other nodes in your cluster:**

    scope launch 192.168.100.17 192.168.100.20 192.168.100.21
    scope launch 192.168.100.17 192.168.100.18 192.168.100.21
    scope launch 192.168.100.17 192.198.100.19 192.168.100.20


###<a name="net-scope"></a> Weave Net and Scope

If Scope is running on the same machine as the Weave Network, then the probe uses weaveDNS to automatically discover any other apps on the network. Scope does this by registering itself under the address `scope.weave.local`.

Each probe sends its reports to every app registered at this address. If you have weaveDNS set up and running, no further steps are required.

If you don’t want to use weaveDNS, then Scope can be instructed to cluster with other Scope instances on the command line. Hostnames and IP addresses are acceptable, both with and without ports, for example:

    scope launch scope1:4030 192.168.0.12 192.168.0.11:4030

Hostnames will be regularly resolved as A records, and each answer used as a target.

###<a name="docker-compose"></a>Using Docker Compose

To install Scope on your local Docker machine in Standalone Mode using Docker Compose, run the following commands using one of the two fragments below.

    docker-compose up -d

Scope needs to be installed onto every machine that you want to monitor. Once launched, Scope doesn’t require any other configuration and it also doesn’t depend on Weave Net.

After it’s been launched, open your browser to `http://localhost:4040`.

**Docker Compose Format Version 1:**

    scope:
      image: weaveworks/scope:1.2.1
      net: "host"
      pid: "host"
      privileged: true
      labels:
        - "works.weave.role=system"
      volumes:
        - "/var/run/docker.sock:/var/run/docker.sock:rw"
      command:
        - "--probe.docker=true"

**Docker Compose Format Version 2:**

    version: '2'
    services:
      scope:
        image: weaveworks/scope:1.2.1
        network_mode: "host"
        pid: "host"
        privileged: true
        labels:
          - "works.weave.role=system"
        volumes:
          - "/var/run/docker.sock:/var/run/docker.sock:rw"
        command:
          - "--probe.docker=true"

Version 2 of this YAML file supports networks and volumes as defined by any plugins you might be using. See [Compose File Reference](https://docs.docker.com/compose/compose-file/) for more information.

###<a name="docker-compose-cloud"></a>Using Docker Compose with Weave Cloud

**1.** First, obtain a Cloud token from Weave Cloud by signing up at [https://cloud.weave.works](https://cloud.weave.works/).

!['Weave Cloud Token'](images/weave-cloud-token.png)

**2.** Decide on which format version to use:

**Docker Compose Format Version 1:**

    probe:
      image: weaveworks/scope:1.2.1
      net: "host"
      pid: "host"
      privileged: true
      labels:
        - "works.weave.role=system"
      volumes:
        - "/var/run/docker.sock:/var/run/docker.sock:rw"
      command:
        - "--probe.docker=true"
        - "--service-token=${SCOPE_SERVICE_TOKEN}"

**Docker Compose Format Version 2:**

    version: '2'
    services:
      probe:
        image: weaveworks/scope:1.2.1
        network_mode: "host"
        pid: "host"
        privileged: true
        labels:
          - "works.weave.role=system"
        volumes:
          - "/var/run/docker.sock:/var/run/docker.sock:rw"
        command:
          - "--probe.docker=true"
          - "--service-token=${SCOPE_SERVICE_TOKEN}"

Version 2 of this YAML file supports a network mode that may be required with any Docker plugins that you are using. See [Compose File Reference](https://docs.docker.com/compose/compose-file/) for more information.

**3.** Once you've set up a compose file, launch Scope onto your local machine:

    SCOPE_SERVICE_TOKEN=<token>  docker-compose up -d

Where,

* `SCOPE_SERVICE_TOKEN=<token>` is the token you obtained after you signed up for Weave Cloud.

Note that you will need to launch Scope onto every node that you want to monitor.

**4.** Go to [https://cloud.weave.works](https://cloud.weave.works) and click 'View Instance'.

##<a name="k8s"></a>Installing Scope on Kubernetes

###<a name="k8s-weave-cloud"></a>With Weave Cloud (recommended)

Weave Cloud hosts the Scope UI for you, provides secure access control for your team and saves resources such as CPU or memory usage.

Sign up for a [Weave Cloud account](https://cloud.weave.works/) and obtain a token. Replace `<token>` with your token by running this on the master node or on whatever machine that has `kubectl` configured to authenticate to your Kubernetes cluster:

    kubectl apply --namespace kube-system -f 'https://cloud.weave.works/k8s/scope.yaml?service-token=<token>'

**SECURITY NOTE: This allows control of your Kubernetes cluster from Weave Cloud, which is a hosted service.**

###<a name="k8s-standalone"></a>Without Weave Cloud (run Scope in standalone mode)

The simplest way to get the latest release of Scope deployed onto a Kubernetes cluster is by running the following:

    kubectl apply --namespace kube-system -f 'https://cloud.weave.works/k8s/scope.yaml'

This runs a recent Scope image from Dockerhub and launches a probe onto every node as well as a single Scope app. Once launched, Scope doesn’t require any other configuration.

Allowable parameters for the launcher URL:

- `v` - Weave Scope version or tag, e.g. `latest` current release is the default
- `k8s-service-type` - Kubernetes service type (for running Scope in Standalone mode), can be either
`LoadBalancer` or `NodePort`, by default this is unspecified (only internal access)

To download and read the Scope manifest run:

    curl --silent --location --remote-name https://cloud.weave.works/k8s/scope.yaml

**Open Scope in Your Browser**

    kubectl port-forward -n kube-system "$(kubectl get -n kube-system pod --selector=weave-scope-component=app -o jsonpath='{.items..metadata.name}')" 4040

The URL is: http://localhost:4040.

##<a name="ose"></a>Installing Scope on OpenShift

To install Weave Scope on OpenShift, you first need to login as `admin` user:

    oc login -u admin

Next, create a dedicated project for Weave Scope and apply policy permissions to allow Weave Scope to access Kubernetes API:

    oc new-project weave-scope
    oc adm policy add-scc-to-user privileged system:serviceaccount:weave-scope:default
    oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:weave-scope:default

The installation method for Scope on OpenShift is very similar to the one described above for Kubernetes, but instead of `kubectl apply -n kube-system ...` you need to use
`oc apply ...` and install it into the namespace of the `weave-scope` project you have just created, and not the `kube-system` namesapce, i.e.:

#### With Weave Cloud (recommended)

    oc apply -f 'https://cloud.weave.works/k8s/scope.yaml?service-token=<token>'


#### Without Weave Cloud (run Scope in standalone mode)

    oc apply -f 'https://cloud.weave.works/k8s/scope.yaml'

To access standalone Scope app from the browser, please refer to Kubernetes instructions above.

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

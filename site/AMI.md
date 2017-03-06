---
title: Weaveworks ECS AMIs
menu_order: 25
---

# Weaveworks' ECS AMIs

To make [Weave Net](http://weave.works/net) and
[Weave Scope](http://weave.works/scope) easier to use with
[Amazon ECS](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html),
a set of Amazon Machine Images (AMIs) are provided. These AMIs are fully
compatible with the
[ECS-Optimized Amazon Linux AMI](https://aws.amazon.com/marketplace/pp/B00U6QTYI2).

These are the latest supported Weave AMIs for each region:

<!--- This table is machine-parsed by
https://github.com/weaveworks/guides/blob/master/aws-ecs/setup.sh, please do
not remove it and respect the format! -->

| Region         | AMI          |
|----------------|--------------|
| us-east-1      | ami-c63709d1 |
| us-east-2      | ami-4788d222 |
| us-west-1      | ami-28e7b348 |
| us-west-2      | ami-c62f81a6 |
| eu-west-1      | ami-25adf356 |
| eu-central-1   | ami-7fd31410 |
| ap-northeast-1 | ami-cb1eafaa |
| ap-southeast-1 | ami-35822f56 |
| ap-southeast-2 | ami-06300965 |

For more information about Weave AMIs and running them see: 


 * [What's in the Weave ECS AMIs?](#whats-in-ecs-ami)
 * [Deployment Requirements](#deployment-requirements)
  * [Required Open Ports](#required-open-ports)
  * [Additional IAM Action Permissions](#additional-permissions)
  * [Requirements for Peer Discovery](#requirements-for-peer-discovery)
 * [Peer Discovery with Weave Net](#peer-discovery-weave-net)
 * [How to Run Weave Scope](#how-to-run-weave-scope)
  * [Standalone mode](#running-weave-scope-in-standalone-mode)
  * [In Weave Cloud](#running-weave-scope-in-weave-cloud)
 * [Upgrading Weave Scope and Weave Net](#upgrading-weave-scope-and-weave-net)


## <a name="whats-in-ecs-ami"></a>What's in the Weave ECS AMIs?

The latest Weave ECS AMIs are based on Amazon's
[ECS-Optimized Amazon Linux AMI](https://aws.amazon.com/marketplace/pp/B00U6QTYI2),
version `2016.09.b` and also includes:

* [Weave Net 1.8.1](https://github.com/weaveworks/weave/blob/master/CHANGELOG.md#release-181)
* [Weave Scope 1.1.0](https://github.com/weaveworks/scope/blob/master/CHANGELOG.md#release-110)


## <a name="deployment-requirements"></a>Deployment Requirements

### <a name="required-open-ports"></a> Required Open Ports

For `Weave Net` to function properly, ensure that the Amazon ECS container
instances can communicate over these ports: TCP 6783, as well as, UDP 6783 and
UDP 6784.

In addition to those open ports, launching `Weave Scope` in [standalone mode](#running-weave-scope-in-standalone-mode),
requires that all instances are able to communicate over TCP port 4040. More information about
this can be found in [How to Run Weave Scope](#how-to-run-weave-scope).

See the
[relevant section of the `setup.sh`](https://github.com/weaveworks/guides/blob/c2d25d4cfd766ca739444eea06fefc57aa7a59ff/aws-ecs/setup.sh#L115-L120)
script from
[Service Discovery and Load Balancing with Weave on Amazon ECS](http://weave.works/guides/service-discovery-with-weave-aws-ecs.html)
for an example.

### <a name="additional-permissions"></a>Additional IAM Action Permissions

Besides the customary Amazon ECS API actions required by all container instances
(see the [`AmazonEC2ContainerServiceforEC2Role`](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/instance_IAM_role.html) managed policy), any instances using the Weaveworks ECS AMI must also be allowed to perform the following actions:

1. `ec2:DescribeInstances`
2. `ec2:DescribeTags`
3. `autoscaling:DescribeAutoScalingInstances`
4. `ecs:ListServices`
5. `ecs:DescribeTasks`
6. `ecs:DescribeServices`

These extra actions are needed for discovering instance peers (1,2,3) and
creating the ECS views in Weave Scope
(4,5,6). [`weave-ecs-policy.json`](https://github.com/weaveworks/guides/blob/41f1f5a60d39d39b78f0e06e224a7c3bad30c4e8/aws-ecs/data/weave-ecs-policy.json#L16-L18)
(from the
[Weaveworks ECS guide](http://weave.works/guides/service-discovery-with-weave-aws-ecs.html)),
describes the minimal policy definition.

For more information on IAM policies see
[IAM Policies for Amazon EC2](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-policies-for-amazon-ec2.html).

### <a name="requirements-for-peer-discovery"></a>Requirements for Peer Discovery

To form a Weave network, the Amazon ECS container instances must either/or:
* be a member of an
[Auto Scaling Group](http://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/AutoScalingGroup.html).
* have a [tag](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html) with key `weave:peerGroupName`.

## <a name="peer-discovery-weave-net"></a>Peer Discovery with Weave Net

At boot time, an instance running the ECS Weave AMI will try to join other instances to form a Weave network.

* If the instance has a
  [tag](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Using_Tags.html)
  with key `weave:peerGroupName`, it will join other instances with the same tag key and value.
  For instance, if the tag key is `weave:peerGroupName` and the value is `foo` it will try
  to join other instances with tag key `weave:peerGroupName` and tag value `foo`.
  Note that for this to work, the instances need to be tagged at creation-time so that
  the tag is available by the time Weave is launched.
* Otherwise it will join all the other instances in the same
  [Auto Scaling Group](http://docs.aws.amazon.com/AutoScaling/latest/DeveloperGuide/AutoScalingGroup.html).

When running `Weave Scope` in Standalone mode, probes discover apps with the same mechanism.

## <a name="how-to-run-weave-scope"></a>How to Run Weave Scope

There are two methods for running `Weave Scope` within the Weave ECS AMIs:

* [Standalone mode](#running-weave-scope-in-standalone-mode)
* [In Weave Cloud](#running-weave-scope-in-weave-cloud)

You can prevent Weave Scope from automatically starting at boot time by creating
file `/etc/init/scope.override` with contents `manual`.

This can be done at instance initialization time adding the following line to
the
[User Data](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html#user-data-shell-scripts)
of the instance.

~~~bash
echo manual > /etc/weave/scope.override
~~~

###<a name="running-weave-scope-in-standalone-mode"></a>Running `Weave Scope` in Standalone Mode

Running `Weave Scope` in standalone mode is the default mode.

The following occurs on all Amazon ECS container instances:

1. A `Weave Scope` probe is launched that collects instance information.
2. A `Weave Scope` app runs that enables cluster visualization.

Since all instances run an app and show the same information, you don't have to
worry about placing the app, thereby eliminating a
[*Leader election problem*](https://en.wikipedia.org/wiki/Leader_election).

However, running the app on all instances impacts performance, resulting in `N *
N = N^2` connections in the Auto Scaling Group with N instances (i.e. all (N)
probes talk to all (N) apps in every instances). 

To avoid this problem, it is recommended that you run `Weave Scope` in [Weave Cloud](https://cloud.weave.works).

The `Weave Scope` app runs a web-based application, which listens on TCP port
4040 where you can connect with your browser.

`Weave Scope` probes also forward information to the apps on TCP
port 4040. Ensure that your Amazon ECS container instances can talk to each
other on that port before running `Weave Scope` in standalone mode (see
[Required Open Ports](#required-open-ports) for more details).

###<a name="running-weave-scope-in-weave-cloud"></a>Running `Weave Scope` in Weave Cloud

In Weave Cloud, you can visualize Amazon ECS containers as well as monitor Tasks 
and Services all from within in Weave Cloud at [https://cloud.weave.works](https://cloud.weave.works). 
In this case, Amazon ECS container instances run a `Weave Scope` probe and reports
data from the container instances to [Weave Cloud](http://cloud.weave.works).

To configure your ECS container instances to communicate with Weave Cloud,
store the `Weave Scope` cloud token in the`/etc/weave/scope.config`
file.

>Note: The `Weave Scope` cloud token can be found in your Weave Cloud account at [http://cloud.weave.works](http://cloud.weave.works).

For example, this command configures the instance to communicate with Weave
Cloud using token `3hud3h6ys3jhg9bq66n8xxa4b147dt5z`.

~~~bash
echo SERVICE_TOKEN=3hud3h6ys3jhg9bq66n8xxa4b147dt5z >> /etc/weave/scope.config
~~~

You can do this at instance-initialization time using
[User Data](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html#user-data-shell-scripts),
which is similar to how
[ECS Cluster Mapping is configured](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/launch_container_instance.html#instance-launch-user-data-step).

## <a name="upgrading-weave-scope-and-weave-net"></a>Upgrading Weave Scope and Weave Net

The AMIs are updated regularly (~monthly) to include the latest versions of Weave Net and Weave Scope. However, it is possible to upgrade Weave Net and Weave Scope in your running EC2 instances without needing to wait for a new AMI release or by rebuilding your cluster. 

In order to upgrade Scope to the latest released version, run the following commands in each of your instances:

~~~bash
sudo curl -L git.io/scope -o /usr/local/bin/scope
sudo chmod a+x /usr/local/bin/scope
sudo stop scope
sudo start scope
~~~

Upgrade Weave Net to the latest version by running the following commands in each of your instances:


~~~bash
sudo curl -L git.io/weave -o /usr/local/bin/weave
sudo chmod a+x /usr/local/bin/weave
sudo stop weave
sudo start weave
~~~


<!--- Do not change the title, otherwise links to
https://github.com/weaveworks/integrations/tree/master/aws/ecs#creating-your-own-customized-weave-ecs-ami
will break (e.g. from the ECS guide) -->
## Creating Your Own Customized Weave ECS AMI

Clone the integrations repository and then change to the `packer` directory.

~~~bash
git clone https://github.com/weaveworks/integrations
cd aws/ecs/packer
~~~

Download and install [Packer](https://www.packer.io/) version >=0.9 to build the AMI.

Finally, invoke `./build-all-amis.sh` to build the `Weave ECS` images for all
regions. This step installs (in the image) AWS-CLI, jq, Weave Net, Weave Scope, init scripts
for `Weave` and it also updates the ECS agent to use the `Weave Docker API Proxy`.

Customize the image by modifying `template.json` to match your
requirements.

~~~bash
AWS_ACCSS_KEY_ID=XXXX AWS_SECRET_ACCESS_KEY=YYYY  ./build-all-amis.sh
~~~

If building an AMI for a particular region, set the `ONLY_REGION` variable to
that region when invoking the script:

~~~bash
ONLY_REGION=us-east-1 AWS_ACCSS_KEY_ID=XXXX AWS_SECRET_ACCESS_KEY=YYYY  ./build-all-amis.sh
~~~

##Further Reading

Read the
[Service Discovery and Load Balancing with Weave on Amazon ECS](http://weave.works/guides/service-discovery-with-weave-aws-ecs.html)
guide for more information about the AMIs.


**See Also**

 * [Installing Weave Scope](/site/installing.md)
 * [Service Discovery and Load Balancing with Weave on Amazon ECS](http://weave.works/guides/service-discovery-with-weave-aws-ecs.html)
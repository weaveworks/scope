# Scope

[![Circle CI](https://circleci.com/gh/weaveworks/scope/tree/master.svg?style=shield)](https://circleci.com/gh/weaveworks/scope/tree/master) [![Coverage Status](https://coveralls.io/repos/weaveworks/scope/badge.svg)](https://coveralls.io/r/weaveworks/scope) [![Sourcegraph](https://sourcegraph.com/api/repos/github.com/weaveworks/scope/.badges/status.svg)](https://sourcegraph.com/github.com/weaveworks/scope)

![Weave Scope Screenshot](http://blog.weave.works/wp-content/uploads/2015/11/0.10-Deailts-Panel-2-1.png)


## <a name="overview"></a>Overview

Weave Scope automatically generates a map of your containers, enabling you to
intuitively understand, monitor, and control your applications.

## <a name="getting-started"></a>Getting started

```
sudo wget -O /usr/local/bin/scope \
  https://github.com/weaveworks/scope/releases/download/latest_release/scope
sudo chmod a+x /usr/local/bin/scope
sudo scope launch
```

This script will download and run a recent Scope image from the Docker Hub.
Now, open your web browser to **http://localhost:4040**. (If you're using
boot2docker, replace localhost with the output of `boot2docker ip`.)

## <a name="help"></a>Getting help

If you have any questions about, feedback for or problem with Scope we invite
you to:
- send an email to <a href="mailto:weave-users@weave.works">weave-users@weave.works</a>
- <a href="https://github.com/weaveworks/scope/issues/new">file an issue</a>

Your feedback is always welcome!

## <a name="requirements"></a>Requirements

Scope does not need any configuration and does not require the Weave Network.
Scope does need to be running on every machine you want to monitor.

Scope allows anyone with access to the UI control over your containers: as
such, the Scope app endpoint (port 4040) should not be made accessible on
the Internet.  Additionally traffic between the app and the probe is currently
insecure and should not traverse the internet.

## <a name="architecture"></a>Architecture

Weave Scope consists of two components: the app and the probe. These two
components are deployed as a single Docker container using the `scope`
script.

The probe is responsible for gathering information about the host is it running
on. This information is sent to the app in the form of a report. The app is
responsible for processing reports from the probe into usable topologies,
serving the UI, and pushing these topologies to the UI.

```
+--Docker host----------+
|  +--Container------+  |    .---------------.
|  |                 |  |    | Browser       |
|  |  +-----------+  |  |    |---------------|
|  |  | scope-app |<---------|               |
|  |  +-----------+  |  |    |               |
|  |        ^        |  |    |               |
|  |        |        |  |    '---------------'
|  | +-------------+ |  |
|  | | scope-probe | |  |
|  | +-------------+ |  |
|  |                 |  |
|  +-----------------+  |
+-----------------------+
```

## <a name="using-weave-scope-in-standalone-mode"></a>Using Weave Scope in Standalone Mode

When running Scope in a cluster, each probe sends reports to each app.
The App merges the reports from each probe into a more complete report.
You need to run Scope on every machine you want to monitor.

```
+--Docker host----------+      +--Docker host----------+
|  +--Container------+  |      |  +--Container------+  |
|  |                 |  |      |  |                 |  |
|  |  +-----------+  |  |      |  |  +-----------+  |  |
|  |  | scope-app |<-----.    .----->| scope-app |  |  |
|  |  +-----------+  |  | \  / |  |  +-----------+  |  |
|  |        ^        |  |  \/  |  |        ^        |  |
|  |        |        |  |  /\  |  |        |        |  |
|  | +-------------+ |  | /  \ |  | +-------------+ |  |
|  | | scope-probe |-----'    '-----| scope-probe | |  |
|  | +-------------+ |  |      |  | +-------------+ |  |
|  |                 |  |      |  |                 |  |
|  +-----------------+  |      |  +-----------------+  |
+-----------------------+      +-----------------------+
```

If you run Scope on the same machine as the Weave Network, the probe will use
weaveDNS to automatically discover other apps on your network. Scope acheives
this by registering itself under the address **scope.weave.local**. Each probe
will send reports to every app registered under this address. Therefore, if
you have a running weaveDNS setup, you do not need to take any further steps.

If you do not wish to use weaveDNS, you can instruct Scope to cluster with
other Scope instances on the command line. Hostnames and IP addresses are
acceptable, both with and without ports:

```
# scope launch scope1:4030 192.168.0.12 192.168.0.11:4030
```

Hostnames will be regularly resolved as A records, and each answer used as a
target.

## <a name="using-weave-scope-in-cloud-service-mode"></a>Using Weave Scope in Cloud Service Mode

Scope can also be used to feed reports to the Scope Service. The Scope Service
allows you centrally manage and share access to your Scope UI. In this
configuration, you only run the probe locally; the apps are hosted for you.

To get an account on the Scope Service, sign up at [scope.weave.works][]. You
need to run a probe on every machine you want to monitor with Scope. To launch
a probe and send reports to the service, run the following command:

[scope.weave.works]: http://scope.weave.works

```
sudo scope launch --service-token=<token>
```

```
                       .-~~~-.
                 .- ~'`       )_   ___
                /               `-'   )_
               |    scope.weave.works   \
                \                      .'
                  ~-______________..--'
                           ^^
                           ||
                           ||
+--Docker host----------+  ||  +--Docker host----------+
|  +--Container------+  |  ||  |  +--Container------+  |
|  |                 |  |  ||  |  |                 |  |
|  | +-------------+ |  | /  \ |  | +-------------+ |  |
|  | | scope-probe |-----'    '-----| scope-probe | |  |
|  | +-------------+ |  |      |  | +-------------+ |  |
|  |                 |  |      |  |                 |  |
|  +-----------------+  |      |  +-----------------+  |
+-----------------------+      +-----------------------+
```


## <a name="using-weave-scope-with-kubernetes"></a>Using Weave Scope with Kubernetes

To use Scope's Kubernetes integration, you need to start Scope with the
`--probe.kubernetes true` flag.  Scope needs to be installed on all
nodes (master and minions), but this flag should only be enabled on the
Kubernetes master node.

As per the normal requirements, you will need to run Scope on every
machine you want to monitor, as shown in [Getting
Started](#getting-started). However, when launching Scope you
need to pass different arguments to the Kubernetes master and minion
nodes.

On the master node you need to launch Scope with Kubernetes support:

```
sudo scope launch --probe.kubernetes true
```

Depending on your setup, you may find that Kubernetes has renamed your
Docker bridge interface. In this instance you'll need to tell Scope
about the new name when launching it. For example, if your Docker bridge is
named `cbr0`:

```
sudo DOCKER_BRIDGE=cbr0 scope launch --probe.docker.bridge cbr0 --probe.kubernetes true
```

On each minion node you need to launch Scope telling it
to connect to the master node.

```
sudo scope launch --no-app kubernetes-master.my.network
```

Again, if your Docker bridge interface is named differently, you'll
need to pass that to your probe when launching it.

Once the first few reports come in, the UI should begin displaying two
Kubernetes-specific views "Pods", and "Pods by Service".


## <a name="developing"></a>Developing

The build is in two stages. `make deps` installs some tools we use later in
the build. `make` builds the UI build container, builds the UI in said
container, builds the backend build container, builds the app and probe in a
said container, and finally pushes the lot into a Docker image called
**weaveworks/scope**.

```
make deps
make
```

Then, run the local build via

```
./scope launch
```

## <a name="developing"></a>Debugging

Scope has a collection of built in debugging tools to aid Scope delevopers.

- To have the app or probe dump their goroutine stacks, run:
```
pkill -SIGQUIT scope-(app|probe)
docker logs weavescope
```

- The probe is instrumented with various counters and timers. To have it dump
  those values, run:
```
pkill -SIGUSR1 scope-probe
docker logs weavescope
```

- The app and probe both include golang's pprof integration for gathering CPU
  and memory profiles.  To use these with the probe, you must launch Scope with
  the following arguments `scope launch --probe.http.listen :4041`.  You can
  then collect profiles in the usual way:
```
go tool pprof http://localhost:(4040|4041)/debug/pprof/profile
```

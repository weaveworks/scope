# Scope

[![Circle CI](https://circleci.com/gh/weaveworks/scope/tree/master.svg?style=shield)](https://circleci.com/gh/weaveworks/scope/tree/master) [![Coverage Status](https://coveralls.io/repos/weaveworks/scope/badge.svg)](https://coveralls.io/r/weaveworks/scope) [![Slack Status](https://weave-scope-slack.herokuapp.com/badge.svg)](https://weave-scope-slack.herokuapp.com)

Weave Scope automatically generates a map of your application, enabling you to
intuitively understand, monitor, and control your containerized, microservices based application.

### Understand your Docker containers in real-time:

<img src="imgs/topology.png" width="200" alt="Map you architecture" align="right">

Choose an overview of your container infrastructure, or focus on a specific microservice. Easily identify and correct issues to ensure the stability and performance of your containerized applications.

### Contextual details and deep linking:

<img src="imgs/selected.png" width="200" alt="Focus on a single container" align="right">

View contextual metrics, tags and metadata for your containers.  Effortlessly navigate between processes inside your container to hosts your containers run on, arranged in expandable, sortable tables.  Easily to find the container using the most CPU or memory for a given host or service.

### Interact with and manage containers:

<img src="imgs/terminals.png" width="200" alt="Launch a command line." align="right">

Interact with your containers directly: pause, restart and stop containers. Launch a command line. All without leaving the scope browser window.

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
- <a href="https://weave-scope-slack.herokuapp.com">join our  public slack channel</a>
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

Scope comes with built-in Kubernetes support. We recommend to run Scope natively
in your Kubernetes cluster using
[this resource definitions](https://github.com/TheNewNormal/kube-charts/tree/master/weavescope/manifests).

1. If you are running a Kubernetes version lower than 1.1, make sure your
   cluster allows running pods in privileged mode (required by the Scope
   probes). To allow privileged pods, your API Server and all your Kubelets must
   be provided with flag `--allow_privileged` at launch time.

2. Make sure your cluster supports
   [DaemonSets](https://github.com/kubernetes/kubernetes/blob/master/docs/design/daemon.md)
   in your cluster. DaemonSets are needed to ensure that each Kubernetes node
   runs a Scope Probe:

   * To enable them in an existing cluster, make sure to add a
     `--runtime-config=extensions/v1beta1/daemonsets=true` argument to the
     [apiserver](https://github.com/kubernetes/kubernetes/blob/master/docs/admin/kube-apiserver.md)'s configuration
     (normally found at `/etc/kubernetes/manifest/kube-apiserver.manifest`) followed by a
     [restart of the apiserver and controller manager](https://github.com/kubernetes/kubernetes/issues/18656).

   * If you are creating a new cluster, set `KUBE_ENABLE_DAEMONSETS=true` in
     your cluster configuration.

3. Download the resource definitions:

   ```
for I in app-rc app-svc probe-ds; do curl -s -L https://raw.githubusercontent.com/TheNewNormal/kube-charts/master/weavescope/manifests/scope-$I.yaml -o scope-$I.yaml; done
```

4. Tweak the Scope probe configuration at `scope-probe-ds.yaml`, namely:
   * If you have an account at http://scope.weave.works and want to use Scope in
     Cloud Service Mode, uncomment the `--probe.token=foo` argument, substitute `foo`
     by the token found in your account page, and comment out the
     `$(WEAVE_SCOPE_APP_SERVICE_HOST):$(WEAVE_SCOPE_APP_SERVICE_PORT)` argument.

5. Install Scope in your cluster (order is important):

   ```
kubectl create -f scope-app-rc.yaml  # Only if you want to run Scope in Standalone Mode
kubectl create -f scope-app-svc.yaml # Only if you want to run Scope in Standalone Mode
kubectl create -f scope-probe-ds.yaml
```


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

- To have the Scope App or Scope Probe dump their goroutine stacks, run:
```
pkill -SIGQUIT scope-(app|probe)
docker logs weavescope
```

- The Scope Probe is instrumented with various counters and timers. To have it dump
  those values, run:
```
pkill -SIGUSR1 scope-probe
docker logs weavescope
```

- Both the Scope App and the Scope Probe offer
  [http endpoints with profiling information](https://golang.org/pkg/net/http/pprof/).
  These cover things such as CPU usage and memory consumption:
  * The Scope App enables its http profiling endpoints by default, which
    are accessible on the same port the Scope UI is served (4040).
  * The Scope Probe doesn't enable its profiling endpoints by default.
    To enable them, you must launch Scope with `--probe.http.listen addr:port`.
    For instance, launching scope with `scope launch --probe.http.listen :4041`, will
    allow you access the Scope Probe's profiling endpoints on port 4041.

  Then, you can collect profiles in the usual way. For instance:

  * To collect the Memory profile of the Scope App:

    ```
go tool pprof http://localhost:4040/debug/pprof/heap
```
  * To collect the CPU profile of the Scope Probe:

    ```
go tool pprof http://localhost:4041/debug/pprof/profile
```

# Scope Probe Plugins

Scope probe plugins let you insert your own custom metrics into Scope
and get them displayed in the UI.

<img src="../../imgs/plugin.png" width="800" alt="Scope Probe plugin screenshot" align="center">

You can find some examples at the [the example
plugins](https://github.com/weaveworks/scope/tree/master/examples/plugins)
directory. We currently provide two examples:

* A [Python
  plugin](https://github.com/weaveworks/scope/tree/master/examples/plugins/http-requests)
  using [bcc](http://iovisor.github.io/bcc/) to extract incoming HTTP
  request rates per process, without any application-level
  instrumentation requirements and negligible performance toll
  (metrics are obtained in-kernel without any packet copying to
  userspace).  **Note:** This plugin needs a [recent kernel version
  with ebpf
  support](https://github.com/iovisor/bcc/blob/master/INSTALL.md#kernel-configuration). It
  will not compile on current [dlite](https://github.com/nlf/dlite)
  and boot2docker hosts.
* A [Go
  plugin](https://github.com/weaveworks/scope/tree/master/examples/plugins/iowait),
  using [iostat](https://en.wikipedia.org/wiki/Iostat) to provide
  host-level CPU IO wait or idle metrics.

The example plugins can be run by calling `make` in their directory.
This will build the plugin, and immediately run it in the foreground.
To run the plugin in the background, see the `Makefile` for examples
of the `docker run ...` command.

If the running plugin was picked up by Scope, you will see it in the
list of `PLUGINS` in the bottom right of the UI.

## Plugin ID

Each plugin should have an unique ID. It is forbidden to change it
during the plugin's lifetime. The scope probe will get the plugin's ID
from the plugin's socket filename. For example, the socket named
`my-plugin.sock`, the scope probe will deduce the ID as
`my-plugin`. IDs can only contain alphanumeric sequences, optionally
separated with a dash.

## Plugin registration

All plugins should listen for HTTP connections on a unix socket in the
`/var/run/scope/plugins` directory. The scope probe will recursively
scan that directory every 5 seconds, to look for sockets being added
(or removed). It is also valid to put the plugin unix socket in a
sub-directory, in case you want to apply some permissions, or store
other information with the socket.

## Protocol

There are several interfaces a plugin may (or must) implement. Usually
implementing an interface means handling specific requests. These
requests are described below.

### Reporter interface

Plugins _must_ implement the reporter interface. Implementing this
interface means listening for HTTP requests at `/report`.

Add the "reporter" string to the `interfaces` field in the plugin
specification.

#### Report

When the scope probe discovers a new plugin unix socket it will begin
periodically making a `GET` request to the `/report` endpoint. The
report data structure returned from this will be merged into the
probe's report and sent to the app. An example of the report structure
can be viewed at the `/api/report` endpoint of any scope app.

In addition to any data about the topology nodes, the report returned
from the plugin must include some metadata about the plugin itself.

For example:

```json
{
  "Processes": {},
  "Plugins": [
    {
      "id":          "iowait",
      "label":       "IOWait",
      "description": "Adds a graph of CPU IO Wait to hosts",
      "interfaces":  ["reporter"],
      "api_version": "1",
    }
  ]
}
```

Note that the `Plugins` section includes exactly one plugin
description. The plugin description fields are:

* `id` is used to check for duplicate plugins. It is
  required. Described in [the Plugin ID section](#plugin-id).
* `label` is a human readable plugin label displayed in the UI. It is
  required.
* `description` is displayed in the UI.
* `interfaces` is a list of interfaces which this plugin supports.  It
  is required, and must contain at least `["reporter"]`.
* `api_version` is used to ensure both the plugin and the scope probe
  can speak to each other. It is required, and must match the probe.

You may notice a small chicken and egg problem - the plugin reports to
the scope probe what interfaces it supports, but the scope probe can
learn that only by doing a `GET /report` request which will be handled
by the plugin if it implements the "reporter" interface. This is
solved (or worked around) by requiring the plugin to always implements
the "reporter" interface.

### Controller interface

Plugins _may_ implement the controller interface. Implementing the
controller interface means that the plugin can react to HTTP `POST`
control requests sent by the app. The plugin will receive them only
for controls it exposed in its reports. The requests will come to the
`/control` endpoint.

Add the "controller" string to the `interfaces` field in the plugin
specification.

#### Control

The `POST` requests will have a JSON-encoded body with the following contents:

```json
{
  "AppID": "some ID of an app",
  "NodeID": "an ID of the node that had the control activated",
  "Control": "the name of the activated control"
}
```

The body of the response should also be a JSON-encoded data. Usually
the body would be an empty JSON object (so, "{}" after
serialization). If some error happens during handling the control,
then the plugin can send a response with an `error` field set, for
example:

```json
{
  "error": "An error message here"
}
```

Sometimes the control activation can make the control obsolete, so the
plugin may want to hide it (for example, control for stopping the
container should be hidden after the container is stopped). For this
to work, the plugin can send a shortcut report by filling the
`ShortcutReport` field in the response, like for example:

```json
{
  "ShortcutReport": { body of the report here }
}
```

##### How to expose controls

Each topology in the report (be it host, pod, endpoint and so on) has
a set of available controls a node in the topology may want to
show. The following (rather artificial) example shows a topology with
two controls (`ctrl-one` and `ctrl-two`) and two nodes, each having a
different control from the two:

```json
{
  "Host": {
    "controls": {
      "ctrl-one": {
        "id": "ctrl-one",
        "human": "Ctrl One",
        "icon": "fa-futbol-o",
        "rank": 1
      },
      "ctrl-two": {
        "id": "ctrl-two",
        "human": "Ctrl Two",
        "icon": "fa-beer",
        "rank": 2
      }
    },
    "nodes": {
      "host1": {
        "latestControls": {
          "ctrl-one": {
            "timestamp": "2016-07-20T15:51:05Z01:00",
            "value": {
              "dead": false
            }
          }
        }
      },
      "host2": {
        "latestControls": {
          "ctrl-two": {
            "timestamp": "2016-07-20T15:51:05Z01:00",
            "value": {
              "dead": false
            }
          }
        }
      }
    }
  }
}
```

When control "ctrl-one" is activated, the plugin will receive a
request like:

```json
{
  "AppID": "some ID of an app",
  "NodeID": "host1",
  "Control": "ctrl-one"
}
```

A short note about the "icon" field of the topology control - the
value for it can be taken from [Font Awesome
Cheatsheet](http://fontawesome.io/cheatsheet/)

##### Node naming

Very often the controller plugin wants to add some controls to already
existing nodes (like controls for network traffic management to nodes
representing the running Docker container). To achieve that, it is
important to make sure that the node ID in the plugin's report matches
the ID of the node created by the probe. The ID is a
semicolon-separated list of strings.

For containers, images, hosts and others the ID is usually formatted
as `${name};<${tag}>`. The `${name}` variable is usually a name of a
thing the node represents, like an ID of the Docker container or the
hostname. The `${tag}` denotes the type of the node. There is a fixed
set of tags used by the probe:

- host
- container
- container_image
- pod
- service
- deployment
- replica_set

The examples of "tagged" node names:

- The Docker container with full ID
  2299a2ca59dfd821f367e689d5869c4e568272c2305701761888e1d79d7a6f51:
  `2299a2ca59dfd821f367e689d5869c4e568272c2305701761888e1d79d7a6f51;<container>`
- The Docker image with name `docker.io/alpine`:
  `docker.io/alpine;<container_image>`
- The host with name `example.com`: `example.com:<host>`

The fixed set of tags listed above is not a complete set of names a
node can have though. For example, nodes representing processes are
have ID formatted as `${host};${pid}`. Probably the easiest ways to
discover how the nodes are named are:

- Read the code in
  [report/id.go](https://github.com/weaveworks/scope/blob/master/report/id.go).
- Browse the Weave Scope GUI, select some node and search for an `id`
  key in the `nodeDetails` array in the address bar.
  - For example in the
    `http://localhost:4040/#!/state/{"controlPipe":null,"nodeDetails":[{"id":"example.com;<host>","label":"example.com","topologyId":"hosts"}],…`
    URL, you can find the `example.com;<host>` which is an ID of the node
    representing the host.
  - Mentally substitute the `<SLASH>` with `/`. This can appear in
    Docker image names, so `docker.io/alpine` in the address bar will
    be `docker.io<SLASH>alpine`.

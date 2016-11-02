---
title: Generating Custom Metrics with Plugins
menu_order: 80
---

The following topics are discussed:

 * [Official Plugins](#official-plugins)
 * [Plugins Internals](#plugins-internals)
  * [Plugin ID](#plugin-id)
  * [Plugin Registration](#plugin-registration)
  * [Reporter Interface](#reporter-interface)
  * [Controller Interface](#controller-interface)

With a Scope probe plugin, you can insert custom metrics into Scope and have them display in the user interface together with the Scope's standard set of metrics.

![Custom Metrics With Plugins](images/plugin-features.png)

## <a id="official-plugins"></a>Official Plugins

You can find all the official plugins at [Weaveworks Plugins](https://github.com/weaveworks-plugins).

* [IOWait](https://github.com/weaveworks-plugins/scope-iowait): a Go plugin using [iostat](https://en.wikipedia.org/wiki/Iostat) to provide host-level CPU IO wait or idle metrics.

* [HTTP Statistics](https://github.com/weaveworks-plugins/scope-http-statistics): A Python plugin using [bcc](http://iovisor.github.io/bcc/) to track multiple metrics about HTTP per process, without any application-level instrumentation requirements and negligible performance toll. This plugin is a work in progress, as of now it implements two metrics (for more information read the [plugin documentation](https://github.com/weaveworks-plugins/scope-http-statistics)):
	* Number of HTTP requests per seconds.
	* Number of HTTP responses code per second (per code).

> **Note:** This plugin needs a [recent kernel version with ebpf support](https://github.com/iovisor/bcc/blob/master/INSTALL.md#kernel-configuration). It will not compile on current [dlite](https://github.com/nlf/dlite) and boot2docker hosts.

* [Traffic Control](https://github.com/weaveworks-plugins/scope-traffic-control): This plugin allows the user to modify latency and packet loss for a specific container via buttons in the UI's container detailed view.

* [Volume Count](https://github.com/weaveworks-plugins/scope-volume-count): This plugin is a Python application that asks docker for the the number of mounted volumes for each container, providing container-level count.

If the running plugin was picked up by Scope, you will see it in the list of `PLUGINS` in the bottom right of the UI.

## <a id="plugins-internals"></a>Plugins Internals

This section explains the fundamental parts of the plugins structure necessary to understand how a plugin communicates with Scope.
You can find more practical examples in [Weaveworks Plugins](https://github.com/weaveworks-plugins) repositories.

### <a id="plugin-id"></a>Plugin ID

Each plugin should have an unique ID. It is forbidden to change it
during the plugin's lifetime. The scope probe will get the plugin's ID
from the plugin's socket filename. For example, the socket named
`my-plugin.sock`, the scope probe will deduce the ID as
`my-plugin`. IDs can only contain alphanumeric sequences, optionally
separated with a dash.

### <a id="plugin-registration"></a>Plugin registration

All plugins must listen for HTTP connections on a Unix socket in the `/var/run/scope/plugins` directory. The Scope probe recursively scans that directory every 5 seconds, to look for any sockets being added (or removed). It is also valid to put the plugin Unix socket into a sub-directory, in case you want to apply some permissions, or store any other information with the socket.

When a new plugin is detected, the scope probe begins requesting reports from it via `GET /report`. So every plugins **must** implement the report interface.
Implementing an interface means handling specific requests.

All plugin endpoints are expected to respond within 500ms, and respond in the JSON format.

### <a id="reporter-interface"></a>Reporter Interface

When the Scope probe discovers a new plugin Unix socket, it begins to periodically make a `GET` request to the `/report` endpoint. The report data structure returned from this will be merged into the probe's report and sent to the app. An example of the report structure can be viewed at the `/api/report` endpoint of any Scope app.

In addition to any data about the topology nodes, the report returned from the plugin must include some metadata about the plugin itself.

For example:

```json
{
  ...,
  "Plugins": [
    {
      "id":          "plugin-id",
      "label":       "Human Friendly Name",
      "description": "Plugin's brief description",
      "interfaces":  ["reporter"],
      "api_version": "1",
    }
  ]
}
```

> **Note:** The `Plugins` section includes exactly one plugin description. The plugin description fields are:

* `id` is used to check for duplicate plugins. It is required.
* `label` is a human readable plugin label displayed in the UI. It is required.
* `description` is displayed in the UI. It is required.
* `interfaces` is a list of interfaces which this plugin supports. It is required, and must contain at least `["reporter"]`.
* `api_version` is used to ensure both the plugin and the scope probe can speak to each other. It is required, and must match the probe's value.

### <a id="controller-interface"></a>Controller Interface

Plugins _may_ implement the controller interface. Implementing the
controller interface means that the plugin can react to HTTP `POST`
control requests sent by the app. The plugin will receive them only
for controls it exposed in its reports. The requests will come to the
`/control` endpoint.

Add the "controller" string to the interfaces field in the plugin specification.

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

#### How to expose controls

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



 **See Also**

  * [Building Scope](/site/building.md)



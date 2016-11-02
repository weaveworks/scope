---
title: Generating Custom Metrics with Plugins
menu_order: 80
---

The following topics are discussed:

 * [Official Plugins](#official-plugins)
 * [Plugins Internals](#plugins-internals)
  * [Plugin ID](#plugin-id)
  * [Plugin Registration](#plugin-registration)
  * [Reporting](#interfaces-interface)
  * [Other Interfaces](#other-interfaces)

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

### <a id="reporting"></a>Reporting Interface

When the Scope probe discovers a new plugin Unix socket, it begins to periodically make a `GET` request to the `/report` endpoint. The report data structure returned from this will be merged into the probe's report and sent to the app. An example of the report structure can be viewed at the `/api/report` endpoint of any Scope app.

In addition to any data about the topology nodes, the report returned from the plugin must include some metadata about the plugin itself.

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

> **Note:** The `Plugins` section includes exactly one plugin description. The plugin description fields are:

* `id` is used to check for duplicate plugins. It is required.
* `label` is a human readable plugin label displayed in the UI. It is required.
* `description` is displayed in the UI. It is required.
* `interfaces` is a list of interfaces which this plugin supports. It is required, and must contain at least `["reporter"]`.
* `api_version` is used to ensure both the plugin and the scope probe can speak to each other. It is required, and must match the probe's value.

### <a id="other-interfaces"></a>Other Interfaces

Currently the only interface a plugin can fulfill is `reporter`.

 **See Also**

  * [Building Scope](/site/building.md)



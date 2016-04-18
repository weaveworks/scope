# Scope Probe Plugins

Scope probe plugins let you insert your own custom metrics into Scope and get them displayed in the UI.

You can find some examples at the
[the example plugins](https://github.com/weaveworks/scope/tree/master/examples/plugins)
directory. We currently provide two examples:
* A
  [Python plugin](https://github.com/weaveworks/scope/tree/master/examples/plugins/http-requests)
  using [bcc](http://iovisor.github.io/bcc/) to extract incoming HTTP request
  rates per process, without any application-level instrumentation requirements.
* A
  [Go plugin](https://github.com/weaveworks/scope/tree/master/examples/plugins/iovisor),
  using [iostat](https://en.wikipedia.org/wiki/Iostat) to provide host-level CPU IO wait
  metrics.

The example plugins can be run by calling `make` in their directory.
This will build the plugin, and immediately run it in the foreground.
To run the plugin in the background, see the `Makefile` for examples
of the `docker run ...` command.


## <a id="protocol"></a>Protocol

All plugins should listen for HTTP connections on a unix socket in the
`/var/run/scope/plugins` directory. The scope probe will recursively scan that
directory every 5 seconds, to look for sockets being added (or removed). It is
also valid to put the plugin unix socket in a sub-directory, in case you want
to apply some permissions, or store other information with the socket.

When a new plugin is detected, the scope probe will begin requesting
reports from it via `GET /report`.

All plugin endpoints are expected to respond within 500ms, and respond in the JSON format.

### <a id="report"></a>Report

When the scope probe discovers a new plugin unix socket it will begin
periodically making a `GET` request to the `/report` endpoint. The
report data structure returned from this will be merged into the
probe's report and sent to the app. An example of the report structure
can be viewed at the `/api/report` endpoint of any scope app.

In addition to any data about the topology nodes, the report returned
from the plugin must include some information about the plugin.

For example:

```json
{
  "Processes: { ... },
  ...
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

Note that the "Plugins" section includes exactly one plugin
description. The plugin description fields are:
"interfaces" including "reporter".

The fields are:

* `id` is used to check for duplicate plugins. It is required.
* `label` is a human readable plugin label displayed in the UI. It is required.
* `description` is displayed in the UI
* `interfaces` is a list of interfaces which this plugin supports. It is required, and must equal `["reporter"]`.
* `api_version` is used to ensure both the plugin and the scope probe can speak to each other. It is required, and must match the probe.

### <a id="interfaces"></a>Interfaces

Currently the only interface a plugin can fulfill is `reporter`.

# Scope Plugins

## <a id="protocol"></a>Protocol

All plugins should listen for HTTP connections on a unix socket in the
`/var/run/scope/plugins` directory. The scope probe will recursively scan that
directory every 5 seconds, to look for sockets being added (or removed). It is
also valid to put the plugin unix socket in a sub-directory, in case you want
to apply some permissions, or store other information with the socket.

When a new plugin is detected, the scope probe will conduct a basic
[Handshake](#handshake) by requesting `GET /`.

All plugin endpoints are expected to respond within 500ms, and respond in the JSON format.

### <a id="handshake"></a>Handshake

When the scope probe discovers a new plugin unix socket it needs to know some
information about the plugin. To learn this it will make a GET request for the
`/` endpoint.

An example response is:

```json
{
	"name":        "iowait",
	"description": "Adds a graph of CPU IO Wait to hosts",
	"interfaces":  []string{"reporter"},
	"api_version": "1",
}
```

The fields are:

* `name` is used to check for duplicate plugins, and displayed in the UI
* `description` is displayed in the UI
* `interfaces` tells the scope probe which endpoints this plugin supports
* `api_version` is used to ensure both the plugin and the scope probe can speak to each other

### <a id="interfaces"></a>Interfaces

Currently the only interface a plugin can fulfill is `reporter`.

#### <a id="reporter"></a>Reporter

The `reporter` interface allows a plugin to add information into the probe report. This could include more nodes, or new fields on existing nodes.

Endpoints:

* GET /report
  This endpoint should return a scope probe-style report. For an example of the
datastructure see `/api/report` on any scope instance. At the moment the plugin
is limited to adding nodes or fields to existing topologies (Endpoint, Process,
Container, etc), along with `metadata_templates` and `metric_templates` to
display more information. For an example of adding a metric to the hosts, see
[the example iowait
plugin.](https://github.com/weaveworks/scope/tree/master/example/plugins/iowait)

## Release 0.6.0

New features:
- Probes now push data to the app, instead of the app pulling it.
  [#342](https://github.com/weaveworks/scope/pull/342)
- Allow probe and app to be started independently, via --no-app and
  --no-probe flags.
  [#345](https://github.com/weaveworks/scope/pull/345)
- Close details pane when changing topology view.
  [#297](https://github.com/weaveworks/scope/issues/297)
- Add support for --probe.foo=bar style flags, in addition to
  --probe.foo bar, which is already supported.
  [#347](https://github.com/weaveworks/scope/pull/347)
- Added X-Scope-Probe-ID header to identify probes when sending
  information to the app.
  [#351](https://github.com/weaveworks/scope/pull/351)

Bug fixes:
- Update scope script to work with master version of weave, where DNS
  has been embedded in the router.
  [#321](https://github.com/weaveworks/scope/issues/321)
- Fixed regression where process names weren't appearing for Darwin
  probes.
  [#320](https://github.com/weaveworks/scope/pull/320)
- Fixed rendering bug resulting in orphaned nodes.
  [#339](https://github.com/weaveworks/scope/pull/339)
- App now only logs to stderr, to match the probe.
  [#343](https://github.com/weaveworks/scope/pull/343)
- Use relative paths for all URLs in the UI.
  [#344](https://github.com/weaveworks/scope/pull/344)
- Removed temporary containers created by the scope script.
  [#348](https://github.com/weaveworks/scope/issues/348)

Experimental features:
- Added support for pcap based packet sniffing, to provide bandwidth
  usage information. It can be enabled via the --capture flag. When
  enabled the probe will monitor packets for a portion of the time, and
  estimate bandwidth usage. Network throughput will be affected if
  capture is enabled.
  [#317](https://github.com/weaveworks/scope/pull/317)


## Release 0.5.0

New features:
- Aggregate all connection information into a single table in the details
  dialog.
  [#298](https://github.com/weaveworks/scope/pull/298)
- Renamed binaries to scope-app and scope-probe
  [#293](https://github.com/weaveworks/scope/pull/293)
- Group containers topology by name only, and not version
  [#291](https://github.com/weaveworks/scope/issues/291)
- Make intra-scope communication traverse the weave network if present.
  [#71](https://github.com/weaveworks/scope/issues/71)

Bug fixes:
- Reduced memory usage
  [#266](https://github.com/weaveworks/scope/issues/266)


## Release 0.4.0

New features:
- Include kernel version and uptime in host details.
  [#274](https://github.com/weaveworks/scope/pull/274)
- Include command line and number of threads in process details.
  [#272](https://github.com/weaveworks/scope/pull/272)
- Include Docker port mapping, entrypoint, memory usage and creation
  date in container details.
  [#262](https://github.com/weaveworks/scope/pull/262)
- Order tables in details panel from least granular to most granular.
  [#261](https://github.com/weaveworks/scope/issues/261)
- Show all container images (even ones without active connections)
  in the containers-by-image view.
  [#230](https://github.com/weaveworks/scope/issues/230)
- Produce process and container views by merging endpoint topology with their
  respective topologies, such that the origins in the details panel always
  aggregate correctly. [#228](https://github.com/weaveworks/scope/issues/228)
- In containers view, show "Uncontained" nodes for each host if they have
  active connections. [#127](https://github.com/weaveworks/scope/issues/127)
- Show connection status in the UI.
  [#162](https://github.com/weaveworks/scope/issues/162)

Bug fixes:
- Reduce CPU usage by caching walks of /proc.
  [#284](https://github.com/weaveworks/scope/issues/284)
- Trim whitespace of process names such that the details panel functions
  correctly in the process-by-name view.
  [#281](https://github.com/weaveworks/scope/issues/281)
- Correctly scope addresses on Docker bridge such that processes on different
  hosts are not incorrectly shown as communicating.
  [#264](https://github.com/weaveworks/scope/pull/264)
- Correctly show connections between nodes which traverse a Docker port
  mapping. [#245](https://github.com/weaveworks/scope/issues/245)
- Make scope script fail if docker run fails.
  [#214](https://github.com/weaveworks/scope/issues/214)
- Prevent left over nodes in the UI when the connection is dropped or Scope is
  restarted. [#162](https://github.com/weaveworks/scope/issues/162)


## Release 0.3.0

- Show containers, even when they aren't communicating.
- Expand topology selectors more descriptive, and remove the grouping button.
- Fix overflow rendering bugs in the detail pane.
- Render pseudonodes with less saturation.

## Release 0.2.0

- Initial release.

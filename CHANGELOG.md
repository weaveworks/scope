## Release 0.9.0

New features:
- Add basic Kubernetes views for pods and services
  [#441](https://github.com/weaveworks/scope/pull/441)
- Support for Weave 1.2
  [#574](https://github.com/weaveworks/scope/pull/574)
- Add containers-by-hostname view
  [#545](https://github.com/weaveworks/scope/pull/545)
- Make `scope launch` work from remote hosts, with an appropriately defined DOCKER_HOST
  [#524](https://github.com/weaveworks/scope/pull/524)
- Increase DNS poll frequency such that Scope clusters up more quickly
  [#524](https://github.com/weaveworks/scope/pull/524)
- Add `scope command` for printing the Docker commands used to run Scope
  [#553](https://github.com/weaveworks/scope/pull/553)
- Include some basic documentation on how to run Scope
  [#572](https://github.com/weaveworks/scope/pull/572)
- Warn if the users tries to run Scope on Docker versions <1.5.0
  [#557](https://github.com/weaveworks/scope/pull/557)
- Add support for loading the Scope UI from https endpoings
  [#572](https://github.com/weaveworks/scope/pull/572)

Big fixes:
- Correctly track short-lived connections from the internet
  [#493](https://github.com/weaveworks/scope/pull/493)
- Fix a corner case where short-lived connections between containers are incorrectly attributed
  [#577](https://github.com/weaveworks/scope/pull/577)
- Ensure service credentials are sent when doing intial probe<->app handshake
  [#564](https://github.com/weaveworks/scope/pull/564)
- Sort reverse-DNS-resolved names to mitigate some UI fluttering
  [#562](https://github.com/weaveworks/scope/pull/562)
- Don't leak goroutines in the probe
  [#531](https://github.com/weaveworks/scope/issue/531)
- Rerun background conntrack processes if they fail
  [#581](https://github.com/weaveworks/scope/issue/581)

## Release 0.8.0

New features:
- Show message in the UI when topologies exceed size limits.
  [#474](https://github.com/weaveworks/scope/issues/474)
- Provide container image information in detail pane for containers.
  [#398](https://github.com/weaveworks/scope/issues/398)
- When filtering out system containers, also filter out pseudo nodes, if they were only connected to system containers.
  [#483](https://github.com/weaveworks/scope/issues/483)
- Show number of filtered nodes in status pane.
  [#509](https://github.com/weaveworks/scope/issues/509)

Bug fixes:
- Prevent detail pane from hiding nodes on click-to-focus.
  [#495](https://github.com/weaveworks/scope/issues/495)
- Stop radial view from bouncing in some circumstances.
  [#496](https://github.com/weaveworks/scope/issues/496)
- Make NAT tracking component more resilient to failure.
  [#506](https://github.com/weaveworks/scope/issues/506)
- Prevent duplicate reports from reaching the same app.
  [#463](https://github.com/weaveworks/scope/issues/463)
- Improve consistency of edge directionality in some use-cases.
  [#373](https://github.com/weaveworks/scope/issues/373)
- Ensure probe, app, and container shut down cleanly.
  [#424](https://github.com/weaveworks/scope/issues/424)
  [#478](https://github.com/weaveworks/scope/issues/478)

## Release 0.7.0

New features:
- Show short-lived connections in the containers view.
  [#356](https://github.com/weaveworks/scope/issues/356)
  [#447](https://github.com/weaveworks/scope/issues/447)
- Details pane:
  1. Add more information:
    - Docker labels.
      [#400](https://github.com/weaveworks/scope/pull/400)
    - Weave IPs/hostnames/MACs and Docker IPs.
      [#394](https://github.com/weaveworks/scope/pull/394)
      [#396](https://github.com/weaveworks/scope/pull/396)
    - Host/container context when ambiguous.
      [#387](https://github.com/weaveworks/scope/pull/387)
  2. Present it in a more intuitive way:
    - Show hostnames instead of IPs when possible.
      [#404](https://github.com/weaveworks/scope/pull/404)
      [#451](https://github.com/weaveworks/scope/pull/451)
    - Merge all the connection-related information into a single table.
      [#322](https://github.com/weaveworks/scope/issues/322)
    - Include relevant information in the table titles.
      [#387](https://github.com/weaveworks/scope/pull/387)
    - Stop including empty fields.
      [#370](https://github.com/weaveworks/scope/issues/370)
- Allow filtering out system containers (e.g. Weave and Scope containers) and
  unconnected containers. System containers are filtered out by default.
  [#420](https://github.com/weaveworks/scope/pull/420)
  [#337](https://github.com/weaveworks/scope/issues/337)
  [#454](https://github.com/weaveworks/scope/issues/454)
  [#457](https://github.com/weaveworks/scope/issues/457)
- Improve rendering by making edges' directions flow from client to server.
  [#355](https://github.com/weaveworks/scope/pull/355)
- Highlight selected node
  [#473](https://github.com/weaveworks/scope/pull/473)
- Animate Edges during UI transtions
  [#445](https://github.com/weaveworks/scope/pull/445)
- New status bar on the bottom left of the UI
  [#487](https://github.com/weaveworks/scope/pull/487)
- Show more information for pseudo nodes where possible - such as processes for uncontained, and connected to/from the internet.
  [#249](https://github.com/weaveworks/scope/issues/249)
  [#401](https://github.com/weaveworks/scope/pull/401)
  [#426](https://github.com/weaveworks/scope/pull/426)
- Truncate node names and text in the details pane.
  [#429](https://github.com/weaveworks/scope/pull/429)
  [#430](https://github.com/weaveworks/scope/pull/430)
- Amazon ECS: Stop displaying mangled container names, display the original Container Definition name instead.
  [#456](https://github.com/weaveworks/scope/pull/456)
- Annotate processes in containers with the container name, in the *Applications* view.
  [#331](https://github.com/weaveworks/scope/issues/331)
- Improve graph transitions between updates.
  [#379](https://github.com/weaveworks/scope/pull/379)
- Reduce CPU usage of Scope probes
  [#470](https://github.com/weaveworks/scope/pull/470)
  [#484](https://github.com/weaveworks/scope/pull/484)
- Make report propagation more reliable
  [#459](https://github.com/weaveworks/scope/pull/459)
- Support Weave 1.1 status interface
  [#389](https://github.com/weaveworks/scope/pull/389)

Bug fixes:
- *Trying to reconnect..* in UI even though its connected.
  [#392](https://github.com/weaveworks/scope/pull/392)
- The *Applications* view goes blank after a few seconds.
  [#442](https://github.com/weaveworks/scope/pull/442)
- Frequently getting disconnected lines in UI
  [#460](https://github.com/weaveworks/scope/issues/460)
- Panic due to closing request body
  [#480](https://github.com/weaveworks/scope/pull/480)

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

## Release 0.13.1

Bug Fixes:
- Make pipes work with scope.weave.works
  [#1099](https://github.com/weaveworks/scope/pull/1099)
  [#1085](https://github.com/weaveworks/scope/pull/1085)
  [#994](https://github.com/weaveworks/scope/pull/994)

## Release 0.13.0

Note: This release come with big performance improvements, cutting the probe's CPU usage by 70% and the app's CPU usage by up to 85%. See detailed performance improvement-related changes below:

Performance improvements:
- Improve codec performance
	[#916](https://github.com/weaveworks/scope/pull/916)
  [#1002](https://github.com/weaveworks/scope/pull/1002)
  [#1005](https://github.com/weaveworks/scope/pull/1005)
  [#980](https://github.com/weaveworks/scope/pull/980)
- Reduce amount of objects allocated by the codec
	[#1000](https://github.com/weaveworks/scope/pull/1000)
- Refactor app for multitenancy
	[#997](https://github.com/weaveworks/scope/pull/997)
- Improve performance of docker stats obtention
	[#989](https://github.com/weaveworks/scope/pull/989)
- Rate-limit reading proc files
	[#912](https://github.com/weaveworks/scope/pull/912)
  [#905](https://github.com/weaveworks/scope/pull/905)
- Compile k8s selectors once (not for each pod)
	[#918](https://github.com/weaveworks/scope/pull/918)
- Fix reading of network namespace inodes
	[#898](https://github.com/weaveworks/scope/pull/898)

New features and enhancements:
- Node shapes for different topologies, e.g. heptagons for Kubernetes pods
  [#884](https://github.com/weaveworks/scope/pull/884)
	[#1006](https://github.com/weaveworks/scope/pull/1006)
	[#1037](https://github.com/weaveworks/scope/pull/1037)
- Force-relayout button that may help with topology layouts that have lots of edge crossings
	[#981](https://github.com/weaveworks/scope/pull/981)
- Download button to save the current node graph as SVG file
	[#1027](https://github.com/weaveworks/scope/pull/1027)
- Replace Show More buttons with carets w/ counts
  [#1012](https://github.com/weaveworks/scope/pull/1012)
	[#1029](https://github.com/weaveworks/scope/pull/1029)
- Improve contrast of default view
	[#979](https://github.com/weaveworks/scope/pull/979)
- High contrast mode button for viewing scope on projectors
	[#954](https://github.com/weaveworks/scope/pull/954)
	[#984](https://github.com/weaveworks/scope/pull/984)
- Gather file descriptors as process metric
	[#961](https://github.com/weaveworks/scope/pull/961)
- Show Docker Labels in their own table in details panel
	[#904](https://github.com/weaveworks/scope/pull/904)
	[#965](https://github.com/weaveworks/scope/pull/965)
- Improve highlighting of selected topology
	[#936](https://github.com/weaveworks/scope/pull/936)
  [#964](https://github.com/weaveworks/scope/pull/964)
- Details: only show important metadata by default, expand the rest
	[#946](https://github.com/weaveworks/scope/pull/946)
- Reorder the children tables in the details panel by importance
	[#941](https://github.com/weaveworks/scope/pull/941)
- Shorten docker container and image IDs in the details panel.
	[#930](https://github.com/weaveworks/scope/pull/930)
- Shorten some details panel labels which were truncated
	[#940](https://github.com/weaveworks/scope/pull/940)
- Sparklines update every second and show 60sec history
	[#795](https://github.com/weaveworks/scope/pull/795)
- Add Container Count column to container images table
	[#919](https://github.com/weaveworks/scope/pull/919)
- Periodically check for newer versions of scope.
	[#907](https://github.com/weaveworks/scope/pull/907)
- Rename Applications -> Process, sort topologies by rank.
	[#866](https://github.com/weaveworks/scope/pull/866)
- Rename 'by hostname' to 'by dns name'
	[#856](https://github.com/weaveworks/scope/pull/856)
- Add container uptime and restart count to details panel.
	[#853](https://github.com/weaveworks/scope/pull/853)
- Use connection directions from conntrack for improved layout flow
	[#967](https://github.com/weaveworks/scope/pull/967)
- Support for container controls in Kubernetes
  [#1043](https://github.com/weaveworks/scope/pull/1043)
- Add debug logging
	[#935](https://github.com/weaveworks/scope/pull/935)

Bug fixes:
- Use TCP for weave dns to fix autoclustering
  [#1038](https://github.com/weaveworks/scope/pull/1038)
- Add ping/pong to websocket protocol to prevent websocket connections being dropped when traversing loadbalancers
	[#995](https://github.com/weaveworks/scope/pull/995)
- Handle closing of docker events channel gracefully
	[#1014](https://github.com/weaveworks/scope/pull/1014)
- Don't show blank IPs metadata row for containers with no IP
	[#960](https://github.com/weaveworks/scope/pull/960)
- Remove pointer math (comparison) from render caching, as it is unreliable
	[#962](https://github.com/weaveworks/scope/pull/962)
- set TERM=xterm on execs to work around docker issue 9299
	[#969](https://github.com/weaveworks/scope/pull/969)
- Fix weave tagger crash
	[#976](https://github.com/weaveworks/scope/pull/976)
- Use Sirupsen/logrus logger in the Weave tagger
	[#974](https://github.com/weaveworks/scope/pull/974)
- Fix JSON encoding for fixedprobe
	[#975](https://github.com/weaveworks/scope/pull/975)
- Don't render any metrics/metadata for uncontained node
	[#956](https://github.com/weaveworks/scope/pull/956)
- Update go-dockerclient to fix bug with docker 1.10
	[#952](https://github.com/weaveworks/scope/pull/952)
- Show nice column labels when no children have metrics
	[#950](https://github.com/weaveworks/scope/pull/950)
- Fixes process-by-name layout with ./foo and /foo nodes
	[#948](https://github.com/weaveworks/scope/pull/948)
- Deal with starting / stopping weave whilst scope is running
	[#867](https://github.com/weaveworks/scope/pull/867)
- Remove host links that link to themselves in details panel
	[#917](https://github.com/weaveworks/scope/pull/917)
- Just show the untruncated label in the tooltip on children
	[#911](https://github.com/weaveworks/scope/pull/911)
- Taking a read lock twice only works most of the time.
	[#889](https://github.com/weaveworks/scope/pull/889)
- Details panel table header looks up label in all rows
	[#895](https://github.com/weaveworks/scope/pull/895)
- Fixes some fields overflowing badly in details panel in Chrome 48
	[#892](https://github.com/weaveworks/scope/pull/892)
- Stop details cards popping up over the terminal.
	[#882](https://github.com/weaveworks/scope/pull/882)
- Fixes host node/details panel color mismatch
	[#880](https://github.com/weaveworks/scope/pull/880)
- Don't log expected websocket errors
	[#1024](https://github.com/weaveworks/scope/pull/1024)
- Overwrite /etc/weave/apps, because it might already exist
	[#959](https://github.com/weaveworks/scope/pull/959)
- Log a warning when reporters or taggers take too long to generate
	[#944](https://github.com/weaveworks/scope/pull/944)
- Minor refactor of backend metadata and metric rendering
	[#920](https://github.com/weaveworks/scope/pull/920)
- Add some tests, and a zero-value for report.Sets
	[#903](https://github.com/weaveworks/scope/pull/903)

Build improvements and cleanup:
- Disable checkpointing in tests.
	[#1031](https://github.com/weaveworks/scope/pull/1031)
- Turn off GC for builds.
	[#1023](https://github.com/weaveworks/scope/pull/1023)
- Bump template name to get latest version of docker.
	[#998](https://github.com/weaveworks/scope/pull/998)
- Fixes building scope outside of a container.
	[#901](https://github.com/weaveworks/scope/pull/901)
- Don't need sudo when DOCKER_HOST is tcp.
	[#888](https://github.com/weaveworks/scope/pull/888)
- Disable npm progress to speed up build
	[#894](https://github.com/weaveworks/scope/pull/894)
- Refactoring deepequal to satisfy linter
	[#890](https://github.com/weaveworks/scope/pull/890)

Documentation:
- Document how to obtain profiles without `go tool pprof`
	[#993](https://github.com/weaveworks/scope/pull/993)
- Use short URL for scope download
	[#1018](https://github.com/weaveworks/scope/pull/1018)
- Added note about docker and go dependency to the readme
	[#966](https://github.com/weaveworks/scope/pull/966)
- Update readme and images.
	[#885](https://github.com/weaveworks/scope/pull/885)
- Update approach to trigger signal dumps
	[#883](https://github.com/weaveworks/scope/pull/883)


## Release 0.12.0

New features and enhancements:
- New, interactive contextual details panel
  [#752](https://github.com/weaveworks/scope/pull/752)
- Gather per-process CPU and memory metrics
  [#767](https://github.com/weaveworks/scope/pull/767)
- k8s: Use service account token by default and improve error logging
  [#808](https://github.com/weaveworks/scope/pull/808)
- k8s: Filter out pause as a system container to declutter view
  [#823](https://github.com/weaveworks/scope/pull/823)
- k8s: Render container names from label "io.kubernetes.container.name"
  [#810](https://github.com/weaveworks/scope/pull/810)
- Probes now use TLS against scope.weave.works by default
  [#785](https://github.com/weaveworks/scope/pull/785)
- Allow dismissing a disconnected terminal w/ \<esc\>
  [#819](https://github.com/weaveworks/scope/pull/819)

Bug fixes:
- General k8s fixups
  [#834](https://github.com/weaveworks/scope/pull/834)
- Use argv\[0\] for process name, differentiate scope app and probe.
  [#796](https://github.com/weaveworks/scope/pull/796)
- Don't panic if you don't understand the message on the control WS.
  [#793](https://github.com/weaveworks/scope/pull/793)
- Highlight a single unconnected node on hover.
  [#790](https://github.com/weaveworks/scope/pull/790)
- Fixes to Terminal resizing and key support
  [#766](https://github.com/weaveworks/scope/pull/766)
  [#780](https://github.com/weaveworks/scope/pull/780)
  [#817](https://github.com/weaveworks/scope/pull/817)
- Correctly collapse nodes in the Container Images view when they use non-standard port.
  [#824](https://github.com/weaveworks/scope/pull/824)
- Stop scope crashing chrome when we get "long" edges.
  [#837](https://github.com/weaveworks/scope/pull/837)
- Fix node controls so they behave independently across nodes
  [#797](https://github.com/weaveworks/scope/pull/797)

Build improvements and cleanup:
- Update to latest tools.git
  [#816](https://github.com/weaveworks/scope/pull/816)
- Update to latest go-dockerclient
  [#788](https://github.com/weaveworks/scope/pull/788)
- Speed up builds
  [#775](https://github.com/weaveworks/scope/pull/775)
  [#789](https://github.com/weaveworks/scope/pull/789)
- Speed up tests
  [#799](https://github.com/weaveworks/scope/pull/799)
  [#807](https://github.com/weaveworks/scope/pull/807)
- Split and move xfer package.
  [#794](https://github.com/weaveworks/scope/pull/794)
- Add more tests to procspy
  [#751](https://github.com/weaveworks/scope/pull/751)
  [#781](https://github.com/weaveworks/scope/pull/781)
- Build example app in container.
  [#831](https://github.com/weaveworks/scope/pull/831)
- Various improvements to build & test
  [#829](https://github.com/weaveworks/scope/pull/829)

## Release 0.11.1

Bug fix:
- Scrape /proc/PID/net/tcp6 such that we see both ends of local connections
  [change](https://github.com/weaveworks/scope/commit/550f21511a2da20717c6de6172b5bf2e9841d905)

## Release 0.11.0

New features:

- Add a terminal to the UI with the ability to `attach` to, or `exec` a shell in, a Docker container.
  [#650](https://github.com/weaveworks/scope/pull/650)
  [#735](https://github.com/weaveworks/scope/pull/735)
  [#726](https://github.com/weaveworks/scope/pull/726)
- Added `scope version` command
  [#750](https://github.com/weaveworks/scope/pull/750)
- Various CPU usage reductions for probe
  [#742](https://github.com/weaveworks/scope/pull/742)
  [#741](https://github.com/weaveworks/scope/pull/741)
  [#737](https://github.com/weaveworks/scope/pull/737)
- Show hostname of app you are connected to in the bottom right of the UI
  [#707](https://github.com/weaveworks/scope/pull/707)
- Add host memory and CPU usage metrics to the details panel
  [#711](https://github.com/weaveworks/scope/pull/711)
- Add json support to app POST /api/report
  [#722](https://github.com/weaveworks/scope/pull/722)
- Update the docker version we embed into the scope image to 1.6.2 in sync with weave 1.3 changes.
  [#702](https://github.com/weaveworks/scope/pull/702)
- Show a spinner while node details are loading
  [#691](https://github.com/weaveworks/scope/pull/691)
- Deterministic coloring of nodes based on rank and label
  [#694](https://github.com/weaveworks/scope/pull/694)

Bug fixes:

- Mitigate one-line-of-nodes layouts (when graph has few connections), layout in rectangle instead
  [#679](https://github.com/weaveworks/scope/pull/679)
- When filtering unconnected nodes in the processes view, also filter nodes that are only connected to themselves.
  [#706](https://github.com/weaveworks/scope/pull/706)
- Correctly hide container based on docker labels on the container image.
  [#705](https://github.com/weaveworks/scope/pull/705)
- Show details for stopped container in the default view, by not applying filters to the node details endpoints.
  [#704](https://github.com/weaveworks/scope/pull/704)
  [#701](https://github.com/weaveworks/scope/pull/701)
- Fix render issues in Safari
  [#686](https://github.com/weaveworks/scope/pull/686)
- Take default topology option if missing in URL
  [#678](https://github.com/weaveworks/scope/pull/678)
- Dont treat missing node as UI error
  [#677](https://github.com/weaveworks/scope/pull/677)
- Unset previous details when deselecting a node
  [#675](https://github.com/weaveworks/scope/pull/675)
- Add x to close details panel again
  [#673](https://github.com/weaveworks/scope/pull/673)

Documentation:

- Add basic security warning.
  [#703](https://github.com/weaveworks/scope/pull/703)
- Add basic kubernetes how-to to the readme
  [#669](https://github.com/weaveworks/scope/pull/669)
- Document debug options for developers
  [#723](https://github.com/weaveworks/scope/pull/723)
- Add 'getting help' section and update screenshot
  [#709](https://github.com/weaveworks/scope/pull/709)

Build improvements and cleanup:

- Don't go get weave, git clone it so weave build errors don't affect Scope.
  [#743](https://github.com/weaveworks/scope/pull/743)
- Reduce image size and build time by merging scope probe and app binaries.
  [#732](https://github.com/weaveworks/scope/pull/732)
- Cleaning up some dead code around edges and edgemetadata
  [#730](https://github.com/weaveworks/scope/pull/730)
- Make `make` build  the UI
  [#728](https://github.com/weaveworks/scope/pull/728)
- Omit controls field from json if emtpy.
  [#725](https://github.com/weaveworks/scope/pull/725)
- JS to ES2015
  [#712](https://github.com/weaveworks/scope/pull/712)
- Upgraded react to 0.14.3
  [#687](https://github.com/weaveworks/scope/pull/687)
- Cleaned up node-details-table
  [#676](https://github.com/weaveworks/scope/pull/676)
- Fix react key warning
  [#672](https://github.com/weaveworks/scope/pull/672)

## Release 0.10.0

Notes:
- Due to the Scope UI now being able to start/stop/restart Docker
  containers, it is not wise to have it accessible to untrusted
  parties.

New features:
- Add lifecycle controls (start/stop/restart) for Docker containers
  [#598](https://github.com/weaveworks/scope/pull/598)
  [#642](https://github.com/weaveworks/scope/pull/642)
- Add sparklines to the UI for some metrics
  [#622](https://github.com/weaveworks/scope/pull/622)
- Show a message when the selected topology is empty
  [#505](https://github.com/weaveworks/scope/pull/505)

Bug fixes:
- Change node layout incrementally to reduce re-layouts
  [#593](https://github.com/weaveworks/scope/pull/593)
- Improve responsiveness of UI updates to container state changes
  [#628](https://github.com/weaveworks/scope/pull/628)
  [#640](https://github.com/weaveworks/scope/pull/640)
- Handle DNS Resolution to a set of names
  [#639](https://github.com/weaveworks/scope/pull/639)
- Correctly show node counts for sub-topologies
  [#621](https://github.com/weaveworks/scope/issues/621)
- Allow scope to start after being upgraded
  [#617](https://github.com/weaveworks/scope/pull/617)
- Prevent a stranded pseudo-nodes from appearing in the container view
  [#627](https://github.com/weaveworks/scope/pull/627)
  [#674](https://github.com/weaveworks/scope/pull/674)
- Parallelise and improve the testing infrastructure
  [#614](https://github.com/weaveworks/scope/pull/614)
  [#618](https://github.com/weaveworks/scope/pull/618)
  [#644](https://github.com/weaveworks/scope/pull/644)

## Release 0.9.0

New features:
- Add basic Kubernetes views for pods and services
  [#441](https://github.com/weaveworks/scope/pull/441)
- Support for Weave 1.2
  [#574](https://github.com/weaveworks/scope/pull/574)
- Add containers-by-hostname view
  [#545](https://github.com/weaveworks/scope/pull/545)
- Build using Go 1.5, with vendored dependencies
  [#584](https://github.com/weaveworks/scope/pull/584)
- Make `scope launch` work from remote hosts, with an appropriately defined DOCKER_HOST
  [#524](https://github.com/weaveworks/scope/pull/524)
- Increase DNS poll frequency such that Scope clusters more quickly
  [#524](https://github.com/weaveworks/scope/pull/524)
- Add `scope command` for printing the Docker commands used to run Scope
  [#553](https://github.com/weaveworks/scope/pull/553)
- Include some basic documentation on how to run Scope
  [#572](https://github.com/weaveworks/scope/pull/572)
- Warn if the user tries to run Scope on Docker versions <1.5.0
  [#557](https://github.com/weaveworks/scope/pull/557)
- Add support for loading the Scope UI from https endpoints
  [#572](https://github.com/weaveworks/scope/pull/572)
- Add support for probe sending reports to https endpoints
  [#575](https://github.com/weaveworks/scope/pull/575)

Bug fixes:
- Correctly track short-lived connections from the internet
  [#493](https://github.com/weaveworks/scope/pull/493)
- Fix a corner case where short-lived connections between containers are incorrectly attributed
  [#577](https://github.com/weaveworks/scope/pull/577)
- Ensure service credentials are sent when doing initial probe<->app handshake
  [#564](https://github.com/weaveworks/scope/pull/564)
- Sort reverse-DNS-resolved names to mitigate some UI fluttering
  [#562](https://github.com/weaveworks/scope/pull/562)
- Don't leak goroutines in the probe
  [#531](https://github.com/weaveworks/scope/issue/531)
- Rerun background conntrack processes if they fail
  [#581](https://github.com/weaveworks/scope/issue/581)
- Build and test using Go 1.5 and vendor all dependencies
  [#584](https://github.com/weaveworks/scope/pull/584)
- Fix "close on nil channel" error on shutdown
  [#599](https://github.com/weaveworks/scope/issues/599)

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

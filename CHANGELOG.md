## Release 0.16.1

This is a bugfix release. In addition, the security of Scope probes can be
hardened through the new flag `--probe.no-controls` with which controls will be
disabled. Without controls, Scope users won't be able to open terminals,
start/stop containers, show logs ...

New features and enhancements:
- Allow disabling controls in probes
	[#1627](https://github.com/weaveworks/scope/pull/1627)

Performance improvements:
- Use a slice instead of a persistent list for temporary accumulation of lists
	[#1660](https://github.com/weaveworks/scope/pull/1660)

Bug fixes:
- Fix ensure plugins dir exists (close #1649)
	[#1649](https://github.com/weaveworks/scope/issue/1649)
- Fixes terminal wrapping by syncing docker/term.js terminal widths.
	[#1648](https://github.com/weaveworks/scope/pull/1648)
- Wrongly attributed local side in outbound internet connections
	[#1598](https://github.com/weaveworks/scope/issues/1598)
- Cannot port-forward app from kubernetes with command in documentation
	[#1526](https://github.com/weaveworks/scope/issues/1526)
- Force some known column widths to prevent truncation of others
	[#1641](https://github.com/weaveworks/scope/pull/1641)

Documentation:
- Don't prepend `scope launch` with sudo
	[#1606](https://github.com/weaveworks/scope/pull/1606)
- Clarify instructions for using Scope with Weave Cloud
	[#1611](https://github.com/weaveworks/scope/pull/1611)
- readded signup page
	[#1604](https://github.com/weaveworks/scope/pull/1604)
- weave cloud screen captures
	[#1603](https://github.com/weaveworks/scope/pull/1603)

Internal improvements and cleanup:
- Promote fixprobe and delete rest of experimental
	[#1646](https://github.com/weaveworks/scope/pull/1646)
- refactor some timing helpers into a common lib
	[#1642](https://github.com/weaveworks/scope/pull/1642)
- Print DNS errors
	[#1607](https://github.com/weaveworks/scope/pull/1607)
- Make it easier to disable weave integrations
	[#1610](https://github.com/weaveworks/scope/pull/1610)
- Add dry run flag to scope, so when launched we can check the args are valid.
	[#1609](https://github.com/weaveworks/scope/pull/1609)
- Helper for reading & writing from binary
	[#1600](https://github.com/weaveworks/scope/pull/1600)
- Updates to vendoring document
	[#1595](https://github.com/weaveworks/scope/pull/1595)

Weave Cloud related changes:
- Count memcache requests even if they time out
	[#1662](https://github.com/weaveworks/scope/pull/1662)
- Adding a static report file mode.
	[#1659](https://github.com/weaveworks/scope/pull/1659)
- Bump memcache expiration
	[#1640](https://github.com/weaveworks/scope/pull/1640)
- Fixes to memcache support
	[#1628](https://github.com/weaveworks/scope/pull/1628)
- Refactor caching layers in dynamo collector
	[#1616](https://github.com/weaveworks/scope/pull/1616)
- Rework Scope metrics according to Prometheus conventions.
	[#1615](https://github.com/weaveworks/scope/pull/1615)
- Fix nil pointer error when memcache not enabled
	[#1612](https://github.com/weaveworks/scope/pull/1612)
- Add backoff to the consul client
	[#1608](https://github.com/weaveworks/scope/pull/1608)
- Query memcached from dynamo db collector
	[#1602](https://github.com/weaveworks/scope/pull/1602)
- Use histograms over summaries
	[#1665](https://github.com/weaveworks/scope/pull/1665)


## Release 0.16.0

Highlights:

* New network filter to quickly ascertain what networks your containers belong to.

New features and enhancements:
- Network view
	[#1528](https://github.com/weaveworks/scope/pull/1528)
	[#1593](https://github.com/weaveworks/scope/pull/1593)
- Label deployment nodes with replica count
	[#1530](https://github.com/weaveworks/scope/pull/1530)
- Add flag to disable reporting of processes (and procspied endpoints)
	[#1511](https://github.com/weaveworks/scope/pull/1511)
- Add pod status to summary table
	[#1523](https://github.com/weaveworks/scope/pull/1523)
- Add filters for pseudo nodes.
	[#1581](https://github.com/weaveworks/scope/pull/1581)

Performance improvements:
- Fast start the dns resolution ticker to improve first report latency.
	[#1508](https://github.com/weaveworks/scope/pull/1508)

Bug fixes:
- Fix tall search box in Firefox
	[#1583](https://github.com/weaveworks/scope/pull/1583)
- Probe reporter stuck
	[#1576](https://github.com/weaveworks/scope/issues/1576)
- Container in multiple networks not showing all connections
	[#1573](https://github.com/weaveworks/scope/issues/1573)
- scope probe connects to localhost & prod even when given explicit hostnames
	[#1566](https://github.com/weaveworks/scope/issues/1566)
- Fix Docker for Mac check
	[#1551](https://github.com/weaveworks/scope/pull/1551)
- If k8s objects only have one container, show that container's metrics on them
	[#1473](https://github.com/weaveworks/scope/pull/1473)
- Don't ever store NEW conntrack flows (only ever store updates).
	[#1541](https://github.com/weaveworks/scope/pull/1541)
- Pods with > 1 container making connections do not show any connections
	[#1494](https://github.com/weaveworks/scope/issues/1494)
- Missing edges when using Docker's IPAM driver
	[#1563](https://github.com/weaveworks/scope/issues/1563)
- Duplicate stack in "by image" view
	[#1521](https://github.com/weaveworks/scope/issues/1521)

Documentation:
- Clarify kubectl version matching
	[#1582](https://github.com/weaveworks/scope/pull/1582)
- updated Weave Cloud and clarified setup
	[#1586](https://github.com/weaveworks/scope/pull/1586)

Internal improvements and cleanup:
- Add Identity middleware
	[#1574](https://github.com/weaveworks/scope/pull/1574)
- Rewrite net/http.Request.{URL.Path,RequestURI} consistently
	[#1555](https://github.com/weaveworks/scope/pull/1555)
- Add Marathon JSON for launching on minimesos cluster
	[#1509](https://github.com/weaveworks/scope/pull/1509)
- Circle integration for auto docs publishing.
	[#1517](https://github.com/weaveworks/scope/pull/1517)
- Tag scope images on docker hub as we do in service
	[#1572](https://github.com/weaveworks/scope/pull/1572)
- Scope slow: improve error messages for debugging
	[#1534](https://github.com/weaveworks/scope/pull/1534)
- circle.yml: deploy non-master branches
	[#1535](https://github.com/weaveworks/scope/pull/1535)
- Add docker hub badge
	[#1540](https://github.com/weaveworks/scope/pull/1540)
- Increase test replicas
	[#1529](https://github.com/weaveworks/scope/pull/1529)
- Ignore IPv6 addresses in Docker reporter
	[#1552](https://github.com/weaveworks/scope/pull/1552)

Weave Cloud related changes:
- Add probe version header to probe requests
	[#1564](https://github.com/weaveworks/scope/pull/1564)
- Fetch non-cached reports in parallel
	[#1554](https://github.com/weaveworks/scope/pull/1554)
- Various fix ups for multitenancy
	[#1533](https://github.com/weaveworks/scope/pull/1533)
- Use NATS for shortcut reports in the service.
	[#1568](https://github.com/weaveworks/scope/pull/1568)
- If we don't get a path name from the router, make one up from the url.
	[#1570](https://github.com/weaveworks/scope/pull/1570)
- Log errors in response to http requests.
	[#1569](https://github.com/weaveworks/scope/pull/1569)
- Put reports in S3; add in process caching
	[#1545](https://github.com/weaveworks/scope/pull/1545)
- Use smart merger in the DynamoDB collector.
	[#1543](https://github.com/weaveworks/scope/pull/1543)
- Allow user to specify table name and queue prefix.
	[#1538](https://github.com/weaveworks/scope/pull/1538)
- Get route name before munging request
	[#1590](https://github.com/weaveworks/scope/pull/1590)


## Release 0.15.0

Highlights:

This release comes with:
  * Search: new smart search field that allows you to filter what you can see by
    container names, all kinds of metadata, e.g., IP addresses, and metric
    comparisons, e.g., CPU > 50%.
  * Enhanced Kubernetes Visualization: namespace filters, ReplicaSet/Deployment
    views, extra metadata, better navigation, show Pod logs, delete Pods,
	bugfixes and more ...
  * Scope App performance improvements: ~3X reduction in CPU consumption.


New features and enhancements:
- New search field
	[#1429](https://github.com/weaveworks/scope/pull/1429)
	[#1499](https://github.com/weaveworks/scope/pull/1499)
- Kubernetes improvements:
  - Deployment and Replica Set views
		[#1436](https://github.com/weaveworks/scope/pull/1436)
  - Add scale up/down controls on deployments, replica sets, and replication controllers
		[#1451](https://github.com/weaveworks/scope/pull/1451)
  - Filter by Kubernetes Namespaces
		[#1386](https://github.com/weaveworks/scope/pull/1386)
  - Remove App->Probe deployment ordering restriction
		[#1433](https://github.com/weaveworks/scope/pull/1433)
  - Show Pod IP and # container in the children table in details panel.
		[#1435](https://github.com/weaveworks/scope/pull/1435)
		[#1409](https://github.com/weaveworks/scope/pull/1409)
  - Add pod delete controls
		[#1368](https://github.com/weaveworks/scope/pull/1368)
  - Show the k8s load balancer IP if it is set
		[#1378](https://github.com/weaveworks/scope/pull/1378)
  - Show number of pods in service
		[#1352](https://github.com/weaveworks/scope/pull/1352)
  - Filter GKE system containers
		[#1438](https://github.com/weaveworks/scope/pull/1438)
- Show k8s labels and container env vars in the details panel
	[#1342](https://github.com/weaveworks/scope/pull/1342)
	[#1465](https://github.com/weaveworks/scope/pull/1465)
- Implement `scope help`
	[#1357](https://github.com/weaveworks/scope/pull/1357)
	[#1419](https://github.com/weaveworks/scope/pull/1419)
- Add swarm-agent, swarm-agent master to system container filter
	[#1356](https://github.com/weaveworks/scope/pull/1356)
- Add control for removing stopped docker containers.
	[#1290](https://github.com/weaveworks/scope/pull/1290)
- Add a button to download the report as JSON
	[#1365](https://github.com/weaveworks/scope/pull/1365)
- Use reverse-resolved DNS info in the connections table.
	[#1359](https://github.com/weaveworks/scope/pull/1359)
- Add a 'Unmanaged' node to k8s views which included non-k8s containers.
	[#1350](https://github.com/weaveworks/scope/pull/1350)
- Support docker rename events
	[#1332](https://github.com/weaveworks/scope/pull/1332)
- Strip image version from parent links
	[#1348](https://github.com/weaveworks/scope/pull/1348)
- Add Docker for Mac support
	[#1448](https://github.com/weaveworks/scope/pull/1448)

Performance improvements:
- Scope App:
  - A log(n) complexity report merger
		[#1418](https://github.com/weaveworks/scope/pull/1418)
        [#1447](https://github.com/weaveworks/scope/pull/1447)
  - Don't merge nodes in the rendering pipeline
		[#1398](https://github.com/weaveworks/scope/pull/1398)
  - Pass nil for the decorator in the rendering pipeline when possible
		[#1397](https://github.com/weaveworks/scope/pull/1397)
- Scope Probe:
  - Precompute base of the container nodes
		[#1456](https://github.com/weaveworks/scope/pull/1456)

Bug fixes:
- Correctly attribute DNAT-ed short-lived connections
	[#1410](https://github.com/weaveworks/scope/pull/1410)
- Don't attribute conntracked connections to k8s pause containers.
	[#1415](https://github.com/weaveworks/scope/pull/1415)
- Don't show kubernetes views if not running kubernetes
	[#1364](https://github.com/weaveworks/scope/issues/1364)
- Missing pod names in kubernetes' pod view and Pause containers don't show as children of pods
	[#1412](https://github.com/weaveworks/scope/pull/1412)
- Fix grouped node count for filtered children nodes
	[#1371](https://github.com/weaveworks/scope/pull/1371)
- Don't show container labels on container images
	[#1374](https://github.com/weaveworks/scope/pull/1374)
- `docker rm -f`ed containers linger
	[#1072](https://github.com/weaveworks/scope/issues/1072)
- Somehow internet node goes missing, yet edges are there
	[#1304](https://github.com/weaveworks/scope/pull/1304)
- Node IDs with / leads to redirect loop when scope is mounted under a path with slash redirect
	[#1335](https://github.com/weaveworks/scope/issues/1335)
- Ignore conntracked connections on which we never saw an update
	[#1466](https://github.com/weaveworks/scope/issues/466)
- Containers incorrectly attributed to host
	[#1472](https://github.com/weaveworks/scope/issues/1472)
- k8s: Unexpected edge to the Internet node
	[#1469](https://github.com/weaveworks/scope/issues/1469)
- When user supplies IP addr on command line, we don't try to connect to localhost
	[#1477](https://github.com/weaveworks/scope/issues/1477)
- Wrong host labels on container nodes
	[#1501](https://github.com/weaveworks/scope/issues/1501)

Documentation:
- Restructured Scope Docs
	[#1416](https://github.com/weaveworks/scope/pull/1416)
	[#1479](https://github.com/weaveworks/scope/pull/1479)
- Add ECS instructions and badge to README
	[#1392](https://github.com/weaveworks/scope/pull/1392)
- Document how to access the Scope UI in k8s
	[#1426](https://github.com/weaveworks/scope/pull/1426)
- Update readme to express that daemon sets won't schedule on unschedulable nodes prior to kubernetes 1.2
	[#1434](https://github.com/weaveworks/scope/pull/1434)


Internal improvements and cleanup:
- Migrate from Flux to Redux
	[#1388](https://github.com/weaveworks/scope/pull/1388)
- Add kubernetes checkpoint flag
	[#1391](https://github.com/weaveworks/scope/pull/1391)
- Add generic path rewrite middleware
	[#1381](https://github.com/weaveworks/scope/pull/1381)
- Report hostname and version in probe struct, and version in host node.
	[#1377](https://github.com/weaveworks/scope/pull/1377)
- Reorganise the render/ package
	[#1360](https://github.com/weaveworks/scope/pull/1360)
- Asset fingerprinting
	[#1354](https://github.com/weaveworks/scope/pull/1354)
- Upgrade to go1.6.2
	[#1362](https://github.com/weaveworks/scope/pull/1362)
- Add buffer to mockPublisher channel to prevent deadlock between Publish() and Stop()
	[#1358](https://github.com/weaveworks/scope/pull/1358)
- Add explicit group node summariser instead of doing it in the other summaries
	[#1327](https://github.com/weaveworks/scope/pull/1327)
- Don't build codecs for render/ package anymore.
	[#1345](https://github.com/weaveworks/scope/pull/1345)
- Measure report sizes
	[#1458](https://github.com/weaveworks/scope/pull/1458)

## Release 0.14.0

Highlights:

This release comes with two main new features.
  * Probe plugins: Now you can create your HTTP-based plugin to provide new metrics
    and display them in Scope. You can read more about it and see some examples
    [here](https://github.com/weaveworks/scope/tree/master/examples/plugins).
  * Metrics on canvas: Metrics are now displayed on the nodes and not just on
    the details panel, starting with CPU and memory consumption.

Also, the performance of the UI has been considerably improved and the 100-node
rendering limit has been lifted.


New features and enhancements:
- Probe plugins
	[#1126](https://github.com/weaveworks/scope/pull/1126)
	[#1277](https://github.com/weaveworks/scope/pull/1277)
	[#1280](https://github.com/weaveworks/scope/pull/1280)
	[#1283](https://github.com/weaveworks/scope/pull/1283)
- Metrics on canvas
	[#1105](https://github.com/weaveworks/scope/pull/1105)
	[#1204](https://github.com/weaveworks/scope/pull/1204)
	[#1225](https://github.com/weaveworks/scope/pull/1225)
	[#1243](https://github.com/weaveworks/scope/issues/1243)
- Node details panel improvements
  - Add connection tables
	[#1017](https://github.com/weaveworks/scope/pull/1017)
	[#1248](https://github.com/weaveworks/scope/pull/1248)
  - Layout: make better use of column space
		[#1272](https://github.com/weaveworks/scope/pull/1272)
  - Sparklines
    - Update every second and show 60sec history
			[#795](https://github.com/weaveworks/scope/pull/795)
    - Apply format to tooltips in hovers
			[#1230](https://github.com/weaveworks/scope/pull/1230)
  - Sort numerical entries (e.g. image counts, process IDs) as expected
		[#1125](https://github.com/weaveworks/scope/pull/1125)
  - Remove load5 and load15 metrics
		[#1274](https://github.com/weaveworks/scope/pull/1274)
- Graph view improvements
  - Node filtering improvements
      - Introduce three-way filtering selectors (e.g. choose from _System containers_, _Application containers_ or _Both_)
			[#1159](https://github.com/weaveworks/scope/pull/1159)
      - Maintain node-filtering selection across subviews (e.g. _Containers by ID_ and _Containers by Image_)
			[#1237](https://github.com/weaveworks/scope/pull/1237)
  - Refine maximum length of node names
		[#1263](https://github.com/weaveworks/scope/issues/1263)
		[#1255](https://github.com/weaveworks/scope/pull/1255)
  - Refine border-width of nodes
		[#1138](https://github.com/weaveworks/scope/pull/1138)
		[#1120](https://github.com/weaveworks/scope/pull/1120)
  - Cache pan/zoom per topology
	[#1261](https://github.com/weaveworks/scope/pull/1261)
- Enable launching terminals in hosts
	[#1208](https://github.com/weaveworks/scope/pull/1208)
- Allow pausing the UI through a button
	[#1106](https://github.com/weaveworks/scope/pull/1106)
- Split the internet node for incoming vs outgoing connections.
	[#566](https://github.com/weaveworks/scope/pull/566)
- Show k8s pod status
	[#1289](https://github.com/weaveworks/scope/pull/1289)
- Allow customizing Scope's hostname in Weave Net with `scope launch --weave.hostname`
	[#1041](https://github.com/weaveworks/scope/pull/1041)
- Rename `--weave.router.addr` to `--weave.addr` in the probe for consistency with the app
	[#1060](https://github.com/weaveworks/scope/issues/1060)
- Support new `sha256:` Docker image identifiers
	[#1161](https://github.com/weaveworks/scope/pull/1161)
	[#1184](https://github.com/weaveworks/scope/pull/1184)
- Handle server disconnects gracefully in the UI
	[#1140](https://github.com/weaveworks/scope/pull/1140)


Performance improvements:
  - Performance improvements for UI canvas
	[#1186](https://github.com/weaveworks/scope/pull/1186)
	[#1236](https://github.com/weaveworks/scope/pull/1236)
	[#1239](https://github.com/weaveworks/scope/pull/1239)
	[#1262](https://github.com/weaveworks/scope/pull/1262)
	[#1259](https://github.com/weaveworks/scope/pull/1259)
  - Reduce CPU consumption if UI cannot connect to backend
	[#1229](https://github.com/weaveworks/scope/pull/1229)


Bug Fixes:
- Scope app doesn't correctly expire old reports
	[#1286](https://github.com/weaveworks/scope/issues/1286)
- Container nodes appear without a host label
	[#1065](https://github.com/weaveworks/scope/issues/1065)
- Resizing the window and zooming in/out can confuse window size
	[#1180](https://github.com/weaveworks/scope/issues/1096)
- Link from container -> Pod doesn't work
    [#1180](https://github.com/weaveworks/scope/issues/1293)
- Various websocket and pipe fixes.
	[#1172](https://github.com/weaveworks/scope/pull/1172)
	[#1175](https://github.com/weaveworks/scope/pull/1175)
- Make `--app-only` only run the app and not probe
	[#1067](https://github.com/weaveworks/scope/pull/1067)
- Exported SVG file throws "CANT" error in Adobe Illustrator
	[#1144](https://github.com/weaveworks/scope/issues/1144)
- Docker labels not rendering correctly
	[#1284](https://github.com/weaveworks/scope/issues/1284)
- Error when parsing kernel version in `/proc` background reader
	[#1136](https://github.com/weaveworks/scope/issues/1136)
- Opening the terminal doesn't open work for some containers
	[#1195](https://github.com/weaveworks/scope/issues/1195)
- Terminals: Try to figure what shell to use instead of simply running `/bin/sh`
	[#1069](https://github.com/weaveworks/scope/pull/1069)
- Fix embedded logo size for Safari
	[#1084](https://github.com/weaveworks/scope/pull/1084)
- Don't read from app.Version before we initialise it
	[#1163](https://github.com/weaveworks/scope/pull/1163)
- Don't show multiple pseudo nodes in the host view for the same IP
	[#1155](https://github.com/weaveworks/scope/issues/1155)
- Fix race conditions detected by race detector from Go 1.6
	[#1192](https://github.com/weaveworks/scope/issues/1192)
	[#1087](https://github.com/weaveworks/scope/issues/1087)


Documentation:
- Provide Docker Compose examples for launching the Scope probe with the Scope Cloud Service
	[#1146](https://github.com/weaveworks/scope/pull/1146)

Experimental features:
- Update demo for tracer
	[#1157](https://github.com/weaveworks/scope/pull/1157)


Service-mode related changes:
- Add `/api/probes` endpoint
	[#1265](https://github.com/weaveworks/scope/pull/1265)
- Multitenancy-support improvements
	[#996](https://github.com/weaveworks/scope/pull/996)
	[#1150](https://github.com/weaveworks/scope/pull/1150)
	[#1200](https://github.com/weaveworks/scope/pull/1200)
	[#1241](https://github.com/weaveworks/scope/pull/1241)
	[#1209](https://github.com/weaveworks/scope/pull/1209)
	[#1232](https://github.com/weaveworks/scope/pull/1232)


Internal improvements and cleanup:
- Make node/edge highlighter objects immutable in app store
	[#1173](https://github.com/weaveworks/scope/pull/1173)
- Make cached edge processing more robust
	[#1254](https://github.com/weaveworks/scope/pull/1254)
- Make app-store's topologies object immutable
	[#1167](https://github.com/weaveworks/scope/pull/1167)
- Fix TestCollector test
	[#1070](https://github.com/weaveworks/scope/pull/1070)
- Update docker client, to get better state strings in the UI
	[#1235](https://github.com/weaveworks/scope/pull/1235)
- Upgrade to go1.6
	[#1077](https://github.com/weaveworks/scope/pull/1077)
- React/lodash/babel upgrades + updated linting (linted)
	[#1171](https://github.com/weaveworks/scope/pull/1171)
- Remove address topology
	[#1127](https://github.com/weaveworks/scope/pull/1127)
- Add vendoring docs
	[#1180](https://github.com/weaveworks/scope/pull/1180)
- Fix make client-start
	[#1210](https://github.com/weaveworks/scope/pull/1210)
- Downgrade react-motion
	[#1183](https://github.com/weaveworks/scope/pull/1183)
- Make bin/release work on a mac.
	[#887](https://github.com/weaveworks/scope/pull/887)
- Add various middleware to app.
	[#1234](https://github.com/weaveworks/scope/pull/1234)
- Make unconteinerized build work on OSX
	[#1028](https://github.com/weaveworks/scope/pull/1028)
- Remove codecgen-generated file before building package
	[#1135](https://github.com/weaveworks/scope/pull/1135)
- Build/install packages before invoking codecgen
	[#1042](https://github.com/weaveworks/scope/pull/1042)
- circle.yml: add variable $DOCKER_ORGANIZATION
	[#1083](https://github.com/weaveworks/scope/pull/1083)
- circle.yml: deploy on a personal hub account
	[#1055](https://github.com/weaveworks/scope/pull/1055)
- circle.yml: disable GCE builds when credentials are missing
	[#1054](https://github.com/weaveworks/scope/pull/1054)
- Clean out all the JS in the client build dir.
	[#1205](https://github.com/weaveworks/scope/pull/1205)
- Remove temporary files in the build container to shrink it down by ~100MB
	[#1206](https://github.com/weaveworks/scope/pull/1206)
- Update tools & build container to check for spelling mistakes
	[#1199](https://github.com/weaveworks/scope/pull/1199)
- Fix a couple of minor issue for goreportcard and add badge for it.
	[#1203](https://github.com/weaveworks/scope/pull/1203)

## Release 0.13.1

Bug Fixes:
- Make pipes work with scope.weave.works
  [#1099](https://github.com/weaveworks/scope/pull/1099)
  [#1085](https://github.com/weaveworks/scope/pull/1085)
  [#994](https://github.com/weaveworks/scope/pull/994)
- Don't panic when checking the version fails
  [#1117](https://github.com/weaveworks/scope/pull/1117)

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
- Don't treat missing node as UI error
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
- Omit controls field from json if empty.
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
  [#531](https://github.com/weaveworks/scope/issues/531)
- Rerun background conntrack processes if they fail
  [#581](https://github.com/weaveworks/scope/issues/581)
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

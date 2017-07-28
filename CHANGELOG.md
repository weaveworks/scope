
## Release 1.6.1

This is a re-release of 1.6.0. The official build for 1.6.0 inadvertently
included outdated versions of some components, which introduced security issues.

## Release 1.6.0

Highlights:
- New Kubernetes Controllers view
- Add Kubernetes Stateful Sets and Cron Jobs
- Various small improvements and performance work

New features and enhancements:
- kubernetes: Add StatefulSets and CronJobs
	[#2724](https://github.com/weaveworks/scope/pull/2724)
- Make Resource view nodes clickable
	[#2679](https://github.com/weaveworks/scope/pull/2679)
- Keep topology nav visible if selected
	[#2709](https://github.com/weaveworks/scope/pull/2709)
- Show multiple relatives in the nodes-grid view
	[#2648](https://github.com/weaveworks/scope/pull/2648)
- Remove type filter in controllers view
	[#2670](https://github.com/weaveworks/scope/pull/2670)
- Remove replica sets
	[#2661](https://github.com/weaveworks/scope/pull/2661)
- add k8s combined view
	[#2552](https://github.com/weaveworks/scope/pull/2552)
- Gather Weave Net plugin and proxy info from report
	[#2719](https://github.com/weaveworks/scope/pull/2719)


Performance improvements:
- optimisation: don't copy report stream unnecessarily
	[#2736](https://github.com/weaveworks/scope/pull/2736)
- new full reports are more important than old and shortcut reports
	[#2743](https://github.com/weaveworks/scope/pull/2743)
- increase default conntrack buffer size
	[#2739](https://github.com/weaveworks/scope/pull/2739)
- Use Kubernetes node name to filter pods if possible
	[#2556](https://github.com/weaveworks/scope/pull/2556)
- refactor: remove unnecessary and dead Copy()
	[#2675](https://github.com/weaveworks/scope/pull/2675)
- performance: only color connected once
	[#2635](https://github.com/weaveworks/scope/pull/2635)
- fast network membership check
	[#2625](https://github.com/weaveworks/scope/pull/2625)
- memoize isKnownServices for improved performance
	[#2617](https://github.com/weaveworks/scope/pull/2617)
- faster matching of known services
	[#2613](https://github.com/weaveworks/scope/pull/2613)

Bug fixes and minor improvements:
- k8s: Use 'DaemonSet', 'StatefulSet' etc instead of 'Daemon Set', 'Stateful Set'
	[#2757](https://github.com/weaveworks/scope/pull/2757)
- maximize report publishing timeout
	[#2756](https://github.com/weaveworks/scope/pull/2756)
- do not back off on timeouts when sending reports
	[#2746](https://github.com/weaveworks/scope/pull/2746)
- Fix Pods number in graph not updating (minor label)
	[#2728](https://github.com/weaveworks/scope/pull/2728)
- defend against nils
	[#2734](https://github.com/weaveworks/scope/pull/2734)
- Fix New Version notification not showing
	[#2720](https://github.com/weaveworks/scope/pull/2720)
- render: In minor labels, display '0 things' instead of blank if zero things present
	[#2726](https://github.com/weaveworks/scope/pull/2726)
- Reset nodes in frontend when scope-app restarted
	[#2713](https://github.com/weaveworks/scope/pull/2713)
- Keep topo nav visible if subnav selected
	[#2710](https://github.com/weaveworks/scope/pull/2710)
- don't miss, or fail to forget, initial connections
	[#2704](https://github.com/weaveworks/scope/pull/2704)
- bump tcptracer-bpf version
	[#2705](https://github.com/weaveworks/scope/pull/2705)
- fix ebpf init race segfault
	[#2695](https://github.com/weaveworks/scope/pull/2695)
- Make graph layout zoom limits constant
	[#2678](https://github.com/weaveworks/scope/pull/2678)
- Last line of defense against overlapping nodes in graph layout
	[#2688](https://github.com/weaveworks/scope/pull/2688)
- Fix `yarn pack` ignoring directory cli flag
	[#2694](https://github.com/weaveworks/scope/pull/2694)
- Show table overflow only if limit exceeded by 2+
	[#2683](https://github.com/weaveworks/scope/pull/2683)
- render/pod: Fix a typo in Map2Parent where UnmanagedID will always be used for noParentsPseudoID
	[#2685](https://github.com/weaveworks/scope/pull/2685)
- don't show container count in host detail panel image list
	[#2682](https://github.com/weaveworks/scope/pull/2682)
- correct determination of a host's container images
	[#2680](https://github.com/weaveworks/scope/pull/2680)
- Prevents 6 digit pids from being truncated in details panel/table mode
	[#2666](https://github.com/weaveworks/scope/pull/2666)
- correct polarity of initial connections
	[#2645](https://github.com/weaveworks/scope/pull/2645)
- ensure connections from /proc/net/tcp{,6} get the right pid
	[#2639](https://github.com/weaveworks/scope/pull/2639)
- Avoid race conditions in DNSSnooper's cached domains
	[#2637](https://github.com/weaveworks/scope/pull/2637)
- Fix issues with union types
	[#2633](https://github.com/weaveworks/scope/pull/2633)
- Fix typo in site/plugins.md
	[#2624](https://github.com/weaveworks/scope/pull/2624)
- correct `nodeSummaryGroupSpec`
	[#2631](https://github.com/weaveworks/scope/pull/2631)
- Ignore ipv6
	[#2622](https://github.com/weaveworks/scope/pull/2622)
- Fix the table sorting order bug for numerical values
	[#2587](https://github.com/weaveworks/scope/pull/2587)
- Fix zoom for `npm start`
	[#2605](https://github.com/weaveworks/scope/pull/2605)
- fix error when docker DAEMON is running with user namespace enabled.
	[#2582](https://github.com/weaveworks/scope/pull/2582)
- Do not read tcp6 files if TCP version 6 isn't supported
	[#2604](https://github.com/weaveworks/scope/pull/2604)
- Elide token-only credentials in cli arguments
	[#2593](https://github.com/weaveworks/scope/pull/2593)
- do not filter endpoints by procspied/ebpf in renderers
	[#2652](https://github.com/weaveworks/scope/pull/2652)

Internal improvements and cleanup:
- Pass build tags to unit tests
	[#2618](https://github.com/weaveworks/scope/pull/2618)
- Allows to skip client build when doing make prog/scope
	[#2732](https://github.com/weaveworks/scope/pull/2732)
- only pass WEAVESCOPE_DOCKER_ARGS to actual probe/app start
	[#2715](https://github.com/weaveworks/scope/pull/2715)
- Set package.json version to 0.0.0
	[#2692](https://github.com/weaveworks/scope/pull/2692)
- simplify connection join
	[#2714](https://github.com/weaveworks/scope/pull/2714)
- EbpfTracker refactoring / cleanup
	[#2699](https://github.com/weaveworks/scope/pull/2699)
- Yarn prefixes version with `v` when packing
	[#2691](https://github.com/weaveworks/scope/pull/2691)
- don't use eBPF in a couple of tests
	[#2690](https://github.com/weaveworks/scope/pull/2690)
- Update README/Makefile/package.json to use yarn
	[#2676](https://github.com/weaveworks/scope/pull/2676)
- render/pod: Remove unused options and incorrect code
	[#2673](https://github.com/weaveworks/scope/pull/2673)
- Use new k8s go client
	[#2659](https://github.com/weaveworks/scope/pull/2659)
- Update github.com/weaveworks/common & dependencies (needs go1.8)
	[#2570](https://github.com/weaveworks/scope/pull/2570)
- Publish updated OpenShift instructions (close #2485)
	[#2657](https://github.com/weaveworks/scope/pull/2657)
- make integration tests pass with latest Weave Net release (2.0)
	[#2641](https://github.com/weaveworks/scope/pull/2641)
- Improved rendering order of nodes/edges in Graph View
	[#2623](https://github.com/weaveworks/scope/pull/2623)
- refactor: extract a couple of heavily used constants
	[#2632](https://github.com/weaveworks/scope/pull/2632)
- Use latest go1.8.3
	[#2626](https://github.com/weaveworks/scope/pull/2626)
- Use Go 1.8
	[#2621](https://github.com/weaveworks/scope/pull/2621)
- Use 127.0.0.1 instead of localhost, more
	[#2554](https://github.com/weaveworks/scope/pull/2554)
- Moved highlighted nodes/edges info to selectors
	[#2584](https://github.com/weaveworks/scope/pull/2584)
- rationalise report set usage
	[#2671](https://github.com/weaveworks/scope/pull/2671)
- ignore endpoints with >1 adjacency in process rendering
	[#2668](https://github.com/weaveworks/scope/pull/2668)
- Honor DOCKER_* env variables in probe and app
	[#2649](https://github.com/weaveworks/scope/pull/2649)

Weave Cloud related changes:
- Back off when writing to Dynamo and S3
	[#2723](https://github.com/weaveworks/scope/pull/2723)
- Time travel redesign
	[#2651](https://github.com/weaveworks/scope/pull/2651)
- Make API calls with time travel timestamp
	[#2600](https://github.com/weaveworks/scope/pull/2600)

## Release 1.5.1

Bugfix patch release

Bug fixes:
- initial connections have wrong polarity
	[#2644](https://github.com/weaveworks/scope/issues/2644)
- connection to dead process associated with different process
	[#2638](https://github.com/weaveworks/scope/pull/2638)

## Release 1.5.0

Highlights:
- More accurate and cheaper connection tracking with eBPF, which now is enabled by default.
- Bug fixes and performance improvements.

New features and enhancements:
- Enable eBPF tracking by default
	[#2535](https://github.com/weaveworks/scope/pull/2535)
- Elide url passwords in cli arguments
	[#2568](https://github.com/weaveworks/scope/pull/2568)

Performance improvements:
- drop addr and port from Endpoint.Latest map
	[#2581](https://github.com/weaveworks/scope/pull/2581)
- parallel reduce
	[#2561](https://github.com/weaveworks/scope/pull/2561)
- don't read all of /proc when probe.proc.spy=false
	[#2557](https://github.com/weaveworks/scope/pull/2557)
- optimise: don't sort in NodeSet.ForEach
	[#2548](https://github.com/weaveworks/scope/pull/2548)
- encode empty ps.Maps as nil
	[#2547](https://github.com/weaveworks/scope/pull/2547)

Bug fixes:
- re-target app clients when name resolution changes
	[#2579](https://github.com/weaveworks/scope/pull/2579)
- correct type for "Observed Gen."
	[#2572](https://github.com/weaveworks/scope/pull/2572)
- Back off upon errored kubernetes api requests
	[#2562](https://github.com/weaveworks/scope/pull/2562)
- Close eBPF tracker cleanly
	[#2541](https://github.com/weaveworks/scope/pull/2541)
- Simplify connection tracker init and fix procfs scan fallback
	[#2539](https://github.com/weaveworks/scope/pull/2539)
- Guard against null DaemonSet store
	[#2538](https://github.com/weaveworks/scope/pull/2538)

Internal improvements and cleanup:
- es6ify server.js and include in eslint
	[#2560](https://github.com/weaveworks/scope/pull/2560)
- Fix prog/main_test.go
	[#2567](https://github.com/weaveworks/scope/pull/2567)
- Fix incomplete dependencies for `make scope/prog`
	[#2563](https://github.com/weaveworks/scope/pull/2563)
- bump package.json version to current scope version
	[#2555](https://github.com/weaveworks/scope/pull/2555)
- simplify connection join
	[#2559](https://github.com/weaveworks/scope/pull/2559)
- Use map helpers
	[#2546](https://github.com/weaveworks/scope/pull/2546)
- add copyreport utility
	[#2542](https://github.com/weaveworks/scope/pull/2542)

Weave Cloud related changes:
- Time travel control
	[#2524](https://github.com/weaveworks/scope/pull/2524)
- Add app capabilities to /api endpoint
	[#2575](https://github.com/weaveworks/scope/pull/2575)


## Release 1.4.0

Highlights:
- New Docker Swarm view
- New Kubernetes DaemonSets view 
- Probe performance improvements
- Many bugfixes

New features and enhancements:
- Add Docker Swarm view
	[#2444](https://github.com/weaveworks/scope/pull/2444)
	[#2452](https://github.com/weaveworks/scope/pull/2452)
	[#2450](https://github.com/weaveworks/scope/pull/2450)
- Kuebrnetes: add daemonsets
	[#2526](https://github.com/weaveworks/scope/pull/2526)	
- Canvas zoom control
	[#2513](https://github.com/weaveworks/scope/pull/2513)
- Consistent resource consumption info in the resource view
	[#2499](https://github.com/weaveworks/scope/pull/2499)
- k8s: show all namespaces by default
	[#2522](https://github.com/weaveworks/scope/pull/2522)
- Hide container image status for pseudo nodes
	[#2520](https://github.com/weaveworks/scope/pull/2520)
- Break out some Azure-based services from "The Internet"
	[#2521](https://github.com/weaveworks/scope/pull/2521)
- Remove zoom on double-click
	[#2457](https://github.com/weaveworks/scope/pull/2457)
- allow disabling of weaveDNS advertising/lookup
	[#2445](https://github.com/weaveworks/scope/pull/2445)

Performance improvements:
- process walker perfs: optimize readLimits and readStats
	[#2491](https://github.com/weaveworks/scope/pull/2491)
- proc walker: optimize open file counter
	[#2456](https://github.com/weaveworks/scope/pull/2456)
- eliminate excessive calls to mtime.Now()
	[#2486](https://github.com/weaveworks/scope/pull/2486)
- Msgpack perf: write psMap out directly
	[#2466](https://github.com/weaveworks/scope/pull/2466)
- proc_linux: don't exec `getNetNamespacePathSuffix()` on every walk
	[#2453](https://github.com/weaveworks/scope/pull/2453)
- gzip: change compression level to the default
	[#2437](https://github.com/weaveworks/scope/pull/2437)

Bug fixes:
- Let conntrack track non-NATed short-lived connections
	[#2527](https://github.com/weaveworks/scope/pull/2527)
- Re-enable pod shortcut reports
	[#2528](https://github.com/weaveworks/scope/pull/2528)
- ebpf connection tracker: perf map fixes
	[#2507](https://github.com/weaveworks/scope/pull/2507)
- ebpf: handle fdinstall events from tcptracer-bpf (aka "accept before kretprobe" issue)
	[#2518](https://github.com/weaveworks/scope/pull/2518)	
- Fix arrow heads positioning
	[#2505](https://github.com/weaveworks/scope/pull/2505)
- Avoid null dereferences in ECS client
	[#2514](https://github.com/weaveworks/scope/pull/2514)
	[#2515](https://github.com/weaveworks/scope/pull/2515)
- api_topologies: Don't put namespace filters on containers-by-dns/image
	[#2506](https://github.com/weaveworks/scope/pull/2506)
- Log specific error when deployments are not supported
	[#2501](https://github.com/weaveworks/scope/pull/2501)
- Missing namespace option in url state breaks filters
	[#2490](https://github.com/weaveworks/scope/issues/2490)
- Metric selector not showing pinned metric highlighted
	[#2467](https://github.com/weaveworks/scope/issues/2467)
- Fixed view mode switching keyboard shortcuts
	[#2471](https://github.com/weaveworks/scope/pull/2471)
- don't lie about reachable address
	[#2443](https://github.com/weaveworks/scope/pull/2443)
- Fix node highlight for all shapes
	[#2430](https://github.com/weaveworks/scope/pull/2430)
- View mode selector not responding well to resize
	[#2396](https://github.com/weaveworks/scope/issues/2396)
- Empty metric selector appearing as a dot
	[#2425](https://github.com/weaveworks/scope/issues/2425)
- Cloud node border too thin comparing to other nodes
	[#2417](https://github.com/weaveworks/scope/issues/2417)
- Table-mode: origin of details panel is not where clicked
	[#1754](https://github.com/weaveworks/scope/issues/1754)
- Table-mode: tooltip for "The Internet" is missing minor label
	[#1884](https://github.com/weaveworks/scope/issues/1884)
- Fixes loading of viewState from localStorage into URL
	[#2409](https://github.com/weaveworks/scope/pull/2409)
- Don't reset zoom on refresh layout
	[#2407](https://github.com/weaveworks/scope/pull/2407)
- Hide the opened help panel when clicking on the search bar icon
	[#2406](https://github.com/weaveworks/scope/pull/2406)

Documentation:
- Report data structure documentation
	[#2025](https://github.com/weaveworks/scope/pull/2025)
- Add multicolumn-table documentation
	[#2516](https://github.com/weaveworks/scope/pull/2516)
- Update k8s installation instructions
	[#2512](https://github.com/weaveworks/scope/pull/2512)
	[#2519](https://github.com/weaveworks/scope/pull/2519)
- Update install docs
	[#2257](https://github.com/weaveworks/scope/pull/2257)	
- Add plugin mention to scope readme
	[#2454](https://github.com/weaveworks/scope/pull/2454)
- Fix disabling Scope in the ECS AMI
	[#2435](https://github.com/weaveworks/scope/pull/2435)
- Add AMI docs into main docs, modified weave token instructions in one place
	[#2307](https://github.com/weaveworks/scope/pull/2307)
	[#2416](https://github.com/weaveworks/scope/pull/2416)
	[#2415](https://github.com/weaveworks/scope/pull/2415)

Internal improvements and cleanup:
- Reduce the number of places topologies are explicitly listed
	[#2436](https://github.com/weaveworks/scope/pull/2436)
- Use prop-types library to silence PropTypes deprecation warning
	[#2498](https://github.com/weaveworks/scope/pull/2498)
- Update node libraries
	[#2292](https://github.com/weaveworks/scope/pull/2292)
- Added search type variable
	[#2493](https://github.com/weaveworks/scope/pull/2493)
- Add website preview via Netlify
	[#2480](https://github.com/weaveworks/scope/pull/2480)
- Consisten spacing in Markdown headings
	[#2438](https://github.com/weaveworks/scope/pull/2438)
- only lint files in git ls-files, not .git/*
	[#2477](https://github.com/weaveworks/scope/pull/2477)
- publish master to dockerhub (again)
	[#2449](https://github.com/weaveworks/scope/pull/2449)
- scope script: Allow 'user' part of image name to be given by DOCKERHUB_USER env var
	[#2447](https://github.com/weaveworks/scope/pull/2447)
- Make various anonymous fields named
	[#2419](https://github.com/weaveworks/scope/pull/2419)
- vendor: update gobpf and tcptracer-bpf
	[#2428](https://github.com/weaveworks/scope/pull/2428)
- extras/dialer updates and fixes
	[#2350](https://github.com/weaveworks/scope/pull/2350)
- Update tcptracer-bpf and re-enable test 311
	[#2411](https://github.com/weaveworks/scope/pull/2411)
- Add check for old options
	[#2405](https://github.com/weaveworks/scope/pull/2405)
- shfmt: fix shell formatting
	[#2533](https://github.com/weaveworks/scope/pull/2533)

Weave Cloud related changes:
- close s3 response body to prevent leaks
	[#2442](https://github.com/weaveworks/scope/pull/2442)
- Add service images widget
	[#2487](https://github.com/weaveworks/scope/pull/2487)
- Add weavenet metrics to billing
	[#2504](https://github.com/weaveworks/scope/pull/2504)
- Calculate viewport dimensions from the scope-app div
	[#2473](https://github.com/weaveworks/scope/pull/2473)
- Added mixpanel tracking for some basic events
	[#2462](https://github.com/weaveworks/scope/pull/2462)
- Add NodeSeconds to billing emitter
	[#2422](https://github.com/weaveworks/scope/pull/2422)


## Release 1.3.0

Highlights:
- New resource usage view
- New Arrows in graph view to indicate connection directions
- [Weave Cloud Agent Docker-certified image](https://store.docker.com/images/f18f278a-54c1-4f25-b252-6e11112776c5)
- eBPF connection tracking (enabled with --probe.ebpf.connections=true)

New features and enhancements:
- Resource usage view
	[#2296](https://github.com/weaveworks/scope/pull/2296)
	[#2390](https://github.com/weaveworks/scope/pull/2390)
- Edge arrows
	[#2317](https://github.com/weaveworks/scope/pull/2317)
	[#2342](https://github.com/weaveworks/scope/pull/2342)
- Add eBPF connection tracking
	[#2135](https://github.com/weaveworks/scope/pull/2135)
	[#2327](https://github.com/weaveworks/scope/pull/2327)
	[#2336](https://github.com/weaveworks/scope/pull/2336)
	[#2366](https://github.com/weaveworks/scope/pull/2366)
-	View multiple Kubernetes namespaces at once
 	[#2404](https://github.com/weaveworks/scope/pull/2404)
- Exclude pause containers when rendering k8s topologies
	[#2338](https://github.com/weaveworks/scope/pull/2338)
- When k8s present, allow filtering of containers by namespace
	[#2285](https://github.com/weaveworks/scope/pull/2285)
	[#2348](https://github.com/weaveworks/scope/pull/2348)
	[#2362](https://github.com/weaveworks/scope/pull/2362)
- Add ECS Service scale up/down controls
	[#2197](https://github.com/weaveworks/scope/pull/2197)
- Improve error reporting when invoking weave script
	[#2335](https://github.com/weaveworks/scope/pull/2335)
- Add options to hide args and env vars
	[#2306](https://github.com/weaveworks/scope/pull/2306)
	[#2311](https://github.com/weaveworks/scope/pull/2311)
	[#2310](https://github.com/weaveworks/scope/pull/2310)
- Add loading indicator on topology option change
	[#2272](https://github.com/weaveworks/scope/pull/2272)
- report playback
	[#2301](https://github.com/weaveworks/scope/pull/2301)
- Show loading indicator on topology changes
	[#2232](https://github.com/weaveworks/scope/pull/2232)

Performance improvements:
- Map decode optimisations
	[#2364](https://github.com/weaveworks/scope/pull/2364)
- Remove LatestMap, to reduce memory allocation
	[#2351](https://github.com/weaveworks/scope/pull/2351)
- Decode via byte slice for memcache and file read
	[#2331](https://github.com/weaveworks/scope/pull/2331)
- quantise reports
	[#2305](https://github.com/weaveworks/scope/pull/2305)
- Layout rendering dynamic optimizations
	[#2221](https://github.com/weaveworks/scope/pull/2221)
	[#2265](https://github.com/weaveworks/scope/pull/2265)

Bug fixes:
- Pinned metric temporarily not displayed on mouse leave
	[#2397](https://github.com/weaveworks/scope/issues/2397)
- Search doesn't consider nodes of unloaded topologies
	[#2395](https://github.com/weaveworks/scope/issues/2393)
- Help panel height overflow in Containers view
	[#2352](https://github.com/weaveworks/scope/issues/2352)
- "Save canvas as SVG" button shown in table mode
	[#2354](https://github.com/weaveworks/scope/pull/2354)
- process with no cmd are shown with no name
	[#2315](https://github.com/weaveworks/scope/issues/2315)
- Throb animation is called on graph nodes even when the search query doesn't change
	[#2255](https://github.com/weaveworks/scope/issues/2255)
- pod names missing
	[#2258](https://github.com/weaveworks/scope/issues/2258)
- parse --probe-only as intended
	[#2300](https://github.com/weaveworks/scope/pull/2300)
- Graph view zoom states get reset when switching to table view
	[#2254](https://github.com/weaveworks/scope/issues/2254)
- graph not rendered top-down, despite lack of cycles
	[#2267](https://github.com/weaveworks/scope/issues/2267)
- Hide Uncontained filter in DNS view not hiding uncontained
	[#2170](https://github.com/weaveworks/scope/issues/2170)

Documentation:
- Documentation improvements
	[#2252](https://github.com/weaveworks/scope/pull/2252)
- Removed missed merge text and made terminology consistent
	[#2289](https://github.com/weaveworks/scope/pull/2289)

Internal improvements and cleanup:
- integration test: disable flaky test 311
	[#2380](https://github.com/weaveworks/scope/pull/2380)
- Add job to trigger service-ui build
	[#2376](https://github.com/weaveworks/scope/pull/2376)
- Use yarn package manager
	[#2368](https://github.com/weaveworks/scope/pull/2368)
- integration tests: list containers for debugging
	[#2346](https://github.com/weaveworks/scope/pull/2346)
- scope: use same Docker args for early dry run
	[#2326](https://github.com/weaveworks/scope/pull/2326)
	[#2358](https://github.com/weaveworks/scope/pull/2358)
- Bump react version
	[#2339](https://github.com/weaveworks/scope/pull/2339)
- integration: disable tests with internet edge
	[#2314](https://github.com/weaveworks/scope/pull/2314)
- Secure integration tests
	[#2312](https://github.com/weaveworks/scope/pull/2312)
- integration: restart docker daemon after each test
	[#2298](https://github.com/weaveworks/scope/pull/2298)
- Changed ui-build-pkg job to use a docker container
	[#2281](https://github.com/weaveworks/scope/pull/2281)
- integration tests: fix scripts
	[#2225](https://github.com/weaveworks/scope/pull/2225)
- circle.yml: Fix ui upload step so it doesn't build twice
	[#2266](https://github.com/weaveworks/scope/pull/2266)

Weave Cloud related changes:
- Create cloud agent image
	[#2284](https://github.com/weaveworks/scope/pull/2284)
	[#2277](https://github.com/weaveworks/scope/pull/2277)
	[#2278](https://github.com/weaveworks/scope/pull/2278)
- Container Seconds should not be Container Nanoseconds
	[#2372](https://github.com/weaveworks/scope/pull/2372)
- Clear client polling and nodes state on dismount
	[#2361](https://github.com/weaveworks/scope/pull/2361)
- Fluent Billing Emitter
	[#2359](https://github.com/weaveworks/scope/pull/2359)
- Correct dynamoDB metric label
	[#2344](https://github.com/weaveworks/scope/pull/2344)
- Add logic to turn off network requests when Scope dismounts
	[#2290](https://github.com/weaveworks/scope/pull/2290)
	[#2340](https://github.com/weaveworks/scope/pull/2340)
- Load contrast stylesheet
	[#2256](https://github.com/weaveworks/scope/pull/2256)
- Consolidate API requests into single helper; added CSRF header
	[#2260](https://github.com/weaveworks/scope/pull/2260)
- Add logic to remove non-transferrable state when switching Cloud instances
	[#2237](https://github.com/weaveworks/scope/pull/2237)


## Release 1.2.1
This is a minor patch release.

Documentation
- Uploaded new cloud token screenshot
	[#2248](https://github.com/weaveworks/scope/pull/2248)
- Updated cloud description
	[#2249](https://github.com/weaveworks/scope/pull/2249)

Bugfixes
- Fix help menu not opening from 'search' hint
	[#2230](https://github.com/weaveworks/scope/pull/2230)
- Re-factor API URL generation code
	[#2202](https://github.com/weaveworks/scope/pull/2202)

Improvements
- Reintroduce probe checkpoint flags for kernel version and OS
	[#2224](https://github.com/weaveworks/scope/pull/2224)
- Upgraded xterm.js to 2.2.3
	[#2126](https://github.com/weaveworks/scope/pull/2126)
- Allow random seed in dialer
	[#2206](https://github.com/weaveworks/scope/pull/2206)
- Rename ECS Service node ids to be cluster;serviceName
	[#2186](https://github.com/weaveworks/scope/pull/2186)

## Release 1.2.0

Highlights:
- Performance improvements (both in UI and probes).
- Scope now requires Docker version >= 1.10.

New features and enhancements:
- ECS: service details panel should list its tasks
	[#2041](https://github.com/weaveworks/scope/issues/2041)
- Prioritize ecs topologies on initial load if available
	[#2105](https://github.com/weaveworks/scope/pull/2105)
- Add control status icon to Terminal header
	[#2087](https://github.com/weaveworks/scope/pull/2087)
- scope launch script improvements
	[#2077](https://github.com/weaveworks/scope/pull/2077)
	[#2093](https://github.com/weaveworks/scope/pull/2093)
- Maintain focus on hovered node table rows
	[#2115](https://github.com/weaveworks/scope/pull/2115)
- Add control to reset local view state
	[#2080](https://github.com/weaveworks/scope/pull/2080)
- Check that conntrack events are enabled in the kernel
	[#2112](https://github.com/weaveworks/scope/pull/2112)
- Hardcode 127.0.0.1 as loopback IP for default target
	[#2103](https://github.com/weaveworks/scope/pull/2103)
- prog/main: use flags.app.port for default target
	[#2096](https://github.com/weaveworks/scope/pull/2096)

Performance improvements:
- Graph layout optimizations
	[#2128](https://github.com/weaveworks/scope/pull/2128)
	[#2179](https://github.com/weaveworks/scope/pull/2179)
	[#2180](https://github.com/weaveworks/scope/pull/2180)
	[#2210](https://github.com/weaveworks/scope/pull/2210)
- Disable XML in conntrack parsing
	[#2095](https://github.com/weaveworks/scope/pull/2095)
	[#2118](https://github.com/weaveworks/scope/pull/2118)

Bug fixes:
- ECS reporter throttled by AWS API
	[#2050](https://github.com/weaveworks/scope/issues/2050)
- Already closed connections showing up in the containers tab
	[#2181](https://github.com/weaveworks/scope/issues/2181)
- Node details spinner Chrome display bug fix
	[#2177](https://github.com/weaveworks/scope/pull/2177)
- fix error when docker daemon is running with user namespace enabled.
	[#2161](https://github.com/weaveworks/scope/pull/2161)
	[#2176](https://github.com/weaveworks/scope/pull/2176)
- DNSSnooper: Support Dot1Q and limit decoding errors
	[#2155](https://github.com/weaveworks/scope/issues/2155)
- Contrast mode not working
	[#2165](https://github.com/weaveworks/scope/issues/2165)
	[#2138](https://github.com/weaveworks/scope/issues/2138)
- Scope does not create special nodes within the same VPC
	[#2163](https://github.com/weaveworks/scope/issues/2163)
- default view fails to select 'application containers only'
	[#2120](https://github.com/weaveworks/scope/issues/2120)
- ECS: Missing link to task on container details panel
	[#2040](https://github.com/weaveworks/scope/issues/2040)
- kubernetes reporter is broken on katacoda
	[#2049](https://github.com/weaveworks/scope/pull/2049)
- probe's procspy does not report netcat's half-duplex long-lived connections
	[#1972](https://github.com/weaveworks/scope/issues/1972)
- Sparkline component throws errors when a container is turned off
	[#2072](https://github.com/weaveworks/scope/pull/2072)
- Graph/table buttons don't resize nicely
	[#2056](https://github.com/weaveworks/scope/issues/2056)
- JS error on edges with lots of waypoints
	[#1187](https://github.com/weaveworks/scope/issues/1187)
- Fix two bugs caused by transition to D3 v4
	[#2048](https://github.com/weaveworks/scope/pull/2048)
- Popped out terminal styles don't quite align with in-scope terminal styles
	[#2209](https://github.com/weaveworks/scope/issues/2209)
- Radii of rounded-corner shape don't quite align
	[#2212](https://github.com/weaveworks/scope/issues/2212)

Documentation:
- Fix Scope arguments in Docker Compose installation docs
	[#2143](https://github.com/weaveworks/scope/pull/2143)
- Document how to run tests on website
	[#2131](https://github.com/weaveworks/scope/pull/2131)
- Follow redirections in curl when getting k8s resources
	[#2067](https://github.com/weaveworks/scope/pull/2067)

Internal improvements and cleanup:
- Embed and require Docker >= 1.10
	[#2190](https://github.com/weaveworks/scope/pull/2190)
- don't attempt to make 'make clean' work on old checkouts
	[#2189](https://github.com/weaveworks/scope/pull/2189)
- Fix linter errors
	[#2068](https://github.com/weaveworks/scope/pull/2068)
	[#2166](https://github.com/weaveworks/scope/pull/2166)
- Fix ownership issues with client/build-external
	[#2153](https://github.com/weaveworks/scope/pull/2153)
- Allow Scope UI to be installed as a Node module
	[#2144](https://github.com/weaveworks/scope/pull/2144)
	[#2159](https://github.com/weaveworks/scope/pull/2159)
- Upgrade container base image to alpine:3.5
	[#2158](https://github.com/weaveworks/scope/pull/2158)
- Use Sass instead of Less
	[#2141](https://github.com/weaveworks/scope/pull/2141)
- probe: refactor probeMain
	[#2148](https://github.com/weaveworks/scope/pull/2148)
- Update to go1.7.4
	[#2147](https://github.com/weaveworks/scope/pull/2147)
- Bump tools subtree and fix integration tests
	[#2136](https://github.com/weaveworks/scope/pull/2136)
- Add support for generic multicolumn tables
	[#2109](https://github.com/weaveworks/scope/pull/2109)
- extras/dialer: move dialer.go to sub directory
	[#2108](https://github.com/weaveworks/scope/pull/2108)
- Forward OS/Kernel version to checkpoint
	[#2101](https://github.com/weaveworks/scope/pull/2101)
- Fix force-push to master
	[#2094](https://github.com/weaveworks/scope/pull/2094)
- Upgraded eslint & eslint-config-airbnb
	[#2058](https://github.com/weaveworks/scope/pull/2058)
	[#2084](https://github.com/weaveworks/scope/pull/2084)
	[#2089](https://github.com/weaveworks/scope/pull/2089)
- ecs reporter: Fix some log lines that were passing *string instead of string
	[#2060](https://github.com/weaveworks/scope/pull/2060)
- Add flag for logging headers
	[#2086](https://github.com/weaveworks/scope/pull/2086)
- Add extras/dialer
	[#2082](https://github.com/weaveworks/scope/pull/2082)
- Remove wcloud
	[#2081](https://github.com/weaveworks/scope/pull/2081)
- Add client linting to CI config
	[#2076](https://github.com/weaveworks/scope/pull/2076)
- Importing lodash util functions explicitly
	[#2053](https://github.com/weaveworks/scope/pull/2053)
- procspy: use a Reader to copy the background reader buffer
	[#2020](https://github.com/weaveworks/scope/pull/2020)
- Use newly-created 'common' repo
	[#2061](https://github.com/weaveworks/scope/pull/2061)
- Fix all the npm library versions
	[#2057](https://github.com/weaveworks/scope/pull/2057)
- linter: fix punctuation and capitalization
	[#2021](https://github.com/weaveworks/scope/pull/2021)
- Using `webpack-dev-middleware` instead of `webpack-dev-server` directly
	[#2034](https://github.com/weaveworks/scope/pull/2034)
- Create `latest_release` Docker image tag during release process
	[#2216](https://github.com/weaveworks/scope/issues/2216)

Weave Cloud related changes:
- Deploy to quay when merging to master
	[#2134](https://github.com/weaveworks/scope/pull/2134)
- Removed leading slash from getAllNodes() api request
	[#2124](https://github.com/weaveworks/scope/pull/2124)
- Correctly instrument websocket handshakes
	[#2074](https://github.com/weaveworks/scope/pull/2074)


## Release 1.1.0

Highlights:
- New ECS view which allows visualizing your tasks and services in Amazon's EC2 Container Service.
- Custom label-based container filters can be defined through `--app.container-label-filter`

New features and enhancements:
- Add ECS views
	[#2026](https://github.com/weaveworks/scope/pull/2026)
- Add custom label-based filters in container view
	[#1895](https://github.com/weaveworks/scope/pull/1895)
- Improve plugin errors tooltip
	[#2022](https://github.com/weaveworks/scope/pull/2022)
- Add anti-dance heuristics (and feature flags)
	[#1993](https://github.com/weaveworks/scope/pull/1993)
- Table-mode: sort ips numerically
	[#2007](https://github.com/weaveworks/scope/pull/2007)
- increase black/white text contrast in contrast mode
	[#2006](https://github.com/weaveworks/scope/pull/2006)
- Improve view-node-in-topo button usability
	[#1926](https://github.com/weaveworks/scope/pull/1926)
- Hide Weave topology if empty
	[#2035](https://github.com/weaveworks/scope/pull/2035)

Performance improvements:
- Add graph complexity check on page load
	[#1994](https://github.com/weaveworks/scope/pull/1994)

Bug fixes:
- plug goroutine leak in control
	[#2003](https://github.com/weaveworks/scope/pull/2003)
- Fix details panel not closing on canvas click
	[#1998](https://github.com/weaveworks/scope/pull/1998)
- Empty publicpath needed for relative paths of scope
	[#2043](https://github.com/weaveworks/scope/pull/2043)

Documentation:
- Use intuitive standalone service name in compose
	[#2019](https://github.com/weaveworks/scope/pull/2019)
- Fix kubectl port-forward command to access the scope app locally
	[#2010](https://github.com/weaveworks/scope/pull/2010)
- Update website plugins documentation
	[#2008](https://github.com/weaveworks/scope/pull/2008)

Internal improvements and cleanup:
- Combined external and prod webpack config files
	[#2014](https://github.com/weaveworks/scope/pull/2014)
- Update package.json
	[#2017](https://github.com/weaveworks/scope/pull/2017)
- Move plugins to the new organization
	[#1906](https://github.com/weaveworks/scope/pull/1906)
- Change webpack local config to use source maps
	[#2011](https://github.com/weaveworks/scope/pull/2011)
- middleware/errorhandler: Implement Hijacker so it works with ws proxy
	[#1971](https://github.com/weaveworks/scope/pull/1971)
- Fix time-dependant test (stop testing docker client library)
	[#2005](https://github.com/weaveworks/scope/pull/2005)
- Give time to the overlay test backoff collectors to finish
	[#1995](https://github.com/weaveworks/scope/pull/1995)
- Update D3 to version 4.4.0
	[#2028](https://github.com/weaveworks/scope/pull/2028)

Weave Cloud related changes:
- Add OpenTracing support to TimeRequestHistogram
	[#2023](https://github.com/weaveworks/scope/pull/2023)

## Release 1.0.0

Highlights:
- New Weave Net View which allows visualizing and troubleshooting your Weave Network.
- New nodes for well-known services. The Internet node is now broken down in individual nodes for known cloud services.
- Improved terminals, with proper resizing, scroll locking and better visuals.
- Refined UI with particularly improved connection information.
- Lots of squashed bugs.

New features and enhancements:
- New Weave Net view
	[#1182](https://github.com/weaveworks/scope/pull/1182)
	[#1973](https://github.com/weaveworks/scope/pull/1973)
	[#1981](https://github.com/weaveworks/scope/pull/1981)
- Show well-known services
	[#1863](https://github.com/weaveworks/scope/pull/1863)
	[#1881](https://github.com/weaveworks/scope/pull/1881)
	[#1887](https://github.com/weaveworks/scope/pull/1887)
	[#1897](https://github.com/weaveworks/scope/pull/1897)
- Terminal improvements
  - Resize TTYs
		[#1966](https://github.com/weaveworks/scope/pull/1966)
		[#1979](https://github.com/weaveworks/scope/pull/1979)
		[#1976](https://github.com/weaveworks/scope/pull/1976)
  - Enable scroll locking on the terminal
		[#1932](https://github.com/weaveworks/scope/pull/1932)
  - Adds tooltip to terminal-popout button
		[#1790](https://github.com/weaveworks/scope/pull/1790)
  - Clarify terminal is child window of details panel.
		[#1903](https://github.com/weaveworks/scope/pull/1903)
  - Use login shells in terminals
		[#1821](https://github.com/weaveworks/scope/pull/1821)
- Miscellaneous UI improvements
  - show more details of a node's internet connections
		[#1875](https://github.com/weaveworks/scope/pull/1875)
  - Close help dialog when the canvas is clicked
		[#1960](https://github.com/weaveworks/scope/pull/1960)
  - Improve metadata table 'date' format
		[#1927](https://github.com/weaveworks/scope/pull/1927)
  - Add a new search section to the help popover
		[#1919](https://github.com/weaveworks/scope/pull/1919)
  - Add label_minor to tooltips in connections table
		[#1912](https://github.com/weaveworks/scope/pull/1912)
  - Add localstorage support for saving view state
		[#1853](https://github.com/weaveworks/scope/pull/1853)
  - Makes services the initial topology if available
		[#1823](https://github.com/weaveworks/scope/pull/1823)
  - Add image information table to container details panel
		[#1942](https://github.com/weaveworks/scope/pull/1942)
- Allow user to specify URLs on the command line, and use that to allow per-target tokens.
	[#1901](https://github.com/weaveworks/scope/pull/1901)
- Apply filters from current view to details panel
	[#1904](https://github.com/weaveworks/scope/pull/1904)
- Increase timestamp precision
	[#1933](https://github.com/weaveworks/scope/pull/1933)
- Add prometheus metrics endpoint to probes.
	[#1915](https://github.com/weaveworks/scope/pull/1915)
- Allow users to specify conntrack buffer size.
	[#1896](https://github.com/weaveworks/scope/pull/1896)
- Plugins: Add support for table based controls
	[#1818](https://github.com/weaveworks/scope/pull/1818)

Performance improvements:
- make smartMerger.Merge merge reports in parallel
	[#1827](https://github.com/weaveworks/scope/pull/1827)

Bug fixes:
- Goroutine leak in scope app
	[#1916](https://github.com/weaveworks/scope/issues/1916)
	[#1920](https://github.com/weaveworks/scope/pull/1920)
- CPU Usage is not accurate on hosts
	[#1664](https://github.com/weaveworks/scope/issues/664)
- Certain query strings would contain a && instead of &
	[#1953](https://github.com/weaveworks/scope/pull/1953)
- Metrics on canvas get stuck
	[#1829](https://github.com/weaveworks/scope/issues/1829)
- conntrack not used even though it's working
	[#1826](https://github.com/weaveworks/scope/issues/1826)
- pod counts and details panel lists do not respect namespace
	[#1824](https://github.com/weaveworks/scope/issues/1824)
- Discard short-lived connections to/from Pods in the host net
	[#1944](https://github.com/weaveworks/scope/pull/1944)
- probe: Stats gathering can be started twice
	[#1799](https://github.com/weaveworks/scope/issues/1799)
- Visual bug where empty span shows up
	[#1945](github.com/weaveworks/scope/issues/1945)
- inbound internet connection counts are too fine-grained
	[#1867](https://github.com/weaveworks/scope/issues/1867)
- IP address truncated in Internet node details panel connection list
	[#1862](https://github.com/weaveworks/scope/issues/1862)
- Incorrect number of connections shown on internet nodes
	[#1495](https://github.com/weaveworks/scope/issues/1495)
- details panel connection counts are too high
	[#1842](https://github.com/weaveworks/scope/issues/1842)
- inbound internet connections reverse-resolved incorrectly
	[#1847](https://github.com/weaveworks/scope/issues/1847)
- Scope hangs after browser reload if current topology goes away
	[#1880](https://github.com/weaveworks/scope/issues/1880)
- node names in connection list truncated unnecessarily
	[#1882](https://github.com/weaveworks/scope/issues/1882)
- numeric values in details panel tables should be right aligned
	[#1794](https://github.com/weaveworks/scope/issues/1794)
- Plugin status line is broken
	[#1825](https://github.com/weaveworks/scope/issues/1825)
- Table-mode: non-metric columns are sorted alphabetically reverse
	[#1802](https://github.com/weaveworks/scope/issues/1802)
- Fix argument escaping in Scope
	[#1950](https://github.com/weaveworks/scope/pull/1950)
- Image details panel shows truncated image name instead of ID
	[#1835](https://github.com/weaveworks/scope/issues/1835)
- Truncated tooltips
	[#1139](https://github.com/weaveworks/scope/issues/1139)
- Incorrect height of terminal window in Safari
	[#1986](https://github.com/weaveworks/scope/issues/1986)

Documentation:
- Simplify k8s instructions
	[#1886](https://github.com/weaveworks/scope/pull/1886)
- Improve installation documentation
	[#1838](https://github.com/weaveworks/scope/pull/1838)
- Update Scope version in documentation
	[#1859](https://github.com/weaveworks/scope/pull/1859)

Internal improvements and cleanup:
- Gracefully shutdown app, letting active http requests finish with timeout
	[#1839](https://github.com/weaveworks/scope/pull/1839)
- middleware/errorhandler: Fix a bug which meant it never works
	[#1958](https://github.com/weaveworks/scope/pull/1958)
- middleware: Add an ErrorHandler middleware used to serve an alternate handler on a certain error code
	[#1954](https://github.com/weaveworks/scope/pull/1954)
- Update client deps to use Node v6.9.0
	[#1959](https://github.com/weaveworks/scope/pull/1959)
- Change term.js library to xterm.js
	[#1948](https://github.com/weaveworks/scope/pull/1948)
- Fix linter errors on unkeyed fields
	[#1922](https://github.com/weaveworks/scope/pull/1922)
- Fix linter error for string in context.WithValue
	[#1921](https://github.com/weaveworks/scope/pull/1921)
- Update tools subtree
	[#1937](https://github.com/weaveworks/scope/pull/1937)
- Fix circle.yml to actually deploy external ui changes
	[#1910](https://github.com/weaveworks/scope/pull/1910)
- Extend logging middleware to optionally only log failed HTTP requests
	[#1909](https://github.com/weaveworks/scope/pull/1909)
- Add option to scope to have static content served from S3 instead
	[#1908](https://github.com/weaveworks/scope/pull/1908)
- Upgrade to go1.7
	[#1797](https://github.com/weaveworks/scope/pull/1797)
- circleci: push traffic control plugin image to docker hub
	[#1858](https://github.com/weaveworks/scope/pull/1858)
- refactor: extract pluralization
	[#1855](https://github.com/weaveworks/scope/pull/1855)
- use go-dockerclient's Client.Stats
	[#1833](https://github.com/weaveworks/scope/pull/1833)
- Print logs to debug shutdown integration test
	[#1888](https://github.com/weaveworks/scope/pull/1888)
- Allow a nil RouteMatcher in instrumentation
	[#1852](https://github.com/weaveworks/scope/pull/1852)

Weave Cloud related changes:
- Don't reencode reports in the collector
	[#1819](https://github.com/weaveworks/scope/pull/1819)

## Release 0.17.1

This is a minor patch release.

New features and enhancements:
- Extend kubernetes client flags to match kubectl
	[#1813](https://github.com/weaveworks/scope/pull/1813)

Bug fixes:
- Fix node label overlap
	[#1812](https://github.com/weaveworks/scope/pull/1812)
- Fix `scope stop` on Docker for Mac
	[#1811](https://github.com/weaveworks/scope/pull/1811)


## Release 0.17.0

Highlights:
- New Table Mode as an alternative to Scope's classic graph view. It provides
  higher information density, proving particularly useful when there are many
  nodes in the graph view.
- Considerable performance enhancements: the CPU efficiency of the Scope App has
  increased in more than 50% and the Scope probes over 25%.


New features and enhancements:
- Table mode
	[#1673](https://github.com/weaveworks/scope/pull/1673)
	[#1747](https://github.com/weaveworks/scope/pull/1747)
	[#1753](https://github.com/weaveworks/scope/pull/1753)
	[#1774](https://github.com/weaveworks/scope/pull/1774)
	[#1775](https://github.com/weaveworks/scope/pull/1775)
	[#1784](https://github.com/weaveworks/scope/pull/1784)
- Loading indicator
	[#1485](https://github.com/weaveworks/scope/pull/1485)
- Don't show weavescope logo when running in a frame
	[#1734](https://github.com/weaveworks/scope/pull/1734)
- Reduce horizontal gap between nodes in topology views
	[#1693](https://github.com/weaveworks/scope/pull/1693)
- Elide service-token when logging commandline arguments
	[#1782](https://github.com/weaveworks/scope/pull/1782)
- Don't complain when stopping Scope if it wasn't running
	[#1783](https://github.com/weaveworks/scope/pull/1783)
- Silence abnormal websocket close
	[#1768](https://github.com/weaveworks/scope/pull/1768)
- Eliminate stats log noise from stopped containers
	[#1687](https://github.com/weaveworks/scope/pull/1687)
	[#1798](https://github.com/weaveworks/scope/pull/1798)
- Hide uncontained/unmanaged by default
	[#1694](https://github.com/weaveworks/scope/pull/1694)

Performance improvements:
- Remove and optimize more Copy()s
	[#1739](https://github.com/weaveworks/scope/pull/1739)
- Use slices instead of linked lists for Metric
	[#1732](https://github.com/weaveworks/scope/pull/1732)
- Don't Copy() self on Merge()
	[#1728](https://github.com/weaveworks/scope/pull/1728)
- Improve performance of immutable maps
	[#1720](https://github.com/weaveworks/scope/pull/1720)
- Custom encoder for latest maps
	[#1709](https://github.com/weaveworks/scope/pull/1709)

Bug fixes:
- Connections inside a container shown as going between containers
	[#1733](https://github.com/weaveworks/scope/issues/1733)
- Probes leak two goroutines when closing attach/exec window
	[#1767](https://github.com/weaveworks/scope/issues/1767)
- Scale node labels with the node's size.
	[#1773](https://github.com/weaveworks/scope/pull/1773)
- Kubernetes infra containers seem to resurface in latest on 1.3
	[#1750](https://github.com/weaveworks/scope/issues/1750)
- Search icon is above text
	[#1715](https://github.com/weaveworks/scope/issues/1715)
- Highlighting is unpredictable
	[#1756](https://github.com/weaveworks/scope/pull/1520)
- Details panel truncates port to four digits
	[#1711](https://github.com/weaveworks/scope/issues/1711)
- Stopped containers not shown with their names
	[#1691](https://github.com/weaveworks/scope/issues/1691)
- Terminals don't support quote characters from intl keyboard layouts
	[#1403](https://github.com/weaveworks/scope/issues/1403)

Internal improvements and cleanup:
- Launcher script: Fix inconsistent whitespace
	[#1776](https://github.com/weaveworks/scope/pull/1776)
- Lint fixes
	[#1751](https://github.com/weaveworks/scope/pull/1751)
- Add browser console logging for websocket to render times
	[#1742](https://github.com/weaveworks/scope/pull/1742)
- circle.yml: deploy master with non-upstream hub accounts
	[#1655](https://github.com/weaveworks/scope/pull/1655)
	[#1710](https://github.com/weaveworks/scope/pull/1710)
- Delete unused instrumentation code
	[#1722](https://github.com/weaveworks/scope/pull/1722)
- Update version of build tools
	[#1685](https://github.com/weaveworks/scope/pull/1685)
- Add flag for block profiling
	[#1681](https://github.com/weaveworks/scope/pull/1681)

Weave Cloud related changes:
- Also serve UI under /ui
	[#1752](https://github.com/weaveworks/scope/pull/1752)
- Name our routes, so /metrics gives more sensible aggregations
	[#1723](https://github.com/weaveworks/scope/pull/1723)
- Add options for storing memcached reports with different compression levels
	[#1684](https://github.com/weaveworks/scope/pull/1684)

## Release 0.16.2

Bug fixes:
- Scope fails to launch on fresh Docker for Mac installs
	[#1755](https://github.com/weaveworks/scope/issues/1755)

## Release 0.16.1

This is a bugfix release. In addition, the security of the Scope probe can be hardened by disabling
controls with the new `--probe.no-controls` flag, which prevents users from
opening terminals, starting/stopping containers, viewing logs, etc.

New features and enhancements:
- Allow disabling controls in probes
	[#1627](https://github.com/weaveworks/scope/pull/1627)
- Make it easier to disable weave integrations
	[#1610](https://github.com/weaveworks/scope/pull/1610)
- Print DNS errors
	[#1607](https://github.com/weaveworks/scope/pull/1607)
- Add dry run flag to scope, so when launched we can check the args are valid.
	[#1609](https://github.com/weaveworks/scope/pull/1609)

Performance improvements:
- Use a slice instead of a persistent list for temporary accumulation of lists
	[#1660](https://github.com/weaveworks/scope/pull/1660)

Bug fixes:
- Should check if probe is already running when launch in standalone mode on Docker for Mac
	[#1679](https://github.com/weaveworks/scope/issues/1679)
- Fixes network bars position when a node is selected.
	[#1667](https://github.com/weaveworks/scope/pull/1667)
- Scope fails to launch on latest Docker for Mac (beta18)
	[#1650](https://github.com/weaveworks/scope/pull/1650)
	[#1669](https://github.com/weaveworks/scope/pull/1669)
- Fixes terminal wrapping by syncing docker/term.js terminal widths.
	[#1648](https://github.com/weaveworks/scope/pull/1648)
- Wrongly attributed local side in outbound internet connections
	[#1598](https://github.com/weaveworks/scope/issues/1598)
- Cannot port-forward app from kubernetes with command in documentation
	[#1526](https://github.com/weaveworks/scope/issues/1526)
- Force some known column widths to prevent truncation of others
	[#1641](https://github.com/weaveworks/scope/pull/1641)

Documentation:
- Replace wget in instructions with curl, as it's more widely avail. on macs
	[#1670](https://github.com/weaveworks/scope/pull/1670)
- Don't prepend `scope launch` with sudo
	[#1606](https://github.com/weaveworks/scope/pull/1606)
- Clarify instructions for using Scope with Weave Cloud
	[#1611](https://github.com/weaveworks/scope/pull/1611)
- Re-added signup page
	[#1604](https://github.com/weaveworks/scope/pull/1604)
- weave cloud screen captures
	[#1603](https://github.com/weaveworks/scope/pull/1603)

Internal improvements and cleanup:
- Lint shellscripts from tools
	[#1658](https://github.com/weaveworks/scope/pull/1658)
- Promote fixprobe and delete rest of experimental
	[#1646](https://github.com/weaveworks/scope/pull/1646)
- refactor some timing helpers into a common lib
	[#1642](https://github.com/weaveworks/scope/pull/1642)
- Helper for reading & writing from binary
	[#1600](https://github.com/weaveworks/scope/pull/1600)
- Updates to vendoring document
	[#1595](https://github.com/weaveworks/scope/pull/1595)

Weave Cloud related changes:
- Store a histogram of report sizes
	[#1668](https://github.com/weaveworks/scope/pull/1668)
- Wire up continuous delivery
	[#1654](https://github.com/weaveworks/scope/pull/1654)
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

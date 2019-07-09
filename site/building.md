---
title: Developing and Debugging
menu_order: 90
search_type: Documentation
---

The following topics are discussed:

 * [Developing](#developing)
 * [Debugging](#debugging)
 * [Profiling](#profiling)


## <a name="developing"></a>Developing

Building Scope from source depends on the latest version of [docker](https://www.docker.com/), so please install that before proceeding.

The build process is automated using `make`, which builds the UI build container, builds the UI in said container, builds the backend build container, builds the app and probe in said container, and finally pushes the lot into a Docker image called **weaveworks/scope**.

    make

Then, run the local build via:

    ./scope launch

If needed, install the tools used for managing dependencies, managing releases, and doing coverage analysis via:

    make deps

Scope unit tests for `probe` and `app` components can be run via:

    make tests

Similarly the frontent client tests can be run via:

    make client-test


>**Note:** The tools from `make deps` depend on a local install of
[Go](https://golang.org).

## <a name="debugging"></a>Debugging

Scope has a collection of built-in debugging tools to aid Scope developers.


To get debug information in the logs launch Scope with `--debug`:

    scope launch --debug
    docker logs weavescope


To have the Scope App or Scope Probe dump their goroutine stacks, run:

    kill -QUIT $(pgrep -f scope-(app|probe))
    docker logs weavescope


The Scope Probe is instrumented with various counters and timers. To have it dump those values, run:


    kill -USR1 $(pgrep -f scope-probe)
    docker logs weavescope

If you run with `--probe.http.listen` enabled, these are exposed as Prometheus metrics instead, via http at `/metrics`.

## <a name="profiling"></a>Profiling

Both the Scope App and the Scope Probe offer [HTTP endpoints with profiling information](https://golang.org/pkg/net/http/pprof/).

These cover things such as CPU usage and memory consumption:

  * The Scope App enables its HTTP profiling endpoints by default, which are accessible on the same port the Scope UI is served (4040).
  * The Scope Probe doesn't enable its profiling endpoints by default. To enable them, you must launch Scope with `--probe.http.listen addr:port`. For instance, launching Scope with `scope launch --probe.http.listen :4041`, will allow you access the Scope Probe's profiling endpoints on port 4041.

Then, you can collect profiles in the usual way. For instance:

To collect the memory profile of the Scope App:

    go tool pprof http://localhost:4040/debug/pprof/heap

To collect the CPU profile of the Scope Probe:

    go tool pprof http://localhost:4041/debug/pprof/profile

To collect a blocking profile of the Scope App, make sure you have launched
Scope with `--app.block.profile.rate=N` (where `N` is the number of
nanoseconds between samples) and then:

    go tool pprof http://localhost:4040/debug/pprof/block

If you don't have `go` installed, you can use a Docker container instead:

To collect the memory profile of the Scope App:

    docker run --net=host -v $PWD:/root/pprof golang go tool pprof http://localhost:4040/debug/pprof/heap

To collect the CPU profile of the Scope Probe:

    docker run --net=host -v $PWD:/root/pprof golang go tool pprof http://localhost:4041/debug/pprof/profile

You will find the output profiles in your working directory.

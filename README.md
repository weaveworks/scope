# Scope

![Weave Scope Screenshot](http://weave.works/scope/assets/img/feature-1.png)

## Overview

Weave Scope automatically generates a map of your containers, enabling you to
intuitively understand, monitor, and control your applications.

Please note that the code, and especially the building and running story, is in
a **pre-alpha** state. Please take a look, but don't be surprised if you hit
bugs or missing pieces.

## Building

### In-place

To build the binaries in-place,

```
make build
```

Note that this doesn't build or include the latest version of the user
interface. The UI is decoupled, living in `client` and following a node/gulp
workflow. To build that and include it in the application binary,

```
make client
make static
make build
```

Or, as a shortcut,

```
make dist
```

### Docker container

To build a Docker container,

```
make docker
```

## Running

### Manually

1. Launch a probe process on each physical host you intend to monitor, via `sudo probe`
2. Launch an app process, and configure it to talk to probes, via `app -probes="probe-host-1:4030,probe-host-2:4030"`.
3. Load the user interface, via **http://app-host:4040**

### As a Docker container

TODO


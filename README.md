# Scope

[![Circle CI](https://circleci.com/gh/weaveworks/scope/tree/master.svg?style=shield)](https://circleci.com/gh/weaveworks/scope/tree/master) [![Coverage Status](https://coveralls.io/repos/weaveworks/scope/badge.svg)](https://coveralls.io/r/weaveworks/scope) [![Sourcegraph](https://sourcegraph.com/api/repos/github.com/weaveworks/scope/.badges/status.svg)](https://sourcegraph.com/github.com/weaveworks/scope)

![Weave Scope Screenshot](http://weave.works/img/header-image-left.png)


## Overview

Weave Scope automatically generates a map of your containers, enabling you to
intuitively understand, monitor, and control your applications.

Please note that the code, and especially the building and running story, is
in a **prerelease** state. Please take a look, but don't be surprised if you
hit bugs or missing pieces.


## Getting started

```
sudo wget -O /usr/local/bin/scope \
  https://github.com/weaveworks/scope/releases/download/latest_release/scope
sudo chmod a+x /usr/local/bin/scope
sudo scope launch
```

This script will download and run a recent Scope image from the Docker Hub.
Now, open your web browser to **http://localhost:4040**. (If you're using
boot2docker, replace localhost with the output of `boot2docker ip`.)


## Build

```
make deps
make
```

This will produce a Docker image called **weaveworks/scope**.

Note that the repository contains a copy of the compiled UI. To build a fresh
UI from the source in the client subdirectory, and re-build the Docker
container,

```
make scope_ui_build.tar
make static
make
```


## Run

```
./scope launch
```

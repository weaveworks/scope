# Scope

[![Circle CI](https://circleci.com/gh/weaveworks/scope/tree/master.svg?style=shield)](https://circleci.com/gh/weaveworks/scope/tree/master) [![Coverage Status](https://coveralls.io/repos/weaveworks/scope/badge.svg)](https://coveralls.io/r/weaveworks/scope) [![Sourcegraph](https://sourcegraph.com/api/repos/github.com/weaveworks/scope/.badges/status.svg)](https://sourcegraph.com/github.com/weaveworks/scope)

![Weave Scope Screenshot](http://weave.works/img/header-image-left.png)


## Overview

Weave Scope automatically generates a map of your containers, enabling you to
intuitively understand, monitor, and control your applications.


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

The build is in five stages. `make deps` installs some tools we use later in
the build. `make frontend` builds a UI build image with all NPM dependencies.
`make static` compiles the UI into `static.go` which is part of the repository
for convenience. `make backend` builds the backend Go app which then includes
the static files. The final `make` pushes the app into a Docker image called
**weaveworks/scope**.

```
make deps
make frontend
make static
make backend
make
```

## Run

```
./scope launch
```

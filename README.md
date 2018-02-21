# OpenEBS Scope - Troubleshooting & Monitoring for Docker & Kubernetes

[![Circle CI](https://circleci.com/gh/weaveworks/scope/tree/master.svg?style=shield)](https://circleci.com/gh/weaveworks/scope/tree/master)
[![Coverage Status](https://coveralls.io/repos/weaveworks/scope/badge.svg)](https://coveralls.io/r/weaveworks/scope)
[![Go Report Card](https://goreportcard.com/badge/github.com/weaveworks/scope)](https://goreportcard.com/report/github.com/weaveworks/scope)
[![Docker Pulls](https://img.shields.io/docker/pulls/weaveworks/scope.svg?maxAge=604800)](https://hub.docker.com/r/weaveworks/scope/)
[![GoDoc](https://godoc.org/github.com/weaveworks/scope?status.svg)](https://godoc.org/github.com/weaveworks/scope)

OpenEBS Scope automatically generates a map of your application, enabling you to
intuitively understand, monitor, and control your containerized, microservices based application.

### Understand your Docker containers in real-time

<img src="imgs/topology.png" width="200" alt="Map you architecture" align="right">

Choose an overview of your container infrastructure, or focus on a specific microservice. Easily identify and correct issues to ensure the stability and performance of your containerized applications.

### Contextual details and deep linking

<img src="imgs/selected.png" width="200" alt="Focus on a single container" align="right">

View contextual metrics, tags and metadata for your containers.  Effortlessly navigate between processes inside your container to hosts your containers run on, arranged in expandable, sortable tables.  Easily find the container using the most CPU or memory for a given host or service.

### Interact with and manage containers

<img src="imgs/terminals.png" width="200" alt="Launch a command line." align="right">

Interact with your containers directly: pause, restart and stop containers. Launch a command line. All without leaving the scope browser window.

### Extend and customize via plugins

Add custom details or interactions for your hosts, containers and/or processes by creating Scope plugins; or just choose from some that others have already written at the Github [Weaveworks Scope Plugins](https://github.com/weaveworks-plugins/) organization.

## <a name="getting-started"></a>Getting started

```
sudo curl -L git.io/scope -o /usr/local/bin/scope
sudo chmod a+x /usr/local/bin/scope
scope launch
```

This script downloads and runs a recent Scope image from Docker Hub.
Now, open your web browser to **http://localhost:4040**. (If you're using
boot2docker, replace localhost with the output of `boot2docker ip`.)

**YAML** for installing Scope on [Kubernetes](https://github.com/openebs/scope/blob/master/scope_k8s.yaml), [DCOS](https://www.weave.works/docs/scope/latest/installing/#dcos) or [ECS](https://www.weave.works/docs/scope/latest/installing/#ecs), see [the docs](https://www.weave.works/docs/scope/latest/introducing/).


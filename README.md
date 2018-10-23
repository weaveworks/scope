# Weave Scope - Troubleshooting & Monitoring for Docker & Kubernetes

[![Circle CI](https://circleci.com/gh/weaveworks/scope/tree/master.svg?style=shield)](https://circleci.com/gh/weaveworks/scope/tree/master)
[![Coverage Status](https://coveralls.io/repos/weaveworks/scope/badge.svg)](https://coveralls.io/r/weaveworks/scope)
[![Go Report Card](https://goreportcard.com/badge/github.com/weaveworks/scope)](https://goreportcard.com/report/github.com/weaveworks/scope)
[![Slack Status](https://slack.weave.works/badge.svg)](https://slack.weave.works)
[![Docker Pulls](https://img.shields.io/docker/pulls/weaveworks/scope.svg?maxAge=604800)](https://hub.docker.com/r/weaveworks/scope/)
[![GoDoc](https://godoc.org/github.com/weaveworks/scope?status.svg)](https://godoc.org/github.com/weaveworks/scope)

Weave Scope automatically generates a map of your application, enabling you to
intuitively understand, monitor, and control your containerized, microservices-based application.

### Understand your Docker containers in real time

<img src="imgs/topology.png" width="200" alt="Map you architecture" align="right">

Choose an overview of your container infrastructure, or focus on a specific microservice. Easily identify and correct issues to ensure the stability and performance of your containerized applications.

### Contextual details and deep linking

<img src="imgs/selected.png" width="200" alt="Focus on a single container" align="right">

View contextual metrics, tags, and metadata for your containers.  Effortlessly navigate between processes inside your container to hosts your containers run on, arranged in expandable, sortable tables.  Easily find the container using the most CPU or memory for a given host or service.

### Interact with and manage containers

<img src="imgs/terminals.png" width="200" alt="Launch a command line." align="right">

Interact with your containers directly: pause, restart, and stop containers. Launch a command line. All without leaving the scope browser window.

### Extend and customize via plugins

Add custom details or interactions for your hosts, containers, and/or processes by creating Scope plugins. Or, just choose from some that others have already written at the GitHub [Weaveworks Scope Plugins](https://github.com/weaveworks-plugins/) organization.

## <a name="getting-started"></a>Getting Started

```console
sudo curl -L git.io/scope -o /usr/local/bin/scope
sudo chmod a+x /usr/local/bin/scope
scope launch
```

This script downloads and runs a recent Scope image from Docker Hub.
Now, open your web browser to **http://localhost:4040**. (If you're using
boot2docker, replace localhost with the output of `boot2docker ip`.)

For instructions on installing Scope on [Kubernetes](https://www.weave.works/docs/scope/latest/installing/#k8s), [DCOS](https://www.weave.works/docs/scope/latest/installing/#dcos), or [ECS](https://www.weave.works/docs/scope/latest/installing/#ecs), see [the docs](https://www.weave.works/docs/scope/latest/introducing/).

## <a name="help"></a>Getting Help

If you have any questions, feedback, or problems with Scope:

- Read [the Weave Scope docs](https://www.weave.works/docs/scope/latest/introducing/).
- Invite yourself to the <a href="https://weaveworks.github.io/community-slack/" target="_blank">Weave community</a> Slack.
- Ask a question on the [#scope](https://weave-community.slack.com/messages/scope/) Slack channel.
- Join the [Weave User Group](https://www.meetup.com/pro/Weave/) and get invited to online talks, hands-on training and meetups in your area.
- Send an email to [Scope community group](https://groups.google.com/forum/#!forum/scope-community).
- Join (and read up on) the regular [Scope community meetings](https://docs.google.com/document/d/103_60TuEkfkhz_h2krrPJH8QOx-vRnPpbcCZqrddE1s/edit).
- [File an issue](https://github.com/weaveworks/scope/issues/new).

Your feedback is always welcome!

### License
Scope is licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for the full license text.  
Find more details about the licenses of vendored code in [VENDORED_CODE.md](VENDORED_CODE.md).
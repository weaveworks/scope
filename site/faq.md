---
title: Frequently asked questions
menu_order: 95
search_type: Documentation
---

# Frequently asked questions

## Running Scope in a Kubernetes setting

This is answered in [our
docs](https://www.weave.works/docs/scope/latest/installing/#k8s). you can
choose between two options, either clone the source of Weave Scope, or use
the YAML manifests from `cloud.weave.works`.

## Disabling Scope Write Access

Can be done by using the `probe.no-controls` option and set it to false for the scope agents. This can be done in the scope deployment manifest under the `weave-scope-agent`'s argument section with `—probe.no-control=true`.

## RBAC and Weave Scope OSS

OSS Scope has no user concept, this is only available in Weave Cloud. To limit the access to the UI,

- setup a reverse proxy with auth and block access to non admin users,
- capture the calls with something like Chrome network console to get the endpoints to know which requests to authenticate in the proxy server.
- you can use Basic HTTP Auth since Scope 1.10.0 - just use these command line
  arguments:

  ```cli
  -app.basicAuth
        Enable basic authentication for app
  -app.basicAuth.password string
        Password for basic authentication (default "admin")
  -app.basicAuth.username string
        Username for basic authentication (default "admin")
  ```

## ARM Support

- It required patches, @adivyoseph (on [#scope](https://weave-community.slack.com/messages/scope/)) had done some work on this.
- [#2110](https://github.com/weaveworks/scope/issues/2110) says that scope's CI builds ARM32 (but not ARM64) for test-builds at least.
- @errordeveloper says: It should be easy to add arm64 in CI, You can try and enable builds in ci on a branch.. In theory, you just need to build for `GOARCH=arm64`.

## Data Storage

OSS Scope reports aren't persistent and the probe keeps the last 15 seconds of metrics in memory.

## API Endpoints

Scope exposes the following endpoints that can be used by external monitoring services.

- `/api` - Scope status and configuration
- `/api/probes` - basic status of Scope probes
- `/api/report` - returns a full JSON report
- `/api/topology` - information on all topologies
- `/api/topology/[TOPOLOGY]` -  information on all nodes belonging to `TOPOLOGY` topology
- `/api/topology/[TOPOLOGY]/[NODE_ID]` - information on specific node `NODE_ID` in topology `TOPOLOGY` (currently `NODE_ID` must be an internal Scope node ID obtained from the URL field `selectedNodeId` when selecting that node in the UI - see [#3122](https://github.com/weaveworks/scope/issues/3122) for a proposal of a better solution)

## Using a different port

You can use `scope launch --app.http.address=127.0.0.1:9000` to run the
http server on another port (in this case 9000).

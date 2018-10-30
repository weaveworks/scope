---
title: Frequently asked questions
menu_order: 95
search_type: Documentation
---

# Frequently asked questions

## Running Scope in a Kubernetes setting

A simple way to get Scope running in a Kubernetes setting is to

1. Clone the Scope repo:

   ```sh
   git clone https://github.com/weaveworks/scope
   cd scope
   ```

1. Spin up a cluster wherever it suits you.
   [Minikube](https://github.com/kubernetes/minikube) is a simple option.
1. Run

   ```sh
   kubectl apply -f example/k8s
   ```

   to deploy Scope to your cluster.
1. Port-forward to access `weave-scope-app`:

   ```sh
   kubectl port-forward svc/weave-scope-app -n weave 4040:80
   ```

1. Point your browser to <http://127.0.0.1:4040.>

## Disabling Scope Write Access

Can be done by using the `probe.no-controls` option and set it to false for the scope agents. This can be done in the scope deployment manifest under the `weave-scope-agent`'s argument section with `—probe.no-control=true`.

## RBAC and Weave Scope OSS

OSS Scope has no user concept, this is only available in Weave Cloud. To limit the access to the UI,

- setup a reverse proxy with auth and block access to non admin users,
- capture the calls with something like Chrome network console to get the endpoints to know which requests to authenticate in the proxy server.
- you can use Basic HTTP Auth since Scope 1.10.0 - just use the `BASIC_AUTH_USERNAME` and `BASIC_AUTH_PASSWORD` environment variables.
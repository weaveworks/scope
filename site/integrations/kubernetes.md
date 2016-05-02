---
title: Visualizing Kubernetes Clusters with Scope
menu_order: 10
---


Before visualizing Kubernetes clusters with Scope, it is recommended that you run Scope natively in your Kubernetes cluster using [these resource definitions](https://github.com/TheNewNormal/kube-charts/tree/master/weavescope/manifests).

**1. Ensure that the cluster allows privileged pods.**

This is required by the Scope probes. By default, privileged pods are allowed from Kubernetes 1.1 and up. If you are running an earlier version or a non-default configuration, ensure that your API Server and all of your Kubelets are launched with the flag `--allow_privileged`.

**2. Your cluster must support [DaemonSets](https://github.com/kubernetes/kubernetes/blob/master/docs/design/daemon.md).**

DaemonSets are necessary to ensure that each Kubernetes node can run a Scope Probe:

To enable them in an existing cluster, add the `--runtime-config=extensions/v1beta1/daemonsets=true` argument to the [apiserver](https://github.com/kubernetes/kubernetes/blob/master/docs/admin/kube-apiserver.md)'s configuration. This is normally found in the `/etc/kubernetes/manifest/kube-apiserver.manifest`file after a restart of [the apiserver and controller manager](https://github.com/kubernetes/kubernetes/issues/18656) has occurred.

If you are creating a new cluster, set `KUBE_ENABLE_DAEMONSETS=true` in your cluster configuration.

**3. Download the resource definitions:**

    for I in app-rc app-svc probe-ds; do
     curl -s -L https://raw.githubusercontent.com/TheNewNormal/kube-charts/master/weavescope/manifests/scope-$I.yaml -o       scope-$I.yaml
    done


**4. Tweak the Scope probe configuration in `scope-probe-ds.yaml`, namely:**

If you have an account at (scope.weave.works)[https://scope.weave.works] and want to use Scope in Cloud Service Mode, uncomment the `--probe.token=foo` argument, substitute `foo` by the token found in your account page, and comment out the `$(WEAVE_SCOPE_APP_SERVICE_HOST):$(WEAVE_SCOPE_APP_SERVICE_PORT)` argument.

**5. Finally, install Scope in your cluster (order is important):**

      kubectl create -f scope-app-rc.yaml  # Only if you want to run Scope in Standalone Mode
      kubectl create -f scope-app-svc.yaml # Only if you want to run Scope in Standalone Mode
      kubectl create -f scope-probe-ds.yaml

      
**See Also**

 * [Understanding Weave Scope](/site/how-it-works.md)
 * [Managing Views](/site/manage-views.md)
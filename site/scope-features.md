---
title: Feature Overview
menu_order: 15
---

Browse the current feature set for Scope with links to relevant indepth topics: 

 * [Topology Mapping](#topology-mapping)
 * [Flexible Filtering](#flexible-filtering)
 * [Real-time App and Container Metrics](#real-time-app-and-container-metrics)
 * [Interact With and Manage Containers](#interact-with-and-manage-containers)
 * [Troubleshoot Apps](#troubleshoot-apps)
 * [Weave Net Integration](#weave-net-integration)
 * [Interact with Kubernetes Clusters](#interact-with-kubernetes-clusters)
 * [Generate Custom Metrics with the Plugin API](#generate-custom-metrics-using-the-plugin-api)


##<a name="topology-mapping"></a> Topology Mapping

Scope builds logical topologies and node groups of your infrastructure. It displays an intelligent map of your app, so that different node types can be easily viewed and specific microservices can be drilled down on and inspected within your app.  

The following views are provided:

* Process type
  * By Name
* Container
  * By DNS Name
  * By Image
* Hosts
* Kubernetes Name Space Label

See [Managing Views](/site/manage-views.md) for more information. 

##<a name="flexible-view-filtering"></a> Flexible Filtering

Nodes can be filtered by CPU, Memory and Open Files. From the Container view, additional filters enable you to sort by System Containers, Application Containers or by Stopped vs.Running Containers or Both. Easily find the container using the most CPU or memory for a given host or service. In Host view, filter by CPU, Load or Memory consumption. Mouse-over any of the nodes to view its filtered metric at a glance. 

##<a name="real-time-app-and-container-metrics"></a>Real-time App and Container Metrics

View contextual metrics, tags and metadata for your containers.  Effortlessly navigate between processes inside your container to the hosts that your containers run on, arranged in expandable, sortable tables.  

Choose an overview of your container infrastructure, or focus on a specific microservice. Identify and correct issues to ensure the stability and performance of your containerized applications.

##<a name="interact-with-and-manage-containers"></a>Interact With and Manage Containers

Interact with your containers directly: pause, restart and stop containers without having to leave the Scope browser window.

##<a name="troubleshoot-apps"></a>Troubleshoot Apps

A convenient terminal window is provided that enables you to interact with your app and to troubleshoot and diagnose any issues all within the same context. 

##<a name="interact-with-kubernetes-clusters"></a>Interact with Kubernetes Clusters


##<a name="generate-custom-metrics-using-the-plugin-api"></a>Generate Custom Metrics using the Plugin API

Scope includes a Plugin API, so that custom metrics may be generated and integrated with the Scope UI. 

For information on how to generate your own metrics in Scope, see [Generating Custom Metrics with Plugins](/site/scope-plugins.md).

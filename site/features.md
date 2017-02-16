---
title: Feature Overview
menu_order: 15
---

Browse the current feature set for Scope with links to relevant indepth topics:

 * [Topology Mapping](#topology-mapping)
 * [Flexible Filtering](#flexible-filtering)
 * [Powerful Search](#powerful-search)
 * [Real-time App and Container Metrics](#real-time-app-and-container-metrics)
 * [Interact With and Manage Containers](#interact-with-and-manage-containers)
 * [Troubleshoot Apps](#troubleshoot-apps)


##<a name="topology-mapping"></a>Topology Mapping

Scope builds logical topologies of your application and infrastructure.  A topology is a collection of nodes and edges, where nodes can represent objects like processes, container or hosts.  In Scope, edges represent TCP connections between nodes.  Scope displays an intelligent map of your app, so that different node types can be easily viewed and specific microservices can be drilled down on and inspected.

!['Automatic Topology Mapping'](images/topology-map.png)

##<a name="flexible-filtering"></a>Flexible Filtering

Nodes can be filtered by various properties, for example in the Container view, you can filter by System Containers vs Application Containers or by Stopped vs. Running Containers. Scope can also display various metrics such as CPU and Memory usage in the nodes, for example allowing you to easily find the container using the most CPU or memory.  Mouse-over any of the nodes to view its filtered metric at a glance.

###<a name="powerful-search"></a> Powerful Search

Powerful search capability enables you to find node types, containers and processes by name, label or even path.  The search functionality supports simple operands so that you can for example, find processes consuming a certain Memory limit or nodes using too much CPU. You can also stack filters to create custom, complex search criterion. 

!['Powerful Search'](images/search.png)

##<a name="real-time-app-and-container-metrics"></a>Real-time App and Container Metrics

View contextual metrics, tags and metadata for your containers.  Effortlessly navigate between processes inside your container to the hosts that your containers run on, arranged in expandable, sortable tables.

Choose an overview of your container infrastructure, or focus on a specific microservice. Identify and correct issues to ensure the stability and performance of your containerized applications.

##<a name="interact-with-and-manage-containers"></a>Interact With and Manage Containers

Interact with your containers directly: pause, restart and stop containers without having to leave the Scope browser window.

##<a name="troubleshoot-apps"></a>Troubleshoot Apps

A convenient terminal window is provided that enables you to interact with your app and to troubleshoot and diagnose any issues all within the same context.

##<a name="generate-custom-metrics-using-the-plugin-api"></a>Generate Custom Metrics using the Plugin API

Scope includes a Plugin API, so that custom metrics may be generated and integrated with the Scope UI.

For information on how to generate your own metrics in Scope, see [Generating Custom Metrics with Plugins](/site/plugins.md).

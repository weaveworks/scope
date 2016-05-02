---
title: Introducing Weave Scope
menu_order: 10
---


Weave Scope is a browser-based infrastructure, app diagnostics and container visualization tool. It provides a top down view into your app as well as your entire infrastructure, and allows you to diagnose any problems with your distributed containerized app, in real time, as it being deployed to a cloud provider.  

Scope can be used in standalone mode, where you deploy it locally onto your hosts, or you can use Weave Scope as a Service, see [Installing Weave Scope](/site/installing-scope.md) for more information. 

##Getting Microservices Under Control

Microservices-based architecture poses significant challenges when deployed to Docker containers on the cloud. But with Weave Scope, visualizing network bottlenecks, troubleshooting CPU consumption and memory leaks is simplified. With Views, you can quickly examine various metrics about your containerized app. 

!['Microservices Under Control'](microservices-under-control.png)

##Automatic Topologies and Intelligent Grouping

Weave Scope generates a map of your processes, containers and hosts, so that you can understand, monitor, and control your applications.

It uses established APIs (for example: /proc, conntrack, the Docker API etc.) to gather information about the hosts on which it is running. And it does this without any modifications to your application or needing a custom kernel module.

Information collected by Weave Scope’s probes is used to build a topology of the applications and the containers running on the hosts.  Reports are generated locally by the probe (agent) and then sent to the app via HTTP, where they are merged and processed to form cluster-wide views.

!['Automatic Topologies and Intelligent Grouping'](automatic-topologies-grouping.png)

Intelligently group nodes to help diagnose problems in your app as well as to clarify and model its infrastructure set up.  For example, you can use the Docker ‘containers-grouped-by-hostname’ view to identify and dedupe replicas of containers that back multiple services, and show which services within your Docker infrastructure are in communication. 

For information on automated topologies and grouping, see [Managing Views](/site/manage-views.md).

##Developer Friendly: Contextual Details and Metrics

Drill down on nodes in any topology, and view contextual details for your containers, hosts and processes.  Metrics have been brought front and center, the tags and metadata user interface redefined with the children of your container/host arranged in expandable, sortable tables.  For example, the container consuming the most CPU or memory on a given host can be determined quickly with just a few clicks.

!['Developer Friendly: Contextual Details and Metrics'](contextual-details.png)

Weave Scope accumulates and stores 15-second sparklines of metrics such as memory and CPU usage, These provide a snapshot so that you can quickly understand how your containers are performing.

You can drill up and down between the various topologies that Scope natively understands. This enables you to easily navigate from containers, to hosts, processes, images, pods and services - and back again.

Scope can display metrics from Weave Net, Docker Containers, and the Linux Kernel. 

For more information on contextual details and metrics, See [XXXX]().

##Real-time Container Monitoring

With Weave Scope you can control the entire container lifecycle across your cluster hosts from a single UI.  Start, stop, pause and restart containers from the details panel, and toggle filters for stopped containers in the containers view.

Controlling Scope containers goes beyond simple lifecycle operations. Scope can also attach and exec. Attach executes a Docker attach against the container and lets you interact with it live.  Exec runs a shell in the container so that you can debug your running app in real-time. 

See, [Troubleshooting Containerized Apps](/site/app-troubleshooting.md).

##Orchestrator Aware and Cloud Integrations

Weave Scope can visualize your app within: 

 * Docker Compose
 * Kubernetes
 * AWS Cloud Formation


For instructions on Weave Scope integrations, see [Integrations](/site/integrations.md).

See Also

 * [Installing Weave Scope](/site/installing-scope.md)
 * [Feature Overview](/site/scope-features.md)


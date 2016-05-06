---
title: Understanding Weave Scope
menu_order: 70
---

The following topics are discussed:

* [Service Mode](#scope-reporting-in-service-mode)
* [Stand-alone Mode](#scope-reporting-in-stand-alone-mode)
* [Managing Clusters with Scope and weaveDNS](#managing-clusters-with-scope-and-weavedns)

Weave Scope consists of two components: the app and the probe. The components are deployed as a single Docker container using the scope script. The probe is responsible for gathering information about the host on which it is running. This information is sent to the app in the form of a report. The app processes reports from the probe into usable topologies, serving the UI, as well as pushing these topologies to the UI.


## <a name="service-mode"></a>Service Mode

Scope can also be used to feed reports to the Scope Service. The Scope Service allows you centrally manage and share access to your Scope UI. In this configuration, the probe is run locally and the apps are hosted for you.

!['Service Mode Reporting'](images/service-mode.png)

##<a name="signing-up-for-weave-scope-cloud-service"></a>Signing Up for Weave Scope Cloud Service

With Scope as a Service you can centrally manage and share access to your Scope user interface. In this configuration, the probe is run locally on your hosts, and the scope apps are hosted for you.

To register with the service, go to [scope.weave.works](https://scope.weave.works), enter your email address and wait for the confirmation email that contains your service token.

![`Weave Scope` Early Access Form](images/scope-early-access.png)


>**Note**: Candidates for the Early Access Program are evaluated on a case per case basis. It may take a few days before you receive a confirmation email. To accelerate the process, please send an email to `help@weave.works` which explains your use-case.

The service token is also accessible from the Scope as a Service main page after you've logged in:

![Weave Scope Configure Your App](images/scope-cloud-main-page.png)


##<a name="stand-alone-mode"></a>Standalone Mode

When running Scope in a cluster, each probe sends its reports to a dedicated app. The app merges the reports from its probe into a comprehensive report that is sent to the browser.  To visualize your entire infrastructure and apps running on that infrastructure, Scope must be launched on to every machine you are using.

!['Stand-alone Reporting'](images/stand-alone.png)

##<a name="managing-clusters-with-scope-and-weavedns"></a>Managing Clusters with Scope and weaveDNS

If Scope is running on the same machine as the Weave Network, then the probe uses weaveDNS to automatically discover any other apps on the network. Scope does this by registering itself under the address scope.weave.local.

Each probe sends its reports to every app registered at this address. If you have weaveDNS set up and running, no further steps are necessary.

If you don’t want to use weaveDNS, then Scope can be instructed to cluster with other Scope instances on the command line. Hostnames and IP addresses are acceptable, both with and without ports, for example:

    # scope launch scope1:4030 192.168.0.12 192.168.0.11:4030

Hostnames will be regularly resolved as A records, and each answer used as a target.


###Disabling Automatic Updates

Scope periodically checks with our servers to see if a new version is available. You can disable this by setting:

    CHECKPOINT_DISABLE=true scope launch

For more information, see [Go Checkpoint](https://github.com/weaveworks/go-checkpoint).



**See Also**

 * [Installing Weave Scope](/site/installing-scope.md)


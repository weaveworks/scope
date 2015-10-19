# Weave Scope

Weave Scope automatically generates a map of your containers, enabling you to
intuitively understand, monitor, and control your applications.

Scope does not need any configuration and does not require the Weave Network.
You need to run Scope on every machine you want to monitor.

## Architecture

Weave Scope consists of two components: the App and the Probe. These two
components are deployed as a single Docker container using the ```scope```
script:

```
sudo scope launch
```

The Probe is responsible for gathering information about the host is it running
on. This information is sent to the app in the form of a report. The App is
responsible for processing reports from the probe into useable topologies,
serving the UI and pushing these topologies to the UI.

```
+-------------------------------+            +-------------------------------+
|                               |            |                               |
| Docker Host                   |            |  Your laptop                  |
|                               |            |                               |
| +---------------------------+ |            |   +-----------------------+   |
| |                           | |            |   |                       |   |
| | Weave Scope Container     | |       -------->| Your web browser      |   |
| |                           | |      /     |   |                       |   |
| | +-----------------------+ | |     /      |   +-----------------------+   |
| | |                       | | |    /       |                               |
| | | Scope-app             |<-------        +-------------------------------+
| | |                       | | |
| | +------------+----------+ | |
| |              ^            | |
| |              |            | |
| | +------------+----------+ | |
| | |                       | | |
| | | Scope-probe           | | |
| | |                       | | |
| | +-----------------------+ | |
| |                           | |
| +---------------------------+ |
|                               |
+-------------------------------+
```

## Multi host setup

When running Scope in a cluster, each Probe sends reports to each App.
The App merges the reports from each probe into a more complete report.
You need to run Scope on every machine you want to monitor.

```
+-------------------------------+            +-------------------------------+
|                               |            |                               |
| Docker Host 1                 |            | Docker Host 2                 |
|                               |            |                               |
| +---------------------------+ |            | +---------------------------+ |
| |                           | |            | |                           | |
| | Weave Scope Container     | |            | | Weave Scope Container     | |
| |                           | |            | |                           | |
| | +-----------------------+ | |            | | +-----------------------+ | |
| | |                       | | |            | | |                       | | |
| | | Scope-app             |<------      ------>| Scope-app             | | |
| | |                       | | |   \    /   | | |                       | | |
| | +------------+----------+ | |    \  /    | | +------------+----------+ | |
| |              ^            | |     \/     | |              ^            | |
| |              |            | |     /\     | |              |            | |
| | +------------+----------+ | |    /  \    | | +------------+----------+ | |
| | |                       | | |   /    \   | | |                       | | |
| | | Scope-probe           +-------      -------+ Scope-probe           | | |
| | |                       | | |            | | |                       | | |
| | +-----------------------+ | |            | | +-----------------------+ | |
| |                           | |            | |                           | |
| +---------------------------+ |            | +---------------------------+ |
|                               |            |                               |
+-------------------------------+            +-------------------------------+
```

If you run Scope on the same machine as the Weave Network, the Probe will use
WeaveDNS to automatically discover other Apps on your network. Scope acheives
this by registering itself under the address scope.weave.local; each Probe
will send reports to every App register for this address. Therefore, if
you have a running WeaveDNS setup, you do not need to take any further steps.

If you do not wish to use WeaveDNS, you can instruct Scope to cluster with
other Scope instances on the command line. Hostnames and IP addresses are
acceptable, both with and without ports:

```
# weave launch scope1:4030 192.168.0.12 192.168.0.11:4030
```

## Using Scope Service

Scope can also be used to feed reports to the Scope Service. The Scope
Service allows you centrally manage and share access to your Scope UI.  In this
configuration, you only run the Probe locally - the Apps are hosted for you.

To get an account on the Scope Service, sign up at scope.weave.works. You need
to run a Probe on every machine you want to monitor with Scope. To launch a
Probe and send reports to the service, run the following command:

```
sudo scope launch --service-token=<service token>
```

```
                                 ___    ,'""""'.
                             ,"""   """"'      `.
                            ,'                   `._
                           ,'                       `"""'.
                          ,'                              `.
                         ,'                                :
                       ,'                                  `.
                 ,""""'                                     `"""".
               ,'                                                `.
              ,'               scope.weave.works                  `.
         ,-""'                                                    `.
        ,'                                                          `.
        ;            ,'                ','                            `.
       ,'            ;                 ,'                              `.
      ,'             ;                 (                    ,           `.
      ;              `                `.                  ,'           ,'
      ;               \                 `.                ,'    ,'    ,'
      '.               '                 `.             ,''.   ,' `--'
       `"-----"" _.-'' .-'`-.:..___...--' `-._      ,-"'   `-'
                                      /\      `"""""
                                      ||
+-----------------------------+       ||       +-----------------------------+
|                             |       ||       |                             |
| Docker Host                 |       ||       | Docker Host                 |
|                             |       ||       |                             |
| +-------------------------+ |       ||       | +-------------------------+ |
| |                         | |       ||       | |                         | |
| | Weave Scope Container   | |       ||       | | Weave Scope Container   | |
| |                         | |       ||       | |                         | |
| | +------------+--------+ | |      /  \      | | +------------+--------+ | |
| | |                     | | |     /    \     | | |                     | | |
| | | Scope-probe         +---------      ---------+ Scope-probe         | | |
| | |                     | | |                | | |                     | | |
| | +---------------------+ | |                | | +---------------------+ | |
| |                         | |                | |                         | |
| +-------------------------+ |                | +-------------------------+ |
|                             |                |                             |
+-----------------------------+                +-----------------------------+
````

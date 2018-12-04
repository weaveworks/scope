# Quick overview

## Information gatherers
These implement the `Reporter` interface

- `awsecs` Deals with talking to AWS ECS to get stats and info
- `docker` Inspects the docker status
- `endpoint` Gathers connection data 
- `host` Gets data from the host os, including things like CPU and mem stats
- `kubernetes` Gathers data from k8s
- `overlay` Talks to Weave Net for network stats from the overlay network
- `process` Is code that looks up running process and stats form the os

## Utility and control
- `appclient` Deals with generating and sending reports
- `controls` Utility code for control messages and the like
- `plugins` allows plugins to be added to the probe.

